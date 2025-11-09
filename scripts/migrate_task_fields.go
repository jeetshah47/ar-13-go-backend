package main

import (
	"context"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	dynamodbpkg "github.com/ar-13-go-backend/pkg/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize DynamoDB
	_, err = dynamodbpkg.InitializeDynamoDB(cfg.AWSRegion)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB: %v", err)
	}

	client := dynamodbpkg.GetClient()
	tableName := dynamodbpkg.GetTableName("tasks")

	ctx := context.Background()

	log.Println("Starting migration of task fields...")
	log.Println("This will:")
	log.Println("  1. Migrate 'duration' field to 'deadline' field")
	log.Println("  2. Ensure all tasks have the new field structure")
	log.Println()

	// Scan all tasks with pagination
	var lastEvaluatedKey map[string]types.AttributeValue
	totalScanned := 0
	totalUpdated := 0
	totalErrors := 0

	for {
		scanInput := &dynamodb.ScanInput{
			TableName: aws.String(tableName),
		}

		if lastEvaluatedKey != nil {
			scanInput.ExclusiveStartKey = lastEvaluatedKey
		}

		page, err := client.Scan(ctx, scanInput)
		if err != nil {
			log.Fatalf("Failed to scan tasks: %v", err)
		}

		totalScanned += len(page.Items)

		// Process each item
		for _, item := range page.Items {
			// Check if item has 'duration' field but not 'deadline'
			hasDuration := false
			hasDeadline := false
			var durationValue types.AttributeValue

			if durationVal, exists := item["duration"]; exists {
				hasDuration = true
				durationValue = durationVal
			}

			if _, exists := item["deadline"]; exists {
				hasDeadline = true
			}

			// Get task ID
			var taskID string
			if idVal, ok := item["id"].(*types.AttributeValueMemberS); ok {
				taskID = idVal.Value
			} else {
				log.Printf("Warning: Task without ID found, skipping")
				continue
			}

			// Migrate duration to deadline if needed
			if hasDuration && !hasDeadline {
				// Copy duration value to deadline
				updateExpr := "SET #deadline = :deadline"
				exprAttrNames := map[string]string{
					"#deadline": "deadline",
				}
				exprAttrValues := map[string]types.AttributeValue{
					":deadline": durationValue,
				}

				// Also remove the old duration field
				updateExpr += " REMOVE #duration"
				exprAttrNames["#duration"] = "duration"

				_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
					TableName:                 aws.String(tableName),
					Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: taskID}},
					UpdateExpression:          aws.String(updateExpr),
					ExpressionAttributeNames:  exprAttrNames,
					ExpressionAttributeValues: exprAttrValues,
				})

				if err != nil {
					log.Printf("Error updating task %s: %v", taskID, err)
					totalErrors++
				} else {
					log.Printf("✓ Migrated task %s: duration -> deadline", taskID)
					totalUpdated++
				}
			} else if hasDuration && hasDeadline {
				// Both exist - remove old duration field
				updateExpr := "REMOVE #duration"
				exprAttrNames := map[string]string{
					"#duration": "duration",
				}

				_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
					TableName:                aws.String(tableName),
					Key:                      map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: taskID}},
					UpdateExpression:         aws.String(updateExpr),
					ExpressionAttributeNames: exprAttrNames,
				})

				if err != nil {
					log.Printf("Error removing duration field from task %s: %v", taskID, err)
					totalErrors++
				} else {
					log.Printf("✓ Cleaned up task %s: removed old duration field", taskID)
					totalUpdated++
				}
			}
		}

		// Check if there are more pages
		if page.LastEvaluatedKey != nil && len(page.LastEvaluatedKey) > 0 {
			lastEvaluatedKey = page.LastEvaluatedKey
		} else {
			// No more pages
			break
		}

		// Small delay to avoid throttling
		time.Sleep(100 * time.Millisecond)
	}

	log.Println()
	log.Println("Migration Summary:")
	log.Printf("  Total tasks scanned: %d", totalScanned)
	log.Printf("  Tasks updated: %d", totalUpdated)
	log.Printf("  Errors: %d", totalErrors)
	log.Println()
	log.Println("Migration completed!")
}

