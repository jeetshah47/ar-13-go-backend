package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ActivityLogRepo handles activity log data operations with MongoDB
type ActivityLogRepo struct {
	*MongoBaseRepo
}

// NewActivityLogRepo creates a new MongoDB activity log repository
func NewActivityLogRepo() *ActivityLogRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &ActivityLogRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "activity_logs"),
	}
}

// Add adds an activity log
func (r *ActivityLogRepo) Add(ctx context.Context, log *models.ActivityLogBase) error {
	now := time.Now()

	if log.ID == "" {
		log.ID = fmt.Sprintf("%s-%s-%d", log.EntityType, log.EntityID, now.Unix())
	}

	// Initialize Created and CreatedAt if they are zero values
	if log.Created.IsZero() {
		log.Created = now
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}

	return r.InsertOne(ctx, log)
}

// GetByEntity gets activity logs for a specific entity
func (r *ActivityLogRepo) GetByEntity(ctx context.Context, entityType models.ActivityLogEntityType, entityID string) ([]models.ActivityLogBase, error) {
	filter := bson.M{"entityId": entityID}
	items, err := r.FindAll(ctx, filter, nil, bson.M{"createdAt": -1}) // Sort by createdAt descending
	if err != nil {
		return nil, err
	}

	logs := make([]models.ActivityLogBase, 0, len(items))
	for _, item := range items {
		var log models.ActivityLogBase
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetByEntityType gets activity logs by entity type
func (r *ActivityLogRepo) GetByEntityType(ctx context.Context, entityType models.ActivityLogEntityType, limit *int) ([]models.ActivityLogBase, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}

	filter := bson.M{"entityType": string(entityType)}
	items, err := r.FindAll(ctx, filter, limitInt64, bson.M{"createdAt": -1})
	if err != nil {
		return nil, err
	}

	logs := make([]models.ActivityLogBase, 0, len(items))
	for _, item := range items {
		var log models.ActivityLogBase
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetByID gets an activity log by ID
func (r *ActivityLogRepo) GetByID(ctx context.Context, activityLogID string) (*models.ActivityLogBase, error) {
	// Query by id field - check both root level and nested "model.id" and _id
	// (MongoDB may store embedded structs as nested objects)
	filter := bson.M{
		"$or": []bson.M{
			{"id": activityLogID},
			{"model.id": activityLogID},
			{"_id": activityLogID},
		},
	}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		// If not found by ID, try to parse the ID format: entityType-entityId-timestamp
		// and search by entityType and entityId combination as fallback
		// This handles cases where ID format might have changed
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	// Decode into a map first for better compatibility with MongoDB document structure
	var rawDoc bson.M
	if err := result.Decode(&rawDoc); err != nil {
		return nil, err
	}

	// Convert map to BSON bytes, then unmarshal to struct
	bsonBytes, err := bson.Marshal(rawDoc)
	if err != nil {
		return nil, err
	}

	var log models.ActivityLogBase
	if err := bson.Unmarshal(bsonBytes, &log); err != nil {
		return nil, err
	}

	return &log, nil
}

