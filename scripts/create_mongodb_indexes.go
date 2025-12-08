// MongoDB Index Optimization Strategy
//
// This script creates only ESSENTIAL indexes to minimize write overhead while maintaining query performance.
//
// Index Overhead Considerations:
// - Each index adds write overhead (every write must update the index)
// - More indexes = slower writes, more storage, more memory usage
// - Only create indexes for fields that are actually queried
//
// Index Strategy:
// 1. PRIMARY KEY: Always index 'id' field (unique) - required for GetByID operations
// 2. QUERY INDEXES: Only index fields used in QueryByIndex operations (replaces DynamoDB GSI)
// 3. NO SORTING INDEXES: Removed unless sorting large result sets (>1000 docs)
// 4. NO COMPOUND INDEXES: Removed unless queries frequently combine multiple fields
// 5. NO ARRAY INDEXES: Removed unless array elements are queried directly
//
// To add indexes later:
// - Monitor slow queries using MongoDB profiler
// - Add indexes only for queries that are slow and frequently executed
// - Test write performance impact before adding production indexes

package main

import (
	"context"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/pkg/mongodb"
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db := mongodb.GetDatabase()

	log.Println("Creating MongoDB indexes...")
	log.Println("Note: Only essential indexes are created to minimize write overhead.")
	log.Println("      Add additional indexes later if query patterns require them.")
	log.Println("")

	// Create indexes for users collection
	if err := createUserIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create users indexes: %v", err)
	}

	// Create indexes for projects collection
	if err := createProjectIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create projects indexes: %v", err)
	}

	// Create indexes for tasks collection
	if err := createTaskIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create tasks indexes: %v", err)
	}

	// Create indexes for notifications collection
	if err := createNotificationIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create notifications indexes: %v", err)
	}

	// Create indexes for calendar_events collection
	if err := createCalendarEventIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create calendar_events indexes: %v", err)
	}

	// Create indexes for leaveRequests collection
	if err := createLeaveRequestIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create leaveRequests indexes: %v", err)
	}

	// Create indexes for activity_logs collection
	if err := createActivityLogIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create activity_logs indexes: %v", err)
	}

	// Create indexes for project_details collection
	if err := createProjectDetailsIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create project_details indexes: %v", err)
	}

	// Create indexes for user_account_links collection
	if err := createUserAccountLinkIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create user_account_links indexes: %v", err)
	}

	// Create indexes for signupInvitations collection
	if err := createSignupInvitationIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create signupInvitations indexes: %v", err)
	}

	// Create indexes for role_permissions collection
	if err := createRolePermissionIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create role_permissions indexes: %v", err)
	}

	// Create indexes for task_statuses collection
	if err := createTaskStatusIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create task_statuses indexes: %v", err)
	}

	// Create indexes for info-portal collection
	if err := createInfoPortalIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create info-portal indexes: %v", err)
	}

	// Create indexes for audit_logs collection
	if err := createAuditLogIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create audit_logs indexes: %v", err)
	}

	log.Println("✅ All indexes created successfully!")
}

func createUserIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")
	// CRITICAL: id (primary key) and email (used in GetByEmail query)
	// REMOVED: role and createdAt indexes - not queried directly, add write overhead
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true).SetName("email_unique"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Users indexes created (id, email)")
	return nil
}

func createProjectIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("projects")
	// CRITICAL: id (primary key) only
	// REMOVED: ownerId, membersIds, created - not queried via indexes (only scans), add write overhead
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Projects indexes created (id only)")
	return nil
}

func createTaskIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("tasks")
	// CRITICAL: id (primary key) and projectId (used in GetAll by project)
	// REMOVED: assignTo, status, deadline, compound - not queried via indexes, add write overhead
	// Note: If you frequently query tasks by assignTo or status, add those indexes later
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"projectId": 1},
			Options: options.Index().SetName("projectId_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Tasks indexes created (id, projectId)")
	return nil
}

func createNotificationIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("notifications")
	// CRITICAL: id (primary key) and userId (used in multiple queries)
	// REMOVED: compound userId+read - GetUnread() queries by userId then filters in memory,
	//          so compound index isn't used. For small collections, in-memory filtering is fast.
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"userId": 1},
			Options: options.Index().SetName("userId_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Notifications indexes created (id, userId)")
	return nil
}

func createCalendarEventIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("calendar_events")
	// CRITICAL: id (primary key) only
	// REMOVED: userId+start, start+end - GetByMonth uses scan with filter, not indexed query
	// Note: If you add indexed date range queries later, add start/end indexes then
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Calendar events indexes created (id only)")
	return nil
}

func createLeaveRequestIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("leaveRequests")
	// CRITICAL: id (primary key), userId (GetByUserID - frequently used)
	// REMOVED: status index - GetByStatus() is admin-only and infrequent.
	//          For small collections (10-15 users), scan with filter is acceptable.
	//          Can add back if GetByStatus becomes a performance bottleneck.
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"userId": 1},
			Options: options.Index().SetName("userId_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Leave requests indexes created (id, userId)")
	return nil
}

func createActivityLogIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("activity_logs")
	// CRITICAL: id (primary key) and entityId (used in GetByEntity)
	// REMOVED: compound entityId+timestamp - sorting can be done in memory for small result sets.
	//          Add back if activity logs per entity exceed 1000+ and sorting becomes slow.
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"entityId": 1},
			Options: options.Index().SetName("entityId_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Activity logs indexes created (id, entityId)")
	return nil
}

func createProjectDetailsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("project_details")
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"projectId": 1},
			Options: options.Index().SetUnique(true).SetName("projectId_unique"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Project details indexes created")
	return nil
}

func createUserAccountLinkIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("user_account_links")
	// CRITICAL: id (primary key), userId (GetByUserID, GetByUserIDAndProvider)
	// REMOVED: compound provider+providerId - GetByProvider() uses ScanItems() not QueryByIndex(),
	//          so this index isn't used for queries. Also field name is "providerUserId" not "providerId".
	//          Uniqueness can be enforced at application level. If GetByProvider() is optimized later
	//          to use indexed queries, add this index then.
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"userId": 1},
			Options: options.Index().SetName("userId_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ User account links indexes created (id, userId)")
	return nil
}

func createSignupInvitationIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("signupInvitations")
	// CRITICAL: id (primary key), email (GetByEmail), token (GetByToken)
	// REMOVED: TTL index on expiresAt - field is actually "linkExpiry", and expiration
	//          is checked manually in ValidateSignupToken(). TTL indexes add write overhead.
	//          If automatic cleanup is needed, use a scheduled job instead.
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetName("email_idx"),
		},
		{
			Keys:    map[string]interface{}{"token": 1},
			Options: options.Index().SetUnique(true).SetName("token_unique"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Signup invitations indexes created (id, email, token)")
	return nil
}

func createRolePermissionIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("role_permissions")
	// CRITICAL: id (primary key) and role (used in GetByRole)
	// REMOVED: compound role+permission unique - uniqueness can be enforced at application level.
	//          Role permissions are rarely written, so duplicate prevention in code is acceptable.
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"role": 1},
			Options: options.Index().SetName("role_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Role permissions indexes created (id, role)")
	return nil
}

func createTaskStatusIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("task_statuses")
	// CRITICAL: id (primary key) only
	// REMOVED: value and order indexes - unnecessary for small collection (6 documents)
	// - Collection is not queried by 'value' field in application code
	// - Sorting by 'order' can be done in memory for 6 documents
	// - This minimizes write overhead and storage usage
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Task statuses indexes created (id only)")
	return nil
}

func createInfoPortalIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("info-portal")
	// CRITICAL: id (primary key) only
	// REMOVED: nested folder/page indexes - info portal uses single document with nested arrays
	// Nested array indexes add significant write overhead and are rarely needed
	// If you need to query by folder/page ID frequently, consider restructuring data
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Info portal indexes created (id only)")
	return nil
}

func createAuditLogIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("audit_logs")
	// CRITICAL: id (primary key), userId (GetByUserID), path (GetByPath), method (GetByMethod), requestTime (sorting)
	// Note: requestTime index is important for date range queries and sorting recent logs
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true).SetName("id_unique"),
		},
		{
			Keys:    map[string]interface{}{"userId": 1},
			Options: options.Index().SetName("userId_idx"),
		},
		{
			Keys:    map[string]interface{}{"path": 1},
			Options: options.Index().SetName("path_idx"),
		},
		{
			Keys:    map[string]interface{}{"method": 1},
			Options: options.Index().SetName("method_idx"),
		},
		{
			Keys:    map[string]interface{}{"requestTime": -1},
			Options: options.Index().SetName("requestTime_idx"),
		},
		{
			Keys:    map[string]interface{}{"statusCode": 1},
			Options: options.Index().SetName("statusCode_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}
	log.Println("  ✅ Audit logs indexes created (id, userId, path, method, requestTime, statusCode)")
	return nil
}

