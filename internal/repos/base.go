package repos

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// BaseRepo provides common repository functionality (DEPRECATED - Use DynamoBaseRepo instead)
// This is kept for backward compatibility but should not be used for new code
type BaseRepo struct {
	// Deprecated: Use DynamoBaseRepo instead
}

// NewBaseRepo creates a new base repository (DEPRECATED)
// Use NewDynamoBaseRepo instead
func NewBaseRepo(collectionName string) *BaseRepo {
	return &BaseRepo{}
}

// SetTimestamps sets created and updated timestamps
func (r *BaseRepo) SetTimestamps(data map[string]interface{}, isUpdate bool) {
	now := time.Now()
	if !isUpdate {
		data["created"] = now
	}
	data["updated"] = now
}

// GetCollection returns the collection reference (DEPRECATED)
func (r *BaseRepo) GetCollection() interface{} {
	return nil
}

// GetByID gets a document by ID (DEPRECATED)
func (r *BaseRepo) GetByID(ctx context.Context, id string) (interface{}, error) {
	return nil, fmt.Errorf("BaseRepo is deprecated, use DynamoBaseRepo instead")
}

// DeleteByID deletes a document by ID (DEPRECATED)
func (r *BaseRepo) DeleteByID(ctx context.Context, id string) error {
	return fmt.Errorf("BaseRepo is deprecated, use DynamoBaseRepo instead")
}

// Exists checks if a document exists (DEPRECATED)
func (r *BaseRepo) Exists(ctx context.Context, id string) (bool, error) {
	return false, fmt.Errorf("BaseRepo is deprecated, use DynamoBaseRepo instead")
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
