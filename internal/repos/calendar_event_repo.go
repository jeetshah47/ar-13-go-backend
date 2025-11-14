package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CalendarEventRepo handles calendar event data operations with MongoDB
type CalendarEventRepo struct {
	*MongoBaseRepo
}

// NewCalendarEventRepo creates a new MongoDB calendar event repository
func NewCalendarEventRepo() *CalendarEventRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &CalendarEventRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "calendar_events"),
	}
}

// GetByMonth gets calendar events for a given month
func (r *CalendarEventRepo) GetByMonth(ctx context.Context, month, year int) ([]models.CalendarEvent, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)

	filter := bson.M{
		"start": bson.M{
			"$gte": startDate,
			"$lt":  endDate,
		},
	}

	items, err := r.FindAll(ctx, filter, nil, bson.M{"start": 1})
	if err != nil {
		return nil, err
	}

	events := make([]models.CalendarEvent, 0, len(items))
	for _, item := range items {
		var event models.CalendarEvent
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &event); err != nil {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

// GetByID gets a calendar event by ID
func (r *CalendarEventRepo) GetByID(ctx context.Context, id string) (*models.CalendarEvent, error) {
	result := r.MongoBaseRepo.GetByID(ctx, id)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var event models.CalendarEvent
	if err := result.Decode(&event); err != nil {
		return nil, err
	}
	return &event, nil
}

// Add creates a new calendar event
func (r *CalendarEventRepo) Add(ctx context.Context, event *models.CalendarEvent) error {
	if event.ID == "" {
		event.ID = fmt.Sprintf("event-%d", time.Now().UnixNano())
	}
	event.Created = time.Now()

	return r.InsertOne(ctx, event)
}

// Update updates a calendar event
func (r *CalendarEventRepo) Update(ctx context.Context, event *models.CalendarEvent) error {
	now := time.Now()
	event.Updated = &now

	updates := bson.M{
		"title":       event.Title,
		"category":    event.Category,
		"priority":    event.Priority,
		"start":       event.Start,
		"end":         event.End,
		"isRepeating": event.IsRepeating,
		"createdBy":   event.CreatedBy,
		"updated":     event.Updated,
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
	if event.InvitedMemberIds != nil {
		updates["invitedMemberIds"] = event.InvitedMemberIds
	}
	if event.Invites != nil {
		updates["invites"] = event.Invites
	}
	if event.Duration != nil {
		updates["duration"] = *event.Duration
	}
	if event.GoogleMeetLink != nil {
		updates["googleMeetLink"] = *event.GoogleMeetLink
	}

	return r.UpdateOne(ctx, event.ID, updates)
}

// Delete deletes a calendar event
func (r *CalendarEventRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

