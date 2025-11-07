package services

import (
	"context"
	"strings"
	"sync"

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
	projectRepo *repos.ProjectRepo
	userRepo    *repos.UserRepo
	taskRepo    *repos.TaskRepo
}

// NewDashboardService creates a new dashboard service
func NewDashboardService() *DashboardService {
	return &DashboardService{
		projectRepo: repos.NewProjectRepo(),
		userRepo:    repos.NewUserRepo(),
		taskRepo:    repos.NewTaskRepo(),
	}
}

// GetAllStats gets all dashboard statistics
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
	employeesWithWorkload, err := s.calculateEmployeeWorkload(ctx, employeeRes.employees, allProjects)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"projects":       projectRes.projects,
		"employees":      employeesWithWorkload,
		"totalProjects":  len(projectRes.projects),
		"totalEmployees": len(employeesWithWorkload),
	}, nil
}

// calculateEmployeeWorkload calculates workload data for employees
func (s *DashboardService) calculateEmployeeWorkload(ctx context.Context, employees []models.User, projects []models.Project) ([]EmployeeWithWorkload, error) {
	// Build a map to count tasks per user efficiently
	taskCounts := make(map[string]WorkloadData)

	// Initialize task counts for all employees
	for _, user := range employees {
		taskCounts[user.ID] = WorkloadData{}
	}

	// Fetch all tasks for all projects and count
	for _, project := range projects {
		tasks, err := s.taskRepo.GetAll(ctx, project.ID)
		if err != nil {
			continue
		}

		for _, task := range tasks {
			// Skip tasks with no assignments
			if len(task.AssignTo) == 0 {
				continue
			}

			// Normalize status to lowercase for comparison
			status := strings.ToLower(strings.TrimSpace(task.Status))

			// Count tasks for each assigned user
			for _, assignID := range task.AssignTo {
				// Skip empty assign IDs
				if assignID == "" {
					continue
				}

				if counts, exists := taskCounts[assignID]; exists {
					counts.TotalTasks++

					// Count by status (case-insensitive)
					switch status {
					case "backlog", "todo", "to-do":
						counts.BacklogTasks++
					case "inprogress", "in-progress", "in_progress", "in progress":
						counts.TasksInProgress++
					case "inreview", "in-review", "in_review", "in review":
						counts.TasksInReview++
					case "pending":
						counts.PendingTasks++
					}

					// Count active tasks (not completed or cancelled)
					if status != "completed" && status != "cancelled" && status != "canceled" {
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
