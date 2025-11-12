package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	dynamodbpkg "github.com/ar-13-go-backend/pkg/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// BackupService handles DynamoDB backup operations
type BackupService struct {
	client *dynamodb.Client
}

// NewBackupService creates a new backup service
func NewBackupService() *BackupService {
	return &BackupService{
		client: dynamodbpkg.GetClient(),
	}
}

// BackupData represents the structure of a backup
type BackupData struct {
	Collection string           `json:"collection"`
	Documents  []BackupDocument `json:"documents"`
	BackedUpAt time.Time        `json:"backedUpAt"`
}

// BackupDocument represents a document with its data and subcollections
type BackupDocument struct {
	ID             string                 `json:"id"`
	Data           map[string]interface{} `json:"data"`
	SubCollections map[string]BackupData  `json:"subCollections,omitempty"`
}

// TableMapping maps collection names to DynamoDB table names
var tableMapping = map[string]string{
	"users":             "users",
	"projects":          "projects",
	"tasks":             "tasks",
	"calendar_events":   "calendar_events",
	"notifications":     "notifications",
	"vacations":         "leaveRequests",
	"activity_logs":     "activity_logs",
	"info_portal":       "info-portal",
	"project_details":   "project_details",
	"user_account_links": "userAccountLinks",
	"signup_invitations": "signupInvitations",
}

// BackupAllCollections backs up all DynamoDB tables to JSON files
func (s *BackupService) BackupAllCollections(ctx context.Context, backupDir string) (map[string]string, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Get all table names
	tables, err := s.getAllCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	backupFiles := make(map[string]string)
	timestamp := time.Now().Format("20060102_150405")

	// Backup each table
	for _, collectionName := range tables {
		tableName := tableMapping[collectionName]
		if tableName == "" {
			tableName = collectionName
		}

		backupFile, err := s.backupTable(ctx, collectionName, tableName, backupDir, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to backup table %s: %w", tableName, err)
		}

		if backupFile != "" {
			backupFiles[collectionName] = backupFile
		}
	}

	return backupFiles, nil
}

// backupTable backs up a single DynamoDB table
func (s *BackupService) backupTable(ctx context.Context, collectionName, tableName, backupDir, timestamp string) (string, error) {
	// Get full table name
	fullTableName := dynamodbpkg.GetTableName(tableName)

	// Scan all items from the table (with pagination)
	var allItems []map[string]types.AttributeValue
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName: aws.String(fullTableName),
		}

		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := s.client.Scan(ctx, input)
		if err != nil {
			return "", fmt.Errorf("failed to scan table %s: %w", fullTableName, err)
		}

		allItems = append(allItems, result.Items...)

		// Check if there are more items
		if result.LastEvaluatedKey == nil || len(result.LastEvaluatedKey) == 0 {
			break
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}

	// Convert DynamoDB items to backup documents
	documents := make([]BackupDocument, 0, len(allItems))
	for _, item := range allItems {
		// Convert DynamoDB attribute values to regular Go types
		data, err := s.convertDynamoItemToMap(item)
		if err != nil {
			return "", fmt.Errorf("failed to convert item: %w", err)
		}

		// Extract ID from the item
		id := ""
		if idVal, ok := item["id"]; ok {
			if idStr, ok := idVal.(*types.AttributeValueMemberS); ok {
				id = idStr.Value
			}
		}

		if id == "" {
			// Skip items without ID
			continue
		}

		documents = append(documents, BackupDocument{
			ID:   id,
			Data: data,
		})
	}

	// Create backup data structure
	backupData := BackupData{
		Collection: collectionName,
		Documents:  documents,
		BackedUpAt: time.Now(),
	}

	// Generate filename
	filename := fmt.Sprintf("%s_%s.json", collectionName, timestamp)
	filePath := filepath.Join(backupDir, filename)

	// Write to JSON file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(backupData); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return filePath, nil
}

// convertDynamoItemToMap converts a DynamoDB item to a regular map[string]interface{}
func (s *BackupService) convertDynamoItemToMap(item map[string]types.AttributeValue) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range item {
		converted, err := s.convertAttributeValue(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert attribute %s: %w", key, err)
		}
		result[key] = converted
	}

	return result, nil
}

// convertAttributeValue converts a DynamoDB AttributeValue to a Go value
func (s *BackupService) convertAttributeValue(av types.AttributeValue) (interface{}, error) {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		return v.Value, nil
	case *types.AttributeValueMemberN:
		// Try to parse as int64 first, then float64
		var intVal int64
		if err := attributevalue.Unmarshal(av, &intVal); err == nil {
			return intVal, nil
		}
		var floatVal float64
		if err := attributevalue.Unmarshal(av, &floatVal); err == nil {
			return floatVal, nil
		}
		// If unmarshaling fails, return the string value
		return v.Value, nil
	case *types.AttributeValueMemberBOOL:
		return v.Value, nil
	case *types.AttributeValueMemberNULL:
		return nil, nil
	case *types.AttributeValueMemberB:
		return v.Value, nil
	case *types.AttributeValueMemberSS:
		return v.Value, nil
	case *types.AttributeValueMemberNS:
		return v.Value, nil
	case *types.AttributeValueMemberBS:
		return v.Value, nil
	case *types.AttributeValueMemberL:
		list := make([]interface{}, 0, len(v.Value))
		for _, item := range v.Value {
			converted, err := s.convertAttributeValue(item)
			if err != nil {
				return nil, err
			}
			list = append(list, converted)
		}
		return list, nil
	case *types.AttributeValueMemberM:
		result := make(map[string]interface{})
		for key, value := range v.Value {
			converted, err := s.convertAttributeValue(value)
			if err != nil {
				return nil, err
			}
			result[key] = converted
		}
		return result, nil
	default:
		// Fallback: try to unmarshal using attributevalue package
		var result interface{}
		if err := attributevalue.Unmarshal(av, &result); err == nil {
			return result, nil
		}
		return nil, fmt.Errorf("unsupported attribute value type: %T", av)
	}
}

// getAllCollections gets all DynamoDB tables
func (s *BackupService) getAllCollections(ctx context.Context) ([]string, error) {
	// Known DynamoDB tables (collection names)
	knownTables := []string{
		"users",
		"projects",
		"tasks",
		"calendar_events",
		"notifications",
		"vacations",
		"activity_logs",
		"info_portal",
		"project_details",
		"user_account_links",
		"signup_invitations",
	}

	return knownTables, nil
}
