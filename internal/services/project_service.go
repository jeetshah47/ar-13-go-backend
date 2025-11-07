package services

import (
	"context"
	"errors"
	"strings"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// ProjectService handles project business logic
type ProjectService struct {
	projectRepo *repos.ProjectRepo
	taskRepo    *repos.TaskRepo
}

// NewProjectService creates a new project service
func NewProjectService() *ProjectService {
	return &ProjectService{
		projectRepo: repos.NewProjectRepo(),
		taskRepo:    repos.NewTaskRepo(),
	}
}

// GetAll gets all projects
func (s *ProjectService) GetAll(ctx context.Context, limit *int) ([]models.Project, error) {
	return s.projectRepo.GetAll(ctx, limit)
}

// GetByID gets a project by ID
func (s *ProjectService) GetByID(ctx context.Context, id string) (*models.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

// Add creates a new project
func (s *ProjectService) Add(ctx context.Context, project *models.Project) error {
	return s.projectRepo.Add(ctx, project)
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

	return s.projectRepo.Update(ctx, project)
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

	return s.projectRepo.Delete(ctx, id)
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
	ByStatus        map[string]int `json:"byStatus"`        // Status -> count
	ByPriority      map[string]int `json:"byPriority"`      // Priority -> count
	CompletionRate  float64        `json:"completionRate"`  // Percentage
	TotalTimeSpent  int            `json:"totalTimeSpent"`  // Total time spent in minutes
	AssignedUsers   int            `json:"assignedUsers"`   // Number of unique users assigned
	TasksByAssignee map[string]int `json:"tasksByAssignee"` // UserID -> task count
}

// ProjectWithStatistics represents a project with its task statistics
type ProjectWithStatistics struct {
	models.Project
	Statistics ProjectTaskStatistics `json:"statistics"`
}

// GetAllWithStatistics gets all projects with their task statistics
func (s *ProjectService) GetAllWithStatistics(ctx context.Context, limit *int) ([]ProjectWithStatistics, error) {
	// Get all projects
	projects, err := s.projectRepo.GetAll(ctx, limit)
	if err != nil {
		return nil, err
	}

	// Calculate statistics for each project
	result := make([]ProjectWithStatistics, 0, len(projects))
	for _, project := range projects {
		statistics, err := s.calculateProjectStatistics(ctx, project.ID)
		if err != nil {
			// If there's an error calculating statistics, continue with empty statistics
			statistics = ProjectTaskStatistics{
				ByStatus:        make(map[string]int),
				ByPriority:      make(map[string]int),
				TasksByAssignee: make(map[string]int),
			}
		}

		result = append(result, ProjectWithStatistics{
			Project:    project,
			Statistics: statistics,
		})
	}

	return result, nil
}

// calculateProjectStatistics calculates task statistics for a project
func (s *ProjectService) calculateProjectStatistics(ctx context.Context, projectID string) (ProjectTaskStatistics, error) {
	stats := ProjectTaskStatistics{
		ByStatus:        make(map[string]int),
		ByPriority:      make(map[string]int),
		TasksByAssignee: make(map[string]int),
	}

	// Get all tasks for the project
	tasks, err := s.taskRepo.GetAll(ctx, projectID)
	if err != nil {
		return stats, err
	}

	// Track unique assigned users
	assignedUserSet := make(map[string]bool)

	// Process each task
	for _, task := range tasks {
		stats.TotalTasks++

		// Normalize status to lowercase for consistent counting
		status := strings.ToLower(strings.TrimSpace(task.Status))
		priority := strings.ToLower(strings.TrimSpace(task.Priority))

		// Count by status (case-insensitive)
		stats.ByStatus[status]++
		switch status {
		case "backlog", "todo", "to-do":
			stats.BacklogTasks++
		case "inprogress", "in-progress", "in_progress", "in progress":
			stats.TasksInProgress++
		case "inreview", "in-review", "in_review", "in review":
			stats.TasksInReview++
		case "pending":
			stats.PendingTasks++
		case "completed":
			stats.CompletedTasks++
		case "cancelled", "canceled":
			stats.CancelledTasks++
		}

		// Count active tasks (not completed or cancelled)
		if status != "completed" && status != "cancelled" && status != "canceled" {
			stats.ActiveTasks++
		}

		// Count by priority
		if priority != "" {
			stats.ByPriority[priority]++
		}

		// Track assigned users and count tasks per assignee
		for _, assignID := range task.AssignTo {
			if assignID != "" {
				if !assignedUserSet[assignID] {
					assignedUserSet[assignID] = true
				}
				stats.TasksByAssignee[assignID]++
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

	return stats, nil
}
