package models

import "time"

// LeaveRequestType represents the type of leave request
type LeaveRequestType string

const (
	LeaveRequestTypeVacation     LeaveRequestType = "vacation"
	LeaveRequestTypeSickLeave    LeaveRequestType = "sick_leave"
	LeaveRequestTypeWorkRemotely LeaveRequestType = "work_remotely"
)

// LeaveRequestStatus represents the status of a leave request
type LeaveRequestStatus string

const (
	LeaveRequestStatusPending   LeaveRequestStatus = "pending"
	LeaveRequestStatusApproved  LeaveRequestStatus = "approved"
	LeaveRequestStatusRejected  LeaveRequestStatus = "rejected"
	LeaveRequestStatusCancelled LeaveRequestStatus = "cancelled"
)

// DurationType represents the duration type
type DurationType string

const (
	DurationTypeDays  DurationType = "days"
	DurationTypeHours DurationType = "hours"
)

// WorkingHours represents working hours for remote work
type WorkingHours struct {
	From string `json:"from" firestore:"from" bson:"from"` // Time format like "9:00 AM"
	To   string `json:"to" firestore:"to" bson:"to"`       // Time format like "1:00 PM"
}

// LeaveRequest represents a leave request
type LeaveRequest struct {
	Model
	UserID         string             `json:"userId" firestore:"userId" bson:"userId"`
	RequestType    LeaveRequestType   `json:"requestType" firestore:"requestType" bson:"requestType"`
	StartDate      time.Time          `json:"startDate" firestore:"startDate" bson:"startDate"`
	EndDate        *time.Time         `json:"endDate,omitempty" firestore:"endDate,omitempty" bson:"endDate,omitempty"`
	Duration       float64            `json:"duration" firestore:"duration" bson:"duration"`
	DurationType   DurationType       `json:"durationType" firestore:"durationType" bson:"durationType"`
	Status         LeaveRequestStatus `json:"status" firestore:"status" bson:"status"`
	Comments       *string            `json:"comments,omitempty" firestore:"comments,omitempty" bson:"comments,omitempty"`
	RequestedAt    time.Time          `json:"requestedAt" firestore:"requestedAt" bson:"requestedAt"`
	ReviewedBy     *string            `json:"reviewedBy,omitempty" firestore:"reviewedBy,omitempty" bson:"reviewedBy,omitempty"`
	ReviewedAt     *time.Time         `json:"reviewedAt,omitempty" firestore:"reviewedAt,omitempty" bson:"reviewedAt,omitempty"`
	ReviewComments *string            `json:"reviewComments,omitempty" firestore:"reviewComments,omitempty" bson:"reviewComments,omitempty"`
	WorkingHours   *WorkingHours      `json:"workingHours,omitempty" firestore:"workingHours,omitempty" bson:"workingHours,omitempty"`
}

// VacationSummary represents a vacation summary for a user
type VacationSummary struct {
	UserID           string  `json:"userId"`
	UserName         string  `json:"userName"`
	UserEmail        string  `json:"userEmail"`
	UserAvatar       *string `json:"userAvatar,omitempty"`
	VacationDays     float64 `json:"vacationDays"`
	SickLeaveDays    float64 `json:"sickLeaveDays"`
	WorkRemotelyDays float64 `json:"workRemotelyDays"`
	PendingRequests  int     `json:"pendingRequests"`
}

// LeaveRequestResponse represents a leave request with user details
type LeaveRequestResponse struct {
	LeaveRequest
	UserDetails     User  `json:"userDetails"`
	ReviewerDetails *User `json:"reviewerDetails,omitempty"`
}
