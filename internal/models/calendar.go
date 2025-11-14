package models

import "time"

// RepeatFrequency represents the frequency of event repetition
type RepeatFrequency string

const (
	RepeatFrequencyDaily   RepeatFrequency = "daily"
	RepeatFrequencyWeekly  RepeatFrequency = "weekly"
	RepeatFrequencyMonthly RepeatFrequency = "monthly"
)

// EventType represents the type of calendar event
type EventType string

const (
	EventTypeOffline EventType = "offline"
	EventTypeOnline  EventType = "online"
)

// CalendarEvent represents a calendar event
type CalendarEvent struct {
	Model
	Title                 string           `json:"title" firestore:"title" bson:"title"`
	Category              string           `json:"category" firestore:"category" bson:"category"`
	Priority              string           `json:"priority" firestore:"priority" bson:"priority"`
	Start                 time.Time        `json:"start" firestore:"start" bson:"start"`
	End                   time.Time        `json:"end" firestore:"end" bson:"end"`
	Time                  *string          `json:"time,omitempty" firestore:"time,omitempty" bson:"time,omitempty"` // HH:MM format
	Description           *string          `json:"description,omitempty" firestore:"description,omitempty" bson:"description,omitempty"`
	IsRepeating           bool             `json:"isRepeating" firestore:"isRepeating" bson:"isRepeating"`
	RepeatFrequency       *RepeatFrequency `json:"repeatFrequency,omitempty" firestore:"repeatFrequency,omitempty" bson:"repeatFrequency,omitempty"`
	RepeatDays            []string         `json:"repeatDays,omitempty" firestore:"repeatDays,omitempty" bson:"repeatDays,omitempty"`
	CreatedBy             string           `json:"createdBy" firestore:"createdBy" bson:"createdBy"`
	AddToGoogleCalendar   *bool            `json:"addToGoogleCalendar,omitempty" firestore:"addToGoogleCalendar,omitempty" bson:"addToGoogleCalendar,omitempty"`
	GoogleCalendarEventID *string          `json:"googleCalendarEventId,omitempty" firestore:"googleCalendarEventId,omitempty" bson:"googleCalendarEventId,omitempty"`
	EventType             *EventType       `json:"eventType,omitempty" firestore:"eventType,omitempty" bson:"eventType,omitempty"`               // offline or online
	InvitedMemberIds      []string         `json:"invitedMemberIds,omitempty" firestore:"invitedMemberIds,omitempty" bson:"invitedMemberIds,omitempty"` // User IDs for online events
	Invites               []string         `json:"invites,omitempty" firestore:"invites,omitempty" bson:"invites,omitempty"`                   // Email addresses for event invites
	Duration              *int             `json:"duration,omitempty" firestore:"duration,omitempty" bson:"duration,omitempty"`                 // Duration in minutes for online events
	GoogleMeetLink        *string          `json:"googleMeetLink,omitempty" firestore:"googleMeetLink,omitempty" bson:"googleMeetLink,omitempty"`     // Google Meet link for online events
}
