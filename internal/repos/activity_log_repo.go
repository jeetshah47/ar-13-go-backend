package repos

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/firebase"
	"google.golang.org/api/iterator"
)

// ActivityLogRepo handles activity log data operations
type ActivityLogRepo struct {
	*BaseRepo
}

// NewActivityLogRepo creates a new activity log repository
func NewActivityLogRepo() *ActivityLogRepo {
	return &ActivityLogRepo{
		BaseRepo: NewBaseRepo("activityLogs"),
	}
}

// getSubCollection gets the sub-collection for an entity type
func (r *ActivityLogRepo) getSubCollection(entityType models.ActivityLogEntityType) *firestore.CollectionRef {
	subCollectionName := ""
	switch entityType {
	case models.ActivityLogEntityTypeTask:
		subCollectionName = "taskActivityLogs"
	case models.ActivityLogEntityTypeProject:
		subCollectionName = "projectActivityLogs"
	case models.ActivityLogEntityTypeUser:
		subCollectionName = "userActivityLogs"
	case models.ActivityLogEntityTypeCalendarEvent:
		subCollectionName = "calendarEventActivityLogs"
	}

	return firebase.GetCollection("activityLogs").
		Doc("logs").
		Collection(subCollectionName)
}

// Add adds an activity log
func (r *ActivityLogRepo) Add(ctx context.Context, log *models.ActivityLogBase) error {
	subCollection := r.getSubCollection(log.EntityType)
	newDocRef := subCollection.NewDoc()
	log.ID = newDocRef.ID

	data := map[string]interface{}{
		"id":         log.ID,
		"entityType": string(log.EntityType),
		"entityId":   log.EntityID,
		"action":     string(log.Action),
		"createdAt":  log.CreatedAt,
		"createdBy":  log.CreatedBy,
		"created":    log.Created,
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

	_, err := newDocRef.Set(ctx, data)
	return err
}

// GetByEntity gets activity logs for a specific entity
func (r *ActivityLogRepo) GetByEntity(ctx context.Context, entityType models.ActivityLogEntityType, entityID string) ([]models.ActivityLogBase, error) {
	subCollection := r.getSubCollection(entityType)
	iter := subCollection.Where("entityId", "==", entityID).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx)

	var logs []models.ActivityLogBase
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

		var log models.ActivityLogBase
		if err := doc.DataTo(&log); err != nil {
			return nil, err
		}
		log.ID = doc.Ref.ID
		logs = append(logs, log)
	}

	return logs, nil
}

// GetByEntityType gets activity logs by entity type
func (r *ActivityLogRepo) GetByEntityType(ctx context.Context, entityType models.ActivityLogEntityType, limit *int) ([]models.ActivityLogBase, error) {
	subCollection := r.getSubCollection(entityType)
	query := subCollection.OrderBy("createdAt", firestore.Desc)
	if limit != nil {
		query = query.Limit(*limit)
	}

	iter := query.Documents(ctx)
	var logs []models.ActivityLogBase

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

		var log models.ActivityLogBase
		if err := doc.DataTo(&log); err != nil {
			return nil, err
		}
		log.ID = doc.Ref.ID
		logs = append(logs, log)
	}

	return logs, nil
}
