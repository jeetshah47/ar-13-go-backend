package models

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeProjectCreated       NotificationType = "PROJECT_CREATED"
	NotificationTypeTaskCreated          NotificationType = "TASK_CREATED"
	NotificationTypeTaskAssigned         NotificationType = "TASK_ASSIGNED"
	NotificationTypeProjectUpdated       NotificationType = "PROJECT_UPDATED"
	NotificationTypeTaskUpdated          NotificationType = "TASK_UPDATED"
	NotificationTypeLeaveRequestCreated  NotificationType = "LEAVE_REQUEST_CREATED"
	NotificationTypeLeaveRequestApproved NotificationType = "LEAVE_REQUEST_APPROVED"
	NotificationTypeLeaveRequestRejected NotificationType = "LEAVE_REQUEST_REJECTED"
	NotificationTypeUserLogin            NotificationType = "USER_LOGIN"
	NotificationTypeUserLogout           NotificationType = "USER_LOGOUT"
)

// RelatedEntityType represents the type of related entity
type RelatedEntityType string

const (
	RelatedEntityTypeProject      RelatedEntityType = "PROJECT"
	RelatedEntityTypeTask         RelatedEntityType = "TASK"
	RelatedEntityTypeLeaveRequest RelatedEntityType = "LEAVE_REQUEST"
	RelatedEntityTypeUser         RelatedEntityType = "USER"
)

// Notification represents a notification
type Notification struct {
	Model
	Title             string            `json:"title" firestore:"title" bson:"title"`
	Message           string            `json:"message" firestore:"message" bson:"message"`
	Type              NotificationType  `json:"type" firestore:"type" bson:"type"`
	UserID            string            `json:"userId" firestore:"userId" bson:"userId"`
	RelatedEntityID   string            `json:"relatedEntityId" firestore:"relatedEntityId" bson:"relatedEntityId"`
	RelatedEntityType RelatedEntityType `json:"relatedEntityType" firestore:"relatedEntityType" bson:"relatedEntityType"`
	IsRead            bool              `json:"isRead" firestore:"isRead" bson:"isRead"`
	CreatedAt         time.Time         `json:"createdAt" firestore:"createdAt" bson:"createdAt"`
}
