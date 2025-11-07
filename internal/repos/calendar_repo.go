package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"google.golang.org/api/iterator"
)

// CalendarEventRepo handles calendar event data operations
type CalendarEventRepo struct {
	*BaseRepo
}

// NewCalendarEventRepo creates a new calendar event repository
func NewCalendarEventRepo() *CalendarEventRepo {
	return &CalendarEventRepo{
		BaseRepo: NewBaseRepo("calendarEvents"),
	}
}

// GetByMonth gets calendar events for a given month
func (r *CalendarEventRepo) GetByMonth(ctx context.Context, month, year int) ([]models.CalendarEvent, error) {
	// Create start date: first day of the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	// Create end date: first day of next month (exclusive)
	endDate := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)

	iter := r.collection.Where("start", ">=", startDate).
		Where("start", "<", endDate).
		Documents(ctx)

	var events []models.CalendarEvent
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		// Convert time fields from strings/timestamps to time.Time
		if err := ConvertTimeFieldsInMap(data, []string{"start", "end", "created", "updated"}); err != nil {
			return nil, err
		}

		var event models.CalendarEvent
		if err := doc.DataTo(&event); err != nil {
			return nil, err
		}
		event.ID = doc.Ref.ID
		events = append(events, event)
	}

	return events, nil
}

// GetByID gets a calendar event by ID
func (r *CalendarEventRepo) GetByID(ctx context.Context, id string) (*models.CalendarEvent, error) {
	doc, err := r.collection.Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, nil
	}

	data := doc.Data()
	// Convert time fields from strings/timestamps to time.Time
	if err := ConvertTimeFieldsInMap(data, []string{"start", "end", "created", "updated"}); err != nil {
		return nil, err
	}

	var event models.CalendarEvent
	if err := doc.DataTo(&event); err != nil {
		return nil, err
	}
	event.ID = doc.Ref.ID
	return &event, nil
}

// Add creates a new calendar event
func (r *CalendarEventRepo) Add(ctx context.Context, event *models.CalendarEvent) error {
	newDocRef := r.collection.NewDoc()
	event.ID = newDocRef.ID
	event.Created = time.Now()

	data := map[string]interface{}{
		"id":          event.ID,
		"title":       event.Title,
		"category":    event.Category,
		"priority":    event.Priority,
		"start":       event.Start,
		"end":         event.End,
		"isRepeating": event.IsRepeating,
		"createdBy":   event.CreatedBy,
		"created":     event.Created,
	}
	if event.Time != nil {
		data["time"] = *event.Time
	}
	if event.Description != nil {
		data["description"] = *event.Description
	}
	if event.RepeatFrequency != nil {
		data["repeatFrequency"] = string(*event.RepeatFrequency)
	}
	if event.RepeatDays != nil {
		data["repeatDays"] = event.RepeatDays
	}
	if event.AddToGoogleCalendar != nil {
		data["addToGoogleCalendar"] = *event.AddToGoogleCalendar
	}
	if event.GoogleCalendarEventID != nil {
		data["googleCalendarEventId"] = *event.GoogleCalendarEventID
	}
	if event.EventType != nil {
		data["eventType"] = string(*event.EventType)
	}
	if event.InvitedMemberIds != nil && len(event.InvitedMemberIds) > 0 {
		data["invitedMemberIds"] = event.InvitedMemberIds
	}
	if event.Duration != nil {
		data["duration"] = *event.Duration
	}
	if event.GoogleMeetLink != nil {
		data["googleMeetLink"] = *event.GoogleMeetLink
	}

	_, err := newDocRef.Set(ctx, data)
	return err
}

// Update updates a calendar event
func (r *CalendarEventRepo) Update(ctx context.Context, event *models.CalendarEvent) error {
	now := time.Now()
	event.Updated = &now

	data := map[string]interface{}{
		"id":          event.ID,
		"title":       event.Title,
		"category":    event.Category,
		"priority":    event.Priority,
		"start":       event.Start,
		"end":         event.End,
		"isRepeating": event.IsRepeating,
		"createdBy":   event.CreatedBy,
		"updated":     now,
	}
	if event.Time != nil {
		data["time"] = *event.Time
	}
	if event.Description != nil {
		data["description"] = *event.Description
	}
	if event.RepeatFrequency != nil {
		data["repeatFrequency"] = string(*event.RepeatFrequency)
	}
	if event.RepeatDays != nil {
		data["repeatDays"] = event.RepeatDays
	}
	if event.AddToGoogleCalendar != nil {
		data["addToGoogleCalendar"] = *event.AddToGoogleCalendar
	}
	if event.GoogleCalendarEventID != nil {
		data["googleCalendarEventId"] = *event.GoogleCalendarEventID
	}
	if event.EventType != nil {
		data["eventType"] = string(*event.EventType)
	}
	if event.InvitedMemberIds != nil && len(event.InvitedMemberIds) > 0 {
		data["invitedMemberIds"] = event.InvitedMemberIds
	}
	if event.Duration != nil {
		data["duration"] = *event.Duration
	}
	if event.GoogleMeetLink != nil {
		data["googleMeetLink"] = *event.GoogleMeetLink
	}

	_, err := r.collection.Doc(event.ID).Set(ctx, data)
	return err
}

// Delete deletes a calendar event
func (r *CalendarEventRepo) Delete(ctx context.Context, id string) error {
	_, err := r.collection.Doc(id).Delete(ctx)
	return err
}
