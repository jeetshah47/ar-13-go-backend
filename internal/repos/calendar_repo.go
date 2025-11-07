package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
)

// CalendarEventRepo handles calendar event data operations
// TODO: Migrate to DynamoDB
type CalendarEventRepo struct {
	*DynamoBaseRepo
}

// NewCalendarEventRepo creates a new calendar event repository
func NewCalendarEventRepo() *CalendarEventRepo {
	return &CalendarEventRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("calendar_events"),
	}
}

// GetByMonth gets calendar events for a given month
func (r *CalendarEventRepo) GetByMonth(ctx context.Context, month, year int) ([]models.CalendarEvent, error) {
	// TODO: Implement DynamoDB scan with filter by date range
	items, err := r.ScanItems(ctx, nil)
	if err != nil {
		return nil, err
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)

	var events []models.CalendarEvent
	for _, item := range items {
		var event models.CalendarEvent
		if err := UnmarshalItem(item, &event); err != nil {
			continue
		}
		// Filter by date range
		if event.Start.After(startDate) && event.Start.Before(endDate) {
			events = append(events, event)
		}
	}

	return events, nil
}

// GetByID gets a calendar event by ID
func (r *CalendarEventRepo) GetByID(ctx context.Context, id string) (*models.CalendarEvent, error) {
	item, err := r.DynamoBaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var event models.CalendarEvent
	if err := UnmarshalItem(item, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// Add creates a new calendar event
func (r *CalendarEventRepo) Add(ctx context.Context, event *models.CalendarEvent) error {
	// TODO: Generate ID if not set
	if event.ID == "" {
		event.ID = fmt.Sprintf("event-%d", time.Now().UnixNano())
	}
	event.Created = time.Now()

	data := map[string]interface{}{
		"id":          event.ID,
		"title":       event.Title,
		"category":    event.Category,
		"priority":    event.Priority,
		"start":       event.Start.Format(time.RFC3339),
		"end":         event.End.Format(time.RFC3339),
		"isRepeating": event.IsRepeating,
		"createdBy":   event.CreatedBy,
		"created":     event.Created.Format(time.RFC3339),
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

	return r.PutItem(ctx, data)
}

// Update updates a calendar event
func (r *CalendarEventRepo) Update(ctx context.Context, event *models.CalendarEvent) error {
	now := time.Now()
	event.Updated = &now

	updates := map[string]interface{}{
		"title":       event.Title,
		"category":    event.Category,
		"priority":    event.Priority,
		"start":       event.Start.Format(time.RFC3339),
		"end":         event.End.Format(time.RFC3339),
		"isRepeating": event.IsRepeating,
		"createdBy":   event.CreatedBy,
	}
	if event.Time != nil {
		updates["time"] = *event.Time
	}
	if event.Description != nil {
		updates["description"] = *event.Description
	}
	if event.RepeatFrequency != nil {
		updates["repeatFrequency"] = string(*event.RepeatFrequency)
	}
	if event.RepeatDays != nil {
		updates["repeatDays"] = event.RepeatDays
	}
	if event.AddToGoogleCalendar != nil {
		updates["addToGoogleCalendar"] = *event.AddToGoogleCalendar
	}
	if event.GoogleCalendarEventID != nil {
		updates["googleCalendarEventId"] = *event.GoogleCalendarEventID
	}
	if event.EventType != nil {
		updates["eventType"] = string(*event.EventType)
	}
	if event.InvitedMemberIds != nil && len(event.InvitedMemberIds) > 0 {
		updates["invitedMemberIds"] = event.InvitedMemberIds
	}
	if event.Duration != nil {
		updates["duration"] = *event.Duration
	}
	if event.GoogleMeetLink != nil {
		updates["googleMeetLink"] = *event.GoogleMeetLink
	}

	return r.UpdateItem(ctx, event.ID, updates)
}

// Delete deletes a calendar event
func (r *CalendarEventRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
