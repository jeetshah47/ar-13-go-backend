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

