// fix_drawing_list_index.go - Fixes the drawing list indexes and cleans up null IDs
//
// This script:
//   - Drops existing non-sparse unique indexes on drawing_categories and drawing_types
//   - Recreates them as sparse indexes (allows multiple nulls but enforces uniqueness when present)
//   - Cleans up any documents with null or missing IDs
//
// Usage:
//
//	go run scripts/fix_drawing_list_index.go
//
// Environment Variables Required:
//
//	MONGODB_URI         MongoDB connection string (default: "mongodb://localhost:27017")
//	MONGODB_DATABASE    MongoDB database name (default: "ar13_backend")
package main

import (
	"context"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB
	_, err = mongodb.InitializeMongoDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer mongodb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db := mongodb.GetDatabase()
	categoryCollection := db.Collection("drawing_categories")
	typeCollection := db.Collection("drawing_types")

	log.Println("Fixing drawing list indexes and cleaning up null IDs...")
	log.Println("")

	// Fix categories collection
	if err := fixCategoryIndexes(ctx, categoryCollection); err != nil {
		log.Fatalf("Failed to fix category indexes: %v", err)
	}

	// Fix types collection
	if err := fixTypeIndexes(ctx, typeCollection); err != nil {
		log.Fatalf("Failed to fix type indexes: %v", err)
	}

	log.Println("✅ All indexes fixed and null IDs cleaned up!")
}

func fixCategoryIndexes(ctx context.Context, collection *mongo.Collection) error {
	log.Println("Fixing drawing_categories collection...")

	// Step 1: Clean up documents with null or missing IDs
	log.Println("  Cleaning up documents with null/missing IDs...")
	filter := bson.M{
		"$or": []bson.M{
			{"id": bson.M{"$exists": false}},
			{"id": nil},
			{"id": ""},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var docsToFix []bson.M
	if err := cursor.All(ctx, &docsToFix); err != nil {
		return err
	}

	if len(docsToFix) > 0 {
		log.Printf("  Found %d documents with null/missing IDs", len(docsToFix))
		for _, doc := range docsToFix {
			newID := uuid.New().String()
			update := bson.M{
				"$set": bson.M{
					"id": newID,
				},
			}
			// Use _id for update since id might be missing
			if docID, ok := doc["_id"]; ok {
				_, err := collection.UpdateOne(ctx, bson.M{"_id": docID}, update)
				if err != nil {
					log.Printf("    Warning: Failed to fix document: %v", err)
					continue
				}
				log.Printf("    ✅ Fixed document: assigned ID %s", newID)
			}
		}
	} else {
		log.Println("  No documents with null/missing IDs found")
	}

	// Step 2: Drop existing index if it exists
	log.Println("  Dropping existing id_unique index...")
	_, err = collection.Indexes().DropOne(ctx, "id_unique")
	if err != nil && err.Error() != "index not found" {
		log.Printf("    Warning: Failed to drop index (may not exist): %v", err)
	} else {
		log.Println("    ✅ Index dropped")
	}

	// Step 3: Create new sparse unique index
	log.Println("  Creating sparse unique index on 'id' field...")
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "id", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetSparse(true).
			SetName("id_unique"),
	}
	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	log.Println("    ✅ Sparse unique index created")

	return nil
}

func fixTypeIndexes(ctx context.Context, collection *mongo.Collection) error {
	log.Println("Fixing drawing_types collection...")

	// Step 1: Clean up documents with null or missing IDs
	log.Println("  Cleaning up documents with null/missing IDs...")
	filter := bson.M{
		"$or": []bson.M{
			{"id": bson.M{"$exists": false}},
			{"id": nil},
			{"id": ""},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var docsToFix []bson.M
	if err := cursor.All(ctx, &docsToFix); err != nil {
		return err
	}

	if len(docsToFix) > 0 {
		log.Printf("  Found %d documents with null/missing IDs", len(docsToFix))
		for _, doc := range docsToFix {
			newID := uuid.New().String()
			update := bson.M{
				"$set": bson.M{
					"id": newID,
				},
			}
			// Use _id for update since id might be missing
			if docID, ok := doc["_id"]; ok {
				_, err := collection.UpdateOne(ctx, bson.M{"_id": docID}, update)
				if err != nil {
					log.Printf("    Warning: Failed to fix document: %v", err)
					continue
				}
				log.Printf("    ✅ Fixed document: assigned ID %s", newID)
			}
		}
	} else {
		log.Println("  No documents with null/missing IDs found")
	}

	// Step 2: Drop existing index if it exists
	log.Println("  Dropping existing id_unique index...")
	_, err = collection.Indexes().DropOne(ctx, "id_unique")
	if err != nil && err.Error() != "index not found" {
		log.Printf("    Warning: Failed to drop index (may not exist): %v", err)
	} else {
		log.Println("    ✅ Index dropped")
	}

	// Step 3: Create new sparse unique index
	log.Println("  Creating sparse unique index on 'id' field...")
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "id", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetSparse(true).
			SetName("id_unique"),
	}
	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}
	log.Println("    ✅ Sparse unique index created")

	return nil
}

