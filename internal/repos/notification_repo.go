package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// NotificationRepo handles notification data operations with MongoDB
type NotificationRepo struct {
	*MongoBaseRepo
}

// NewNotificationRepo creates a new MongoDB notification repository
func NewNotificationRepo() *NotificationRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &NotificationRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "notifications"),
	}
}

// GetByID gets a notification by ID
func (r *NotificationRepo) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	result := r.MongoBaseRepo.GetByID(ctx, id)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var notification models.Notification
	if err := result.Decode(&notification); err != nil {
		return nil, err
	}

	return &notification, nil
}

// GetAll gets all notifications for a user
func (r *NotificationRepo) GetAll(ctx context.Context, userID string) ([]models.Notification, error) {
	filter := bson.M{"userId": userID}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	notifications := make([]models.Notification, 0, len(items))
	for _, item := range items {
		var notification models.Notification
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &notification); err != nil {
			continue
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetUnread gets unread notifications for a user
func (r *NotificationRepo) GetUnread(ctx context.Context, userID string) ([]models.Notification, error) {
	filter := bson.M{"userId": userID, "isRead": false}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	notifications := make([]models.Notification, 0, len(items))
	for _, item := range items {
		var notification models.Notification
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &notification); err != nil {
			continue
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetCount gets notification count for a user
func (r *NotificationRepo) GetCount(ctx context.Context, userID string) (total, unread int, err error) {
	filter := bson.M{"userId": userID}
	totalCount, err := r.CountDocuments(ctx, filter)
	if err != nil {
		return 0, 0, err
	}

	unreadFilter := bson.M{"userId": userID, "isRead": false}
	unreadCount, err := r.CountDocuments(ctx, unreadFilter)
	if err != nil {
		return 0, 0, err
	}

	return int(totalCount), int(unreadCount), nil
}

// Add creates a new notification
func (r *NotificationRepo) Add(ctx context.Context, notification *models.Notification) error {
	now := time.Now()
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}
	notification.Created = now
	notification.CreatedAt = now

	return r.InsertOne(ctx, notification)
}

// MarkAsRead marks a notification as read
func (r *NotificationRepo) MarkAsRead(ctx context.Context, id string) error {
	updates := bson.M{"isRead": true}
	return r.UpdateOne(ctx, id, updates)
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	filter := bson.M{"userId": userID, "isRead": false}
	updates := bson.M{"isRead": true}
	_, err := r.UpdateMany(ctx, filter, updates)
	return err
}

// Delete deletes a notification
func (r *NotificationRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// DeleteAllForUser deletes all notifications for a user
func (r *NotificationRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	filter := bson.M{"userId": userID}
	_, err := r.DeleteMany(ctx, filter)
	return err
}

