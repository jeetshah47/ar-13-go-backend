package main

import (
	"context"
	"fmt"
	"log"
	"os"

	dynamodbpkg "github.com/ar-13-go-backend/pkg/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Table names that should exist in DynamoDB
var requiredTables = []string{
	"users",
	"notifications",
	"projects",
	"calendar_events",
	"tasks",
	"project_details",
	"activity_logs",
	"leaveRequests",
	"signupInvitations",
	"userAccountLinks",
	"info-portal",
}

func main() {
	// Get region from environment or use default
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	fmt.Printf("Testing DynamoDB connection in region: %s\n\n", region)

	// Initialize DynamoDB client
	client, err := dynamodbpkg.InitializeDynamoDB(region)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}

	fmt.Println("✅ DynamoDB client initialized successfully\n")

	// Test connection by listing tables
	ctx := context.Background()
	result, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		log.Fatalf("Failed to list tables (check AWS credentials): %v", err)
	}

	fmt.Printf("Found %d table(s) in DynamoDB:\n", len(result.TableNames))
	for _, tableName := range result.TableNames {
		fmt.Printf("  - %s\n", tableName)
	}
	fmt.Println()

	// Check if required tables exist
	fmt.Println("Checking required tables...")
	existingTables := make(map[string]bool)
	for _, tableName := range result.TableNames {
		existingTables[tableName] = true
	}

	missingTables := []string{}
	for _, tableName := range requiredTables {
		if existingTables[tableName] {
			fmt.Printf("  ✅ %s - exists\n", tableName)
		} else {
			fmt.Printf("  ❌ %s - MISSING\n", tableName)
			missingTables = append(missingTables, tableName)
		}
	}

	fmt.Println()

	if len(missingTables) > 0 {
		fmt.Printf("⚠️  Warning: %d table(s) are missing:\n", len(missingTables))
		for _, tableName := range missingTables {
			fmt.Printf("  - %s\n", tableName)
		}
		fmt.Println("\nPlease create these tables using the commands in DYNAMODB_TABLES.md")
		os.Exit(1)
	}

	// Test a simple operation on users table
	fmt.Println("Testing basic operation on 'users' table...")
	_, err = client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String("users"),
	})
	if err != nil {
		log.Fatalf("Failed to describe 'users' table: %v", err)
	}
	fmt.Println("✅ Successfully accessed 'users' table")

	fmt.Println("\n🎉 All checks passed! DynamoDB connection is working correctly.")
}

