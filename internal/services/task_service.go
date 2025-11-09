package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/email"
)

// TaskService handles task business logic
type TaskService struct {
	taskRepo    *repos.TaskRepo
	userRepo    *repos.UserRepo
	emailClient *email.Client
	cacheSvc    *CacheService
}

// NewTaskService creates a new task service
func NewTaskService(cfg *config.Config) *TaskService {
	var emailClient *email.Client
	if cfg != nil {
		emailClient = email.NewClient(cfg)
	}
	return &TaskService{
		taskRepo:    repos.NewTaskRepo(),
		userRepo:    repos.NewUserRepo(),
		emailClient: emailClient,
		cacheSvc:    NewCacheService(),
	}
}

// populateActivityLogUsers populates user details for activity logs
func (s *TaskService) populateActivityLogUsers(ctx context.Context, activityLogs []models.ActivityLog) ([]models.ActivityLog, error) {
	if len(activityLogs) == 0 {
		return activityLogs, nil
	}

	userCache := make(map[string]*models.User)
	result := make([]models.ActivityLog, 0, len(activityLogs))

	for _, log := range activityLogs {
		// Populate user details if UserID is present
		if log.UserID != "" {
			user, exists := userCache[log.UserID]
			if !exists {
				fetchedUser, err := s.userRepo.GetByID(ctx, log.UserID)
				if err == nil && fetchedUser != nil {
					user = fetchedUser
					userCache[log.UserID] = user
				}
			}
			if user != nil {
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

// GetAll gets all tasks for a project
func (s *TaskService) GetAll(ctx context.Context, projectID string) ([]models.Task, error) {
	return s.taskRepo.GetAll(ctx, projectID)
}

// GetByID gets a task by ID with populated activity log user details
func (s *TaskService) GetByID(ctx context.Context, projectID, taskID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}

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
	if err := s.taskRepo.Add(ctx, task); err != nil {
		return err
	}
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
	}
	// Invalidate caches that depend on tasks
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	_ = s.cacheSvc.InvalidateDashboardStats(ctx)
	return nil
}

// Update updates a task
func (s *TaskService) Update(ctx context.Context, task *models.Task) error {
	// Check if task exists
	existing, err := s.taskRepo.GetByID(ctx, task.ProjectID, task.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("task not found")
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
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

	return s.taskRepo.UpdateDeadline(ctx, projectID, taskID, deadline)
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

	return s.taskRepo.UpdateDescription(ctx, projectID, taskID, description)
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

	if err := s.taskRepo.UpdateStatus(ctx, projectID, taskID, status); err != nil {
		return err
	}
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

	if err := s.taskRepo.UpdateProgress(ctx, projectID, taskID, progress); err != nil {
		return err
	}
	// Invalidate caches that depend on tasks
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	_ = s.cacheSvc.InvalidateDashboardStats(ctx)
	return nil
}

// AddTimeSpent adds a time spent entry
func (s *TaskService) AddTimeSpent(ctx context.Context, projectID, taskID string, timeSpent models.TimeSpent) error {
	return s.taskRepo.AddTimeSpent(ctx, projectID, taskID, timeSpent)
}

// UpdateTimeSpent updates a time spent entry
func (s *TaskService) UpdateTimeSpent(ctx context.Context, projectID, taskID string, index int, timeSpent models.TimeSpent) error {
	return s.taskRepo.UpdateTimeSpent(ctx, projectID, taskID, index, timeSpent)
}

// RemoveTimeSpent removes a time spent entry
func (s *TaskService) RemoveTimeSpent(ctx context.Context, projectID, taskID string, index int) error {
	return s.taskRepo.RemoveTimeSpent(ctx, projectID, taskID, index)
}

// GetTimeSpent gets time spent entries
func (s *TaskService) GetTimeSpent(ctx context.Context, projectID, taskID string) ([]models.TimeSpent, error) {
	return s.taskRepo.GetTimeSpent(ctx, projectID, taskID)
}

// AddFileAttachment adds a file attachment
func (s *TaskService) AddFileAttachment(ctx context.Context, projectID, taskID string, attachment models.FileAttachment) error {
	return s.taskRepo.AddFileAttachment(ctx, projectID, taskID, attachment)
}

// RemoveFileAttachment removes a file attachment
func (s *TaskService) RemoveFileAttachment(ctx context.Context, projectID, taskID string, index int) error {
	return s.taskRepo.RemoveFileAttachment(ctx, projectID, taskID, index)
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

	// Assign task to user (replacing any existing assignment)
	task.AssignTo = &userID
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
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
