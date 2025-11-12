package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ar-13-go-backend/internal/config"
	dynamodbpkg "github.com/ar-13-go-backend/pkg/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TableMapping maps collection names to DynamoDB table names
var tableMapping = map[string]string{
	"users":              "users",
	"projects":           "projects",
	"tasks":              "tasks",
	"calendar_events":    "calendar_events",
	"notifications":      "notifications",
	"vacations":          "leaveRequests",
	"activity_logs":      "activity_logs",
	"info_portal":        "info-portal",
	"project_details":    "project_details",
	"user_account_links": "userAccountLinks",
	"signup_invitations": "signupInvitations",
}

func main() {
	// Check for confirmation flag
	confirmFlag := false
	if len(os.Args) > 1 && os.Args[1] == "--confirm" {
		confirmFlag = true
	}

	if !confirmFlag {
		fmt.Println("⚠️  WARNING: This script will DELETE ALL DATA from all DynamoDB tables!")
		fmt.Println()
		fmt.Println("Tables that will be cleared:")
		for collectionName, tableName := range tableMapping {
			fmt.Printf("  - %s (table: %s)\n", collectionName, tableName)
		}
		fmt.Println()
		fmt.Println("This operation CANNOT be undone!")
		fmt.Println()
		fmt.Print("Type 'DELETE ALL DATA' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(confirmation)

		if confirmation != "DELETE ALL DATA" {
			fmt.Println("❌ Confirmation failed. Operation cancelled.")
			os.Exit(1)
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize DynamoDB
	client, err := dynamodbpkg.InitializeDynamoDB(cfg.AWSRegion)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB: %v", err)
	}
	fmt.Println("✅ DynamoDB initialized")
	fmt.Println()

	ctx := context.Background()

	// Delete data from each table
	totalDeleted := 0
	for collectionName, tableName := range tableMapping {
		fullTableName := dynamodbpkg.GetTableName(tableName)
		fmt.Printf("🗑️  Deleting data from %s (table: %s)...\n", collectionName, fullTableName)

		deleted, err := deleteAllFromTable(ctx, client, fullTableName)
		if err != nil {
			log.Printf("❌ Error deleting from %s: %v\n", collectionName, err)
			continue
		}

		fmt.Printf("   ✅ Deleted %d item(s) from %s\n", deleted, collectionName)
		totalDeleted += deleted
	}

	fmt.Println()
	fmt.Printf("🎉 Deletion complete! Total items deleted: %d\n", totalDeleted)
}

// deleteAllFromTable deletes all items from a DynamoDB table
func deleteAllFromTable(ctx context.Context, client *dynamodb.Client, tableName string) (int, error) {
	// Get table schema to determine key attributes
	tableDesc, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to describe table: %w", err)
	}

	// Extract key attribute names from the schema
	keyAttributes := make(map[string]bool)
	for _, keySchema := range tableDesc.Table.KeySchema {
		keyAttributes[*keySchema.AttributeName] = true
	}

	totalDeleted := 0
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		// Scan items from the table
		scanInput := &dynamodb.ScanInput{
			TableName: aws.String(tableName),
		}

		if lastEvaluatedKey != nil {
			scanInput.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := client.Scan(ctx, scanInput)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to scan table: %w", err)
		}

		if len(result.Items) == 0 {
			break
		}

		// Delete items in batches (DynamoDB batch limit is 25)
		const batchSize = 25
		for i := 0; i < len(result.Items); i += batchSize {
			end := i + batchSize
			if end > len(result.Items) {
				end = len(result.Items)
			}

			batch := result.Items[i:end]
			writeRequests := make([]types.WriteRequest, len(batch))

			for j, item := range batch {
				// Extract the key attributes from the item
				key := make(map[string]types.AttributeValue)
				for attrName := range keyAttributes {
					if attrVal, ok := item[attrName]; ok {
						key[attrName] = attrVal
					} else {
						return totalDeleted, fmt.Errorf("item at index %d missing required key attribute '%s'", i+j, attrName)
					}
				}

				writeRequests[j] = types.WriteRequest{
					DeleteRequest: &types.DeleteRequest{
						Key: key,
					},
				}
			}

			// Execute batch delete
			batchResult, err := client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{
					tableName: writeRequests,
				},
			})
			if err != nil {
				return totalDeleted, fmt.Errorf("batch delete failed: %w", err)
			}

			// Handle unprocessed items (retry if needed)
			if unprocessed, ok := batchResult.UnprocessedItems[tableName]; ok && len(unprocessed) > 0 {
				// Retry unprocessed items
				retryResult, err := client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
					RequestItems: map[string][]types.WriteRequest{
						tableName: unprocessed,
					},
				})
				if err != nil {
					return totalDeleted, fmt.Errorf("retry batch delete failed: %w", err)
				}

				// If still unprocessed, log warning but continue
				if retryUnprocessed, ok := retryResult.UnprocessedItems[tableName]; ok && len(retryUnprocessed) > 0 {
					log.Printf("   ⚠️  Warning: %d items could not be deleted from %s (may need manual cleanup)", len(retryUnprocessed), tableName)
				}
			}

			totalDeleted += len(batch)
		}

		// Check if there are more items
		if result.LastEvaluatedKey == nil || len(result.LastEvaluatedKey) == 0 {
			break
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}

	return totalDeleted, nil
}

