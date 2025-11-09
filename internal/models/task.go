package models

import "time"

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeStatusChanged      ActivityType = "status_changed"
	ActivityTypeTimeSpentAdded     ActivityType = "time_spent_added"
	ActivityTypeTimeSpentUpdated   ActivityType = "time_spent_updated"
	ActivityTypeTimeSpentRemoved   ActivityType = "time_spent_removed"
	ActivityTypeFileUploaded       ActivityType = "file_uploaded"
	ActivityTypeFileRemoved        ActivityType = "file_removed"
	ActivityTypeTaskCreated        ActivityType = "task_created"
	ActivityTypeTaskUpdated        ActivityType = "task_updated"
	ActivityTypeTaskAssigned       ActivityType = "task_assigned"
	ActivityTypeDescriptionUpdated ActivityType = "description_updated"
	ActivityTypeDeadlineUpdated    ActivityType = "deadline_updated"
	ActivityTypeProgressUpdated    ActivityType = "progress_updated"
)

// TimeSpent represents time spent on a task
type TimeSpent struct {
	Date        string  `json:"date" firestore:"date"`           // ISO date string (YYYY-MM-DD)
	TimeSpent   int     `json:"timeSpent" firestore:"timeSpent"` // Time in minutes
	UserID      string  `json:"userId" firestore:"userId"`
	Description *string `json:"description,omitempty" firestore:"description,omitempty"`
}

// FileAttachment represents a file attachment
type FileAttachment struct {
	FileName     string    `json:"fileName" firestore:"fileName"`
	OriginalName string    `json:"originalName" firestore:"originalName"`
	FileSize     int64     `json:"fileSize" firestore:"fileSize"` // Size in bytes
	MimeType     string    `json:"mimeType" firestore:"mimeType"`
	UploadDate   time.Time `json:"uploadDate" firestore:"uploadDate"`
	UploadedBy   string    `json:"uploadedBy" firestore:"uploadedBy"`
	FileURL      string    `json:"fileUrl" firestore:"fileUrl"`
}

// LinkAttachment represents a link attachment
type LinkAttachment struct {
	URL          string    `json:"url" firestore:"url"`
	Title        *string   `json:"title,omitempty" firestore:"title,omitempty"`
	Description  *string   `json:"description,omitempty" firestore:"description,omitempty"`
	ThumbnailURL *string   `json:"thumbnailUrl,omitempty" firestore:"thumbnailUrl,omitempty"`
	AddedBy      string    `json:"addedBy" firestore:"addedBy"`
	AddedDate    time.Time `json:"addedDate" firestore:"addedDate"`
}

// ActivityLog represents an activity log entry
type ActivityLog struct {
	ID          string                 `json:"id" firestore:"id"`
	Type        ActivityType           `json:"type" firestore:"type"`
	Timestamp   time.Time              `json:"timestamp" firestore:"timestamp"`
	UserID      string                 `json:"userId" firestore:"userId"`
	UserName    *string                `json:"userName,omitempty" firestore:"userName,omitempty"`
	User        *User                  `json:"user,omitempty" firestore:"-"` // Populated user details
	Description string                 `json:"description" firestore:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}

// Task represents a task
type Task struct {
	Model
	Subject         string           `json:"subject" firestore:"subject"`
	Code            string           `json:"code" firestore:"code"`
	Status          string           `json:"status" firestore:"status"`
	Deadline        time.Time        `json:"deadline" firestore:"deadline"`
	Priority        string           `json:"priority" firestore:"priority"`
	Progress        *int             `json:"progress,omitempty" firestore:"progress,omitempty"` // Completion percentage (0-100)
	AssignTo        *string          `json:"assignTo,omitempty" firestore:"assignTo,omitempty"`
	ProjectID       string           `json:"projectId" firestore:"projectId"`
	TimeSpent       []TimeSpent      `json:"timeSpent" firestore:"timeSpent"`
	Description     *string          `json:"description,omitempty" firestore:"description,omitempty"`
	FileAttachments []FileAttachment `json:"fileAttachments" firestore:"fileAttachments"`
	ActivityLogs    []ActivityLog    `json:"activityLogs" firestore:"activityLogs"`
}

// TaskDetailResponse represents a task with assignee details
type TaskDetailResponse struct {
	Task
	AssignDetail *User `json:"assignDetail,omitempty" firestore:"-"`
}
