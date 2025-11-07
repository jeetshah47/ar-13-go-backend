package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// NotificationRepo handles notification data operations
type NotificationRepo struct {
	*DynamoBaseRepo
}

// NewNotificationRepo creates a new notification repository
func NewNotificationRepo() *NotificationRepo {
	return &NotificationRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("notifications"),
	}
}

// GetByID gets a notification by ID
func (r *NotificationRepo) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	item, err := r.DynamoBaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var notification models.Notification
	if err := UnmarshalItem(item, &notification); err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetAll gets all notifications for a user
func (r *NotificationRepo) GetAll(ctx context.Context, userID string) ([]models.Notification, error) {
	// Query by userId using GSI
	items, err := r.QueryByIndex(ctx, "userId-index", "userId", userID)
	if err != nil {
		return nil, err
	}

	notifications := make([]models.Notification, 0, len(items))
	for _, item := range items {
		var notification models.Notification
		if err := UnmarshalItem(item, &notification); err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetUnread gets unread notifications for a user
func (r *NotificationRepo) GetUnread(ctx context.Context, userID string) ([]models.Notification, error) {
	// Query by userId using GSI, then filter by isRead
	items, err := r.QueryByIndex(ctx, "userId-index", "userId", userID)
	if err != nil {
		return nil, err
	}

	notifications := make([]models.Notification, 0)
	for _, item := range items {
		// Filter by isRead
		if isRead, ok := item["isRead"].(*types.AttributeValueMemberBOOL); ok && !isRead.Value {
			var notification models.Notification
			if err := UnmarshalItem(item, &notification); err != nil {
				return nil, err
			}
			notifications = append(notifications, notification)
		}
	}

	return notifications, nil
}

// GetCount gets notification count for a user
func (r *NotificationRepo) GetCount(ctx context.Context, userID string) (total, unread int, err error) {
	// Query by userId using GSI
	items, err := r.QueryByIndex(ctx, "userId-index", "userId", userID)
	if err != nil {
		return 0, 0, err
	}

	total = len(items)
	for _, item := range items {
		// Check if notification is unread
		if isRead, ok := item["isRead"].(*types.AttributeValueMemberBOOL); ok && !isRead.Value {
			unread++
		} else if _, ok := item["isRead"]; !ok {
			// If isRead field doesn't exist, treat as unread
			unread++
		}
	}

	return total, unread, nil
}

// Add creates a new notification
func (r *NotificationRepo) Add(ctx context.Context, notification *models.Notification) error {
	now := time.Now()
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}
	notification.Created = now
	notification.CreatedAt = now
	notification.Updated = nil

	data := map[string]interface{}{
		"id":                notification.ID,
		"title":             notification.Title,
		"message":           notification.Message,
		"type":              string(notification.Type),
		"userId":            notification.UserID,
		"relatedEntityId":   notification.RelatedEntityID,
		"relatedEntityType": string(notification.RelatedEntityType),
		"isRead":            notification.IsRead,
		"createdAt":         notification.CreatedAt.Format(time.RFC3339),
		"created":           notification.Created.Format(time.RFC3339),
	}

	return r.PutItem(ctx, data)
}

// MarkAsRead marks a notification as read
func (r *NotificationRepo) MarkAsRead(ctx context.Context, id string) error {
	updates := map[string]interface{}{
		"isRead": true,
	}
	return r.UpdateItem(ctx, id, updates)
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	// Get all unread notifications
	items, err := r.QueryByIndex(ctx, "userId-index", "userId", userID)
	if err != nil {
		return err
	}

	// Update each unread notification
	for _, item := range items {
		if isRead, ok := item["isRead"].(*types.AttributeValueMemberBOOL); ok && !isRead.Value {
			if id, ok := item["id"].(*types.AttributeValueMemberS); ok {
				updates := map[string]interface{}{
					"isRead": true,
				}
				if err := r.UpdateItem(ctx, id.Value, updates); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Delete deletes a notification
func (r *NotificationRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// DeleteAllForUser deletes all notifications for a user
func (r *NotificationRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	// Get all notifications for user
	items, err := r.QueryByIndex(ctx, "userId-index", "userId", userID)
	if err != nil {
		return err
	}

	// Delete each notification
	for _, item := range items {
		if id, ok := item["id"].(*types.AttributeValueMemberS); ok {
			if err := r.DeleteByID(ctx, id.Value); err != nil {
				return err
			}
		}
	}

	return nil
}
