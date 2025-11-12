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
	Title                 string           `json:"title" firestore:"title"`
	Category              string           `json:"category" firestore:"category"`
	Priority              string           `json:"priority" firestore:"priority"`
	Start                 time.Time        `json:"start" firestore:"start"`
	End                   time.Time        `json:"end" firestore:"end"`
	Time                  *string          `json:"time,omitempty" firestore:"time,omitempty"` // HH:MM format
	Description           *string          `json:"description,omitempty" firestore:"description,omitempty"`
	IsRepeating           bool             `json:"isRepeating" firestore:"isRepeating"`
	RepeatFrequency       *RepeatFrequency `json:"repeatFrequency,omitempty" firestore:"repeatFrequency,omitempty"`
	RepeatDays            []string         `json:"repeatDays,omitempty" firestore:"repeatDays,omitempty"`
	CreatedBy             string           `json:"createdBy" firestore:"createdBy"`
	AddToGoogleCalendar   *bool            `json:"addToGoogleCalendar,omitempty" firestore:"addToGoogleCalendar,omitempty"`
	GoogleCalendarEventID *string          `json:"googleCalendarEventId,omitempty" firestore:"googleCalendarEventId,omitempty"`
	EventType             *EventType       `json:"eventType,omitempty" firestore:"eventType,omitempty"`               // offline or online
	InvitedMemberIds      []string         `json:"invitedMemberIds,omitempty" firestore:"invitedMemberIds,omitempty"` // User IDs for online events
	Invites               []string         `json:"invites,omitempty" firestore:"invites,omitempty"`                   // Email addresses for event invites
	Duration              *int             `json:"duration,omitempty" firestore:"duration,omitempty"`                 // Duration in minutes for online events
	GoogleMeetLink        *string          `json:"googleMeetLink,omitempty" firestore:"googleMeetLink,omitempty"`     // Google Meet link for online events
}
