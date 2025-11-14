package constants

import (
	"strings"
)

// TaskStatus represents the master task status values
type TaskStatus string

const (
	// TaskStatusPending - Task is pending/not started
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusInProgress - Task is currently being worked on
	TaskStatusInProgress TaskStatus = "in_progress"
	// TaskStatusInReview - Task is under review
	TaskStatusInReview TaskStatus = "in_review"
	// TaskStatusCompleted - Task is completed
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusAccepted - Task has been accepted
	TaskStatusAccepted TaskStatus = "accepted"
	// TaskStatusRejected - Task has been rejected
	TaskStatusRejected TaskStatus = "rejected"
)

// AllTaskStatuses returns all valid task status values
func AllTaskStatuses() []TaskStatus {
	return []TaskStatus{
		TaskStatusPending,
		TaskStatusInProgress,
		TaskStatusInReview,
		TaskStatusCompleted,
		TaskStatusAccepted,
		TaskStatusRejected,
	}
}

// IsValidTaskStatus checks if a status string is a valid task status
func IsValidTaskStatus(status string) bool {
	normalized := NormalizeTaskStatus(status)
	return normalized != ""
}

// NormalizeTaskStatus converts various status string formats to the master status value
// Returns empty string if status cannot be normalized to a valid master status
func NormalizeTaskStatus(status string) string {
	if status == "" {
		return ""
	}

	// Convert to lowercase and trim whitespace
	normalized := strings.ToLower(strings.TrimSpace(status))

	// Map old status variants to master statuses
	statusMap := map[string]string{
		// Pending variants
		"pending": string(TaskStatusPending),
		
		// Backlog/Todo variants -> Pending
		"backlog": string(TaskStatusPending),
		"todo":    string(TaskStatusPending),
		"to-do":   string(TaskStatusPending),
		"to do":   string(TaskStatusPending),
		
		// In Progress variants
		"in_progress": string(TaskStatusInProgress),
		"inprogress":  string(TaskStatusInProgress),
		"in-progress": string(TaskStatusInProgress),
		"in progress": string(TaskStatusInProgress),
		
		// In Review variants
		"in_review": string(TaskStatusInReview),
		"inreview":  string(TaskStatusInReview),
		"in-review": string(TaskStatusInReview),
		"in review": string(TaskStatusInReview),
		
		// Completed variants
		"completed": string(TaskStatusCompleted),
		"done":      string(TaskStatusCompleted),
		"finished":  string(TaskStatusCompleted),
		
		// Accepted
		"accepted": string(TaskStatusAccepted),
		
		// Rejected variants
		"rejected": string(TaskStatusRejected),
		"cancelled": string(TaskStatusRejected),
		"canceled":  string(TaskStatusRejected),
	}

	// Check if normalized status exists in map
	if masterStatus, exists := statusMap[normalized]; exists {
		return masterStatus
	}

	// If the status is already a master status, return it
	for _, masterStatus := range AllTaskStatuses() {
		if normalized == string(masterStatus) {
			return normalized
		}
	}

	// Return empty string if status cannot be normalized
	return ""
}

// GetTaskStatusString returns the string value of a TaskStatus constant
func GetTaskStatusString(status TaskStatus) string {
	return string(status)
}

