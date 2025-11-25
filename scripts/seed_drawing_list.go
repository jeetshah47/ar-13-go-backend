// seed_drawing_list.go - Seeds drawing list master data to MongoDB
//
// This script seeds drawing categories and types as master data in the database.
// Drawing categories are stored in the "drawing_categories" collection.
// Drawing types are stored in the "drawing_types" collection.
//
// Usage:
//
//	go run scripts/seed_drawing_list.go
//	go run scripts/seed_drawing_list.go -force  # Force update existing data
//
// Flags:
//
//	-force    Force update existing data (default: false, skips if exists)
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
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Parse command-line flags
	force := flag.Bool("force", false, "Force update existing data")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Display connection info
	log.Println("Connecting to MongoDB...")
	if cfg.MongoDBURI != "" {
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
	}

	log.Println("✅ MongoDB connected successfully")
	defer mongodb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db := mongodb.GetDatabase()
	categoryCollection := db.Collection("drawing_categories")
	typeCollection := db.Collection("drawing_types")

	// Check if data already exists
	if !*force {
		categoryCount, _ := categoryCollection.CountDocuments(ctx, bson.M{})
		typeCount, _ := typeCollection.CountDocuments(ctx, bson.M{})
		if categoryCount > 0 || typeCount > 0 {
			log.Printf("Drawing list data already exists in database (categories: %d, types: %d). Use -force to update.", categoryCount, typeCount)
			log.Println("Skipping seed operation.")
			return
		}
	}

	log.Println("Seeding drawing list master data to MongoDB...")
	log.Println("")

	// Define drawing categories
	now := time.Now()
	categories := []models.DrawingCategory{
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil, // Updated is optional, set to nil initially
			},
			Name:        "Architectural Drawings",
			Description: "Technical drawings of buildings and structures",
			Order:       1,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			Name:        "Structural Drawings",
			Description: "Drawings showing structural elements and systems",
			Order:       2,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			Name:        "MEP Drawings",
			Description: "Mechanical, Electrical, and Plumbing drawings",
			Order:       3,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			Name:        "Site Plans",
			Description: "Drawings showing site layout and landscape",
			Order:       4,
			IsActive:    true,
		},
	}

	// Create sparse unique index on 'id' field for categories
	// Sparse index allows multiple documents with null/missing id values
	// but enforces uniqueness when id is present
	categoryIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "id", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetSparse(true).
			SetName("id_unique"),
	}
	_, err = categoryCollection.Indexes().CreateOne(ctx, categoryIndexModel)
	if err != nil {
		log.Printf("Warning: Failed to create category index (may already exist): %v", err)
	}

	// Insert or update categories
	categoryMap := make(map[string]string) // name -> id mapping
	insertedCategories := 0
	updatedCategories := 0

	for _, category := range categories {
		// Find existing category by name (for force update)
		var existingCategory models.DrawingCategory
		err := categoryCollection.FindOne(ctx, bson.M{"name": category.Name}).Decode(&existingCategory)
		if err == nil && existingCategory.ID != "" {
			// Update existing
			category.ID = existingCategory.ID
			updateTime := time.Now()
			category.Updated = &updateTime
			filter := bson.M{"id": category.ID}
			update := bson.M{
				"$set": bson.M{
					"name":        category.Name,
					"description": category.Description,
					"order":       category.Order,
					"isActive":    category.IsActive,
					"updated":     category.Updated,
				},
			}
			opts := options.Update()
			result, err := categoryCollection.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				log.Printf("Error updating category %s: %v", category.Name, err)
				continue
			}
			if result.ModifiedCount > 0 {
				updatedCategories++
				log.Printf("  🔄 Updated category: %s", category.Name)
			}
		} else {
			// Insert new
			_, err := categoryCollection.InsertOne(ctx, category)
			if err != nil {
				log.Printf("Error inserting category %s: %v", category.Name, err)
				continue
			}
			insertedCategories++
			log.Printf("  ✅ Inserted category: %s", category.Name)
		}
		categoryMap[category.Name] = category.ID
	}

	log.Println("")

	// Define drawing types (mapped to categories)
	types := []models.DrawingType{
		// Architectural Drawings
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["Architectural Drawings"],
			Name:        "Floor Plans",
			Description: "Horizontal cross-section views of buildings",
			Order:       1,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["Architectural Drawings"],
			Name:        "Elevations",
			Description: "Exterior views of building facades",
			Order:       2,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["Architectural Drawings"],
			Name:        "Sections",
			Description: "Vertical cross-section views",
			Order:       3,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["Architectural Drawings"],
			Name:        "Details",
			Description: "Detailed drawings of specific building components",
			Order:       4,
			IsActive:    true,
		},
		// Structural Drawings
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["Structural Drawings"],
			Name:        "Foundation Plans",
			Description: "Drawings showing foundation layout and details",
			Order:       1,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["Structural Drawings"],
			Name:        "Framing Plans",
			Description: "Drawings showing structural framing systems",
			Order:       2,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["Structural Drawings"],
			Name:        "Reinforcement Details",
			Description: "Details of steel reinforcement in concrete",
			Order:       3,
			IsActive:    true,
		},
		// MEP Drawings
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["MEP Drawings"],
			Name:        "Mechanical Plans",
			Description: "HVAC and mechanical system drawings",
			Order:       1,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["MEP Drawings"],
			Name:        "Electrical Plans",
			Description: "Electrical system and wiring diagrams",
			Order:       2,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["MEP Drawings"],
			Name:        "Plumbing Plans",
			Description: "Plumbing system and fixture layouts",
			Order:       3,
			IsActive:    true,
		},
		// Site Plans
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["Site Plans"],
			Name:        "Site Layout",
			Description: "Overall site plan showing building placement",
			Order:       1,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
				Updated: nil,
			},
			CategoryID:  categoryMap["Site Plans"],
			Name:        "Landscape Plans",
			Description: "Landscaping and outdoor space design",
			Order:       2,
			IsActive:    true,
		},
	}

	// Create sparse unique index on 'id' field for types
	// Sparse index allows multiple documents with null/missing id values
	// but enforces uniqueness when id is present
	typeIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "id", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetSparse(true).
			SetName("id_unique"),
	}
	_, err = typeCollection.Indexes().CreateOne(ctx, typeIndexModel)
	if err != nil {
		log.Printf("Warning: Failed to create type index (may already exist): %v", err)
	}

	// Create index on 'categoryId' field for efficient queries
	categoryIDIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "categoryId", Value: 1}},
		Options: options.Index().SetName("categoryId_index"),
	}
	_, err = typeCollection.Indexes().CreateOne(ctx, categoryIDIndexModel)
	if err != nil {
		log.Printf("Warning: Failed to create categoryId index (may already exist): %v", err)
	}

	// Insert or update types
	insertedTypes := 0
	updatedTypes := 0

	for _, drawingType := range types {
		// Find existing type by name and categoryId (for force update)
		var existingType models.DrawingType
		err := typeCollection.FindOne(ctx, bson.M{"name": drawingType.Name, "categoryId": drawingType.CategoryID}).Decode(&existingType)
		if err == nil && existingType.ID != "" {
			// Update existing
			updateTime := time.Now()
			drawingType.ID = existingType.ID
			drawingType.Updated = &updateTime
			filter := bson.M{"id": drawingType.ID}
			update := bson.M{
				"$set": bson.M{
					"categoryId":  drawingType.CategoryID,
					"name":        drawingType.Name,
					"description": drawingType.Description,
					"order":       drawingType.Order,
					"isActive":    drawingType.IsActive,
					"updated":     drawingType.Updated,
				},
			}
			opts := options.Update()
			result, err := typeCollection.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				log.Printf("Error updating type %s: %v", drawingType.Name, err)
				continue
			}
			if result.ModifiedCount > 0 {
				updatedTypes++
				log.Printf("  🔄 Updated type: %s", drawingType.Name)
			}
		} else {
			// Insert new
			_, err := typeCollection.InsertOne(ctx, drawingType)
			if err != nil {
				log.Printf("Error inserting type %s: %v", drawingType.Name, err)
				continue
			}
			insertedTypes++
			log.Printf("  ✅ Inserted type: %s", drawingType.Name)
		}
	}

	log.Println("")
	log.Printf("✅ Drawing list seeding completed!")
	log.Printf("   Categories - Inserted: %d, Updated: %d, Total: %d", insertedCategories, updatedCategories, len(categories))
	log.Printf("   Types - Inserted: %d, Updated: %d, Total: %d", insertedTypes, updatedTypes, len(types))
	log.Println("")
	log.Println("Drawing list master data is now available in MongoDB collections:")
	log.Println("  - drawing_categories")
	log.Println("  - drawing_types")
}

// maskMongoDBURI masks the password in MongoDB URI for safe display
func maskMongoDBURI(uri string) string {
	masked := uri
	if idx := findNthOccurrence(uri, "://", 1); idx > 0 {
		if atIdx := findNthOccurrence(uri, "@", 1); atIdx > idx {
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
