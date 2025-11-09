package repos

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

// convertAssignToArrayToString converts assignTo from array format to string format
// This handles migration from old format (array) to new format (string)
func convertAssignToArrayToString(item map[string]types.AttributeValue) {
	if assignToVal, exists := item["assignTo"]; exists {
		// Check if it's an array (old format)
		if listVal, ok := assignToVal.(*types.AttributeValueMemberL); ok {
			// Convert array to string (take first element)
			if len(listVal.Value) > 0 {
				if firstItem, ok := listVal.Value[0].(*types.AttributeValueMemberS); ok && firstItem.Value != "" {
					item["assignTo"] = &types.AttributeValueMemberS{Value: firstItem.Value}
				} else {
					// Remove assignTo if first element is not a valid string
					delete(item, "assignTo")
				}
			} else {
				// Remove assignTo if array is empty
				delete(item, "assignTo")
			}
		}
		// If it's already a string, leave it as is
	}
}

// convertDurationToDeadline converts duration field to deadline field
// This handles migration from old field name (duration) to new field name (deadline)
func convertDurationToDeadline(item map[string]types.AttributeValue) {
	// If deadline already exists, use it
	if _, exists := item["deadline"]; exists {
		return
	}

	// If duration exists but deadline doesn't, migrate it
	if durationVal, exists := item["duration"]; exists {
		item["deadline"] = durationVal
		// Optionally remove the old duration field after migration
		// delete(item, "duration")
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

	// Convert assignTo from array to string if needed (for backward compatibility)
	convertAssignToArrayToString(item)

	// Convert duration to deadline if needed (for backward compatibility)
	convertDurationToDeadline(item)

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
		// Convert assignTo from array to string if needed (for backward compatibility)
		convertAssignToArrayToString(item)

		var task models.Task
		if err := UnmarshalItem(item, &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetAllByProjectIDs gets all tasks for multiple projects in parallel
// This is optimized to fetch tasks for multiple projects concurrently
func (r *TaskRepo) GetAllByProjectIDs(ctx context.Context, projectIDs []string) (map[string][]models.Task, error) {
	if len(projectIDs) == 0 {
		return make(map[string][]models.Task), nil
	}

	type result struct {
		projectID string
		tasks     []models.Task
		err       error
	}

	resultChan := make(chan result, len(projectIDs))
	var wg sync.WaitGroup

	// Fetch tasks for each project in parallel
	for _, projectID := range projectIDs {
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			tasks, err := r.GetAll(ctx, pid)
			resultChan <- result{
				projectID: pid,
				tasks:     tasks,
				err:       err,
			}
		}(projectID)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	tasksByProject := make(map[string][]models.Task, len(projectIDs))
	for res := range resultChan {
		if res.err != nil {
			return nil, fmt.Errorf("failed to fetch tasks for project %s: %w", res.projectID, res.err)
		}
		tasksByProject[res.projectID] = res.tasks
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
	// AssignTo is now a pointer to string, no initialization needed

	data := map[string]interface{}{
		"id":              task.ID,
		"subject":         task.Subject,
		"code":            task.Code,
		"status":          task.Status,
		"deadline":        task.Deadline.Format(time.RFC3339),
		"priority":        task.Priority,
		"assignTo":        task.AssignTo,
		"progress":        task.Progress,
		"projectId":       task.ProjectID,
		"timeSpent":       task.TimeSpent,
		"fileAttachments": task.FileAttachments,
		"activityLogs":    task.ActivityLogs,
		"created":         task.Created.Format(time.RFC3339),
	}

	if task.Description != nil {
		data["description"] = *task.Description
	}

	if task.Progress != nil {
		data["progress"] = *task.Progress
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
		"deadline":        task.Deadline.Format(time.RFC3339),
		"priority":        task.Priority,
		"assignTo":        task.AssignTo,
		"progress":        task.Progress,
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

// UpdateDeadline updates task deadline
func (r *TaskRepo) UpdateDeadline(ctx context.Context, projectID, taskID string, deadline time.Time) error {
	updates := map[string]interface{}{
		"deadline": deadline.Format(time.RFC3339),
	}
	return r.UpdateItem(ctx, taskID, updates)
}

// UpdateProgress updates task progress
func (r *TaskRepo) UpdateProgress(ctx context.Context, projectID, taskID string, progress int) error {
	// Validate progress is between 0 and 100
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	updates := map[string]interface{}{
		"progress": progress,
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
