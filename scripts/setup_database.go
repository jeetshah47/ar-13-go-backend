// setup_database.go - Sets up database collections, indexes, and master data
//
// This script:
//   - Creates all MongoDB collections (if they don't exist)
//   - Creates indexes for optimal query performance
//   - Seeds master data (roles, permissions, task statuses, drawing categories/types)
//
// Usage:
//
//	go run scripts/setup_database.go
//	go run scripts/setup_database.go -skip-indexes  # Skip index creation
//	go run scripts/setup_database.go -skip-master-data  # Skip master data seeding
//	go run scripts/setup_database.go -skip-collections  # Skip collection creation
//
// Flags:
//
//	-skip-collections   Skip creating collections
//	-skip-indexes       Skip creating indexes
//	-skip-master-data   Skip seeding master data
//	-force              Force recreate indexes and overwrite existing master data
//
// Environment Variables Required:
//
//	MONGODB_URI         MongoDB connection string (default: "mongodb://localhost:27017")
//	MONGODB_DATABASE    MongoDB database name (default: "ar13_backend")
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CollectionConfig defines a collection and its indexes
type CollectionConfig struct {
	Name    string
	Indexes []IndexConfig
}

// IndexConfig defines an index configuration
type IndexConfig struct {
	Keys    bson.M
	Unique  bool
	Options *options.IndexOptions
}

func main() {
	// Parse command-line flags
	skipCollections := flag.Bool("skip-collections", false, "Skip creating collections")
	skipIndexes := flag.Bool("skip-indexes", false, "Skip creating indexes")
	skipMasterData := flag.Bool("skip-master-data", false, "Skip seeding master data")
	force := flag.Bool("force", false, "Force recreate indexes and overwrite existing master data")
	flag.Parse()

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
	fmt.Println("✅ MongoDB initialized successfully")

	ctx := context.Background()
	db := mongodb.GetDatabase()

	// Create collections
	if !*skipCollections {
		createCollections(ctx, db)
	}

	// Create indexes
	if !*skipIndexes {
		createIndexes(ctx, db, *force)
	}

	// Seed master data
	if !*skipMasterData {
		seedMasterData(ctx, *force)
	}

	fmt.Println("\n✅ Database setup completed successfully!")
}

// createCollections ensures all collections exist
func createCollections(ctx context.Context, db *mongo.Database) {
	fmt.Println("\n📦 Creating collections...")

	collections := []string{
		"users",
		"role_permissions",
		"projects",
		"tasks",
		"activity_logs",
		"activity_log_replies",
		"calendar_events",
		"leaveRequests",
		"notifications",
		"drawing_categories",
		"drawing_types",
		"task_statuses",
		"time_tracking_sessions",
		"user_account_links",
		"signupInvitations",
		"project_details",
		"info-portal",
	}

	for _, collName := range collections {
		// MongoDB creates collections automatically on first insert,
		// but we can verify they exist by trying to create them
		// This will not error if the collection already exists
		err := db.CreateCollection(ctx, collName)
		if err != nil {
			// Check if it's a namespace exists error (collection already exists)
			if cmdErr, ok := err.(mongo.CommandError); ok && cmdErr.Code == 48 {
				fmt.Printf("   ✓ Collection '%s' already exists\n", collName)
			} else {
				// For other errors, just log a warning - collection will be created on first insert
				fmt.Printf("   ⚠ Collection '%s': %v (will be created on first insert)\n", collName, err)
			}
		} else {
			fmt.Printf("   ✓ Created collection '%s'\n", collName)
		}
	}

	fmt.Println("✅ Collections setup completed")
}

// createIndexes creates indexes for all collections
func createIndexes(ctx context.Context, db *mongo.Database, force bool) {
	fmt.Println("\n🔍 Creating indexes...")

	// Users collection indexes
	createCollectionIndexes(ctx, db, "users", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"email": 1}, Unique: true},
		{Keys: bson.M{"role": 1}},
		{Keys: bson.M{"createdAt": -1}},
	}, force)

	// Role permissions collection indexes
	createCollectionIndexes(ctx, db, "role_permissions", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"role": 1}},
		{Keys: bson.M{"permission": 1}},
		{Keys: bson.M{"role": 1, "permission": 1}, Unique: true},
		{Keys: bson.M{"createdAt": -1}},
	}, force)

	// Projects collection indexes
	createCollectionIndexes(ctx, db, "projects", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"ownerId": 1}},
		{Keys: bson.M{"membersIds": 1}},
		{Keys: bson.M{"project_code": 1}, Unique: true},
		{Keys: bson.M{"isArchived": 1}},
		{Keys: bson.M{"created": -1}},
	}, force)

	// Tasks collection indexes
	createCollectionIndexes(ctx, db, "tasks", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"projectId": 1}},
		{Keys: bson.M{"assignTo": 1}},
		{Keys: bson.M{"status": 1}},
		{Keys: bson.M{"code": 1}},
		{Keys: bson.M{"deadline": 1}},
		{Keys: bson.M{"created": -1}},
		{Keys: bson.M{"projectId": 1, "status": 1}},
	}, force)

	// Activity logs collection indexes
	createCollectionIndexes(ctx, db, "activity_logs", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"userId": 1}},
		{Keys: bson.M{"timestamp": -1}},
	}, force)

	// Activity log replies collection indexes
	createCollectionIndexes(ctx, db, "activity_log_replies", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"activityLogId": 1}},
		{Keys: bson.M{"userId": 1}},
		{Keys: bson.M{"created": -1}},
	}, force)

	// Calendar events collection indexes
	createCollectionIndexes(ctx, db, "calendar_events", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"createdBy": 1}},
		{Keys: bson.M{"start": 1}},
		{Keys: bson.M{"end": 1}},
		{Keys: bson.M{"created": -1}},
		{Keys: bson.M{"start": 1, "end": 1}},
	}, force)

	// Leave requests collection indexes
	createCollectionIndexes(ctx, db, "leaveRequests", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"userId": 1}},
		{Keys: bson.M{"status": 1}},
		{Keys: bson.M{"requestType": 1}},
		{Keys: bson.M{"startDate": 1}},
		{Keys: bson.M{"created": -1}},
		{Keys: bson.M{"userId": 1, "status": 1}},
	}, force)

	// Notifications collection indexes
	createCollectionIndexes(ctx, db, "notifications", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"userId": 1}},
		{Keys: bson.M{"isRead": 1}},
		{Keys: bson.M{"createdAt": -1}},
		{Keys: bson.M{"userId": 1, "isRead": 1}},
		{Keys: bson.M{"relatedEntityId": 1, "relatedEntityType": 1}},
	}, force)

	// Drawing categories collection indexes
	createCollectionIndexes(ctx, db, "drawing_categories", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"order": 1}},
		{Keys: bson.M{"isActive": 1}},
	}, force)

	// Drawing types collection indexes
	createCollectionIndexes(ctx, db, "drawing_types", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"categoryId": 1}},
		{Keys: bson.M{"order": 1}},
		{Keys: bson.M{"isActive": 1}},
		{Keys: bson.M{"categoryId": 1, "order": 1}},
	}, force)

	// Task statuses collection indexes
	createCollectionIndexes(ctx, db, "task_statuses", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"order": 1}},
		{Keys: bson.M{"name": 1}, Unique: true},
	}, force)

	// Time tracking sessions collection indexes
	createCollectionIndexes(ctx, db, "time_tracking_sessions", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"taskId": 1}},
		{Keys: bson.M{"userId": 1}},
		{Keys: bson.M{"isActive": 1}},
		{Keys: bson.M{"startTime": -1}},
		{Keys: bson.M{"userId": 1, "isActive": 1}},
	}, force)

	// User account links collection indexes
	createCollectionIndexes(ctx, db, "user_account_links", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"userId": 1}, Unique: true},
		{Keys: bson.M{"googleEmail": 1}},
	}, force)

	// Signup invitations collection indexes
	createCollectionIndexes(ctx, db, "signupInvitations", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"token": 1}, Unique: true},
		{Keys: bson.M{"email": 1}},
		{Keys: bson.M{"isUsed": 1}},
		{Keys: bson.M{"expiresAt": 1}},
	}, force)

	// Project details collection indexes
	createCollectionIndexes(ctx, db, "project_details", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"projectId": 1}, Unique: true},
	}, force)

	// Info portal collection indexes
	createCollectionIndexes(ctx, db, "info-portal", []IndexConfig{
		{Keys: bson.M{"id": 1}, Unique: true},
		{Keys: bson.M{"type": 1}},
		{Keys: bson.M{"folderId": 1}},
		{Keys: bson.M{"pageId": 1}},
		{Keys: bson.M{"created": -1}},
	}, force)

	fmt.Println("✅ Indexes setup completed")
}

// createCollectionIndexes creates indexes for a specific collection
func createCollectionIndexes(ctx context.Context, db *mongo.Database, collectionName string, indexes []IndexConfig, force bool) {
	collection := db.Collection(collectionName)

	if force {
		// Drop existing indexes (except _id)
		indexesList, err := collection.Indexes().List(ctx)
		if err == nil {
			for indexesList.Next(ctx) {
				var index bson.M
				if err := indexesList.Decode(&index); err == nil {
					if name, ok := index["name"].(string); ok && name != "_id_" {
						collection.Indexes().DropOne(ctx, name)
					}
				}
			}
		}
	}

	for _, idxConfig := range indexes {
		opts := options.Index()
		if idxConfig.Options != nil {
			opts = idxConfig.Options
		}
		opts.SetUnique(idxConfig.Unique)

		indexModel := mongo.IndexModel{
			Keys:    idxConfig.Keys,
			Options: opts,
		}

		indexName, err := collection.Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			// Index might already exist, which is fine
			if mongo.IsDuplicateKeyError(err) {
				fmt.Printf("   ✓ Index on '%s' already exists\n", collectionName)
			} else {
				fmt.Printf("   ⚠ Index on '%s': %v\n", collectionName, err)
			}
		} else {
			fmt.Printf("   ✓ Created index '%s' on '%s'\n", indexName, collectionName)
		}
	}
}

// seedMasterData seeds all master data
func seedMasterData(ctx context.Context, force bool) {
	fmt.Println("\n🌱 Seeding master data...")

	// Seed role permissions
	seedRolePermissionsWithForce(ctx, force)

	// Seed task statuses
	seedTaskStatuses(ctx, force)

	// Seed drawing categories and types
	seedDrawingData(ctx, force)

	fmt.Println("✅ Master data seeding completed")
}

// seedRolePermissionsWithForce seeds role permissions with force option
func seedRolePermissionsWithForce(ctx context.Context, force bool) {
	fmt.Println("\n   📋 Seeding role permissions...")

	repo := repos.NewRolePermissionRepo()

	// Check if permissions already exist
	existingAdmin, _ := repo.GetByRole(ctx, models.UserRoleAdmin)
	existingStandard, _ := repo.GetByRole(ctx, models.UserRoleStandard)

	if len(existingAdmin) > 0 || len(existingStandard) > 0 {
		if !force {
			fmt.Println("   ⚠️  Role permissions already exist. Use -force to overwrite.")
			return
		}
		// Delete existing permissions
		allExisting, _ := repo.GetAll(ctx)
		for _, perm := range allExisting {
			var rp models.RolePermission
			bsonBytes, _ := bson.Marshal(perm)
			bson.Unmarshal(bsonBytes, &rp)
			repo.Delete(ctx, rp.ID)
		}
		fmt.Println("   ✓ Deleted existing role permissions")
	}

	// Admin permissions
	adminPermissions := []string{
		"projects:read", "projects:write", "projects:delete",
		"tasks:read", "tasks:write", "tasks:delete", "tasks:assign",
		"users:read", "users:write", "users:delete", "users:profile", "users:invite",
		"calendar:read", "calendar:write", "calendar:delete",
		"vacation:read", "vacation:write", "vacation:delete", "vacation:approve",
		"notifications:read", "notifications:write", "notifications:delete",
		"activityLogs:read",
		"dashboard:read",
		"employees:read",
		"infoPortal:read", "infoPortal:write", "infoPortal:delete",
		"googleAccount:read", "googleAccount:write", "googleAccount:link", "googleAccount:unlink",
		"websocket:connect",
		"auth:read",
		"projectDetails:read", "projectDetails:write", "projectDetails:delete",
		"backup:read", "backup:write",
	}

	// Standard permissions
	standardPermissions := []string{
		"projects:read",
		"tasks:read", "tasks:write", "tasks:delete",
		"calendar:read", "calendar:write", "calendar:delete",
		"vacation:read", "vacation:write", "vacation:delete",
		"notifications:read", "notifications:write", "notifications:delete",
		"activityLogs:read",
		"dashboard:read",
		"employees:read",
		"infoPortal:read", "infoPortal:write", "infoPortal:delete",
		"googleAccount:read", "googleAccount:write", "googleAccount:link", "googleAccount:unlink",
		"websocket:connect",
		"auth:read",
		"users:profile",
		"projectDetails:read",
	}

	// Create role permissions
	var allPermissions []models.RolePermission
	now := time.Now()

	// Add admin permissions
	for _, perm := range adminPermissions {
		allPermissions = append(allPermissions, models.RolePermission{
			ID:         uuid.New().String(),
			Role:       models.UserRoleAdmin,
			Permission: perm,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	// Add standard permissions
	for _, perm := range standardPermissions {
		allPermissions = append(allPermissions, models.RolePermission{
			ID:         uuid.New().String(),
			Role:       models.UserRoleStandard,
			Permission: perm,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	// Batch insert
	if err := repo.BatchAdd(ctx, allPermissions); err != nil {
		log.Fatalf("❌ Failed to seed role permissions: %v", err)
	}

	fmt.Printf("   ✅ Seeded %d role permissions (Admin: %d, Standard: %d)\n",
		len(allPermissions), len(adminPermissions), len(standardPermissions))
}

// seedTaskStatuses seeds default task statuses
func seedTaskStatuses(ctx context.Context, force bool) {
	fmt.Println("\n   📊 Seeding task statuses...")

	repo := repos.NewTaskStatusRepo()
	collection := repo.GetCollection()

	// Check if statuses already exist
	count, _ := collection.CountDocuments(ctx, bson.M{})
	if count > 0 {
		if !force {
			fmt.Println("   ⚠️  Task statuses already exist. Use -force to overwrite.")
			return
		}
		// Delete existing statuses
		collection.DeleteMany(ctx, bson.M{})
		fmt.Println("   ✓ Deleted existing task statuses")
	}

	// Default task statuses
	statuses := []map[string]interface{}{
		{
			"id":          uuid.New().String(),
			"name":        "Not Started",
			"order":       1,
			"color":       "#808080",
			"description": "Task has been created but not started",
			"created":     time.Now(),
		},
		{
			"id":          uuid.New().String(),
			"name":        "In Progress",
			"order":       2,
			"color":       "#2196F3",
			"description": "Task is currently being worked on",
			"created":     time.Now(),
		},
		{
			"id":          uuid.New().String(),
			"name":        "On Hold",
			"order":       3,
			"color":       "#FF9800",
			"description": "Task is temporarily paused",
			"created":     time.Now(),
		},
		{
			"id":          uuid.New().String(),
			"name":        "Review",
			"order":       4,
			"color":       "#9C27B0",
			"description": "Task is ready for review",
			"created":     time.Now(),
		},
		{
			"id":          uuid.New().String(),
			"name":        "Completed",
			"order":       5,
			"color":       "#4CAF50",
			"description": "Task has been completed",
			"created":     time.Now(),
		},
		{
			"id":          uuid.New().String(),
			"name":        "Cancelled",
			"order":       6,
			"color":       "#F44336",
			"description": "Task has been cancelled",
			"created":     time.Now(),
		},
	}

	// Convert to interface slice for batch insert
	items := make([]interface{}, len(statuses))
	for i, status := range statuses {
		items[i] = status
	}

	if _, err := collection.InsertMany(ctx, items); err != nil {
		log.Fatalf("❌ Failed to seed task statuses: %v", err)
	}

	fmt.Printf("   ✅ Seeded %d task statuses\n", len(statuses))
}

// seedDrawingData seeds drawing categories and types
func seedDrawingData(ctx context.Context, force bool) {
	fmt.Println("\n   📐 Seeding drawing categories and types...")

	categoryRepo := repos.NewDrawingListRepo()
	typeRepo := repos.NewDrawingTypeRepo()

	// Check if categories already exist
	categoryCount, _ := categoryRepo.CountDocuments(ctx, bson.M{})
	typeCount, _ := typeRepo.CountDocuments(ctx, bson.M{})

	if categoryCount > 0 || typeCount > 0 {
		if !force {
			fmt.Println("   ⚠️  Drawing data already exists. Use -force to overwrite.")
			return
		}
		// Delete existing data
		typeRepo.DeleteMany(ctx, bson.M{})
		categoryRepo.DeleteMany(ctx, bson.M{})
		fmt.Println("   ✓ Deleted existing drawing data")
	}

	now := time.Now()

	// Create categories
	categories := []*models.DrawingCategory{
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
			},
			Name:        "Architectural",
			Description: "Architectural drawings and plans",
			Order:       1,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
			},
			Name:        "Structural",
			Description: "Structural engineering drawings",
			Order:       2,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
			},
			Name:        "MEP",
			Description: "Mechanical, Electrical, and Plumbing drawings",
			Order:       3,
			IsActive:    true,
		},
		{
			Model: models.Model{
				ID:      uuid.New().String(),
				Created: now,
			},
			Name:        "Site",
			Description: "Site plans and surveys",
			Order:       4,
			IsActive:    true,
		},
	}

	// Insert categories
	for _, cat := range categories {
		if err := categoryRepo.AddCategory(ctx, cat); err != nil {
			log.Printf("⚠️  Failed to insert category %s: %v", cat.Name, err)
		}
	}

	// Create types for each category
	types := []*models.DrawingType{
		// Architectural types
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[0].ID, Name: "Floor Plans", Description: "Building floor plans", Order: 1, IsActive: true},
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[0].ID, Name: "Elevations", Description: "Building elevations", Order: 2, IsActive: true},
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[0].ID, Name: "Sections", Description: "Building sections", Order: 3, IsActive: true},
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[0].ID, Name: "Details", Description: "Architectural details", Order: 4, IsActive: true},

		// Structural types
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[1].ID, Name: "Foundation Plans", Description: "Foundation structural plans", Order: 1, IsActive: true},
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[1].ID, Name: "Framing Plans", Description: "Structural framing plans", Order: 2, IsActive: true},
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[1].ID, Name: "Details", Description: "Structural details", Order: 3, IsActive: true},

		// MEP types
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[2].ID, Name: "Mechanical Plans", Description: "HVAC and mechanical plans", Order: 1, IsActive: true},
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[2].ID, Name: "Electrical Plans", Description: "Electrical plans", Order: 2, IsActive: true},
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[2].ID, Name: "Plumbing Plans", Description: "Plumbing plans", Order: 3, IsActive: true},

		// Site types
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[3].ID, Name: "Site Plans", Description: "Site layout plans", Order: 1, IsActive: true},
		{Model: models.Model{ID: uuid.New().String(), Created: now}, CategoryID: categories[3].ID, Name: "Landscape Plans", Description: "Landscape and hardscape plans", Order: 2, IsActive: true},
	}

	// Insert types
	for _, t := range types {
		if err := typeRepo.AddType(ctx, t); err != nil {
			log.Printf("⚠️  Failed to insert type %s: %v", t.Name, err)
		}
	}

	fmt.Printf("   ✅ Seeded %d drawing categories and %d drawing types\n", len(categories), len(types))
}

