package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// AuditLogRepo handles audit log data operations with MongoDB
type AuditLogRepo struct {
	*MongoBaseRepo
}

// NewAuditLogRepo creates a new MongoDB audit log repository
func NewAuditLogRepo() *AuditLogRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &AuditLogRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "audit_logs"),
	}
}

// Add adds an audit log entry
func (r *AuditLogRepo) Add(ctx context.Context, log *models.AuditLog) error {
	now := time.Now()

	if log.ID == "" {
		// Generate ID: method-path-timestamp-userId (if available)
		userIdPart := ""
		if log.UserID != nil && *log.UserID != "" {
			userIdPart = "-" + *log.UserID
		}
		log.ID = fmt.Sprintf("audit-%s-%s-%d%s", log.Method, log.Path, now.Unix(), userIdPart)
	}

	// Initialize Created and RequestTime if they are zero values
	if log.Created.IsZero() {
		log.Created = now
	}
	if log.RequestTime.IsZero() {
		log.RequestTime = now
	}

	return r.InsertOne(ctx, log)
}

// GetByUserID gets audit logs for a specific user
func (r *AuditLogRepo) GetByUserID(ctx context.Context, userID string, limit *int) ([]models.AuditLog, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}

	filter := bson.M{"userId": userID}
	items, err := r.FindAll(ctx, filter, limitInt64, bson.M{"requestTime": -1}) // Sort by requestTime descending
	if err != nil {
		return nil, err
	}

	logs := make([]models.AuditLog, 0, len(items))
	for _, item := range items {
		var log models.AuditLog
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetByPath gets audit logs for a specific path/endpoint
func (r *AuditLogRepo) GetByPath(ctx context.Context, path string, limit *int) ([]models.AuditLog, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}

	filter := bson.M{"path": path}
	items, err := r.FindAll(ctx, filter, limitInt64, bson.M{"requestTime": -1})
	if err != nil {
		return nil, err
	}

	logs := make([]models.AuditLog, 0, len(items))
	for _, item := range items {
		var log models.AuditLog
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetByMethod gets audit logs for a specific HTTP method
func (r *AuditLogRepo) GetByMethod(ctx context.Context, method string, limit *int) ([]models.AuditLog, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}

	filter := bson.M{"method": method}
	items, err := r.FindAll(ctx, filter, limitInt64, bson.M{"requestTime": -1})
	if err != nil {
		return nil, err
	}

	logs := make([]models.AuditLog, 0, len(items))
	for _, item := range items {
		var log models.AuditLog
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetByStatusCode gets audit logs for a specific status code
func (r *AuditLogRepo) GetByStatusCode(ctx context.Context, statusCode int, limit *int) ([]models.AuditLog, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}

	filter := bson.M{"statusCode": statusCode}
	items, err := r.FindAll(ctx, filter, limitInt64, bson.M{"requestTime": -1})
	if err != nil {
		return nil, err
	}

	logs := make([]models.AuditLog, 0, len(items))
	for _, item := range items {
		var log models.AuditLog
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetRecent gets recent audit logs with optional filters
func (r *AuditLogRepo) GetRecent(ctx context.Context, limit int, filters bson.M) ([]models.AuditLog, error) {
	limitInt64 := int64(limit)

	if filters == nil {
		filters = bson.M{}
	}

	items, err := r.FindAll(ctx, filters, &limitInt64, bson.M{"requestTime": -1})
	if err != nil {
		return nil, err
	}

	logs := make([]models.AuditLog, 0, len(items))
	for _, item := range items {
		var log models.AuditLog
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetByDateRange gets audit logs within a date range
func (r *AuditLogRepo) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit *int) ([]models.AuditLog, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}

	filter := bson.M{
		"requestTime": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	items, err := r.FindAll(ctx, filter, limitInt64, bson.M{"requestTime": -1})
	if err != nil {
		return nil, err
	}

	logs := make([]models.AuditLog, 0, len(items))
	for _, item := range items {
		var log models.AuditLog
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

