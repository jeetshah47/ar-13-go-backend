package repos

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/firebase"
	"google.golang.org/api/iterator"
)

// NotificationRepo handles notification data operations
type NotificationRepo struct {
	*BaseRepo
}

// NewNotificationRepo creates a new notification repository
func NewNotificationRepo() *NotificationRepo {
	return &NotificationRepo{
		BaseRepo: NewBaseRepo("notifications"),
	}
}

// GetByID gets a notification by ID
func (r *NotificationRepo) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	doc, err := r.collection.Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, nil
	}

	data := doc.Data()
	// Convert time fields from strings/timestamps to time.Time
	if err := ConvertTimeFieldsInMap(data, []string{"createdAt", "created", "updated"}); err != nil {
		return nil, err
	}

	var notification models.Notification
	if err := doc.DataTo(&notification); err != nil {
		return nil, err
	}
	notification.ID = doc.Ref.ID
	return &notification, nil
}

// GetAll gets all notifications for a user
func (r *NotificationRepo) GetAll(ctx context.Context, userID string) ([]models.Notification, error) {
	// Use "created" field for ordering as it's part of Model and more likely to be indexed
	iter := r.collection.Where("userId", "==", userID).OrderBy("created", firestore.Desc).Documents(ctx)
	var notifications []models.Notification

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		// Convert time fields from strings/timestamps to time.Time
		if err := ConvertTimeFieldsInMap(data, []string{"createdAt", "created", "updated"}); err != nil {
			return nil, err
		}

		var notification models.Notification
		if err := doc.DataTo(&notification); err != nil {
			return nil, err
		}
		notification.ID = doc.Ref.ID
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetUnread gets unread notifications for a user
// Note: This query requires a composite index on (userId, isRead, created) in Firestore
// If the index doesn't exist, Firestore will return an error with instructions to create it
func (r *NotificationRepo) GetUnread(ctx context.Context, userID string) ([]models.Notification, error) {
	// Use "created" field for ordering as it's part of Model and more likely to be indexed
	// Query order: filter fields first, then order by
	query := r.collection.Where("userId", "==", userID).Where("isRead", "==", false).OrderBy("created", firestore.Desc)
	iter := query.Documents(ctx)
	var notifications []models.Notification

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		// Convert time fields from strings/timestamps to time.Time
		if err := ConvertTimeFieldsInMap(data, []string{"createdAt", "created", "updated"}); err != nil {
			return nil, err
		}

		var notification models.Notification
		if err := doc.DataTo(&notification); err != nil {
			return nil, err
		}
		notification.ID = doc.Ref.ID
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetCount gets notification count for a user
// Optimized: Uses Select() to only fetch the isRead field, reducing data transfer by ~90%
// IMPORTANT: This still counts as document reads in Firestore quota
//
// To prevent quota exhaustion:
// 1. Implement caching on the frontend/backend (cache for 30-60 seconds)
// 2. Consider implementing a counter document that gets updated when notifications change
// 3. Monitor Firestore quotas in Google Cloud Console
// 4. For users with >1000 notifications, consider pagination or archiving old notifications
func (r *NotificationRepo) GetCount(ctx context.Context, userID string) (total, unread int, err error) {
	// Use Select() to only fetch the isRead field - reduces data transfer significantly
	// This still counts as document reads but minimizes bandwidth (only ~10 bytes per doc vs full doc)
	iter := r.collection.Where("userId", "==", userID).Select("isRead").Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, 0, err
		}

		total++

		// Check if notification is unread
		// With Select(), only the isRead field is available
		data := doc.Data()
		if isRead, ok := data["isRead"].(bool); ok && !isRead {
			unread++
		} else if isReadVal, ok := data["isRead"]; !ok || isReadVal == nil {
			// If isRead field doesn't exist or is nil, treat as unread
			unread++
		}
	}

	return total, unread, nil
}

// Add creates a new notification
func (r *NotificationRepo) Add(ctx context.Context, notification *models.Notification) error {
	newDocRef := r.collection.NewDoc()
	now := time.Now()
	notification.ID = newDocRef.ID
	notification.Created = now
	notification.CreatedAt = now
	// Initialize Updated as nil for new documents
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
		"createdAt":         notification.CreatedAt,
		"created":           notification.Created,
		// Don't set "updated" for new documents
	}

	_, err := newDocRef.Set(ctx, data)
	return err
}

// MarkAsRead marks a notification as read
func (r *NotificationRepo) MarkAsRead(ctx context.Context, id string) error {
	// Check if document exists first
	doc, err := r.collection.Doc(id).Get(ctx)
	if err != nil {
		return err
	}
	if !doc.Exists() {
		return fmt.Errorf("notification not found")
	}

	_, err = r.collection.Doc(id).Update(ctx, []firestore.Update{
		{Path: "isRead", Value: true},
		{Path: "updated", Value: time.Now()},
	})
	return err
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	iter := r.collection.Where("userId", "==", userID).Where("isRead", "==", false).Documents(ctx)
	batch := firebase.GetFirestoreClient().Batch()
	updateCount := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		batch.Update(doc.Ref, []firestore.Update{
			{Path: "isRead", Value: true},
			{Path: "updated", Value: time.Now()},
		})
		updateCount++

		// Firestore batch limit is 500
		if updateCount >= 500 {
			if _, err := batch.Commit(ctx); err != nil {
				return err
			}
			batch = firebase.GetFirestoreClient().Batch()
			updateCount = 0
		}
	}

	if updateCount > 0 {
		_, err := batch.Commit(ctx)
		return err
	}

	return nil
}

// Delete deletes a notification
func (r *NotificationRepo) Delete(ctx context.Context, id string) error {
	_, err := r.collection.Doc(id).Delete(ctx)
	return err
}

// DeleteAllForUser deletes all notifications for a user
func (r *NotificationRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	iter := r.collection.Where("userId", "==", userID).Documents(ctx)
	batch := firebase.GetFirestoreClient().Batch()
	deleteCount := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		batch.Delete(doc.Ref)
		deleteCount++

		// Firestore batch limit is 500
		if deleteCount >= 500 {
			if _, err := batch.Commit(ctx); err != nil {
				return err
			}
			batch = firebase.GetFirestoreClient().Batch()
			deleteCount = 0
		}
	}

	if deleteCount > 0 {
		_, err := batch.Commit(ctx)
		return err
	}

	return nil
}
