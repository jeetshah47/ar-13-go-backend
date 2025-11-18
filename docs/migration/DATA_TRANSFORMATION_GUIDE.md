# DynamoDB to MongoDB Data Transformation Guide

This guide explains how to transform DynamoDB data format to MongoDB format during migration.

## Overview

DynamoDB uses AttributeValue types, while MongoDB uses native BSON types. This guide shows how to convert between them.

## Data Type Mappings

| DynamoDB Type | MongoDB Type | Notes |
|--------------|-------------|-------|
| `S` (String) | `string` | Direct mapping |
| `N` (Number) | `number` | Convert string to number |
| `BOOL` (Boolean) | `boolean` | Direct mapping |
| `NULL` | `null` | Direct mapping |
| `L` (List) | `array` | Recursively convert items |
| `M` (Map) | `object` | Recursively convert fields |
| `SS` (String Set) | `array` | Convert to array |
| `NS` (Number Set) | `array` | Convert to array of numbers |
| `BS` (Binary Set) | `array` | Convert to array of binaries |

## Transformation Functions

### Basic Type Conversion

```go
package transform

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ConvertAttributeValue converts DynamoDB AttributeValue to MongoDB-compatible value
func ConvertAttributeValue(av types.AttributeValue) (interface{}, error) {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		return v.Value, nil
		
	case *types.AttributeValueMemberN:
		// Try int first, then float64
		if intVal, err := strconv.ParseInt(v.Value, 10, 64); err == nil {
			return intVal, nil
		}
		return strconv.ParseFloat(v.Value, 64)
		
	case *types.AttributeValueMemberBOOL:
		return v.Value, nil
		
	case *types.AttributeValueMemberNULL:
		return nil, nil
		
	case *types.AttributeValueMemberL:
		// Convert list
		result := make([]interface{}, len(v.Value))
		for i, item := range v.Value {
			converted, err := ConvertAttributeValue(item)
			if err != nil {
				return nil, fmt.Errorf("error converting list item %d: %w", i, err)
			}
			result[i] = converted
		}
		return result, nil
		
	case *types.AttributeValueMemberM:
		// Convert map
		result := make(map[string]interface{})
		for key, value := range v.Value {
			converted, err := ConvertAttributeValue(value)
			if err != nil {
				return nil, fmt.Errorf("error converting map field %s: %w", key, err)
			}
			result[key] = converted
		}
		return result, nil
		
	case *types.AttributeValueMemberSS:
		// String set to array
		return v.Value, nil
		
	case *types.AttributeValueMemberNS:
		// Number set to array of numbers
		result := make([]float64, len(v.Value))
		for i, s := range v.Value {
			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing number set item: %w", err)
			}
			result[i] = val
		}
		return result, nil
		
	case *types.AttributeValueMemberBS:
		// Binary set to array
		return v.Value, nil
		
	default:
		return nil, fmt.Errorf("unsupported attribute value type: %T", av)
	}
}

// ConvertDynamoItem converts a DynamoDB item to MongoDB document
func ConvertDynamoItem(item map[string]types.AttributeValue) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	for key, value := range item {
		converted, err := ConvertAttributeValue(value)
		if err != nil {
			return nil, fmt.Errorf("error converting field %s: %w", key, err)
		}
		result[key] = converted
	}
	
	return result, nil
}
```

## Date/Time Conversion

DynamoDB stores dates as RFC3339 strings, MongoDB uses Date objects.

```go
// ConvertDateString converts RFC3339 string to time.Time
func ConvertDateString(dateStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateStr)
}

// ConvertDateField converts a date field in a document
func ConvertDateField(doc map[string]interface{}, fieldName string) error {
	if val, ok := doc[fieldName]; ok {
		if str, ok := val.(string); ok {
			t, err := time.Parse(time.RFC3339, str)
			if err != nil {
				return fmt.Errorf("error parsing date field %s: %w", fieldName, err)
			}
			doc[fieldName] = t
		}
	}
	return nil
}

// ConvertAllDateFields converts all known date fields in a document
func ConvertAllDateFields(doc map[string]interface{}, dateFields []string) error {
	for _, field := range dateFields {
		if err := ConvertDateField(doc, field); err != nil {
			return err
		}
	}
	return nil
}
```

## Collection-Specific Transformations

### Users Collection

```go
func TransformUser(item map[string]types.AttributeValue) (map[string]interface{}, error) {
	doc, err := ConvertDynamoItem(item)
	if err != nil {
		return nil, err
	}
	
	// Convert date fields
	dateFields := []string{"createdAt", "updatedAt"}
	if err := ConvertAllDateFields(doc, dateFields); err != nil {
		return nil, err
	}
	
	// Ensure id field exists and is string
	if id, ok := doc["id"].(string); ok && id != "" {
		doc["id"] = id
	} else {
		return nil, fmt.Errorf("user missing id field")
	}
	
	return doc, nil
}
```

### Tasks Collection

```go
func TransformTask(item map[string]types.AttributeValue) (map[string]interface{}, error) {
	doc, err := ConvertDynamoItem(item)
	if err != nil {
		return nil, err
	}
	
	// Convert date fields
	dateFields := []string{"created", "updated", "deadline"}
	if err := ConvertAllDateFields(doc, dateFields); err != nil {
		return nil, err
	}
	
	// Handle nested timeSpent array
	if timeSpent, ok := doc["timeSpent"].([]interface{}); ok {
		for _, ts := range timeSpent {
			if tsMap, ok := ts.(map[string]interface{}); ok {
				// timeSpent.date is already a string, keep it
				// timeSpent.timeSpent should be a number
				if timeSpentVal, ok := tsMap["timeSpent"].(string); ok {
					if intVal, err := strconv.Atoi(timeSpentVal); err == nil {
						tsMap["timeSpent"] = intVal
					}
				}
			}
		}
	}
	
	// Handle nested fileAttachments array
	if fileAttachments, ok := doc["fileAttachments"].([]interface{}); ok {
		for _, fa := range fileAttachments {
			if faMap, ok := fa.(map[string]interface{}); ok {
				// Convert uploadDate
				if err := ConvertDateField(faMap, "uploadDate"); err != nil {
					// Log but don't fail
					fmt.Printf("Warning: error converting uploadDate: %v\n", err)
				}
			}
		}
	}
	
	// Handle nested activityLogs array
	if activityLogs, ok := doc["activityLogs"].([]interface{}); ok {
		for _, al := range activityLogs {
			if alMap, ok := al.(map[string]interface{}); ok {
				// Convert timestamp
				if err := ConvertDateField(alMap, "timestamp"); err != nil {
					fmt.Printf("Warning: error converting timestamp: %v\n", err)
				}
			}
		}
	}
	
	return doc, nil
}
```

### Projects Collection

```go
func TransformProject(item map[string]types.AttributeValue) (map[string]interface{}, error) {
	doc, err := ConvertDynamoItem(item)
	if err != nil {
		return nil, err
	}
	
	// Convert date fields
	dateFields := []string{"created", "updated", "deadLine"}
	if err := ConvertAllDateFields(doc, dateFields); err != nil {
		return nil, err
	}
	
	// Handle membersIds array (should already be array from conversion)
	if membersIds, ok := doc["membersIds"].([]interface{}); ok {
		// Ensure all are strings
		for i, memberId := range membersIds {
			if str, ok := memberId.(string); ok {
				membersIds[i] = str
			}
		}
		doc["membersIds"] = membersIds
	}
	
	return doc, nil
}
```

## Complete Transformation Script Example

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TransformAndImport transforms DynamoDB data and imports to MongoDB
func TransformAndImport(
	ctx context.Context,
	dynamoItems []map[string]types.AttributeValue,
	collection *mongo.Collection,
	transformFunc func(map[string]types.AttributeValue) (map[string]interface{}, error),
) error {
	
	documents := make([]interface{}, 0, len(dynamoItems))
	
	for i, item := range dynamoItems {
		doc, err := transformFunc(item)
		if err != nil {
			log.Printf("Error transforming item %d: %v", i, err)
			continue // Skip invalid items
		}
		
		// Add _id as ObjectId (MongoDB will generate if not present)
		// But we want to keep the original id field
		documents = append(documents, doc)
	}
	
	if len(documents) == 0 {
		return fmt.Errorf("no valid documents to import")
	}
	
	// Batch insert (MongoDB supports up to 100,000 documents per batch)
	const batchSize = 1000
	for i := 0; i < len(documents); i += batchSize {
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}
		
		batch := documents[i:end]
		_, err := collection.InsertMany(ctx, batch)
		if err != nil {
			// Check for duplicate key errors (id field)
			if mongo.IsDuplicateKeyError(err) {
				log.Printf("Warning: duplicate keys in batch %d-%d, skipping", i, end)
				continue
			}
			return fmt.Errorf("error inserting batch %d-%d: %w", i, end, err)
		}
		
		log.Printf("Imported batch %d-%d (%d documents)", i, end, len(batch))
	}
	
	return nil
}
```

## Handling Special Cases

### Empty Arrays
```go
// Ensure empty arrays are preserved as arrays, not null
if val, ok := doc["membersIds"]; ok && val == nil {
	doc["membersIds"] = []interface{}{}
}
```

### Optional Fields
```go
// Remove null optional fields (optional, MongoDB handles nulls fine)
for key, value := range doc {
	if value == nil {
		// Optionally remove null fields
		// delete(doc, key)
	}
}
```

### Nested Object Conversion
```go
// Convert nested DynamoDB maps recursively
func ConvertNestedMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = ConvertNestedMap(val)
		case []interface{}:
			result[k] = ConvertNestedArray(val)
		default:
			result[k] = v
		}
	}
	return result
}

func ConvertNestedArray(arr []interface{}) []interface{} {
	result := make([]interface{}, len(arr))
	for i, item := range arr {
		switch val := item.(type) {
		case map[string]interface{}:
			result[i] = ConvertNestedMap(val)
		case []interface{}:
			result[i] = ConvertNestedArray(val)
		default:
			result[i] = item
		}
	}
	return result
}
```

## Validation

### Validate Document Structure
```go
func ValidateUserDocument(doc map[string]interface{}) error {
	requiredFields := []string{"id", "name", "email", "role", "createdAt"}
	for _, field := range requiredFields {
		if _, ok := doc[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	
	// Validate email format
	if email, ok := doc["email"].(string); ok {
		if !isValidEmail(email) {
			return fmt.Errorf("invalid email format: %s", email)
		}
	}
	
	return nil
}
```

### Compare Record Counts
```go
func CompareRecordCounts(
	dynamoCount int,
	mongoCollection *mongo.Collection,
) error {
	ctx := context.Background()
	mongoCount, err := mongoCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	
	if int64(dynamoCount) != mongoCount {
		return fmt.Errorf("record count mismatch: DynamoDB=%d, MongoDB=%d", dynamoCount, mongoCount)
	}
	
	log.Printf("Record counts match: %d", dynamoCount)
	return nil
}
```

## Error Handling

### Logging Transformation Errors
```go
type TransformationError struct {
	ItemIndex int
	Field     string
	Error     error
}

func TransformWithErrorTracking(
	items []map[string]types.AttributeValue,
	transformFunc func(map[string]types.AttributeValue) (map[string]interface{}, error),
) ([]interface{}, []TransformationError) {
	
	var documents []interface{}
	var errors []TransformationError
	
	for i, item := range items {
		doc, err := transformFunc(item)
		if err != nil {
			errors = append(errors, TransformationError{
				ItemIndex: i,
				Error:     err,
			})
			continue
		}
		documents = append(documents, doc)
	}
	
	return documents, errors
}
```

## Performance Optimization

### Parallel Transformation
```go
func TransformParallel(
	items []map[string]types.AttributeValue,
	transformFunc func(map[string]types.AttributeValue) (map[string]interface{}, error),
	workers int,
) ([]interface{}, error) {
	
	type result struct {
		doc interface{}
		err error
	}
	
	itemChan := make(chan map[string]types.AttributeValue, len(items))
	resultChan := make(chan result, len(items))
	
	// Start workers
	for w := 0; w < workers; w++ {
		go func() {
			for item := range itemChan {
				doc, err := transformFunc(item)
				resultChan <- result{doc: doc, err: err}
			}
		}()
	}
	
	// Send items
	for _, item := range items {
		itemChan <- item
	}
	close(itemChan)
	
	// Collect results
	var documents []interface{}
	for i := 0; i < len(items); i++ {
		res := <-resultChan
		if res.err != nil {
			return nil, res.err
		}
		documents = append(documents, res.doc)
	}
	
	return documents, nil
}
```

---

## Summary

Key transformation steps:
1. Convert DynamoDB AttributeValue types to Go types
2. Convert date strings to `time.Time` objects
3. Handle nested structures (arrays and maps)
4. Preserve original `id` field
5. Validate document structure
6. Handle errors gracefully
7. Optimize for large datasets with batching and parallel processing

---

**Last Updated**: 2025-01-XX

