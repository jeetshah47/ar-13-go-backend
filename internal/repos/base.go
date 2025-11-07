package repos

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/pkg/firebase"
)

// BaseRepo provides common repository functionality
type BaseRepo struct {
	collection *firestore.CollectionRef
}

// NewBaseRepo creates a new base repository
func NewBaseRepo(collectionName string) *BaseRepo {
	return &BaseRepo{
		collection: firebase.GetCollection(collectionName),
	}
}

// SetTimestamps sets created and updated timestamps
func (r *BaseRepo) SetTimestamps(data map[string]interface{}, isUpdate bool) {
	now := time.Now()
	if !isUpdate {
		data["created"] = now
	}
	data["updated"] = now
}

// GetCollection returns the collection reference
func (r *BaseRepo) GetCollection() *firestore.CollectionRef {
	return r.collection
}

// GetByID gets a document by ID
func (r *BaseRepo) GetByID(ctx context.Context, id string) (*firestore.DocumentSnapshot, error) {
	return r.collection.Doc(id).Get(ctx)
}

// DeleteByID deletes a document by ID
func (r *BaseRepo) DeleteByID(ctx context.Context, id string) error {
	_, err := r.collection.Doc(id).Delete(ctx)
	return err
}

// Exists checks if a document exists
func (r *BaseRepo) Exists(ctx context.Context, id string) (bool, error) {
	doc, err := r.collection.Doc(id).Get(ctx)
	if err != nil {
		return false, err
	}
	return doc.Exists(), nil
}

// ConvertToTime converts a value to time.Time, handling strings, time.Time, Unix timestamps, and Firestore timestamps
// This is a shared utility function for all repositories to handle time conversion from Firestore
func ConvertToTime(val interface{}) (time.Time, error) {
	if val == nil {
		return time.Time{}, fmt.Errorf("value is nil")
	}

	switch v := val.(type) {
	case time.Time:
		return v, nil
	case string:
		// First, try to parse as Unix timestamp (milliseconds or seconds)
		if timestamp, err := strconv.ParseInt(v, 10, 64); err == nil {
			// Check if it's milliseconds (13 digits) or seconds (10 digits)
			if timestamp > 1e12 {
				// Milliseconds
				return time.Unix(0, timestamp*int64(time.Millisecond)), nil
			} else {
				// Seconds
				return time.Unix(timestamp, 0), nil
			}
		}

		// Try to parse the string as time.Time in various formats
		parsedTime, err := time.Parse(time.RFC3339, v)
		if err != nil {
			// Try other common formats
			parsedTime, err = time.Parse("2006-01-02T15:04:05Z07:00", v)
			if err != nil {
				parsedTime, err = time.Parse("2006-01-02 15:04:05", v)
				if err != nil {
					return time.Time{}, fmt.Errorf("failed to parse time string: %w", err)
				}
			}
		}
		return parsedTime, nil
	case int64:
		// Handle Unix timestamp (milliseconds or seconds)
		if v > 1e12 {
			// Milliseconds
			return time.Unix(0, v*int64(time.Millisecond)), nil
		} else {
			// Seconds
			return time.Unix(v, 0), nil
		}
	case int:
		// Handle Unix timestamp (milliseconds or seconds)
		if int64(v) > 1e12 {
			// Milliseconds
			return time.Unix(0, int64(v)*int64(time.Millisecond)), nil
		} else {
			// Seconds
			return time.Unix(int64(v), 0), nil
		}
	case float64:
		// Handle Unix timestamp (milliseconds or seconds) as float
		timestamp := int64(v)
		if timestamp > 1e12 {
			// Milliseconds
			return time.Unix(0, timestamp*int64(time.Millisecond)), nil
		} else {
			// Seconds
			return time.Unix(timestamp, 0), nil
		}
	default:
		// Try to use Firestore Timestamp if available
		// Use reflection or type assertion to check for *firestore.Timestamp
		if ts, ok := val.(interface{ ToTime() time.Time }); ok {
			return ts.ToTime(), nil
		}
		return time.Time{}, fmt.Errorf("cannot convert type %T to time.Time", val)
	}
}

// ConvertTimeFieldsInMap converts all time fields in a data map from strings/timestamps to time.Time
func ConvertTimeFieldsInMap(data map[string]interface{}, timeFields []string) error {
	for _, field := range timeFields {
		if val, ok := data[field]; ok && val != nil {
			convertedTime, err := ConvertToTime(val)
			if err != nil {
				return fmt.Errorf("failed to convert field %s: %w", field, err)
			}
			data[field] = convertedTime
		}
	}
	return nil
}
