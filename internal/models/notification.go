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
	NotificationTypeLeaveRequestUpdated  NotificationType = "LEAVE_REQUEST_UPDATED"
	NotificationTypeLeaveRequestApproved NotificationType = "LEAVE_REQUEST_APPROVED"
	NotificationTypeLeaveRequestRejected NotificationType = "LEAVE_REQUEST_REJECTED"
	NotificationTypeUserLogin            NotificationType = "USER_LOGIN"
	NotificationTypeUserLogout           NotificationType = "USER_LOGOUT"
	NotificationTypeUserCreated          NotificationType = "USER_CREATED"
	NotificationTypeUserUpdated          NotificationType = "USER_UPDATED"
	NotificationTypeCalendarEventCreated  NotificationType = "CALENDAR_EVENT_CREATED"
	NotificationTypeCalendarEventUpdated  NotificationType = "CALENDAR_EVENT_UPDATED"
	NotificationTypeProjectDetailsCreated NotificationType = "PROJECT_DETAILS_CREATED"
	NotificationTypeProjectDetailsUpdated NotificationType = "PROJECT_DETAILS_UPDATED"
	NotificationTypeInfoPortalFolderCreated NotificationType = "INFO_PORTAL_FOLDER_CREATED"
	NotificationTypeInfoPortalFolderUpdated NotificationType = "INFO_PORTAL_FOLDER_UPDATED"
	NotificationTypeInfoPortalPageCreated   NotificationType = "INFO_PORTAL_PAGE_CREATED"
	NotificationTypeInfoPortalPageUpdated   NotificationType = "INFO_PORTAL_PAGE_UPDATED"
	NotificationTypeDrawingListCategoryCreated NotificationType = "DRAWING_LIST_CATEGORY_CREATED"
	NotificationTypeDrawingListCategoryUpdated NotificationType = "DRAWING_LIST_CATEGORY_UPDATED"
	NotificationTypeDrawingListTypeCreated    NotificationType = "DRAWING_LIST_TYPE_CREATED"
	NotificationTypeDrawingListTypeUpdated    NotificationType = "DRAWING_LIST_TYPE_UPDATED"
)

// RelatedEntityType represents the type of related entity
type RelatedEntityType string

const (
	RelatedEntityTypeProject      RelatedEntityType = "PROJECT"
	RelatedEntityTypeTask         RelatedEntityType = "TASK"
	RelatedEntityTypeLeaveRequest RelatedEntityType = "LEAVE_REQUEST"
	RelatedEntityTypeUser         RelatedEntityType = "USER"
	RelatedEntityTypeCalendarEvent RelatedEntityType = "CALENDAR_EVENT"
	RelatedEntityTypeProjectDetails RelatedEntityType = "PROJECT_DETAILS"
	RelatedEntityTypeInfoPortalFolder RelatedEntityType = "INFO_PORTAL_FOLDER"
	RelatedEntityTypeInfoPortalPage   RelatedEntityType = "INFO_PORTAL_PAGE"
	RelatedEntityTypeDrawingListCategory RelatedEntityType = "DRAWING_LIST_CATEGORY"
	RelatedEntityTypeDrawingListType     RelatedEntityType = "DRAWING_LIST_TYPE"
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
