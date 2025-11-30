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
	Date        string  `json:"date" firestore:"date" bson:"date"`                // ISO date string (YYYY-MM-DD)
	TimeSpent   int     `json:"timeSpent" firestore:"timeSpent" bson:"timeSpent"` // Time in minutes
	UserID      string  `json:"userId" firestore:"userId" bson:"userId"`
	Description *string `json:"description,omitempty" firestore:"description,omitempty" bson:"description,omitempty"`
}

// TimeTrackingSession represents an active time tracking session for a task
type TimeTrackingSession struct {
	ID          string     `json:"id" bson:"id"`
	TaskID      string     `json:"taskId" bson:"taskId"`
	ProjectID   string     `json:"projectId" bson:"projectId"`
	UserID      string     `json:"userId" bson:"userId"`
	StartTime   time.Time  `json:"startTime" bson:"startTime"`
	LastActive  time.Time  `json:"lastActive" bson:"lastActive"`
	EndTime     *time.Time `json:"endTime,omitempty" bson:"endTime,omitempty"`
	TotalMinutes int       `json:"totalMinutes" bson:"totalMinutes"` // Accumulated active minutes
	IsActive    bool       `json:"isActive" bson:"isActive"`
	Created     time.Time  `json:"created" bson:"created"`
	Updated     *time.Time `json:"updated,omitempty" bson:"updated,omitempty"`
}

// FileAttachment represents a file attachment
type FileAttachment struct {
	FileName     string    `json:"fileName" firestore:"fileName" bson:"fileName"`
	OriginalName string    `json:"originalName" firestore:"originalName" bson:"originalName"`
	FileSize     int64     `json:"fileSize" firestore:"fileSize" bson:"fileSize"` // Size in bytes
	MimeType     string    `json:"mimeType" firestore:"mimeType" bson:"mimeType"`
	UploadDate   time.Time `json:"uploadDate" firestore:"uploadDate" bson:"uploadDate"`
	UploadedBy   string    `json:"uploadedBy" firestore:"uploadedBy" bson:"uploadedBy"`
	FileURL      string    `json:"fileUrl" firestore:"fileUrl" bson:"fileUrl"`
}

// FileAttachmentResponse represents a file attachment with pre-signed URL for API responses
type FileAttachmentResponse struct {
	FileName      string    `json:"fileName"`
	OriginalName  string    `json:"originalName"`
	FileSize      int64     `json:"fileSize"`
	MimeType      string    `json:"mimeType"`
	UploadDate    time.Time `json:"uploadDate"`
	UploadedBy    string    `json:"uploadedBy"`
	FileURL       string    `json:"fileUrl"`        // Original file path
	PreviewURL    *string   `json:"previewUrl,omitempty"` // Pre-signed URL for images (1 hour expiry)
	DownloadURL   *string   `json:"downloadUrl,omitempty"` // Pre-signed URL for downloads (1 hour expiry)
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
	ID          string                 `json:"id" firestore:"id" bson:"id"`
	Type        ActivityType           `json:"type" firestore:"type" bson:"type"`
	Timestamp   time.Time              `json:"timestamp" firestore:"timestamp" bson:"timestamp"`
	UserID      string                 `json:"userId" firestore:"userId" bson:"userId"`
	UserName    *string                `json:"userName,omitempty" firestore:"userName,omitempty" bson:"userName,omitempty"`
	User        *User                  `json:"user,omitempty" firestore:"-" bson:"-"` // Populated user details
	Description string                 `json:"description" firestore:"description" bson:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty" bson:"metadata,omitempty"`
}

// Task represents a task
type Task struct {
	Model
	Subject         string           `json:"subject" firestore:"subject" bson:"subject"`
	Code            string           `json:"code" firestore:"code" bson:"code"`
	Status          string           `json:"status" firestore:"status" bson:"status"`
	StartDate       *time.Time       `json:"startDate,omitempty" firestore:"startDate,omitempty" bson:"startDate,omitempty"`
	EndDate         *time.Time       `json:"endDate,omitempty" firestore:"endDate,omitempty" bson:"endDate,omitempty"`
	Deadline        time.Time        `json:"deadline" firestore:"deadline" bson:"deadline"`
	Priority        string           `json:"priority" firestore:"priority" bson:"priority"`
	Progress        *int             `json:"progress,omitempty" firestore:"progress,omitempty" bson:"progress,omitempty"` // Completion percentage (0-100)
	AssignTo        *string          `json:"assignTo,omitempty" firestore:"assignTo,omitempty" bson:"assignTo,omitempty"`
	ProjectID       string           `json:"projectId" firestore:"projectId" bson:"projectId"`
	DrawingID       *string          `json:"drawingId,omitempty" firestore:"drawingId,omitempty" bson:"drawingId,omitempty"` // Optional reference to drawing type ID
	DrawingInfo     *DrawingInfo     `json:"drawingInfo,omitempty" firestore:"-" bson:"-"`                                   // Populated drawing information (not stored in DB)
	TimeSpent       []TimeSpent      `json:"timeSpent" firestore:"timeSpent" bson:"timeSpent"`
	Description     *string          `json:"description,omitempty" firestore:"description,omitempty" bson:"description,omitempty"`
	FileAttachments []FileAttachment `json:"fileAttachments" firestore:"fileAttachments" bson:"fileAttachments"`
	ActivityLogs    []ActivityLog    `json:"activityLogs" firestore:"activityLogs" bson:"activityLogs"`
}

// AssignToUser represents user details in assignTo field
type AssignToUser struct {
	ID   string `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

// TaskDetailResponse represents a task with assignee details
type TaskDetailResponse struct {
	Task
	AssignDetail *User `json:"assignDetail,omitempty" firestore:"-"`
}

// DrawingInfo represents drawing category and type information
type DrawingInfo struct {
	TypeID       string `json:"typeId" bson:"typeId"`
	TypeName     string `json:"typeName" bson:"typeName"`
	CategoryID   string `json:"categoryId" bson:"categoryId"`
	CategoryName string `json:"categoryName" bson:"categoryName"`
}

// TaskWithUserDetails represents a task with assignTo as an object containing user id and name
type TaskWithUserDetails struct {
	Task
	AssignTo    *AssignToUser `json:"assignTo,omitempty" bson:"assignTo,omitempty"`
	DrawingInfo *DrawingInfo  `json:"drawingInfo,omitempty" bson:"drawingInfo,omitempty"` // Populated drawing information
}

// ToTaskWithUserDetails converts a Task to TaskWithUserDetails
// If assignTo is a string (user ID), it will be set to nil (needs to be populated via aggregation)
func (t *Task) ToTaskWithUserDetails() *TaskWithUserDetails {
	return &TaskWithUserDetails{
		Task:     *t,
		AssignTo: nil, // Will be populated by aggregation
	}
}

// ToTask converts TaskWithUserDetails back to Task
func (twud *TaskWithUserDetails) ToTask() *Task {
	task := twud.Task
	// If AssignTo has user details, extract the ID
	if twud.AssignTo != nil && twud.AssignTo.ID != "" {
		task.AssignTo = &twud.AssignTo.ID
	}
	return &task
}

// ToTaskWithUserDetailsSlice converts a slice of Task to TaskWithUserDetails
func ToTaskWithUserDetailsSlice(tasks []Task) []TaskWithUserDetails {
	result := make([]TaskWithUserDetails, len(tasks))
	for i := range tasks {
		result[i] = *tasks[i].ToTaskWithUserDetails()
	}
	return result
}

// ToTaskSlice converts a slice of TaskWithUserDetails to Task
func ToTaskSlice(tasks []TaskWithUserDetails) []Task {
	result := make([]Task, len(tasks))
	for i := range tasks {
		result[i] = *tasks[i].ToTask()
	}
	return result
}
