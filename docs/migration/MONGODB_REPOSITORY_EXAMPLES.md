# MongoDB Repository Implementation Examples

This document provides code examples for implementing MongoDB repositories based on the existing DynamoDB repositories.

## Base MongoDB Repository

### File: `internal/repos/mongodb_base.go`

```go
package repos

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoBaseRepo provides common MongoDB repository functionality
type MongoBaseRepo struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	collectionName string
}

// NewMongoBaseRepo creates a new MongoDB base repository
func NewMongoBaseRepo(client *mongo.Client, dbName, collectionName string) *MongoBaseRepo {
	db := client.Database(dbName)
	collection := db.Collection(collectionName)
	
	return &MongoBaseRepo{
		client:         client,
		database:       db,
		collection:     collection,
		collectionName: collectionName,
	}
}

// GetCollection returns the MongoDB collection
func (r *MongoBaseRepo) GetCollection() *mongo.Collection {
	return r.collection
}

// GetByID gets an item by ID
func (r *MongoBaseRepo) GetByID(ctx context.Context, id string) (*mongo.SingleResult) {
	filter := bson.M{"id": id}
	return r.collection.FindOne(ctx, filter)
}

// Exists checks if an item exists
func (r *MongoBaseRepo) Exists(ctx context.Context, id string) (bool, error) {
	filter := bson.M{"id": id}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteByID deletes an item by ID
func (r *MongoBaseRepo) DeleteByID(ctx context.Context, id string) error {
	filter := bson.M{"id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// InsertOne inserts a single document
func (r *MongoBaseRepo) InsertOne(ctx context.Context, document interface{}) error {
	_, err := r.collection.InsertOne(ctx, document)
	return err
}

// UpdateOne updates a single document
func (r *MongoBaseRepo) UpdateOne(ctx context.Context, id string, update bson.M) error {
	filter := bson.M{"id": id}
	
	// Add updatedAt timestamp
	update["updatedAt"] = time.Now()
	
	updateDoc := bson.M{"$set": update}
	_, err := r.collection.UpdateOne(ctx, filter, updateDoc)
	return err
}

// FindAll finds all documents with optional filter
func (r *MongoBaseRepo) FindAll(ctx context.Context, filter bson.M, limit *int64) ([]map[string]interface{}, error) {
	opts := options.Find()
	if limit != nil {
		opts.SetLimit(*limit)
	}
	opts.SetSort(bson.M{"created": -1}) // Default sort by creation date
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	
	return results, nil
}

// FindOne finds a single document
func (r *MongoBaseRepo) FindOne(ctx context.Context, filter bson.M) (*mongo.SingleResult) {
	return r.collection.FindOne(ctx, filter)
}

// QueryByField queries documents by a specific field (replaces GSI queries)
func (r *MongoBaseRepo) QueryByField(ctx context.Context, fieldName, fieldValue string) ([]map[string]interface{}, error) {
	filter := bson.M{fieldName: fieldValue}
	return r.FindAll(ctx, filter, nil)
}

// BatchGetItems retrieves multiple items by IDs
func (r *MongoBaseRepo) BatchGetItems(ctx context.Context, ids []string) ([]map[string]interface{}, error) {
	if len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}
	
	filter := bson.M{"id": bson.M{"$in": ids}}
	return r.FindAll(ctx, filter, nil)
}

// BatchInsertMany inserts multiple documents
func (r *MongoBaseRepo) BatchInsertMany(ctx context.Context, documents []interface{}) error {
	if len(documents) == 0 {
		return nil
	}
	
	_, err := r.collection.InsertMany(ctx, documents)
	return err
}

// CountDocuments counts documents matching a filter
func (r *MongoBaseRepo) CountDocuments(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}
```

## User Repository Example

### File: `internal/repos/user_repo_mongodb.go`

```go
package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepoMongo handles user data operations with MongoDB
type UserRepoMongo struct {
	*MongoBaseRepo
}

// NewUserRepoMongo creates a new MongoDB user repository
func NewUserRepoMongo(client *mongo.Client, dbName string) *UserRepoMongo {
	return &UserRepoMongo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "users"),
	}
}

// GetByEmail gets a user by email
func (r *UserRepoMongo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	filter := bson.M{"email": email}
	result := r.FindOne(ctx, filter)
	
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	
	var user models.User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}
	
	return &user, nil
}

// GetByID gets a user by ID
func (r *UserRepoMongo) GetByID(ctx context.Context, id string) (*models.User, error) {
	result := r.MongoBaseRepo.GetByID(ctx, id)
	
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	
	var user models.User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}
	
	return &user, nil
}

// GetAll gets all users
func (r *UserRepoMongo) GetAll(ctx context.Context, limit *int) ([]models.User, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}
	
	items, err := r.FindAll(ctx, bson.M{}, limitInt64)
	if err != nil {
		return nil, err
	}
	
	users := make([]models.User, 0, len(items))
	for _, item := range items {
		var user models.User
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &user); err != nil {
			continue // Skip invalid items
		}
		users = append(users, user)
	}
	
	return users, nil
}

// Add creates a new user
func (r *UserRepoMongo) Add(ctx context.Context, user *models.User) error {
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now
	
	// Convert to BSON document
	doc := bson.M{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"phoneNumber": user.PhoneNumber,
		"role":        string(user.Role),
		"password":    user.Password,
		"createdAt":   user.CreatedAt,
		"updatedAt":   user.UpdatedAt,
	}
	
	if user.Designation != nil {
		doc["designation"] = *user.Designation
	}
	
	return r.InsertOne(ctx, doc)
}

// Update updates a user
func (r *UserRepoMongo) Update(ctx context.Context, user *models.User) error {
	updates := bson.M{
		"name":        user.Name,
		"email":       user.Email,
		"phoneNumber": user.PhoneNumber,
		"role":        string(user.Role),
	}
	
	if user.Designation != nil {
		updates["designation"] = *user.Designation
	}
	
	if user.Password != "" {
		updates["password"] = user.Password
	}
	
	return r.UpdateOne(ctx, user.ID, updates)
}

// Delete deletes a user
func (r *UserRepoMongo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// Persists checks if a user exists
func (r *UserRepoMongo) Persists(ctx context.Context, id string) (bool, error) {
	return r.Exists(ctx, id)
}

// BatchGetItems retrieves multiple users by IDs
func (r *UserRepoMongo) BatchGetItems(ctx context.Context, ids []string) (map[string]*models.User, error) {
	items, err := r.BatchGetItems(ctx, ids)
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]*models.User)
	for _, item := range items {
		var user models.User
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &user); err != nil {
			continue
		}
		result[user.ID] = &user
	}
	
	return result, nil
}
```

## Task Repository Example

### File: `internal/repos/task_repo_mongodb.go`

```go
package repos

import (
	"context"
	"errors"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TaskRepoMongo handles task data operations with MongoDB
type TaskRepoMongo struct {
	*MongoBaseRepo
}

// NewTaskRepoMongo creates a new MongoDB task repository
func NewTaskRepoMongo(client *mongo.Client, dbName string) *TaskRepoMongo {
	return &TaskRepoMongo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "tasks"),
	}
}

// GetByID gets a task by ID
func (r *TaskRepoMongo) GetByID(ctx context.Context, projectID, taskID string) (*models.Task, error) {
	result := r.MongoBaseRepo.GetByID(ctx, taskID)
	
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	
	var task models.Task
	if err := result.Decode(&task); err != nil {
		return nil, err
	}
	
	return &task, nil
}

// GetAll gets all tasks for a project
func (r *TaskRepoMongo) GetAll(ctx context.Context, projectID string) ([]models.Task, error) {
	filter := bson.M{"projectId": projectID}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}
	
	tasks := make([]models.Task, 0, len(items))
	for _, item := range items {
		var task models.Task
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &task); err != nil {
			continue
		}
		tasks = append(tasks, task)
	}
	
	return tasks, nil
}

// Add creates a new task
func (r *TaskRepoMongo) Add(ctx context.Context, task *models.Task) error {
	now := time.Now()
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	task.Created = now
	
	// Initialize empty slices if nil
	if task.TimeSpent == nil {
		task.TimeSpent = []models.TimeSpent{}
	}
	if task.FileAttachments == nil {
		task.FileAttachments = []models.FileAttachment{}
	}
	if task.ActivityLogs == nil {
		task.ActivityLogs = []models.ActivityLog{}
	}
	
	doc := bson.M{
		"id":              task.ID,
		"subject":         task.Subject,
		"code":            task.Code,
		"status":          task.Status,
		"deadline":        task.Deadline,
		"priority":        task.Priority,
		"projectId":       task.ProjectID,
		"timeSpent":       task.TimeSpent,
		"fileAttachments": task.FileAttachments,
		"activityLogs":    task.ActivityLogs,
		"created":         task.Created,
	}
	
	if task.AssignTo != nil {
		doc["assignTo"] = *task.AssignTo
	}
	if task.Description != nil {
		doc["description"] = *task.Description
	}
	if task.Progress != nil {
		doc["progress"] = *task.Progress
	}
	
	return r.InsertOne(ctx, doc)
}

// Update updates a task
func (r *TaskRepoMongo) Update(ctx context.Context, task *models.Task) error {
	now := time.Now()
	task.Updated = &now
	
	updates := bson.M{
		"subject":         task.Subject,
		"code":            task.Code,
		"status":          task.Status,
		"deadline":        task.Deadline,
		"priority":        task.Priority,
		"projectId":       task.ProjectID,
		"timeSpent":       task.TimeSpent,
		"fileAttachments": task.FileAttachments,
		"activityLogs":    task.ActivityLogs,
		"updated":         task.Updated,
	}
	
	if task.AssignTo != nil {
		updates["assignTo"] = *task.AssignTo
	}
	if task.Description != nil {
		updates["description"] = *task.Description
	}
	if task.Progress != nil {
		updates["progress"] = *task.Progress
	}
	
	return r.UpdateOne(ctx, task.ID, updates)
}

// Delete deletes a task
func (r *TaskRepoMongo) Delete(ctx context.Context, projectID, taskID string) error {
	return r.DeleteByID(ctx, taskID)
}

// UpdateDeadline updates task deadline
func (r *TaskRepoMongo) UpdateDeadline(ctx context.Context, projectID, taskID string, deadline time.Time) error {
	updates := bson.M{"deadline": deadline}
	return r.UpdateOne(ctx, taskID, updates)
}

// UpdateProgress updates task progress
func (r *TaskRepoMongo) UpdateProgress(ctx context.Context, projectID, taskID string, progress int) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	
	updates := bson.M{"progress": progress}
	return r.UpdateOne(ctx, taskID, updates)
}

// AddTimeSpent adds a time spent entry
func (r *TaskRepoMongo) AddTimeSpent(ctx context.Context, projectID, taskID string, timeSpent models.TimeSpent) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	
	task.TimeSpent = append(task.TimeSpent, timeSpent)
	updates := bson.M{"timeSpent": task.TimeSpent}
	return r.UpdateOne(ctx, taskID, updates)
}
```

## MongoDB Connection Setup

### File: `pkg/mongodb/mongodb.go`

```go
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

// InitializeMongoDB initializes MongoDB client
func InitializeMongoDB(uri, databaseName string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second)
	
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

// Close closes the MongoDB connection
func Close() error {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return mongoClient.Disconnect(ctx)
	}
	return nil
}
```

## Index Creation Example

### File: `scripts/create_mongodb_indexes.go`

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/ar-13-go-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Initialize MongoDB (use your config)
	client, err := mongodb.InitializeMongoDB("mongodb://localhost:27017", "ar13_backend")
	if err != nil {
		log.Fatal(err)
	}
	defer mongodb.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	db := mongodb.GetDatabase()
	
	// Create indexes for users collection
	usersCollection := db.Collection("users")
	usersIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"role": 1},
		},
		{
			Keys: map[string]interface{}{"createdAt": -1},
		},
	}
	
	_, err = usersCollection.Indexes().CreateMany(ctx, usersIndexes)
	if err != nil {
		log.Fatalf("Failed to create users indexes: %v", err)
	}
	
	log.Println("Users indexes created successfully")
	
	// Create indexes for tasks collection
	tasksCollection := db.Collection("tasks")
	tasksIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"projectId": 1},
		},
		{
			Keys: map[string]interface{}{"assignTo": 1},
		},
		{
			Keys: map[string]interface{}{"status": 1},
		},
		{
			Keys: map[string]interface{}{"deadline": 1},
		},
		{
			Keys: map[string]interface{}{"projectId": 1, "status": 1},
		},
	}
	
	_, err = tasksCollection.Indexes().CreateMany(ctx, tasksIndexes)
	if err != nil {
		log.Fatalf("Failed to create tasks indexes: %v", err)
	}
	
	log.Println("Tasks indexes created successfully")
	
	// Repeat for other collections...
}
```

---

## Notes

1. **Error Handling**: Always check for `mongo.ErrNoDocuments` when a document is not found
2. **Context**: Use context with timeout for all operations
3. **BSON Tags**: Add `bson` tags to model structs
4. **Date Handling**: MongoDB stores dates as `time.Time`, no conversion needed
5. **Nested Documents**: MongoDB handles nested structures natively
6. **Transactions**: Use MongoDB transactions for multi-document operations if needed

---

**Last Updated**: 2025-01-XX

