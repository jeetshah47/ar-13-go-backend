package repos

import (
	"context"
	"errors"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TaskRepo handles task data operations with MongoDB
type TaskRepo struct {
	*MongoBaseRepo
}

// NewTaskRepo creates a new MongoDB task repository
func NewTaskRepo() *TaskRepo {
	client := mongodb.GetClient()
	if client == nil {
		panic("MongoDB client is not initialized. Please ensure MongoDB is connected before creating repositories.")
	}
	return &TaskRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, mongodb.GetDatabaseName(), "tasks"),
	}
}

// GetByID gets a task by ID
func (r *TaskRepo) GetByID(ctx context.Context, projectID, taskID string) (*models.Task, error) {
	// Query by id field - check both root level and nested "model.id" 
	// (MongoDB may store embedded structs as nested objects)
	filter := bson.M{
		"$or": []bson.M{
			{"id": taskID},
			{"model.id": taskID},
			{"_id": taskID},
		},
	}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	// Decode into a map first for better compatibility with MongoDB document structure
	var rawDoc bson.M
	if err := result.Decode(&rawDoc); err != nil {
		return nil, err
	}

	// Convert map to BSON bytes, then unmarshal to struct
	// This approach handles type conversions and missing fields better
	bsonBytes, err := bson.Marshal(rawDoc)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := bson.Unmarshal(bsonBytes, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// GetAll gets all tasks for a project
func (r *TaskRepo) GetAll(ctx context.Context, projectID string) ([]models.Task, error) {
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

// GetAllByProjectIDs gets all tasks for multiple projects (batch operation)
func (r *TaskRepo) GetAllByProjectIDs(ctx context.Context, projectIDs []string) (map[string][]models.Task, error) {
	if len(projectIDs) == 0 {
		return make(map[string][]models.Task), nil
	}

	filter := bson.M{"projectId": bson.M{"$in": projectIDs}}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	// Group tasks by project ID
	tasksByProject := make(map[string][]models.Task)
	for _, item := range items {
		var task models.Task
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &task); err != nil {
			continue
		}
		tasksByProject[task.ProjectID] = append(tasksByProject[task.ProjectID], task)
	}

	// Ensure all project IDs are in the map (even if empty)
	for _, projectID := range projectIDs {
		if _, exists := tasksByProject[projectID]; !exists {
			tasksByProject[projectID] = []models.Task{}
		}
	}

	return tasksByProject, nil
}

// Add creates a new task
func (r *TaskRepo) Add(ctx context.Context, task *models.Task) error {
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

	return r.InsertOne(ctx, task)
}

// Update updates a task
func (r *TaskRepo) Update(ctx context.Context, task *models.Task) error {
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

	if task.StartDate != nil {
		updates["startDate"] = *task.StartDate
	}
	if task.EndDate != nil {
		updates["endDate"] = *task.EndDate
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
func (r *TaskRepo) Delete(ctx context.Context, projectID, taskID string) error {
	return r.DeleteByID(ctx, taskID)
}

// UpdateDeadline updates task deadline
func (r *TaskRepo) UpdateDeadline(ctx context.Context, projectID, taskID string, deadline time.Time) error {
	updates := bson.M{"deadline": deadline}
	return r.UpdateOne(ctx, taskID, updates)
}

// UpdateProgress updates task progress
func (r *TaskRepo) UpdateProgress(ctx context.Context, projectID, taskID string, progress int) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	updates := bson.M{"progress": progress}
	return r.UpdateOne(ctx, taskID, updates)
}

// UpdateDescription updates task description
func (r *TaskRepo) UpdateDescription(ctx context.Context, projectID, taskID string, description string) error {
	updates := bson.M{"description": description}
	return r.UpdateOne(ctx, taskID, updates)
}

// UpdateStatus updates task status
func (r *TaskRepo) UpdateStatus(ctx context.Context, projectID, taskID, status string) error {
	updates := bson.M{"status": status}
	return r.UpdateOne(ctx, taskID, updates)
}

// AddTimeSpent adds a time spent entry
func (r *TaskRepo) AddTimeSpent(ctx context.Context, projectID, taskID string, timeSpent models.TimeSpent) error {
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

// UpdateTimeSpent updates a time spent entry
func (r *TaskRepo) UpdateTimeSpent(ctx context.Context, projectID, taskID string, index int, timeSpent models.TimeSpent) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if index < 0 || index >= len(task.TimeSpent) {
		return errors.New("invalid time spent index")
	}

	task.TimeSpent[index] = timeSpent
	updates := bson.M{"timeSpent": task.TimeSpent}
	return r.UpdateOne(ctx, taskID, updates)
}

// RemoveTimeSpent removes a time spent entry
func (r *TaskRepo) RemoveTimeSpent(ctx context.Context, projectID, taskID string, index int) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if index < 0 || index >= len(task.TimeSpent) {
		return errors.New("invalid time spent index")
	}

	task.TimeSpent = append(task.TimeSpent[:index], task.TimeSpent[index+1:]...)
	updates := bson.M{"timeSpent": task.TimeSpent}
	return r.UpdateOne(ctx, taskID, updates)
}

// AddFileAttachment adds a file attachment
func (r *TaskRepo) AddFileAttachment(ctx context.Context, projectID, taskID string, attachment models.FileAttachment) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	task.FileAttachments = append(task.FileAttachments, attachment)
	updates := bson.M{"fileAttachments": task.FileAttachments}
	return r.UpdateOne(ctx, taskID, updates)
}

// RemoveFileAttachment removes a file attachment
func (r *TaskRepo) RemoveFileAttachment(ctx context.Context, projectID, taskID string, index int) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if index < 0 || index >= len(task.FileAttachments) {
		return errors.New("invalid file attachment index")
	}

	task.FileAttachments = append(task.FileAttachments[:index], task.FileAttachments[index+1:]...)
	updates := bson.M{"fileAttachments": task.FileAttachments}
	return r.UpdateOne(ctx, taskID, updates)
}

// GetTimeSpent gets time spent entries
func (r *TaskRepo) GetTimeSpent(ctx context.Context, projectID, taskID string) ([]models.TimeSpent, error) {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	return task.TimeSpent, nil
}

// GetFileAttachments gets file attachments
func (r *TaskRepo) GetFileAttachments(ctx context.Context, projectID, taskID string) ([]models.FileAttachment, error) {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	return task.FileAttachments, nil
}

