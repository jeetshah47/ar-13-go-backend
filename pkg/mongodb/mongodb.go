package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	mongoDB     *mongo.Database
)

// InitializeMongoDB initializes MongoDB client with optimized connection pool settings
func InitializeMongoDB(uri, databaseName string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optimized connection pool settings:
	// - MaxPoolSize: 20 (reduced from 100 to prevent excessive connections)
	// - MinPoolSize: 2 (reduced from 10 to avoid keeping too many idle connections)
	// - MaxConnecting: 5 (limits concurrent connection attempts)
	// - MaxConnIdleTime: 5 minutes (increased from 30s to reduce reconnection overhead)
	// - ConnectTimeout: 10 seconds (timeout for establishing connections)
	// - SocketTimeout: 30 seconds (timeout for socket operations)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(20).                        // Maximum connections in pool (reduced from 100)
		SetMinPoolSize(2).                         // Minimum connections to maintain (reduced from 10)
		SetMaxConnecting(5).                       // Max concurrent connection attempts
		SetMaxConnIdleTime(5 * time.Minute).       // Close idle connections after 5 minutes (increased from 30s)
		SetConnectTimeout(10 * time.Second).       // Timeout for establishing connections
		SetSocketTimeout(30 * time.Second).        // Timeout for socket operations
		SetServerSelectionTimeout(5 * time.Second) // Timeout for server selection

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	mongoClient = client
	mongoDB = client.Database(databaseName)

	return client, nil
}

// GetClient returns the MongoDB client instance
func GetClient() *mongo.Client {
	return mongoClient
}

// GetDatabase returns the MongoDB database instance
func GetDatabase() *mongo.Database {
	return mongoDB
}

// GetDatabaseName returns the MongoDB database name
// Panics if database is not initialized
func GetDatabaseName() string {
	if mongoDB == nil {
		panic("MongoDB database is not initialized. Please ensure MongoDB is connected before creating repositories.")
	}
	return mongoDB.Name()
}

// Close closes the MongoDB connection
func Close() error {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return mongoClient.Disconnect(ctx)
	}
	return nil
}

// HealthCheck checks if MongoDB connection is healthy
func HealthCheck(ctx context.Context) error {
	if mongoClient == nil {
		return fmt.Errorf("MongoDB client not initialized")
	}
	return mongoClient.Ping(ctx, nil)
}

// GetConnectionPoolStats returns connection pool statistics for monitoring
func GetConnectionPoolStats() (map[string]interface{}, error) {
	if mongoClient == nil {
		return nil, fmt.Errorf("MongoDB client not initialized")
	}

	// Get server description to access connection pool stats
	serverStatus := mongoClient.NumberSessionsInProgress()

	stats := map[string]interface{}{
		"sessionsInProgress": serverStatus,
		"clientInitialized":  mongoClient != nil,
	}

	// Try to get more detailed stats from the client
	// Note: The MongoDB driver doesn't expose all pool stats directly,
	// but we can monitor sessions in progress as an indicator
	return stats, nil
}
