package services

import (
	"context"
	"sync"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

const (
	// Default limits for dashboard queries
	DefaultProjectLimit  = 10
	DefaultEmployeeLimit = 20
)

// EmployeeWithWorkload represents an employee with their workload data
type EmployeeWithWorkload struct {
	models.User
	Workload WorkloadData `json:"workload"`
}

// WorkloadData represents employee workload statistics
type WorkloadData struct {
	BacklogTasks    int `json:"backlogTasks"`
	TasksInProgress int `json:"tasksInProgress"`
	TasksInReview   int `json:"tasksInReview"`
	PendingTasks    int `json:"pendingTasks"`
	TotalTasks      int `json:"totalTasks"`
	ActiveTasks     int `json:"activeTasks"`
}

// DashboardService handles dashboard business logic
type DashboardService struct {
	projectRepo repos.ProjectRepository
	userRepo    repos.UserRepository
	taskRepo    repos.TaskRepository
	cacheSvc    CacheServiceInterface
}

// NewDashboardService creates a new dashboard service with dependency injection
func NewDashboardService(
	projectRepo repos.ProjectRepository,
	userRepo repos.UserRepository,
	taskRepo repos.TaskRepository,
	cacheSvc CacheServiceInterface,
) *DashboardService {
	return &DashboardService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
		taskRepo:    taskRepo,
		cacheSvc:    cacheSvc,
	}
}

// NewDashboardServiceWithDefaults creates a new dashboard service with default dependencies
func NewDashboardServiceWithDefaults() *DashboardService {
	return NewDashboardService(
		repos.NewProjectRepo(),
		repos.NewUserRepo(),
		repos.NewTaskRepo(),
		NewCacheService(),
	)
}

// GetAllStats gets all dashboard statistics
// Uses Redis cache to improve performance
func (s *DashboardService) GetAllStats(ctx context.Context, projectLimit, empLimit *int) (map[string]interface{}, error) {
	// Apply default limits if not provided to prevent fetching all data
	if projectLimit == nil {
		defaultLimit := DefaultProjectLimit
		projectLimit = &defaultLimit
	}
	if empLimit == nil {
		defaultLimit := DefaultEmployeeLimit
		empLimit = &defaultLimit
	}

	// Try to get from cache first
	cached, err := s.cacheSvc.GetDashboardStats(ctx, *projectLimit, *empLimit)
	if err == nil && cached != nil {
		return cached, nil
	}
	// If cache miss or error, continue to fetch from DB

	// Use goroutines to fetch projects and employees in parallel
	type projectResult struct {
		projects []models.Project
		err      error
	}

	type employeeResult struct {
		employees []models.User
		err       error
	}

	projectChan := make(chan projectResult, 1)
	employeeChan := make(chan employeeResult, 1)
	var wg sync.WaitGroup

	// Fetch projects in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		projs, err := s.projectRepo.GetAll(ctx, projectLimit)
		projectChan <- projectResult{projects: projs, err: err}
	}()

	// Fetch employees in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		emps, err := s.userRepo.GetAll(ctx, empLimit)
		employeeChan <- employeeResult{employees: emps, err: err}
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	// Collect results
	projectRes := <-projectChan
	employeeRes := <-employeeChan

	// Check for errors
	if projectRes.err != nil {
		return nil, projectRes.err
	}
	if employeeRes.err != nil {
		return nil, employeeRes.err
	}

	// Fetch all projects for accurate workload calculation (not limited)
	allProjects, err := s.projectRepo.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Calculate workload for employees using all projects
	// Optimized to batch fetch all tasks instead of nested loops
	employeesWithWorkload, err := s.calculateEmployeeWorkload(ctx, employeeRes.employees, allProjects)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"projects":       projectRes.projects,
		"employees":      employeesWithWorkload,
		"totalProjects":  len(projectRes.projects),
		"totalEmployees": len(employeesWithWorkload),
	}

	// Cache the result (ignore cache errors)
	_ = s.cacheSvc.SetDashboardStats(ctx, *projectLimit, *empLimit, result)

	return result, nil
}

// calculateEmployeeWorkload calculates workload data for employees
// Optimized to batch fetch all tasks instead of nested loops
func (s *DashboardService) calculateEmployeeWorkload(ctx context.Context, employees []models.User, projects []models.Project) ([]EmployeeWithWorkload, error) {
	// Build a map to count tasks per user efficiently
	taskCounts := make(map[string]WorkloadData)

	// Initialize task counts for all employees
	for _, user := range employees {
		taskCounts[user.ID] = WorkloadData{}
	}

	// Extract project IDs for batch fetching
	if len(projects) > 0 {
		projectIDs := make([]string, 0, len(projects))
		for _, project := range projects {
			projectIDs = append(projectIDs, project.ID)
		}

		// Batch fetch all tasks for all projects in parallel
		tasksByProject, err := s.taskRepo.GetAllByProjectIDs(ctx, projectIDs)
		if err != nil {
			return nil, err
		}

		// Process all tasks from all projects
		for _, tasks := range tasksByProject {
			for _, task := range tasks {
				// Skip tasks with no assignments
				if task.AssignTo == nil || *task.AssignTo == "" {
					continue
				}

				assignID := *task.AssignTo

				// Normalize status to master status value
				normalizedStatus := constants.NormalizeTaskStatus(task.Status)

				// Count tasks for assigned user
				if counts, exists := taskCounts[assignID]; exists {
					counts.TotalTasks++

					// Count by master status
					switch normalizedStatus {
					case constants.GetTaskStatusString(constants.TaskStatusPending):
						counts.PendingTasks++
						counts.BacklogTasks++ // Keep backward compatibility
					case constants.GetTaskStatusString(constants.TaskStatusInProgress):
						counts.TasksInProgress++
					case constants.GetTaskStatusString(constants.TaskStatusInReview):
						counts.TasksInReview++
					}

					// Count active tasks (not completed or rejected)
					if normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusCompleted) &&
						normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusRejected) {
						counts.ActiveTasks++
					}

					taskCounts[assignID] = counts
				}
			}
		}
	}

	// Build the result list with employee details and workload
	var employeesWithWorkload []EmployeeWithWorkload
	for _, user := range employees {
		workload := taskCounts[user.ID]
		employeesWithWorkload = append(employeesWithWorkload, EmployeeWithWorkload{
			User:     user,
			Workload: workload,
		})
	}

	return employeesWithWorkload, nil
}
