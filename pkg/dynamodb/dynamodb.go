package dynamodb

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	dynamoClient *dynamodb.Client
)

// InitializeDynamoDB initializes DynamoDB client
func InitializeDynamoDB(region string) (*dynamodb.Client, error) {
	ctx := context.TODO()

	// Priority 1: Use environment variables (from .env file) if available
	accessKey := strings.Trim(strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID")), `"`)
	secretKey := strings.Trim(strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY")), `"`)

	var cfg aws.Config
	var err error

	if accessKey != "" && secretKey != "" {
		// Use explicit credentials from environment variables
		credProvider := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credProvider),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load config with credentials: %w", err)
		}
	} else {
		// Priority 2: Use default credential chain (shared credentials file, IAM role, etc.)
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load default config: %w", err)
		}
	}

	client := dynamodb.NewFromConfig(cfg)
	dynamoClient = client
	return client, nil
}

// GetClient returns the DynamoDB client instance
func GetClient() *dynamodb.Client {
	return dynamoClient
}

// GetTableName returns the full table name (with optional prefix)
func GetTableName(baseName string) string {
	// You can add environment prefix here if needed
	// e.g., return fmt.Sprintf("%s-%s", os.Getenv("ENV"), baseName)
	return baseName
}
