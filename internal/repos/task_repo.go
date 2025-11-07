package repos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/google/uuid"
)

// TaskRepo handles task data operations
type TaskRepo struct {
	*DynamoBaseRepo
}

// NewTaskRepo creates a new task repository
func NewTaskRepo() *TaskRepo {
	return &TaskRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("tasks"),
	}
}

// GetByID gets a task by ID
// Note: projectID parameter is kept for API compatibility but not used in DynamoDB query
func (r *TaskRepo) GetByID(ctx context.Context, projectID, taskID string) (*models.Task, error) {
	item, err := r.DynamoBaseRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var task models.Task
	if err := UnmarshalItem(item, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// GetAll gets all tasks for a project
func (r *TaskRepo) GetAll(ctx context.Context, projectID string) ([]models.Task, error) {
	items, err := r.QueryByIndex(ctx, "projectId-index", "projectId", projectID)
	if err != nil {
		return nil, err
	}

	tasks := make([]models.Task, 0, len(items))
	for _, item := range items {
		var task models.Task
		if err := UnmarshalItem(item, &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
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
	if task.AssignTo == nil {
		task.AssignTo = []string{}
	}

	data := map[string]interface{}{
		"id":              task.ID,
		"subject":         task.Subject,
		"code":            task.Code,
		"status":          task.Status,
		"duration":        task.Duration.Format(time.RFC3339),
		"priority":        task.Priority,
		"assignTo":        task.AssignTo,
		"projectId":       task.ProjectID,
		"timeSpent":       task.TimeSpent,
		"fileAttachments": task.FileAttachments,
		"activityLogs":    task.ActivityLogs,
		"created":         task.Created.Format(time.RFC3339),
	}

	if task.Description != nil {
		data["description"] = *task.Description
	}

	return r.PutItem(ctx, data)
}

// Update updates a task
func (r *TaskRepo) Update(ctx context.Context, task *models.Task) error {
	now := time.Now()
	task.Updated = &now

	updates := map[string]interface{}{
		"subject":         task.Subject,
		"code":            task.Code,
		"status":          task.Status,
		"duration":        task.Duration.Format(time.RFC3339),
		"priority":        task.Priority,
		"assignTo":        task.AssignTo,
		"projectId":       task.ProjectID,
		"timeSpent":       task.TimeSpent,
		"fileAttachments": task.FileAttachments,
		"activityLogs":    task.ActivityLogs,
		"updated":         task.Updated.Format(time.RFC3339),
	}

	if task.Description != nil {
		updates["description"] = *task.Description
	}

	return r.UpdateItem(ctx, task.ID, updates)
}

// Delete deletes a task
// Note: projectID parameter is kept for API compatibility but not used in DynamoDB query
func (r *TaskRepo) Delete(ctx context.Context, projectID, taskID string) error {
	return r.DeleteByID(ctx, taskID)
}

// UpdateDuration updates task duration
func (r *TaskRepo) UpdateDuration(ctx context.Context, projectID, taskID string, duration time.Time) error {
	updates := map[string]interface{}{
		"duration": duration.Format(time.RFC3339),
	}
	return r.UpdateItem(ctx, taskID, updates)
}

// UpdateDescription updates task description
func (r *TaskRepo) UpdateDescription(ctx context.Context, projectID, taskID string, description string) error {
	updates := map[string]interface{}{
		"description": description,
	}
	return r.UpdateItem(ctx, taskID, updates)
}

// UpdateStatus updates task status
func (r *TaskRepo) UpdateStatus(ctx context.Context, projectID, taskID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	return r.UpdateItem(ctx, taskID, updates)
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
	updates := map[string]interface{}{
		"timeSpent": task.TimeSpent,
	}
	return r.UpdateItem(ctx, taskID, updates)
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
	updates := map[string]interface{}{
		"timeSpent": task.TimeSpent,
	}
	return r.UpdateItem(ctx, taskID, updates)
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
	updates := map[string]interface{}{
		"timeSpent": task.TimeSpent,
	}
	return r.UpdateItem(ctx, taskID, updates)
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
	updates := map[string]interface{}{
		"fileAttachments": task.FileAttachments,
	}
	return r.UpdateItem(ctx, taskID, updates)
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
	updates := map[string]interface{}{
		"fileAttachments": task.FileAttachments,
	}
	return r.UpdateItem(ctx, taskID, updates)
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
