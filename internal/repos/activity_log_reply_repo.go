package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// ActivityLogReplyRepo handles activity log reply data operations with MongoDB
type ActivityLogReplyRepo struct {
	*MongoBaseRepo
}

// NewActivityLogReplyRepo creates a new MongoDB activity log reply repository
func NewActivityLogReplyRepo() *ActivityLogReplyRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &ActivityLogReplyRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "activity_log_replies"),
	}
}

// Add adds an activity log reply
func (r *ActivityLogReplyRepo) Add(ctx context.Context, reply *models.ActivityLogReply) error {
	now := time.Now()

	if reply.ID == "" {
		reply.ID = fmt.Sprintf("reply-%s-%d", reply.ActivityLogID, now.UnixNano())
	}

	// Initialize Created and CreatedAt if they are zero values
	if reply.Created.IsZero() {
		reply.Created = now
	}
	if reply.CreatedAt.IsZero() {
		reply.CreatedAt = now
	}

	return r.InsertOne(ctx, reply)
}

// GetByActivityLogID gets all replies for a specific activity log
func (r *ActivityLogReplyRepo) GetByActivityLogID(ctx context.Context, activityLogID string) ([]models.ActivityLogReply, error) {
	filter := bson.M{"activityLogId": activityLogID}
	items, err := r.FindAll(ctx, filter, nil, bson.M{"createdAt": 1}) // Sort by createdAt ascending (oldest first)
	if err != nil {
		return nil, err
	}

	replies := make([]models.ActivityLogReply, 0, len(items))
	for _, item := range items {
		var reply models.ActivityLogReply
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &reply); err != nil {
			continue
		}
		replies = append(replies, reply)
	}

	return replies, nil
}

