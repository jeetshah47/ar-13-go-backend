package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// ProjectService handles project business logic
type ProjectService struct {
	projectRepo     repos.ProjectRepository
	taskRepo        repos.TaskRepository
	cacheSvc        CacheServiceInterface
	notificationSvc *NotificationService
	websocketService WebSocketServiceInterface
}

// NewProjectService creates a new project service with dependency injection
func NewProjectService(
	projectRepo repos.ProjectRepository,
	taskRepo repos.TaskRepository,
	cacheSvc CacheServiceInterface,
) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
		cacheSvc:    cacheSvc,
	}
}

// NewProjectServiceWithDefaults creates a new project service with default dependencies
func NewProjectServiceWithDefaults() *ProjectService {
	return NewProjectService(
		repos.NewProjectRepo(),
		repos.NewTaskRepo(),
		NewCacheService(),
	)
}

// SetNotificationService sets the notification service for storing notifications
func (s *ProjectService) SetNotificationService(notificationSvc *NotificationService) {
	s.notificationSvc = notificationSvc
}

// SetWebSocketService sets the WebSocket service for sending real-time notifications
func (s *ProjectService) SetWebSocketService(websocketService WebSocketServiceInterface) {
	s.websocketService = websocketService
}

// GetAll gets all projects
func (s *ProjectService) GetAll(ctx context.Context, limit *int) ([]models.Project, error) {
	return s.projectRepo.GetAll(ctx, limit)
}

// GetByID gets a project by ID
func (s *ProjectService) GetByID(ctx context.Context, id string) (*models.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

// GetByUserID gets all projects where the user is either the owner or a member
func (s *ProjectService) GetByUserID(ctx context.Context, userID string) ([]models.Project, error) {
	return s.projectRepo.GetByUserID(ctx, userID)
}

// GetTasksByProjectIDs gets all tasks for multiple projects in a single batch query
// This method is exposed to avoid N+1 queries when fetching tasks for multiple projects
func (s *ProjectService) GetTasksByProjectIDs(ctx context.Context, projectIDs []string) (map[string][]models.Task, error) {
	if s.taskRepo == nil {
		return make(map[string][]models.Task), nil
	}
	return s.taskRepo.GetAllByProjectIDs(ctx, projectIDs)
}

// Add creates a new project
func (s *ProjectService) Add(ctx context.Context, project *models.Project) error {
	if err := s.projectRepo.Add(ctx, project); err != nil {
		return err
	}
	// Invalidate project stats cache
	_ = s.cacheSvc.InvalidateProjectStats(ctx)

	// Send notification to project owner (non-blocking)
	if s.notificationSvc != nil {
		go func() {
			notification := &models.Notification{
				Title:             fmt.Sprintf("Project Created: %s", project.Title),
				Message:           fmt.Sprintf("Project '%s' has been created successfully.", project.Title),
				Type:              models.NotificationTypeProjectCreated,
				UserID:            project.OwnerID,
				RelatedEntityID:   project.ID,
				RelatedEntityType: models.RelatedEntityTypeProject,
				IsRead:            false,
			}

			if err := s.notificationSvc.CreateNotification(context.Background(), notification); err != nil {
				// Log error but don't fail the operation
				_ = err
			}

			// Send WebSocket notification
			if s.websocketService != nil {
				wsData := map[string]interface{}{
					"userId": project.OwnerID,
				}
				_ = s.websocketService.SendToUser(project.OwnerID, "notifications-available", wsData)
			}
		}()
	}

	return nil
}

// Update updates a project
func (s *ProjectService) Update(ctx context.Context, project *models.Project) error {
	// Check if project exists
	existing, err := s.projectRepo.GetByID(ctx, project.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("project not found")
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return err
	}
	// Invalidate project stats cache
	_ = s.cacheSvc.InvalidateProjectStats(ctx)

	// Send notifications to project owner and members (non-blocking)
	if s.notificationSvc != nil {
		go func() {
			// Notify project owner
			ownerNotification := &models.Notification{
				Title:             fmt.Sprintf("Project Updated: %s", project.Title),
				Message:           fmt.Sprintf("Project '%s' has been updated.", project.Title),
				Type:              models.NotificationTypeProjectUpdated,
				UserID:            project.OwnerID,
				RelatedEntityID:   project.ID,
				RelatedEntityType: models.RelatedEntityTypeProject,
				IsRead:            false,
			}
			if err := s.notificationSvc.CreateNotification(context.Background(), ownerNotification); err != nil {
				log.Printf("Failed to create project update notification for owner: %v", err)
			}
			if s.websocketService != nil {
				wsData := map[string]interface{}{"userId": project.OwnerID}
				_ = s.websocketService.SendToUser(project.OwnerID, "notifications-available", wsData)
			}

			// Notify all project members
			for _, memberID := range project.MembersIDs {
				if memberID != project.OwnerID { // Don't notify owner twice
					memberNotification := &models.Notification{
						Title:             fmt.Sprintf("Project Updated: %s", project.Title),
						Message:           fmt.Sprintf("Project '%s' has been updated.", project.Title),
						Type:              models.NotificationTypeProjectUpdated,
						UserID:            memberID,
						RelatedEntityID:   project.ID,
						RelatedEntityType: models.RelatedEntityTypeProject,
						IsRead:            false,
					}
					if err := s.notificationSvc.CreateNotification(context.Background(), memberNotification); err != nil {
						log.Printf("Failed to create project update notification for member %s: %v", memberID, err)
					}
					if s.websocketService != nil {
						wsData := map[string]interface{}{"userId": memberID}
						_ = s.websocketService.SendToUser(memberID, "notifications-available", wsData)
					}
				}
			}
		}()
	}

	return nil
}

// Delete deletes a project
func (s *ProjectService) Delete(ctx context.Context, id string) error {
	// Check if project exists
	existing, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("project not found")
	}

	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return err
	}
	// Invalidate project stats cache
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	return nil
}

// UpdateAgencyContact updates the agency contact for a project
func (s *ProjectService) UpdateAgencyContact(ctx context.Context, projectID string, agencyContact *models.AgencyContact) error {
	// Check if project exists
	existing, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("project not found")
	}

	if err := s.projectRepo.UpdateAgencyContact(ctx, projectID, agencyContact); err != nil {
		return err
	}
	// Invalidate project stats cache
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	return nil
}

// Archive archives or unarchives a project
func (s *ProjectService) Archive(ctx context.Context, projectID string, isArchived bool) error {
	// Check if project exists
	existing, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("project not found")
	}

	if err := s.projectRepo.Archive(ctx, projectID, isArchived); err != nil {
		return err
	}
	// Invalidate project stats cache
	_ = s.cacheSvc.InvalidateProjectStats(ctx)
	return nil
}

// ProjectTaskStatistics represents task statistics for a project
type ProjectTaskStatistics struct {
	TotalTasks      int            `json:"totalTasks"`
	CompletedTasks  int            `json:"completedTasks"`
	ActiveTasks     int            `json:"activeTasks"`
	BacklogTasks    int            `json:"backlogTasks"`
	TasksInProgress int            `json:"tasksInProgress"`
	TasksInReview   int            `json:"tasksInReview"`
	PendingTasks    int            `json:"pendingTasks"`
	CancelledTasks  int            `json:"cancelledTasks"`
	ByStatus              map[string]int `json:"byStatus"`              // Status -> count
	ByPriority            map[string]int `json:"byPriority"`            // Priority -> count
	CompletionRate        float64        `json:"completionRate"`         // Percentage
	TotalTimeSpent        int            `json:"totalTimeSpent"`         // Total time spent in minutes
	AssignedUsers         int            `json:"assignedUsers"`          // Number of unique users assigned
	TasksByAssignee       map[string]int `json:"tasksByAssignee"`        // UserID -> task count
	CompletedTasksByAssignee map[string]int `json:"completedTasksByAssignee"` // UserID -> completed task count
}

// ProjectWithStatistics represents a project with its task statistics
type ProjectWithStatistics struct {
	models.Project
	Statistics ProjectTaskStatistics `json:"statistics"`
}

// GetAllWithStatistics gets all projects with their task statistics
// Optimized to batch fetch all tasks at once instead of N+1 queries
// Uses Redis cache to improve performance
func (s *ProjectService) GetAllWithStatistics(ctx context.Context, limit *int) ([]ProjectWithStatistics, error) {
	// Try to get from cache first (only if no limit specified, as cache key doesn't include limit)
	if limit == nil {
		cached, err := s.cacheSvc.GetProjectStats(ctx)
		if err == nil && cached != nil {
			// Convert cached interface{} slice to ProjectWithStatistics slice
			projects := make([]ProjectWithStatistics, 0, len(cached))
			for _, item := range cached {
				if projectMap, ok := item.(map[string]interface{}); ok {
					var project ProjectWithStatistics
					if data, err := json.Marshal(projectMap); err == nil {
						if err := json.Unmarshal(data, &project); err == nil {
							projects = append(projects, project)
						}
					}
				}
			}
			if len(projects) > 0 {
				return projects, nil
			}
		}
	}

	// Get all projects
	projects, err := s.projectRepo.GetAll(ctx, limit)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return []ProjectWithStatistics{}, nil
	}

	// Extract project IDs for batch fetching
	projectIDs := make([]string, 0, len(projects))
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}

	// Batch fetch all tasks for all projects in parallel
	tasksByProject, err := s.taskRepo.GetAllByProjectIDs(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	// Calculate statistics for each project using pre-fetched tasks
	result := make([]ProjectWithStatistics, 0, len(projects))
	for _, project := range projects {
		tasks := tasksByProject[project.ID]
		statistics := s.calculateProjectStatisticsFromTasks(tasks)

		result = append(result, ProjectWithStatistics{
			Project:    project,
			Statistics: statistics,
		})
	}

	// Cache the result if no limit specified (ignore cache errors)
	if limit == nil {
		cacheData := make([]interface{}, len(result))
		for i := range result {
			cacheData[i] = result[i]
		}
		_ = s.cacheSvc.SetProjectStats(ctx, cacheData)
	}

	return result, nil
}

// GetStatistics calculates task statistics for a project
// This method fetches tasks from DB and calculates statistics
func (s *ProjectService) GetStatistics(ctx context.Context, projectID string) (ProjectTaskStatistics, error) {
	// Get all tasks for the project
	tasks, err := s.taskRepo.GetAll(ctx, projectID)
	if err != nil {
		return ProjectTaskStatistics{
			ByStatus:              make(map[string]int),
			ByPriority:            make(map[string]int),
			TasksByAssignee:       make(map[string]int),
			CompletedTasksByAssignee: make(map[string]int),
		}, err
	}

	return s.calculateProjectStatisticsFromTasks(tasks), nil
}

// calculateProjectStatisticsFromTasks calculates task statistics from a list of tasks
// This is the optimized version that doesn't require a DB call
func (s *ProjectService) calculateProjectStatisticsFromTasks(tasks []models.Task) ProjectTaskStatistics {
	stats := ProjectTaskStatistics{
		ByStatus:              make(map[string]int),
		ByPriority:            make(map[string]int),
		TasksByAssignee:       make(map[string]int),
		CompletedTasksByAssignee: make(map[string]int),
	}

	// Track unique assigned users
	assignedUserSet := make(map[string]bool)

	// Process each task
	for _, task := range tasks {
		stats.TotalTasks++

		// Normalize status to master status value
		normalizedStatus := constants.NormalizeTaskStatus(task.Status)
		priority := strings.ToLower(strings.TrimSpace(task.Priority))

		// Count by master status (use normalized status in ByStatus map)
		if normalizedStatus != "" {
			stats.ByStatus[normalizedStatus]++
		}
		switch normalizedStatus {
		case constants.GetTaskStatusString(constants.TaskStatusPending):
			stats.PendingTasks++
			stats.BacklogTasks++ // Keep backward compatibility
		case constants.GetTaskStatusString(constants.TaskStatusInProgress):
			stats.TasksInProgress++
		case constants.GetTaskStatusString(constants.TaskStatusInReview):
			stats.TasksInReview++
		case constants.GetTaskStatusString(constants.TaskStatusCompleted):
			stats.CompletedTasks++
		case constants.GetTaskStatusString(constants.TaskStatusRejected):
			stats.CancelledTasks++ // Map rejected to cancelled for backward compatibility
		}

		// Count active tasks (not completed or rejected)
		if normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusCompleted) &&
			normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusRejected) {
			stats.ActiveTasks++
		}

		// Count by priority
		if priority != "" {
			stats.ByPriority[priority]++
		}

		// Track assigned user and count tasks per assignee
		if task.AssignTo != nil && *task.AssignTo != "" {
			assignID := *task.AssignTo
			if !assignedUserSet[assignID] {
				assignedUserSet[assignID] = true
			}
			stats.TasksByAssignee[assignID]++
			
			// Track completed tasks per assignee
			if normalizedStatus == constants.GetTaskStatusString(constants.TaskStatusCompleted) {
				stats.CompletedTasksByAssignee[assignID]++
			}
		}

		// Calculate total time spent
		for _, timeSpent := range task.TimeSpent {
			stats.TotalTimeSpent += timeSpent.TimeSpent
		}
	}

	// Set assigned users count
	stats.AssignedUsers = len(assignedUserSet)

	// Calculate completion rate
	if stats.TotalTasks > 0 {
		stats.CompletionRate = (float64(stats.CompletedTasks) / float64(stats.TotalTasks)) * 100
	}

	return stats
}
