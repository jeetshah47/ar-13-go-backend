package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/stretchr/testify/require"
)

// SetupIntegrationTest sets up the environment for integration tests
// This includes database connections, test data, etc.
func SetupIntegrationTest(t *testing.T) (*config.Config, func()) {
	cfg := SetupTestConfig()

	// Override with test database if TEST_MONGODB_URI is set
	if testMongoURI := os.Getenv("TEST_MONGODB_URI"); testMongoURI != "" {
		cfg.MongoDBURI = testMongoURI
		cfg.MongoDBDatabase = "test_db_" + time.Now().Format("20060102150405")
	}

	// Initialize MongoDB for integration tests
	_, err := mongodb.InitializeMongoDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
	require.NoError(t, err, "Failed to initialize MongoDB for integration tests")

	// Cleanup function
	cleanup := func() {
		// Drop test database
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client := mongodb.GetClient()
		if client != nil {
			err := client.Database(cfg.MongoDBDatabase).Drop(ctx)
			if err != nil {
				t.Logf("Warning: Failed to drop test database: %v", err)
			}
		}

		// Close MongoDB connection
		if err := mongodb.Close(); err != nil {
			t.Logf("Warning: Failed to close MongoDB connection: %v", err)
		}
	}

	return cfg, cleanup
}

// SkipIfNoDatabase skips the test if TEST_MONGODB_URI is not set
func SkipIfNoDatabase(t *testing.T) {
	if os.Getenv("TEST_MONGODB_URI") == "" {
		t.Skip("Skipping integration test: TEST_MONGODB_URI not set")
	}
}

// SetupTestData creates test data for integration tests
// This is a helper that can be extended based on test needs
func SetupTestData(t *testing.T, ctx context.Context) {
	// This can be extended to create test users, projects, tasks, etc.
	// For now, it's a placeholder
}

