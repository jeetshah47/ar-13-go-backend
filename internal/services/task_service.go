package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/email"
)

// TaskService handles task business logic
type TaskService struct {
	taskRepo        repos.TaskRepository
	userRepo        repos.UserRepository
	emailClient     EmailClientInterface
	cacheSvc        CacheServiceInterface
	activityLogSvc  ActivityLogServiceInterface
	sseService      SSEServiceInterface
	notificationSvc *NotificationService
}

// NewTaskService creates a new task service with dependency injection
func NewTaskService(
	taskRepo repos.TaskRepository,
	userRepo repos.UserRepository,
	emailClient EmailClientInterface,
	cacheSvc CacheServiceInterface,
	activityLogSvc ActivityLogServiceInterface,
	sseService SSEServiceInterface,
	notificationSvc *NotificationService,
) *TaskService {
	return &TaskService{
		taskRepo:        taskRepo,
		userRepo:        userRepo,
		emailClient:     emailClient,
		cacheSvc:        cacheSvc,
		activityLogSvc:  activityLogSvc,
		sseService:      sseService,
		notificationSvc: notificationSvc,
	}
}

// NewTaskServiceWithDefaults creates a new task service with default implementations
// This is a convenience constructor for backward compatibility
// SSE and Notification services are optional and can be set later via SetSSEService and SetNotificationService
func NewTaskServiceWithDefaults(cfg *config.Config) *TaskService {
	var emailClient EmailClientInterface
	if cfg != nil {
		emailClient = email.NewClient(cfg)
	}
	return NewTaskService(
		repos.NewTaskRepo(),
		repos.NewUserRepo(),
		emailClient,
		NewCacheService(),
		NewActivityLogServiceWithDefaults(),
		nil, // SSE service - can be set later
		nil, // Notification service - can be set later
	)
}

// SetSSEService sets the SSE service for sending real-time notifications
func (s *TaskService) SetSSEService(sseService SSEServiceInterface) {
	s.sseService = sseService
}

// SetNotificationService sets the notification service for storing notifications
func (s *TaskService) SetNotificationService(notificationSvc *NotificationService) {
	s.notificationSvc = notificationSvc
}

// populateActivityLogUsers populates user details for activity logs
// Optimized to use batch operations and caching to reduce database reads
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

	// Batch fetch missing users from MongoDB
	if len(missingUserIDs) > 0 {
		// Use batch get for efficiency
		items, err := s.userRepo.BatchGetItems(ctx, missingUserIDs)
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
			// Process batch results (items is now map[string]*models.User)
			for userID, user := range items {
				if user != nil {
					userCache[userID] = user
					// Cache the user for future requests
					_ = s.cacheSvc.SetUser(ctx, userID, user)
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
func (s *TaskService) createActivityLog(ctx context.Context, taskID string, action models.ActivityLogAction, description *string, remark *string, fields map[string]interface{}) {
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
		Remark:      remark,
		Fields:      fields,
	}

	// Create activity log asynchronously to avoid blocking the main operation
	go func() {
		if err := s.activityLogSvc.Add(context.Background(), activityLog); err != nil {
			log.Printf("Failed to create activity log for task %s: %v", taskID, err)
		}
	}()
}

// normalizeTaskStatus normalizes a task's status to master status value
func (s *TaskService) normalizeTaskStatus(task *models.Task) {
	if task != nil && task.Status != "" {
		if normalized := constants.NormalizeTaskStatus(task.Status); normalized != "" {
			task.Status = normalized
		}
	}
}

// GetAll gets all tasks for a project
func (s *TaskService) GetAll(ctx context.Context, projectID string) ([]models.Task, error) {
	tasks, err := s.taskRepo.GetAll(ctx, projectID)
	if err != nil {
		return nil, err
	}
	// Normalize statuses for all tasks
	for i := range tasks {
		s.normalizeTaskStatus(&tasks[i])
	}
	return tasks, nil
}

// GetAllWithUserDetails gets all tasks for a project with user details in assignTo field
func (s *TaskService) GetAllWithUserDetails(ctx context.Context, projectID string) ([]models.TaskWithUserDetails, error) {
	tasks, err := s.taskRepo.GetAllWithUserDetails(ctx, projectID)
	if err != nil {
		return nil, err
	}
	// Normalize statuses for all tasks
	for i := range tasks {
		s.normalizeTaskStatus(&tasks[i].Task)
	}
	return tasks, nil
}

// GetByID gets a task by ID with populated activity log user details
// Optimized to use caching to reduce database reads
func (s *TaskService) GetByID(ctx context.Context, projectID, taskID string) (*models.Task, error) {
	// Try cache first
	var task models.Task
	err := s.cacheSvc.GetTask(ctx, taskID, &task)
	if err != nil {
		// Not in cache, fetch from database
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

	// Normalize task status
	s.normalizeTaskStatus(&task)

	// Populate user details for activity logs
	if len(task.ActivityLogs) > 0 {
		populatedLogs, err := s.populateActivityLogUsers(ctx, task.ActivityLogs)
		if err == nil {
			task.ActivityLogs = populatedLogs
		}
	}

	return &task, nil
}

// GetByIDWithUserDetails gets a task by ID with user details in assignTo field using aggregation
func (s *TaskService) GetByIDWithUserDetails(ctx context.Context, projectID, taskID string) (*models.TaskWithUserDetails, error) {
	task, err := s.taskRepo.GetByIDWithUserDetails(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}

	// Normalize task status
	s.normalizeTaskStatus(&task.Task)

	// Populate user details for activity logs
	if len(task.ActivityLogs) > 0 {
		populatedLogs, err := s.populateActivityLogUsers(ctx, task.ActivityLogs)
		if err == nil {
			task.ActivityLogs = populatedLogs
		}
	}

	return task, nil
}

// Add creates a new task
func (s *TaskService) Add(ctx context.Context, task *models.Task) error {
	// Normalize status before saving
	s.normalizeTaskStatus(task)
	// Default to pending if status is empty
	if task.Status == "" {
		task.Status = constants.GetTaskStatusString(constants.TaskStatusPending)
	}

	if err := s.taskRepo.Add(ctx, task); err != nil {
		return err
	}

	// Cache the newly created task
	_ = s.cacheSvc.SetTask(ctx, task.ID, task)

	// Create activity log for task creation
	desc := fmt.Sprintf("Task '%s' was created", task.Subject)
	s.createActivityLog(ctx, task.ID, models.ActivityLogActionCreated, &desc, nil, map[string]interface{}{
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
		// Normalize status before saving
		s.normalizeTaskStatus(&tasks[i])
		// Default to pending if status is empty
		if tasks[i].Status == "" {
			tasks[i].Status = constants.GetTaskStatusString(constants.TaskStatusPending)
		}

		if err := s.taskRepo.Add(ctx, &tasks[i]); err != nil {
			return err
		}

		// Create activity log for each task creation
		desc := fmt.Sprintf("Task '%s' was created", tasks[i].Subject)
		s.createActivityLog(ctx, tasks[i].ID, models.ActivityLogActionCreated, &desc, nil, map[string]interface{}{
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
	// Normalize status before saving
	if task.Status != "" {
		s.normalizeTaskStatus(task)
	}

	// Check if task exists (try cache first)
	var existing models.Task
	err := s.cacheSvc.GetTask(ctx, task.ID, &existing)
	if err != nil {
		// Not in cache, fetch from database
		fetchedTask, err := s.taskRepo.GetByID(ctx, task.ProjectID, task.ID)
		if err != nil {
			return err
		}
		if fetchedTask == nil {
			return errors.New("task not found")
		}
		existing = *fetchedTask
		// Normalize existing status for comparison
		s.normalizeTaskStatus(&existing)
	}

	// Validate that the projectId matches the existing task's projectId
	// This prevents accidentally moving tasks between projects
	if existing.ProjectID != task.ProjectID {
		return errors.New("cannot change task projectId - task belongs to a different project")
	}

	// Track changes for activity log
	fields := make(map[string]interface{})
	hasChanges := false
	if existing.Subject != task.Subject {
		fields["subject"] = map[string]interface{}{"old": existing.Subject, "new": task.Subject}
		hasChanges = true
	}
	if existing.Status != task.Status {
		fields["status"] = map[string]interface{}{"old": existing.Status, "new": task.Status}
		hasChanges = true
	}
	if existing.Priority != task.Priority {
		fields["priority"] = map[string]interface{}{"old": existing.Priority, "new": task.Priority}
		hasChanges = true
	}
	if existing.Deadline.Format(time.RFC3339) != task.Deadline.Format(time.RFC3339) {
		fields["deadline"] = map[string]interface{}{"old": existing.Deadline.Format(time.RFC3339), "new": task.Deadline.Format(time.RFC3339)}
		hasChanges = true
	}
	if existing.Progress != nil && task.Progress != nil && *existing.Progress != *task.Progress {
		fields["progress"] = map[string]interface{}{"old": *existing.Progress, "new": *task.Progress}
		hasChanges = true
	}

	// Check description changes
	existingDesc := ""
	if existing.Description != nil {
		existingDesc = *existing.Description
	}
	newDesc := ""
	if task.Description != nil {
		newDesc = *task.Description
	}
	if existingDesc != newDesc {
		fields["description"] = map[string]interface{}{"old": existingDesc, "new": newDesc}
		hasChanges = true
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// Invalidate task cache since it was updated
	_ = s.cacheSvc.InvalidateTask(ctx, task.ID)

	// Create activity log for task update
	if hasChanges {
		desc := fmt.Sprintf("Task '%s' was updated", task.Subject)
		s.createActivityLog(ctx, task.ID, models.ActivityLogActionUpdated, &desc, nil, fields)
	}

	// Send notification to assigned member when task is updated via SSE and store in database (non-blocking)
	if s.notificationSvc != nil && hasChanges && task.AssignTo != nil && *task.AssignTo != "" {
		go func() {
			// Get project details
			projectRepo := repos.NewProjectRepo()
			project, err := projectRepo.GetByID(context.Background(), task.ProjectID)
			if err != nil {
				log.Printf("Failed to get project for task update notification: %v", err)
			}

			projectTitle := "the project"
			if project != nil {
				projectTitle = project.Title
			}

			// Get updater info
			updaterID := s.getUserIDFromContext(ctx)
			updaterName := "Someone"
			if updaterID != "" {
				updaterUser, err := s.userRepo.GetByID(context.Background(), updaterID)
				if err == nil && updaterUser != nil {
					updaterName = updaterUser.Name
				}
			}

			// Build change summary
			changeSummary := "The following changes were made:\n"
			if subjectChange, ok := fields["subject"].(map[string]interface{}); ok {
				changeSummary += fmt.Sprintf("- Subject: '%s' → '%s'\n", subjectChange["old"], subjectChange["new"])
			}
			if statusChange, ok := fields["status"].(map[string]interface{}); ok {
				changeSummary += fmt.Sprintf("- Status: '%s' → '%s'\n", statusChange["old"], statusChange["new"])
			}
			if priorityChange, ok := fields["priority"].(map[string]interface{}); ok {
				changeSummary += fmt.Sprintf("- Priority: '%s' → '%s'\n", priorityChange["old"], priorityChange["new"])
			}
			if deadlineChange, ok := fields["deadline"].(map[string]interface{}); ok {
				changeSummary += fmt.Sprintf("- Deadline: '%s' → '%s'\n", deadlineChange["old"], deadlineChange["new"])
			}
			if progressChange, ok := fields["progress"].(map[string]interface{}); ok {
				changeSummary += fmt.Sprintf("- Progress: %v%% → %v%%\n", progressChange["old"], progressChange["new"])
			}
			if descChange, ok := fields["description"].(map[string]interface{}); ok {
				oldDesc := descChange["old"].(string)
				newDesc := descChange["new"].(string)
				if oldDesc == "" {
					changeSummary += "- Description: Added\n"
				} else if newDesc == "" {
					changeSummary += "- Description: Removed\n"
				} else {
					changeSummary += "- Description: Updated\n"
				}
			}

			message := fmt.Sprintf("%s has updated task '%s' in project '%s'.\n\n%s", updaterName, task.Subject, projectTitle, changeSummary)

			notification := &models.Notification{
				Title:             fmt.Sprintf("Task Updated: %s", task.Subject),
				Message:           message,
				Type:              models.NotificationTypeTaskUpdated,
				UserID:            *task.AssignTo,
				RelatedEntityID:   task.ID,
				RelatedEntityType: models.RelatedEntityTypeTask,
				IsRead:            false,
			}

			// Store notification in database
			if err := s.notificationSvc.CreateNotification(context.Background(), notification); err != nil {
				log.Printf("[SSE-NOTIFICATION] ERROR: Failed to create task update notification: %v", err)
			} else {
				log.Printf("[SSE-NOTIFICATION] Created task update notification in database - notificationId: %s, taskId: %s, userId: %s", notification.ID, task.ID, *task.AssignTo)
			}

			// Send simple notifications-available event via SSE (client will fetch notifications via API)
			if s.sseService != nil {
				log.Printf("[SSE-NOTIFICATION] Preparing to send notifications-available event via SSE - taskId: %s, userId: %s", task.ID, *task.AssignTo)
				sseData := map[string]interface{}{
					"userId": *task.AssignTo,
				}
				if err := s.sseService.SendToUser(*task.AssignTo, "notifications-available", sseData); err != nil {
					log.Printf("[SSE-NOTIFICATION] ERROR: Failed to send notifications-available event via SSE to user %s: %v", *task.AssignTo, err)
				} else {
					log.Printf("[SSE-NOTIFICATION] SUCCESS: Sent notifications-available event via SSE - taskId: %s, userId: %s, type: notifications-available", task.ID, *task.AssignTo)
				}
			} else {
				log.Printf("[SSE-NOTIFICATION] WARNING: SSE service not available, skipping notifications-available event - taskId: %s, userId: %s", task.ID, *task.AssignTo)
			}
		}()
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
	s.createActivityLog(ctx, taskID, models.ActivityLogActionDeleted, &desc, nil, map[string]interface{}{
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
	s.createActivityLog(ctx, taskID, models.ActivityLogActionDeadlineUpdated, &desc, nil, map[string]interface{}{
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
	s.createActivityLog(ctx, taskID, models.ActivityLogActionDescriptionUpdated, &logDesc, nil, fields)

	return nil
}

// UpdateStatus updates task status
func (s *TaskService) UpdateStatus(ctx context.Context, projectID, taskID, status string, remark *string) error {
	// Normalize the incoming status (should already be normalized from handler, but ensure it)
	normalizedStatus := constants.NormalizeTaskStatus(status)
	if normalizedStatus == "" {
		return errors.New("invalid task status")
	}
	status = normalizedStatus

	// Check if task exists
	existing, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("task not found")
	}

	// Normalize existing status for comparison
	s.normalizeTaskStatus(existing)

	// Only update if status is actually changing
	if existing.Status == status {
		return nil // No change, no need to update or log
	}

	if err := s.taskRepo.UpdateStatus(ctx, projectID, taskID, status); err != nil {
		return err
	}

	// Create activity log for status change
	desc := fmt.Sprintf("Status changed from \"%s\" to \"%s\"", existing.Status, status)
	s.createActivityLog(ctx, taskID, models.ActivityLogActionStatusChanged, &desc, remark, map[string]interface{}{
		"oldStatus": existing.Status,
		"newStatus": status,
	})

	// Note: Notifications for task status updates are now handled in the handler
	// to ensure all project members receive notifications and avoid duplicates

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
	s.createActivityLog(ctx, taskID, models.ActivityLogActionProgressUpdated, &desc, nil, map[string]interface{}{
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
	s.createActivityLog(ctx, taskID, models.ActivityLogActionTimeSpentAdded, &desc, nil, fields)

	// Send notification to project owner when member adds time log via SSE and store in database (non-blocking)
	if s.notificationSvc != nil {
		go func() {
			// Get project details
			projectRepo := repos.NewProjectRepo()
			project, err := projectRepo.GetByID(context.Background(), projectID)
			if err != nil || project == nil {
				log.Printf("Failed to get project for time log notification: %v", err)
				return
			}

			// Get member who added the time log
			memberUser, err := s.userRepo.GetByID(context.Background(), timeSpent.UserID)
			if err != nil || memberUser == nil {
				log.Printf("Failed to get member user for time log notification: %v", err)
				return
			}

			// Notify project owner (skip if owner is the one who added the time log)
			if project.OwnerID != "" && project.OwnerID != timeSpent.UserID {
				timeDescription := ""
				if timeSpent.Description != nil && *timeSpent.Description != "" {
					timeDescription = fmt.Sprintf("\nDescription: %s", *timeSpent.Description)
				}

				// Parse date string and format it nicely
				dateStr := timeSpent.Date
				parsedDate, err := time.Parse("2006-01-02", timeSpent.Date)
				if err == nil {
					dateStr = parsedDate.Format("January 2, 2006")
				}

				// Convert minutes to hours
				hours := float64(timeSpent.TimeSpent) / 60.0

				message := fmt.Sprintf("Member %s has logged %.2f hours for task '%s' in project '%s' on %s.%s",
					memberUser.Name,
					hours,
					task.Subject,
					project.Title,
					dateStr,
					timeDescription)

				notification := &models.Notification{
					Title:             fmt.Sprintf("Time Logged: %s", task.Subject),
					Message:           message,
					Type:              models.NotificationTypeTaskUpdated, // Using TaskUpdated as there's no specific type for time log
					UserID:            project.OwnerID,
					RelatedEntityID:   taskID,
					RelatedEntityType: models.RelatedEntityTypeTask,
					IsRead:            false,
				}

				// Store notification in database
				if err := s.notificationSvc.CreateNotification(context.Background(), notification); err != nil {
					log.Printf("[SSE-NOTIFICATION] ERROR: Failed to create time log notification: %v", err)
				} else {
					log.Printf("[SSE-NOTIFICATION] Created time log notification in database - notificationId: %s, taskId: %s, userId: %s, hours: %.2f, date: %s", notification.ID, taskID, project.OwnerID, hours, dateStr)
				}

				// Send simple notifications-available event via SSE (client will fetch notifications via API)
				if s.sseService != nil {
					log.Printf("[SSE-NOTIFICATION] Preparing to send notifications-available event via SSE - taskId: %s, userId: %s", taskID, project.OwnerID)
					sseData := map[string]interface{}{
						"userId": project.OwnerID,
					}
					if err := s.sseService.SendToUser(project.OwnerID, "notifications-available", sseData); err != nil {
						log.Printf("[SSE-NOTIFICATION] ERROR: Failed to send notifications-available event via SSE to project owner %s: %v", project.OwnerID, err)
					} else {
						log.Printf("[SSE-NOTIFICATION] SUCCESS: Sent notifications-available event via SSE - taskId: %s, userId: %s, type: notifications-available", taskID, project.OwnerID)
					}
				} else {
					log.Printf("[SSE-NOTIFICATION] WARNING: SSE service not available, skipping notifications-available event - taskId: %s, userId: %s", taskID, project.OwnerID)
				}
			}
		}()
	}

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
	s.createActivityLog(ctx, taskID, models.ActivityLogActionTimeSpentUpdated, &desc, nil, fields)

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
	s.createActivityLog(ctx, taskID, models.ActivityLogActionTimeSpentRemoved, &desc, nil, fields)

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
	s.createActivityLog(ctx, taskID, models.ActivityLogActionFileUploaded, &desc, nil, map[string]interface{}{
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
	s.createActivityLog(ctx, taskID, models.ActivityLogActionFileRemoved, &desc, nil, fields)

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
		s.createActivityLog(ctx, taskID, models.ActivityLogActionAssigned, &desc, nil, map[string]interface{}{
			"assignedTo": userID,
		})
	} else if oldAssignTo != userID {
		desc := fmt.Sprintf("Task '%s' was reassigned from user '%s' to user '%s'", task.Subject, oldAssignTo, userID)
		s.createActivityLog(ctx, taskID, models.ActivityLogActionAssigned, &desc, nil, map[string]interface{}{
			"oldAssignedTo": oldAssignTo,
			"newAssignedTo": userID,
		})
	}

	// Send notification to assigned user via SSE and store in database (non-blocking)
	if s.notificationSvc != nil {
		go func() {
			// Get project details
			projectRepo := repos.NewProjectRepo()
			project, err := projectRepo.GetByID(context.Background(), projectID)
			if err != nil {
				log.Printf("Failed to get project for task assignment notification: %v", err)
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

			// Store notification in database
			if err := s.notificationSvc.CreateNotification(context.Background(), notification); err != nil {
				log.Printf("[SSE-NOTIFICATION] ERROR: Failed to create task assignment notification: %v", err)
			} else {
				log.Printf("[SSE-NOTIFICATION] Created task assignment notification in database - notificationId: %s, taskId: %s, userId: %s", notification.ID, taskID, userID)
			}

			// Send simple notifications-available event via SSE (client will fetch notifications via API)
			if s.sseService != nil {
				log.Printf("[SSE-NOTIFICATION] Preparing to send notifications-available event via SSE - taskId: %s, userId: %s", taskID, userID)
				sseData := map[string]interface{}{
					"userId": userID,
				}
				if err := s.sseService.SendToUser(userID, "notifications-available", sseData); err != nil {
					log.Printf("[SSE-NOTIFICATION] ERROR: Failed to send notifications-available event via SSE to user %s: %v", userID, err)
				} else {
					log.Printf("[SSE-NOTIFICATION] SUCCESS: Sent notifications-available event via SSE - taskId: %s, userId: %s, type: notifications-available", taskID, userID)
				}
			} else {
				log.Printf("[SSE-NOTIFICATION] WARNING: SSE service not available, skipping notifications-available event - taskId: %s, userId: %s", taskID, userID)
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

// GetStatuses gets all task statuses from the database ordered by order field
func (s *TaskService) GetStatuses(ctx context.Context) ([]map[string]interface{}, error) {
	taskStatusRepo := repos.NewTaskStatusRepo()
	return taskStatusRepo.GetAll(ctx)
}
