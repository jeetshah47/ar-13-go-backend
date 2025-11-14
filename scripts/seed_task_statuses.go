// seed_task_statuses.go - Seeds task status master data to MongoDB
//
// This script seeds task statuses as master data in the database.
// Task statuses are stored in the "task_statuses" collection.
//
// Usage:
//
//	go run scripts/seed_task_statuses.go
//	go run scripts/seed_task_statuses.go -force  # Force update existing statuses
//
// Flags:
//
//	-force    Force update existing statuses (default: false, skips if exists)
//
// Environment Variables Required:
//
//	MONGODB_URI         MongoDB connection string (default: "mongodb://localhost:27017")
//	MONGODB_DATABASE    MongoDB database name (default: "ar13_backend")
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskStatusDocument struct {
	ID          string    `bson:"id" json:"id"`
	Value       string    `bson:"value" json:"value"`
	DisplayName string    `bson:"displayName" json:"displayName"`
	Description string    `bson:"description" json:"description"`
	Category    string    `bson:"category" json:"category"`
	IsActive    bool      `bson:"isActive" json:"isActive"`
	IsCompleted bool      `bson:"isCompleted" json:"isCompleted"`
	Order       int       `bson:"order" json:"order"`
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}

func main() {
	// Parse command-line flags
	force := flag.Bool("force", false, "Force update existing statuses")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Display connection info (without exposing credentials)
	log.Println("Connecting to MongoDB...")
	if cfg.MongoDBURI != "" {
		// Mask password in URI for display
		maskedURI := maskMongoDBURI(cfg.MongoDBURI)
		log.Printf("  URI: %s", maskedURI)
	} else {
		log.Println("  URI: (using default)")
	}
	log.Printf("  Database: %s", cfg.MongoDBDatabase)
	log.Println("")

	// Initialize MongoDB with retry logic
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("Retry attempt %d/%d...", i+1, maxRetries)
			time.Sleep(2 * time.Second)
		}
		
		_, err = mongodb.InitializeMongoDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
		if err == nil {
			break
		}
		
		if i < maxRetries-1 {
			log.Printf("Connection failed: %v", err)
		}
	}
	
	if err != nil {
		log.Fatalf("\n❌ Failed to initialize MongoDB after %d attempts", maxRetries)
		log.Fatalf("Error: %v", err)
		log.Fatalf("\nTroubleshooting tips:")
		log.Fatalf("  1. Check your MONGODB_URI environment variable")
		log.Fatalf("  2. Verify network connectivity and firewall settings")
		log.Fatalf("  3. Ensure MongoDB Atlas IP whitelist includes your IP")
		log.Fatalf("  4. Verify TLS/SSL settings in connection string")
		log.Fatalf("  5. Check MongoDB Atlas cluster status")
	}
	
	log.Println("✅ MongoDB connected successfully")
	defer mongodb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db := mongodb.GetDatabase()
	collection := db.Collection("task_statuses")

	// Check if statuses already exist
	if !*force {
		count, err := collection.CountDocuments(ctx, bson.M{})
		if err != nil {
			log.Fatalf("Failed to check existing statuses: %v", err)
		}
		if count > 0 {
			log.Printf("Task statuses already exist in database (%d documents). Use -force to update.", count)
			log.Println("Skipping seed operation.")
			return
		}
	}

	log.Println("Seeding task statuses to MongoDB...")
	log.Println("")

	// Define task statuses with order
	statuses := []TaskStatusDocument{
		{
			ID:          "pending",
			Value:       constants.GetTaskStatusString(constants.TaskStatusPending),
			DisplayName: "Pending",
			Description: "Task is pending/not started",
			Category:    "active",
			IsActive:     true,
			IsCompleted: false,
			Order:       1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "in_progress",
			Value:       constants.GetTaskStatusString(constants.TaskStatusInProgress),
			DisplayName: "In Progress",
			Description: "Task is currently being worked on",
			Category:    "active",
			IsActive:     true,
			IsCompleted: false,
			Order:       2,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "in_review",
			Value:       constants.GetTaskStatusString(constants.TaskStatusInReview),
			DisplayName: "In Review",
			Description: "Task is under review",
			Category:    "active",
			IsActive:     true,
			IsCompleted: false,
			Order:       3,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "completed",
			Value:       constants.GetTaskStatusString(constants.TaskStatusCompleted),
			DisplayName: "Completed",
			Description: "Task is completed",
			Category:    "completed",
			IsActive:     false,
			IsCompleted: true,
			Order:       4,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "accepted",
			Value:       constants.GetTaskStatusString(constants.TaskStatusAccepted),
			DisplayName: "Accepted",
			Description: "Task has been accepted",
			Category:    "final",
			IsActive:     false,
			IsCompleted: true,
			Order:       5,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "rejected",
			Value:       constants.GetTaskStatusString(constants.TaskStatusRejected),
			DisplayName: "Rejected",
			Description: "Task has been rejected",
			Category:    "final",
			IsActive:     false,
			IsCompleted: false,
			Order:       6,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Create index on 'id' field for uniqueness
	// NOTE: Only 'id' index is needed. 'value' and 'order' indexes are unnecessary:
	// - Collection has only 6 documents, so sorting can be done in memory
	// - Collection is not queried by 'value' field in application code
	// - This minimizes write overhead and storage usage
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("id_unique"),
	}
	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Failed to create index (may already exist): %v", err)
	}

	// Insert or update statuses
	insertedCount := 0
	updatedCount := 0

	for _, status := range statuses {
		filter := bson.M{"id": status.ID}
		update := bson.M{
			"$set": bson.M{
				"value":       status.Value,
				"displayName": status.DisplayName,
				"description": status.Description,
				"category":    status.Category,
				"isActive":    status.IsActive,
				"isCompleted": status.IsCompleted,
				"order":       status.Order,
				"updatedAt":   status.UpdatedAt,
			},
			"$setOnInsert": bson.M{
				"createdAt": status.CreatedAt,
			},
		}

		opts := options.Update().SetUpsert(true)
		result, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Printf("Error upserting status %s: %v", status.ID, err)
			continue
		}

		if result.UpsertedCount > 0 {
			insertedCount++
			log.Printf("  ✅ Inserted: %s (%s)", status.DisplayName, status.Value)
		} else if result.ModifiedCount > 0 {
			updatedCount++
			log.Printf("  🔄 Updated: %s (%s)", status.DisplayName, status.Value)
		} else {
			log.Printf("  ⏭️  No change: %s (%s)", status.DisplayName, status.Value)
		}
	}

	log.Println("")
	log.Printf("✅ Task status seeding completed!")
	log.Printf("   Inserted: %d", insertedCount)
	log.Printf("   Updated: %d", updatedCount)
	log.Printf("   Total: %d", len(statuses))
	log.Println("")
	log.Println("Task statuses are now available in the 'task_statuses' collection.")
}

// maskMongoDBURI masks the password in MongoDB URI for safe display
func maskMongoDBURI(uri string) string {
	// Simple masking: replace password between :// and @
	// Example: mongodb://user:password@host -> mongodb://user:****@host
	masked := uri
	if idx := findNthOccurrence(uri, "://", 1); idx > 0 {
		if atIdx := findNthOccurrence(uri, "@", 1); atIdx > idx {
			// Find the colon after the protocol
			colonIdx := findNthOccurrence(uri[idx:], ":", 1)
			if colonIdx > 0 {
				colonIdx += idx
				if colonIdx < atIdx {
					masked = uri[:colonIdx+1] + "****" + uri[atIdx:]
				}
			}
		}
	}
	return masked
}

// findNthOccurrence finds the nth occurrence of a substring
func findNthOccurrence(s, substr string, n int) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
			if count == n {
				return i
			}
		}
	}
	return -1
}

