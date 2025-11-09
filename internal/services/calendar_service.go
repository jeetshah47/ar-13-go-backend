package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/email"
)

// CalendarEventService handles calendar event business logic
type CalendarEventService struct {
	calendarRepo     *repos.CalendarEventRepo
	userRepo         *repos.UserRepo
	emailClient      *email.Client
	googleAccountSvc *GoogleAccountService
	cacheSvc         *CacheService
}

// NewCalendarEventService creates a new calendar event service
func NewCalendarEventService(cfg *config.Config) *CalendarEventService {
	var emailClient *email.Client
	if cfg != nil {
		emailClient = email.NewClient(cfg)
	}
	return &CalendarEventService{
		calendarRepo:     repos.NewCalendarEventRepo(),
		userRepo:         repos.NewUserRepo(),
		emailClient:      emailClient,
		googleAccountSvc: NewGoogleAccountService(),
		cacheSvc:         NewCacheService(),
	}
}

// GetByMonth gets calendar events for a month
// Uses Redis cache to improve performance
func (s *CalendarEventService) GetByMonth(ctx context.Context, month, year int) ([]models.CalendarEvent, error) {
	// Try to get from cache first
	cached, err := s.cacheSvc.GetCalendarMonth(ctx, year, month)
	if err == nil && cached != nil {
		// Convert cached interface{} slice to CalendarEvent slice
		events := make([]models.CalendarEvent, 0, len(cached))
		for _, item := range cached {
			if eventMap, ok := item.(map[string]interface{}); ok {
				// Convert map to CalendarEvent
				var event models.CalendarEvent
				if data, err := json.Marshal(eventMap); err == nil {
					if err := json.Unmarshal(data, &event); err == nil {
						events = append(events, event)
					}
				}
			}
		}
		if len(events) > 0 {
			return events, nil
		}
	}

	// Cache miss or error - fetch from DB
	events, err := s.calendarRepo.GetByMonth(ctx, month, year)
	if err != nil {
		return nil, err
	}

	// Convert to interface{} slice for caching
	cacheData := make([]interface{}, len(events))
	for i := range events {
		cacheData[i] = events[i]
	}

	// Cache the result (ignore cache errors)
	_ = s.cacheSvc.SetCalendarMonth(ctx, year, month, cacheData)

	return events, nil
}

// GetByID gets a calendar event by ID
func (s *CalendarEventService) GetByID(ctx context.Context, id string) (*models.CalendarEvent, error) {
	return s.calendarRepo.GetByID(ctx, id)
}

// Add creates a new calendar event
func (s *CalendarEventService) Add(ctx context.Context, event *models.CalendarEvent) error {
	// If duration is provided and end time is not set, calculate end time
	if event.Duration != nil && *event.Duration > 0 && event.End.IsZero() {
		event.End = event.Start.Add(time.Duration(*event.Duration) * time.Minute)
	}

	if err := s.calendarRepo.Add(ctx, event); err != nil {
		return err
	}

	// Invalidate cache for the month of this event
	_ = s.cacheSvc.InvalidateCalendarMonth(ctx, event.Start.Year(), int(event.Start.Month()))

	// Sync with Google Calendar if requested
	if event.AddToGoogleCalendar != nil && *event.AddToGoogleCalendar && event.CreatedBy != "" {
		go func() {
			googleEvent, err := s.googleAccountSvc.ConvertCalendarEventToGoogleEvent(context.Background(), event.CreatedBy, event)
			if err != nil {
				log.Printf("Failed to convert calendar event to Google format: %v", err)
				return
			}

			createdEvent, err := s.googleAccountSvc.CreateGoogleCalendarEvent(context.Background(), event.CreatedBy, googleEvent)
			if err != nil {
				log.Printf("Failed to create Google Calendar event: %v", err)
				return
			}

			// Update the event with Google Calendar event ID and Meet link
			event.GoogleCalendarEventID = &createdEvent.ID

			// Extract Google Meet link from conference data
			meetLink := ""
			if createdEvent.HangoutLink != "" {
				meetLink = createdEvent.HangoutLink
			} else if createdEvent.ConferenceData != nil && len(createdEvent.ConferenceData.EntryPoints) > 0 {
				// Get the video entry point (Google Meet link)
				for _, entryPoint := range createdEvent.ConferenceData.EntryPoints {
					if entryPoint.EntryPointType == "video" {
						meetLink = entryPoint.URI
						break
					}
				}
			}

			if meetLink != "" {
				event.GoogleMeetLink = &meetLink
			}

			if updateErr := s.calendarRepo.Update(context.Background(), event); updateErr != nil {
				log.Printf("Failed to update event with Google Calendar ID: %v", updateErr)
			}
		}()
	}

	// Send email notification to event creator (non-blocking)
	if s.emailClient != nil && event.CreatedBy != "" {
		go func() {
			user, err := s.userRepo.GetByID(context.Background(), event.CreatedBy)
			if err != nil || user == nil {
				log.Printf("Failed to get user for calendar event email notification: %v", err)
				return
			}

			// Format event time
			eventTime := ""
			if event.Time != nil {
				eventTime = *event.Time
			} else {
				eventTime = event.Start.Format("15:04")
			}

			// Format event date
			eventDate := event.Start.Format("January 2, 2006")
			if !event.End.IsZero() && event.End != event.Start {
				eventDate += fmt.Sprintf(" - %s", event.End.Format("January 2, 2006"))
			}

			message := fmt.Sprintf("A new calendar event '%s' has been created.\n\nDate: %s\nTime: %s", event.Title, eventDate, eventTime)
			if event.Description != nil && *event.Description != "" {
				message += fmt.Sprintf("\n\nDescription: %s", *event.Description)
			}
			if event.Category != "" {
				message += fmt.Sprintf("\nCategory: %s", event.Category)
			}
			if event.Priority != "" {
				message += fmt.Sprintf("\nPriority: %s", event.Priority)
			}

			title := fmt.Sprintf("Calendar Event Created: %s", event.Title)
			if err := s.emailClient.SendAlertEmail([]string{user.Email}, title, message, "info"); err != nil {
				log.Printf("Failed to send calendar event email: %v", err)
			}
		}()
	}

	return nil
}

// Update updates a calendar event
func (s *CalendarEventService) Update(ctx context.Context, event *models.CalendarEvent) error {
	existing, err := s.calendarRepo.GetByID(ctx, event.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("calendar event not found")
	}

	// Sync with Google Calendar if the event is linked to Google Calendar
	if existing.GoogleCalendarEventID != nil && *existing.GoogleCalendarEventID != "" && event.CreatedBy != "" {
		go func() {
			googleEvent, err := s.googleAccountSvc.ConvertCalendarEventToGoogleEvent(context.Background(), event.CreatedBy, event)
			if err != nil {
				log.Printf("Failed to convert calendar event to Google format: %v", err)
				return
			}

			_, err = s.googleAccountSvc.UpdateGoogleCalendarEvent(context.Background(), event.CreatedBy, *existing.GoogleCalendarEventID, googleEvent)
			if err != nil {
				log.Printf("Failed to update Google Calendar event: %v", err)
				return
			}
		}()
	}

	if err := s.calendarRepo.Update(ctx, event); err != nil {
		return err
	}

	// Invalidate cache for both old and new months
	_ = s.cacheSvc.InvalidateCalendarMonth(ctx, existing.Start.Year(), int(existing.Start.Month()))
	_ = s.cacheSvc.InvalidateCalendarMonth(ctx, event.Start.Year(), int(event.Start.Month()))

	return nil
}

// Delete deletes a calendar event
func (s *CalendarEventService) Delete(ctx context.Context, id string) error {
	existing, err := s.calendarRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("calendar event not found")
	}

	// Delete from Google Calendar if the event is linked
	if existing.GoogleCalendarEventID != nil && *existing.GoogleCalendarEventID != "" && existing.CreatedBy != "" {
		go func() {
			if err := s.googleAccountSvc.DeleteGoogleCalendarEvent(context.Background(), existing.CreatedBy, *existing.GoogleCalendarEventID); err != nil {
				log.Printf("Failed to delete Google Calendar event: %v", err)
			}
		}()
	}

	if err := s.calendarRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache for the month of this event
	_ = s.cacheSvc.InvalidateCalendarMonth(ctx, existing.Start.Year(), int(existing.Start.Month()))

	return nil
}

// ParseMonthYear parses month and year from strings
func ParseMonthYear(monthStr, yearStr string) (month, year int, err error) {
	month, err = strconv.Atoi(monthStr)
	if err != nil {
		return 0, 0, err
	}
	year, err = strconv.Atoi(yearStr)
	if err != nil {
		return 0, 0, err
	}
	return month, year, nil
}
