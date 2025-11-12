package repos

import (
	"context"
	"fmt"
	"time"

	dynamodbpkg "github.com/ar-13-go-backend/pkg/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoBaseRepo provides common DynamoDB repository functionality
type DynamoBaseRepo struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoBaseRepo creates a new DynamoDB base repository
func NewDynamoBaseRepo(tableName string) *DynamoBaseRepo {
	return &DynamoBaseRepo{
		client:    dynamodbpkg.GetClient(),
		tableName: dynamodbpkg.GetTableName(tableName),
	}
}

// GetTableName returns the table name
func (r *DynamoBaseRepo) GetTableName() string {
	return r.tableName
}

// GetByID gets an item by ID
func (r *DynamoBaseRepo) GetByID(ctx context.Context, id string) (map[string]types.AttributeValue, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	return result.Item, nil
}

// Exists checks if an item exists
func (r *DynamoBaseRepo) Exists(ctx context.Context, id string) (bool, error) {
	item, err := r.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return item != nil, nil
}

// DeleteByID deletes an item by ID
func (r *DynamoBaseRepo) DeleteByID(ctx context.Context, id string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

// PutItem puts an item into DynamoDB
func (r *DynamoBaseRepo) PutItem(ctx context.Context, item interface{}) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	return err
}

// UpdateItem updates an item in DynamoDB
func (r *DynamoBaseRepo) UpdateItem(ctx context.Context, id string, updates map[string]interface{}) error {
	// Add updated timestamp
	updates["updatedAt"] = time.Now()

	// Build update expression
	updateExpr := "SET "
	exprAttrNames := make(map[string]string)
	exprAttrValues := make(map[string]types.AttributeValue)

	first := true
	for key, value := range updates {
		if !first {
			updateExpr += ", "
		}
		first = false

		placeholder := fmt.Sprintf("#%s", key)
		valuePlaceholder := fmt.Sprintf(":%s", key)

		updateExpr += placeholder + " = " + valuePlaceholder
		exprAttrNames[placeholder] = key

		av, err := attributevalue.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for %s: %w", key, err)
		}
		exprAttrValues[valuePlaceholder] = av
	}

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.tableName),
		Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
	})

	return err
}

// ScanItems scans all items from the table
func (r *DynamoBaseRepo) ScanItems(ctx context.Context, limit *int32) ([]map[string]types.AttributeValue, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	if limit != nil {
		input.Limit = limit
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// ScanItemsWithFilter scans items from the table with a filter expression
// This is more efficient than scanning all items and filtering in memory
func (r *DynamoBaseRepo) ScanItemsWithFilter(ctx context.Context, filterExpression string, expressionAttributeNames map[string]string, expressionAttributeValues map[string]types.AttributeValue, limit *int32) ([]map[string]types.AttributeValue, error) {
	input := &dynamodb.ScanInput{
		TableName:                 aws.String(r.tableName),
		FilterExpression:          aws.String(filterExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	}

	if limit != nil {
		input.Limit = limit
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// QueryByIndex queries items by a GSI (Global Secondary Index)
func (r *DynamoBaseRepo) QueryByIndex(ctx context.Context, indexName, keyName, keyValue string) ([]map[string]types.AttributeValue, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String(fmt.Sprintf("#%s = :%s", keyName, keyName)),
		ExpressionAttributeNames: map[string]string{
			fmt.Sprintf("#%s", keyName): keyName,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", keyName): &types.AttributeValueMemberS{Value: keyValue},
		},
	})
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// UnmarshalItem unmarshals a DynamoDB item into a struct
func UnmarshalItem(item map[string]types.AttributeValue, target interface{}) error {
	return attributevalue.UnmarshalMap(item, target)
}

// MarshalItem marshals a struct into DynamoDB item
func MarshalItem(item interface{}) (map[string]types.AttributeValue, error) {
	return attributevalue.MarshalMap(item)
}

// BatchGetItems retrieves multiple items by IDs in a single request
// DynamoDB BatchGetItem can retrieve up to 100 items at once
// This method handles batching automatically for larger requests
func (r *DynamoBaseRepo) BatchGetItems(ctx context.Context, ids []string) (map[string]map[string]types.AttributeValue, error) {
	if len(ids) == 0 {
		return make(map[string]map[string]types.AttributeValue), nil
	}

	result := make(map[string]map[string]types.AttributeValue)
	const maxBatchSize = 100 // DynamoDB limit

	// Process in batches of 100
	for i := 0; i < len(ids); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]

		// Build keys for this batch
		keys := make([]map[string]types.AttributeValue, len(batch))
		for j, id := range batch {
			keys[j] = map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: id},
			}
		}

		// Execute batch get
		batchResult, err := r.client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				r.tableName: {
					Keys: keys,
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("batch get failed: %w", err)
		}

		// Process results
		if tableResults, ok := batchResult.Responses[r.tableName]; ok {
			for _, item := range tableResults {
				if idVal, ok := item["id"].(*types.AttributeValueMemberS); ok {
					result[idVal.Value] = item
				}
			}
		}

		// Handle unprocessed keys (retry logic could be added here)
		if len(batchResult.UnprocessedKeys) > 0 {
			// For simplicity, we'll return an error if there are unprocessed keys
			// In production, you might want to retry these
			return nil, fmt.Errorf("some items were not processed in batch get")
		}
	}

	return result, nil
}

// BatchWriteItems writes multiple items in a single request
// DynamoDB BatchWriteItem can write up to 25 items at once
// This method handles batching automatically for larger requests
func (r *DynamoBaseRepo) BatchWriteItems(ctx context.Context, items []interface{}) error {
	if len(items) == 0 {
		return nil
	}

	const maxBatchSize = 25 // DynamoDB limit

	// Process in batches of 25
	for i := 0; i < len(items); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]

		// Build write requests for this batch
		writeRequests := make([]types.WriteRequest, len(batch))
		for j, item := range batch {
			av, err := attributevalue.MarshalMap(item)
			if err != nil {
				return fmt.Errorf("failed to marshal item %d: %w", j, err)
			}
			writeRequests[j] = types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: av,
				},
			}
		}

		// Execute batch write
		batchResult, err := r.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				r.tableName: writeRequests,
			},
		})
		if err != nil {
			return fmt.Errorf("batch write failed: %w", err)
		}

		// Handle unprocessed items (retry logic could be added here)
		if len(batchResult.UnprocessedItems) > 0 {
			// For simplicity, we'll return an error if there are unprocessed items
			// In production, you might want to retry these
			return fmt.Errorf("some items were not processed in batch write")
		}
	}

	return nil
}

// ScanItemsPaginated scans items with pagination support
// Returns items, lastEvaluatedKey (for pagination), and error
func (r *DynamoBaseRepo) ScanItemsPaginated(ctx context.Context, limit *int32, exclusiveStartKey map[string]types.AttributeValue) ([]map[string]types.AttributeValue, map[string]types.AttributeValue, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	if limit != nil {
		input.Limit = limit
	}

	if exclusiveStartKey != nil {
		input.ExclusiveStartKey = exclusiveStartKey
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	return result.Items, result.LastEvaluatedKey, nil
}
