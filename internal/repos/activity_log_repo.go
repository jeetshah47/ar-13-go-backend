package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// ActivityLogRepo handles activity log data operations
type ActivityLogRepo struct {
	*DynamoBaseRepo
}

// NewActivityLogRepo creates a new activity log repository
func NewActivityLogRepo() *ActivityLogRepo {
	return &ActivityLogRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("activity_logs"),
	}
}

// Add adds an activity log
func (r *ActivityLogRepo) Add(ctx context.Context, log *models.ActivityLogBase) error {
	if log.ID == "" {
		log.ID = fmt.Sprintf("%s-%s-%d", log.EntityType, log.EntityID, time.Now().Unix())
	}

	data := map[string]interface{}{
		"id":         log.ID,
		"entityType": string(log.EntityType),
		"entityId":   log.EntityID,
		"action":     string(log.Action),
		"createdAt":  log.CreatedAt.Format(time.RFC3339),
		"createdBy":  log.CreatedBy,
		"created":    log.Created.Format(time.RFC3339),
	}
	if log.Description != nil {
		data["description"] = *log.Description
	}
	if log.Fields != nil {
		data["fields"] = log.Fields
	}
	if log.Metadata != nil {
		data["metadata"] = log.Metadata
	}

	return r.PutItem(ctx, data)
}

// GetByEntity gets activity logs for a specific entity
func (r *ActivityLogRepo) GetByEntity(ctx context.Context, entityType models.ActivityLogEntityType, entityID string) ([]models.ActivityLogBase, error) {
	items, err := r.QueryByIndex(ctx, "entityId-index", "entityId", entityID)
	if err != nil {
		return nil, err
	}

	logs := make([]models.ActivityLogBase, 0, len(items))
	for _, item := range items {
		var log models.ActivityLogBase
		if err := UnmarshalItem(item, &log); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetByEntityType gets activity logs by entity type
func (r *ActivityLogRepo) GetByEntityType(ctx context.Context, entityType models.ActivityLogEntityType, limit *int) ([]models.ActivityLogBase, error) {
	var limitInt32 *int32
	if limit != nil {
		l := int32(*limit)
		limitInt32 = &l
	}

	items, err := r.ScanItems(ctx, limitInt32)
	if err != nil {
		return nil, err
	}

	logs := make([]models.ActivityLogBase, 0)
	for _, item := range items {
		// Filter by entityType
		if et, ok := item["entityType"].(*types.AttributeValueMemberS); ok && et.Value == string(entityType) {
			var log models.ActivityLogBase
			if err := UnmarshalItem(item, &log); err != nil {
				return nil, err
			}
			logs = append(logs, log)
		}
	}

	return logs, nil
}
