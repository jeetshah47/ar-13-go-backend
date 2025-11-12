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
	tableName := dynamodbpkg.GetTableName("calendar_events")

	ctx := context.Background()

	log.Println("Starting migration of calendar_events invites field...")
	log.Println("This will:")
	log.Println("  1. Add 'invites' field (empty array) to calendar events that don't have it")
	log.Println()

	// Scan all calendar events with pagination
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
			log.Fatalf("Failed to scan calendar events: %v", err)
		}

		totalScanned += len(page.Items)

		// Process each item
		for _, item := range page.Items {
			// Check if item already has 'invites' field
			hasInvites := false
			if _, exists := item["invites"]; exists {
				hasInvites = true
			}

			// Get event ID
			var eventID string
			if idVal, ok := item["id"].(*types.AttributeValueMemberS); ok {
				eventID = idVal.Value
			} else {
				log.Printf("Warning: Calendar event without ID found, skipping")
				continue
			}

			// Add invites field if it doesn't exist
			if !hasInvites {
				// Set invites to empty list
				updateExpr := "SET #invites = :invites"
				exprAttrNames := map[string]string{
					"#invites": "invites",
				}
				exprAttrValues := map[string]types.AttributeValue{
					":invites": &types.AttributeValueMemberL{Value: []types.AttributeValue{}}, // Empty list
				}

				_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
					TableName:                 aws.String(tableName),
					Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: eventID}},
					UpdateExpression:          aws.String(updateExpr),
					ExpressionAttributeNames:  exprAttrNames,
					ExpressionAttributeValues: exprAttrValues,
				})

				if err != nil {
					log.Printf("Error updating calendar event %s: %v", eventID, err)
					totalErrors++
				} else {
					log.Printf("✓ Added invites field to calendar event %s", eventID)
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
	log.Printf("  Total calendar events scanned: %d", totalScanned)
	log.Printf("  Calendar events updated: %d", totalUpdated)
	log.Printf("  Errors: %d", totalErrors)
	log.Println()
	log.Println("Migration completed!")
}
