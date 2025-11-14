package models

import "time"

// ActivityLogEntityType represents the type of entity
type ActivityLogEntityType string

const (
	ActivityLogEntityTypeTask          ActivityLogEntityType = "task"
	ActivityLogEntityTypeProject       ActivityLogEntityType = "project"
	ActivityLogEntityTypeUser          ActivityLogEntityType = "user"
	ActivityLogEntityTypeCalendarEvent ActivityLogEntityType = "calendarEvent"
)

// ActivityLogAction represents the action performed
type ActivityLogAction string

const (
	ActivityLogActionCreated            ActivityLogAction = "created"
	ActivityLogActionUpdated            ActivityLogAction = "updated"
	ActivityLogActionDeleted            ActivityLogAction = "deleted"
	ActivityLogActionAssigned           ActivityLogAction = "assigned"
	ActivityLogActionUnassigned         ActivityLogAction = "unassigned"
	ActivityLogActionStatusChanged      ActivityLogAction = "status_changed"
	ActivityLogActionPriorityChanged    ActivityLogAction = "priority_changed"
	ActivityLogActionDescriptionUpdated ActivityLogAction = "description_updated"
	ActivityLogActionDurationUpdated    ActivityLogAction = "duration_updated"
	ActivityLogActionProgressUpdated    ActivityLogAction = "progress_updated"
	ActivityLogActionFileUploaded       ActivityLogAction = "file_uploaded"
	ActivityLogActionFileRemoved        ActivityLogAction = "file_removed"
	ActivityLogActionTimeSpentAdded     ActivityLogAction = "time_spent_added"
	ActivityLogActionTimeSpentUpdated   ActivityLogAction = "time_spent_updated"
	ActivityLogActionTimeSpentRemoved   ActivityLogAction = "time_spent_removed"
	ActivityLogActionMemberAdded        ActivityLogAction = "member_added"
	ActivityLogActionMemberRemoved      ActivityLogAction = "member_removed"
	ActivityLogActionOwnerChanged       ActivityLogAction = "owner_changed"
	ActivityLogActionDeadlineUpdated    ActivityLogAction = "deadline_updated"
	ActivityLogActionRoleChanged        ActivityLogAction = "role_changed"
	ActivityLogActionPasswordChanged    ActivityLogAction = "password_changed"
	ActivityLogActionProfileUpdated     ActivityLogAction = "profile_updated"
	ActivityLogActionEventCreated       ActivityLogAction = "event_created"
	ActivityLogActionEventUpdated       ActivityLogAction = "event_updated"
	ActivityLogActionEventDeleted       ActivityLogAction = "event_deleted"
	ActivityLogActionEventCancelled     ActivityLogAction = "event_cancelled"
	ActivityLogActionOther              ActivityLogAction = "other"
)

// ActivityLogBase represents the base activity log structure
type ActivityLogBase struct {
	Model
	EntityType  ActivityLogEntityType  `json:"entityType" firestore:"entityType" bson:"entityType"`
	EntityID    string                 `json:"entityId" firestore:"entityId" bson:"entityId"`
	Action      ActivityLogAction      `json:"action" firestore:"action" bson:"action"`
	CreatedAt   time.Time              `json:"createdAt" firestore:"createdAt" bson:"createdAt"`
	CreatedBy   string                 `json:"createdBy" firestore:"createdBy" bson:"createdBy"`
	Fields      map[string]interface{} `json:"fields,omitempty" firestore:"fields,omitempty" bson:"fields,omitempty"`
	Description *string                `json:"description,omitempty" firestore:"description,omitempty" bson:"description,omitempty"`
	Remark      *string                `json:"remark,omitempty" firestore:"remark,omitempty" bson:"remark,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty" bson:"metadata,omitempty"`
}

// TaskActivityLog represents a task-specific activity log
type TaskActivityLog struct {
	ActivityLogBase
	ProjectID *string `json:"projectId,omitempty" firestore:"projectId,omitempty"`
}

// ActivityLogResponse represents an activity log with user information
type ActivityLogResponse struct {
	ActivityLogBase
	CreatedByUser *User `json:"createdByUser,omitempty" firestore:"-"`
}
