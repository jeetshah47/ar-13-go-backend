package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/email"
)

// TaskService handles task business logic
type TaskService struct {
	taskRepo       *repos.TaskRepo
	userRepo       *repos.UserRepo
	emailClient    *email.Client
	cacheSvc       *CacheService
	activityLogSvc *ActivityLogService
}

// NewTaskService creates a new task service
func NewTaskService(cfg *config.Config) *TaskService {
	var emailClient *email.Client
	if cfg != nil {
		emailClient = email.NewClient(cfg)
	}
	return &TaskService{
		taskRepo:       repos.NewTaskRepo(),
		userRepo:       repos.NewUserRepo(),
		emailClient:    emailClient,
		cacheSvc:       NewCacheService(),
		activityLogSvc: NewActivityLogService(),
	}
}

// populateActivityLogUsers populates user details for activity logs
// Optimized to use batch operations and caching to reduce DynamoDB reads
func (s *TaskService) populateActivityLogUsers(ctx context.Context, activityLogs []models.ActivityLog) ([]models.ActivityLog, error) {
	if len(activityLogs) == 0 {
		return activityLogs, nil
	}

	// Collect unique user IDs
	userIDSet := make(map[string]bool)
	for _, log := range activityLogs {
		if log.UserID != "" {
			userIDSet[log.UserID] = true
		}
	}

	if len(userIDSet) == 0 {
		return activityLogs, nil
	}

	// Convert set to slice
	userIDs := make([]string, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}

	// Try to fetch users from cache first, then batch fetch missing ones
	userCache := make(map[string]*models.User)
	missingUserIDs := make([]string, 0)

	for _, userID := range userIDs {
		var user models.User
		err := s.cacheSvc.GetUser(ctx, userID, &user)
		if err == nil {
			// Found in cache
			userCache[userID] = &user
		} else {
			// Not in cache, need to fetch
			missingUserIDs = append(missingUserIDs, userID)
		}
	}

	// Batch fetch missing users from DynamoDB
	if len(missingUserIDs) > 0 {
		// Use batch get for efficiency
		userRepo := repos.NewUserRepo()
		items, err := userRepo.BatchGetItems(ctx, missingUserIDs)
		if err != nil {
			// Fallback to individual gets if batch fails
			for _, userID := range missingUserIDs {
				user, err := s.userRepo.GetByID(ctx, userID)
				if err == nil && user != nil {
					userCache[userID] = user
					// Cache the user for future requests
					_ = s.cacheSvc.SetUser(ctx, userID, user)
				}
			}
		} else {
			// Process batch results
			for userID, item := range items {
				var user models.User
				if err := repos.UnmarshalItem(item, &user); err == nil {
					userCache[userID] = &user
					// Cache the user for future requests
					_ = s.cacheSvc.SetUser(ctx, userID, &user)
				}
			}
		}
	}

	// Populate activity logs with user data
	result := make([]models.ActivityLog, 0, len(activityLogs))
	for _, log := range activityLogs {
		if log.UserID != "" {
			if user, exists := userCache[log.UserID]; exists && user != nil {
				// Populate UserName if not already set
				if log.UserName == nil {
					log.UserName = &user.Name
				}
				// Set full user object
				log.User = user
			}
		}
		result = append(result, log)
	}

	return result, nil
}

// getUserIDFromContext extracts user ID from context
func (s *TaskService) getUserIDFromContext(ctx context.Context) string {
	// Get from context using the middleware key
	if userID, ok := ctx.Value(middleware.UserIDKey).(string); ok {
		return userID
	}
	// Fallback: try string key directly
	if userID, ok := ctx.Value("userId").(string); ok {
		return userID
	}
	return ""
}

// createActivityLog creates an activity log entry for a task operation
func (s *TaskService) createActivityLog(ctx context.Context, taskID string, action models.ActivityLogAction, description *string, fields map[string]interface{}) {
	userID := s.getUserIDFromContext(ctx)
	if userID == "" {
		// Skip activity log if no user ID in context
		return
	}

	activityLog := &models.ActivityLogBase{
		EntityType:  models.ActivityLogEntityTypeTask,
		EntityID:    taskID,
		Action:      action,
		CreatedBy:   userID,
		Description: description,
		Fields:      fields,
	}

	// Create activity log asynchronously to avoid blocking the main operation
	go func() {
		if err := s.activityLogSvc.Add(context.Background(), activityLog); err != nil {
			log.Printf("Failed to create activity log for task %s: %v", taskID, err)
		}
	}()
}

// GetAll gets all tasks for a project
func (s *TaskService) GetAll(ctx context.Context, projectID string) ([]models.Task, error) {
	return s.taskRepo.GetAll(ctx, projectID)
}

// GetByID gets a task by ID with populated activity log user details
// Optimized to use caching to reduce DynamoDB reads
func (s *TaskService) GetByID(ctx context.Context, projectID, taskID string) (*models.Task, error) {
	// Try cache first
	var task models.Task
	err := s.cacheSvc.GetTask(ctx, taskID, &task)
	if err != nil {
		// Not in cache, fetch from DynamoDB
		fetchedTask, err := s.taskRepo.GetByID(ctx, projectID, taskID)
		if err != nil {
			return nil, err
		}
		if fetchedTask == nil {
			return nil, nil
		}
		task = *fetchedTask
		// Cache the task for future requests
		_ = s.cacheSvc.SetTask(ctx, taskID, &task)
	}

	// Populate user details for activity logs
	if len(task.ActivityLogs) > 0 {
		populatedLogs, err := s.populateActivityLogUsers(ctx, task.ActivityLogs)
		if err == nil {
			task.ActivityLogs = populatedLogs
		}
	}

	return &task, nil
}

// Add creates a new task
func (s *TaskService) Add(ctx context.Context, task *models.Task) error {
	if err := s.taskRepo.Add(ctx, task); err != nil {
		return err
	}

	// Cache the newly created task
	_ = s.cacheSvc.SetTask(ctx, task.ID, task)

	// Create activity log for task creation
	desc := fmt.Sprintf("Task '%s' was created", task.Subject)
	s.createActivityLog(ctx, task.ID, models.ActivityLogActionCreated, &desc, map[string]interface{}{
		"subject":   task.Subject,
		"status":    task.Status,
		"priority":  task.Priority,
		"projectId": task.ProjectID,
	})

	// Invalidate caches that depend on tasks
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	_ = s.cacheSvc.InvalidateDashboardStats(ctx)
	return nil
}

// AddMultiple creates multiple tasks
func (s *TaskService) AddMultiple(ctx context.Context, tasks []models.Task) error {
	for i := range tasks {
		if err := s.taskRepo.Add(ctx, &tasks[i]); err != nil {
			return err
		}

		// Create activity log for each task creation
		desc := fmt.Sprintf("Task '%s' was created", tasks[i].Subject)
		s.createActivityLog(ctx, tasks[i].ID, models.ActivityLogActionCreated, &desc, map[string]interface{}{
			"subject":   tasks[i].Subject,
			"status":    tasks[i].Status,
			"priority":  tasks[i].Priority,
			"projectId": tasks[i].ProjectID,
		})
	}
	// Invalidate caches that depend on tasks
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	_ = s.cacheSvc.InvalidateDashboardStats(ctx)
	return nil
}

// Update updates a task
func (s *TaskService) Update(ctx context.Context, task *models.Task) error {
	// Check if task exists (try cache first)
	var existing models.Task
	err := s.cacheSvc.GetTask(ctx, task.ID, &existing)
	if err != nil {
		// Not in cache, fetch from DynamoDB
		fetchedTask, err := s.taskRepo.GetByID(ctx, task.ProjectID, task.ID)
		if err != nil {
			return err
		}
		if fetchedTask == nil {
			return errors.New("task not found")
		}
		existing = *fetchedTask
	}

	// Track changes for activity log
	fields := make(map[string]interface{})
	if existing.Subject != task.Subject {
		fields["subject"] = map[string]interface{}{"old": existing.Subject, "new": task.Subject}
	}
	if existing.Status != task.Status {
		fields["status"] = map[string]interface{}{"old": existing.Status, "new": task.Status}
	}
	if existing.Priority != task.Priority {
		fields["priority"] = map[string]interface{}{"old": existing.Priority, "new": task.Priority}
	}
	if existing.Deadline.Format(time.RFC3339) != task.Deadline.Format(time.RFC3339) {
		fields["deadline"] = map[string]interface{}{"old": existing.Deadline.Format(time.RFC3339), "new": task.Deadline.Format(time.RFC3339)}
	}
	if existing.Progress != nil && task.Progress != nil && *existing.Progress != *task.Progress {
		fields["progress"] = map[string]interface{}{"old": *existing.Progress, "new": *task.Progress}
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// Invalidate task cache since it was updated
	_ = s.cacheSvc.InvalidateTask(ctx, task.ID)

	// Create activity log for task update
	if len(fields) > 0 {
		desc := fmt.Sprintf("Task '%s' was updated", task.Subject)
		s.createActivityLog(ctx, task.ID, models.ActivityLogActionUpdated, &desc, fields)
	}

	// Invalidate caches that depend on tasks
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	_ = s.cacheSvc.InvalidateDashboardStats(ctx)
	return nil
}

// Delete deletes a task
func (s *TaskService) Delete(ctx context.Context, projectID, taskID string) error {
	// Check if task exists
	existing, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("task not found")
	}

	if err := s.taskRepo.Delete(ctx, projectID, taskID); err != nil {
		return err
	}

	// Invalidate task cache since it was deleted
	_ = s.cacheSvc.InvalidateTask(ctx, taskID)

	// Create activity log for task deletion
	desc := fmt.Sprintf("Task '%s' was deleted", existing.Subject)
	s.createActivityLog(ctx, taskID, models.ActivityLogActionDeleted, &desc, map[string]interface{}{
		"subject": existing.Subject,
	})

	// Invalidate caches that depend on tasks
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	_ = s.cacheSvc.InvalidateDashboardStats(ctx)
	return nil
}

// UpdateDeadline updates task deadline
func (s *TaskService) UpdateDeadline(ctx context.Context, projectID, taskID string, deadline time.Time) error {
	// Check if task exists
	existing, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("task not found")
	}

	if err := s.taskRepo.UpdateDeadline(ctx, projectID, taskID, deadline); err != nil {
		return err
	}

	// Create activity log for deadline update
	desc := fmt.Sprintf("Deadline for task '%s' was updated", existing.Subject)
	s.createActivityLog(ctx, taskID, models.ActivityLogActionDeadlineUpdated, &desc, map[string]interface{}{
		"oldDeadline": existing.Deadline.Format(time.RFC3339),
		"newDeadline": deadline.Format(time.RFC3339),
	})

	return nil
}

// UpdateDescription updates task description
func (s *TaskService) UpdateDescription(ctx context.Context, projectID, taskID, description string) error {
	// Check if task exists
	existing, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("task not found")
	}

	oldDescription := ""
	if existing.Description != nil {
		oldDescription = *existing.Description
	}

	if err := s.taskRepo.UpdateDescription(ctx, projectID, taskID, description); err != nil {
		return err
	}

	// Create activity log for description update
	var logDesc string
	if oldDescription == "" && description != "" {
		logDesc = fmt.Sprintf("Added description for task '%s'", existing.Subject)
	} else if oldDescription != "" && description == "" {
		logDesc = fmt.Sprintf("Removed description for task '%s'", existing.Subject)
	} else {
		logDesc = fmt.Sprintf("Updated description for task '%s'", existing.Subject)
	}

	fields := map[string]interface{}{
		"oldDescription": oldDescription,
		"newDescription": description,
	}
	s.createActivityLog(ctx, taskID, models.ActivityLogActionDescriptionUpdated, &logDesc, fields)

	return nil
}

// UpdateStatus updates task status
func (s *TaskService) UpdateStatus(ctx context.Context, projectID, taskID, status string) error {
	// Check if task exists
	existing, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("task not found")
	}

	// Only update if status is actually changing
	if existing.Status == status {
		return nil // No change, no need to update or log
	}

	if err := s.taskRepo.UpdateStatus(ctx, projectID, taskID, status); err != nil {
		return err
	}

	// Create activity log for status change
	desc := fmt.Sprintf("Status changed from \"%s\" to \"%s\"", existing.Status, status)
	s.createActivityLog(ctx, taskID, models.ActivityLogActionStatusChanged, &desc, map[string]interface{}{
		"oldStatus": existing.Status,
		"newStatus": status,
	})

	// Invalidate caches that depend on task status
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	_ = s.cacheSvc.InvalidateDashboardStats(ctx)
	return nil
}

// UpdateProgress updates task progress
func (s *TaskService) UpdateProgress(ctx context.Context, projectID, taskID string, progress int) error {
	// Check if task exists
	existing, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("task not found")
	}

	// Validate progress is between 0 and 100
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	oldProgress := 0
	if existing.Progress != nil {
		oldProgress = *existing.Progress
	}

	if err := s.taskRepo.UpdateProgress(ctx, projectID, taskID, progress); err != nil {
		return err
	}

	// Create activity log for progress update
	desc := fmt.Sprintf("Progress for task '%s' was updated from %d%% to %d%%", existing.Subject, oldProgress, progress)
	s.createActivityLog(ctx, taskID, models.ActivityLogActionProgressUpdated, &desc, map[string]interface{}{
		"oldProgress": oldProgress,
		"newProgress": progress,
	})

	// Invalidate caches that depend on tasks
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	_ = s.cacheSvc.InvalidateDashboardStats(ctx)
	return nil
}

// AddTimeSpent adds a time spent entry
func (s *TaskService) AddTimeSpent(ctx context.Context, projectID, taskID string, timeSpent models.TimeSpent) error {
	task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if err := s.taskRepo.AddTimeSpent(ctx, projectID, taskID, timeSpent); err != nil {
		return err
	}

	// Create activity log for time spent addition
	desc := fmt.Sprintf("Time log entry added for task '%s'", task.Subject)
	fields := map[string]interface{}{
		"timeSpent": timeSpent.TimeSpent,
		"date":      timeSpent.Date,
		"userId":    timeSpent.UserID,
	}
	if timeSpent.Description != nil {
		fields["description"] = *timeSpent.Description
	}
	s.createActivityLog(ctx, taskID, models.ActivityLogActionTimeSpentAdded, &desc, fields)

	return nil
}

// UpdateTimeSpent updates a time spent entry
func (s *TaskService) UpdateTimeSpent(ctx context.Context, projectID, taskID string, index int, timeSpent models.TimeSpent) error {
	task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// Get existing time spent entry for comparison
	var oldTimeSpent *models.TimeSpent
	if index >= 0 && index < len(task.TimeSpent) {
		oldTimeSpent = &task.TimeSpent[index]
	}

	if err := s.taskRepo.UpdateTimeSpent(ctx, projectID, taskID, index, timeSpent); err != nil {
		return err
	}

	// Create activity log for time spent update
	desc := fmt.Sprintf("Time log entry updated for task '%s'", task.Subject)
	fields := map[string]interface{}{
		"index":     index,
		"timeSpent": timeSpent.TimeSpent,
		"date":      timeSpent.Date,
		"userId":    timeSpent.UserID,
	}
	if oldTimeSpent != nil {
		fields["oldTimeSpent"] = oldTimeSpent.TimeSpent
		fields["oldDate"] = oldTimeSpent.Date
	}
	if timeSpent.Description != nil {
		fields["description"] = *timeSpent.Description
	}
	s.createActivityLog(ctx, taskID, models.ActivityLogActionTimeSpentUpdated, &desc, fields)

	return nil
}

// RemoveTimeSpent removes a time spent entry
func (s *TaskService) RemoveTimeSpent(ctx context.Context, projectID, taskID string, index int) error {
	task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// Get existing time spent entry before removal
	var timeSpent *models.TimeSpent
	if index >= 0 && index < len(task.TimeSpent) {
		timeSpent = &task.TimeSpent[index]
	}

	if err := s.taskRepo.RemoveTimeSpent(ctx, projectID, taskID, index); err != nil {
		return err
	}

	// Create activity log for time spent removal
	desc := fmt.Sprintf("Time log entry removed from task '%s'", task.Subject)
	fields := map[string]interface{}{"index": index}
	if timeSpent != nil {
		fields["timeSpent"] = timeSpent.TimeSpent
		fields["date"] = timeSpent.Date
		fields["userId"] = timeSpent.UserID
	}
	s.createActivityLog(ctx, taskID, models.ActivityLogActionTimeSpentRemoved, &desc, fields)

	return nil
}

// GetTimeSpent gets time spent entries
func (s *TaskService) GetTimeSpent(ctx context.Context, projectID, taskID string) ([]models.TimeSpent, error) {
	return s.taskRepo.GetTimeSpent(ctx, projectID, taskID)
}

// AddFileAttachment adds a file attachment
func (s *TaskService) AddFileAttachment(ctx context.Context, projectID, taskID string, attachment models.FileAttachment) error {
	task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if err := s.taskRepo.AddFileAttachment(ctx, projectID, taskID, attachment); err != nil {
		return err
	}

	// Create activity log for file upload
	desc := fmt.Sprintf("File '%s' was uploaded to task '%s'", attachment.FileName, task.Subject)
	s.createActivityLog(ctx, taskID, models.ActivityLogActionFileUploaded, &desc, map[string]interface{}{
		"fileName":     attachment.FileName,
		"originalName": attachment.OriginalName,
		"fileSize":     attachment.FileSize,
		"mimeType":     attachment.MimeType,
		"uploadedBy":   attachment.UploadedBy,
	})

	return nil
}

// RemoveFileAttachment removes a file attachment
func (s *TaskService) RemoveFileAttachment(ctx context.Context, projectID, taskID string, index int) error {
	task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// Get existing file attachment before removal
	var attachment *models.FileAttachment
	if index >= 0 && index < len(task.FileAttachments) {
		attachment = &task.FileAttachments[index]
	}

	if err := s.taskRepo.RemoveFileAttachment(ctx, projectID, taskID, index); err != nil {
		return err
	}

	// Create activity log for file removal
	desc := fmt.Sprintf("File was removed from task '%s'", task.Subject)
	fields := map[string]interface{}{"index": index}
	if attachment != nil {
		fields["fileName"] = attachment.FileName
		desc = fmt.Sprintf("File '%s' was removed from task '%s'", attachment.FileName, task.Subject)
	}
	s.createActivityLog(ctx, taskID, models.ActivityLogActionFileRemoved, &desc, fields)

	return nil
}

// GetFileAttachments gets file attachments
func (s *TaskService) GetFileAttachments(ctx context.Context, projectID, taskID string) ([]models.FileAttachment, error) {
	return s.taskRepo.GetFileAttachments(ctx, projectID, taskID)
}

// AssignTask assigns a task to a user
func (s *TaskService) AssignTask(ctx context.Context, projectID, taskID, userID string) error {
	task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// Check if user is already assigned
	if task.AssignTo != nil && *task.AssignTo == userID {
		return nil // Already assigned
	}

	// Track old assignment for activity log
	oldAssignTo := ""
	if task.AssignTo != nil {
		oldAssignTo = *task.AssignTo
	}

	// Assign task to user (replacing any existing assignment)
	task.AssignTo = &userID
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// Create activity log for task assignment
	if oldAssignTo == "" {
		desc := fmt.Sprintf("Task '%s' was assigned to user", task.Subject)
		s.createActivityLog(ctx, taskID, models.ActivityLogActionAssigned, &desc, map[string]interface{}{
			"assignedTo": userID,
		})
	} else if oldAssignTo != userID {
		desc := fmt.Sprintf("Task '%s' was reassigned from user '%s' to user '%s'", task.Subject, oldAssignTo, userID)
		s.createActivityLog(ctx, taskID, models.ActivityLogActionAssigned, &desc, map[string]interface{}{
			"oldAssignedTo": oldAssignTo,
			"newAssignedTo": userID,
		})
	}

	// Send email notification to assigned user (non-blocking)
	if s.emailClient != nil {
		go func() {
			user, err := s.userRepo.GetByID(context.Background(), userID)
			if err != nil || user == nil {
				log.Printf("Failed to get user for email notification: %v", err)
				return
			}

			// Get project details for email
			projectRepo := repos.NewProjectRepo()
			project, err := projectRepo.GetByID(context.Background(), projectID)
			if err != nil {
				log.Printf("Failed to get project for email notification: %v", err)
			}

			projectTitle := "the project"
			if project != nil {
				projectTitle = project.Title
			}

			message := fmt.Sprintf("You have been assigned to task '%s' in project '%s'.", task.Subject, projectTitle)
			if task.Description != nil && *task.Description != "" {
				message += fmt.Sprintf("\n\nDescription: %s", *task.Description)
			}

			notification := &models.Notification{
				Title:             fmt.Sprintf("Task Assigned: %s", task.Subject),
				Message:           message,
				Type:              models.NotificationTypeTaskAssigned,
				UserID:            userID,
				RelatedEntityID:   taskID,
				RelatedEntityType: models.RelatedEntityTypeTask,
				IsRead:            false,
			}

			if err := s.emailClient.SendNotificationEmail(notification, user.Email); err != nil {
				log.Printf("Failed to send task assignment email: %v", err)
			}
		}()
	}

	return nil
}

// ClaimTask claims a task
func (s *TaskService) ClaimTask(ctx context.Context, projectID, taskID, userID string) error {
	return s.AssignTask(ctx, projectID, taskID, userID)
}

// GetAssignableUsers gets users that can be assigned to tasks
func (s *TaskService) GetAssignableUsers(ctx context.Context) ([]models.User, error) {
	userRepo := repos.NewUserRepo()
	return userRepo.GetAll(ctx, nil)
}
