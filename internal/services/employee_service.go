package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// EmployeeTaskCounts represents employee task counts
type EmployeeTaskCounts struct {
	UserID          string `json:"userId"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	BacklogTasks    int    `json:"backlogTasks"`
	TasksInProgress int    `json:"tasksInProgress"`
	TasksInReview   int    `json:"tasksInReview"`
	PendingTasks    int    `json:"pendingTasks"`
	TotalTasks      int    `json:"totalTasks"`
	ActiveTasks     int    `json:"activeTasks"`
}

// EmployeeService handles employee business logic
type EmployeeService struct {
	userRepo *repos.UserRepo
	taskRepo *repos.TaskRepo
}

// NewEmployeeService creates a new employee service
func NewEmployeeService() *EmployeeService {
	return &EmployeeService{
		userRepo: repos.NewUserRepo(),
		taskRepo: repos.NewTaskRepo(),
	}
}

// GetEmployeeList gets employee list with task counts
func (s *EmployeeService) GetEmployeeList(ctx context.Context) ([]EmployeeTaskCounts, error) {
	// Fetch all users once
	users, err := s.userRepo.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Fetch all projects once
	projectRepo := repos.NewProjectRepo()
	projects, err := projectRepo.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Build a map to count tasks per user efficiently
	// Map structure: userID -> {backlogTasks, tasksInProgress, tasksInReview, pendingTasks, totalTasks, activeTasks}
	taskCounts := make(map[string]struct {
		backlogTasks    int
		tasksInProgress int
		tasksInReview   int
		pendingTasks    int
		totalTasks      int
		activeTasks     int
	})

	// Initialize task counts for all users
	for _, user := range users {
		taskCounts[user.ID] = struct {
			backlogTasks    int
			tasksInProgress int
			tasksInReview   int
			pendingTasks    int
			totalTasks      int
			activeTasks     int
		}{0, 0, 0, 0, 0, 0}
	}

	// Fetch all tasks for all projects once and count
	for _, project := range projects {
		tasks, err := s.taskRepo.GetAll(ctx, project.ID)
		if err != nil {
			continue
		}

		for _, task := range tasks {
			// Skip tasks with no assignments
			if task.AssignTo == nil || *task.AssignTo == "" {
				continue
			}

			assignID := *task.AssignTo

			// Normalize status to lowercase for comparison
			status := strings.ToLower(strings.TrimSpace(task.Status))

			// Count tasks for assigned user
			if counts, exists := taskCounts[assignID]; exists {
				counts.totalTasks++

				// Normalize status and count by master status
				normalizedStatus := constants.NormalizeTaskStatus(status)
				switch normalizedStatus {
				case constants.GetTaskStatusString(constants.TaskStatusPending):
					counts.pendingTasks++
					counts.backlogTasks++ // Keep backward compatibility
				case constants.GetTaskStatusString(constants.TaskStatusInProgress):
					counts.tasksInProgress++
				case constants.GetTaskStatusString(constants.TaskStatusInReview):
					counts.tasksInReview++
				}

				// Count active tasks (not completed or rejected)
				if normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusCompleted) &&
					normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusRejected) {
					counts.activeTasks++
				}

				taskCounts[assignID] = counts
			}
		}
	}

	// Build the result list
	var employees []EmployeeTaskCounts
	for _, user := range users {
		counts := taskCounts[user.ID]
		employees = append(employees, EmployeeTaskCounts{
			UserID:          user.ID,
			Name:            user.Name,
			Email:           user.Email,
			BacklogTasks:    counts.backlogTasks,
			TasksInProgress: counts.tasksInProgress,
			TasksInReview:   counts.tasksInReview,
			PendingTasks:    counts.pendingTasks,
			TotalTasks:      counts.totalTasks,
			ActiveTasks:     counts.activeTasks,
		})
	}

	return employees, nil
}

// GetEmployeeTaskCounts gets task counts for a specific employee
func (s *EmployeeService) GetEmployeeTaskCounts(ctx context.Context, userID string) (*EmployeeTaskCounts, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("employee not found")
	}

	// Get all projects to count tasks
	projectRepo := repos.NewProjectRepo()
	projects, err := projectRepo.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	backlogTasks := 0
	tasksInProgress := 0
	tasksInReview := 0
	pendingTasks := 0
	totalTasks := 0
	activeTasks := 0

	for _, project := range projects {
		tasks, err := s.taskRepo.GetAll(ctx, project.ID)
		if err != nil {
			continue
		}

		for _, task := range tasks {
			// Skip tasks with no assignments
			if task.AssignTo == nil || *task.AssignTo == "" {
				continue
			}

			// Check if user is assigned to task
			if *task.AssignTo != userID {
				continue
			}

			// Normalize status
			normalizedStatus := constants.NormalizeTaskStatus(task.Status)

			totalTasks++

			// Count by master status
			switch normalizedStatus {
			case constants.GetTaskStatusString(constants.TaskStatusPending):
				pendingTasks++
				backlogTasks++ // Keep backward compatibility
			case constants.GetTaskStatusString(constants.TaskStatusInProgress):
				tasksInProgress++
			case constants.GetTaskStatusString(constants.TaskStatusInReview):
				tasksInReview++
			}

			// Count active tasks (not completed or rejected)
			if normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusCompleted) &&
				normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusRejected) {
				activeTasks++
			}
		}
	}

	return &EmployeeTaskCounts{
		UserID:          user.ID,
		Name:            user.Name,
		Email:           user.Email,
		BacklogTasks:    backlogTasks,
		TasksInProgress: tasksInProgress,
		TasksInReview:   tasksInReview,
		PendingTasks:    pendingTasks,
		TotalTasks:      totalTasks,
		ActiveTasks:     activeTasks,
	}, nil
}

// EmployeeTaskStats represents detailed employee task statistics and analysis
type EmployeeTaskStats struct {
	UserID      string                `json:"userId"`
	Name        string                `json:"name"`
	Email       string                `json:"email"`
	Period      string                `json:"period"`      // "month", "quarter", or "year"
	PeriodValue string                `json:"periodValue"` // e.g., "2024-01" for month, "2024-Q1" for quarter, "2024" for year
	Overall     TaskStatsOverview     `json:"overall"`
	ByProject   []ProjectTaskStats    `json:"byProject"`
	ByTime      []TimePeriodTaskStats `json:"byTime"`
	Analysis    TaskAnalysis          `json:"analysis"`
}

// TaskStatsOverview represents overall task statistics
type TaskStatsOverview struct {
	TotalTasks         int     `json:"totalTasks"`
	CompletedTasks     int     `json:"completedTasks"`
	ActiveTasks        int     `json:"activeTasks"`
	BacklogTasks       int     `json:"backlogTasks"`
	TasksInProgress    int     `json:"tasksInProgress"`
	TasksInReview      int     `json:"tasksInReview"`
	PendingTasks       int     `json:"pendingTasks"`
	TotalTimeSpent     int     `json:"totalTimeSpent"`     // in minutes
	AverageTimePerTask float64 `json:"averageTimePerTask"` // in minutes
	CompletionRate     float64 `json:"completionRate"`     // percentage
}

// ProjectTaskStats represents task statistics for a specific project
type ProjectTaskStats struct {
	ProjectID   string            `json:"projectId"`
	ProjectName string            `json:"projectName"`
	Stats       TaskStatsOverview `json:"stats"`
}

// TimePeriodTaskStats represents task statistics for a specific time period
type TimePeriodTaskStats struct {
	Period      string            `json:"period"`      // e.g., "2024-01", "2024-Q1", "2024"
	PeriodLabel string            `json:"periodLabel"` // e.g., "January 2024", "Q1 2024", "2024"
	Stats       TaskStatsOverview `json:"stats"`
}

// TaskAnalysis represents analysis of task performance
type TaskAnalysis struct {
	ProductivityTrend     string         `json:"productivityTrend"` // "increasing", "decreasing", "stable"
	MostActiveProject     string         `json:"mostActiveProject"`
	MostActiveProjectID   string         `json:"mostActiveProjectId"`
	AverageCompletionTime float64        `json:"averageCompletionTime"` // in minutes
	PeakProductivityMonth string         `json:"peakProductivityMonth"`
	TaskDistribution      map[string]int `json:"taskDistribution"` // status -> count
}

// GetEmployeeTaskStats gets detailed task statistics and analysis for an employee
func (s *EmployeeService) GetEmployeeTaskStats(ctx context.Context, userID string, period string, periodValue string, projectID *string) (*EmployeeTaskStats, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("employee not found")
	}

	// Get all projects
	projectRepo := repos.NewProjectRepo()
	projects, err := projectRepo.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Create project map for quick lookup
	projectMap := make(map[string]*models.Project)
	for i := range projects {
		projectMap[projects[i].ID] = &projects[i]
	}

	// Parse time period
	var startTime, endTime time.Time

	switch strings.ToLower(period) {
	case "month":
		// periodValue format: "2024-01"
		t, err := time.Parse("2006-01", periodValue)
		if err != nil {
			return nil, errors.New("invalid month format. Expected YYYY-MM")
		}
		startTime = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		endTime = startTime.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case "quarter":
		// periodValue format: "2024-Q1"
		parts := strings.Split(periodValue, "-Q")
		if len(parts) != 2 {
			return nil, errors.New("invalid quarter format. Expected YYYY-QN")
		}
		var year int
		var quarter int
		if _, err := fmt.Sscanf(parts[0], "%d", &year); err != nil {
			return nil, errors.New("invalid year in quarter format")
		}
		if _, err := fmt.Sscanf(parts[1], "%d", &quarter); err != nil || quarter < 1 || quarter > 4 {
			return nil, errors.New("invalid quarter number. Must be 1-4")
		}
		month := (quarter-1)*3 + 1
		startTime = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endTime = startTime.AddDate(0, 3, 0).Add(-time.Nanosecond)
	case "year":
		// periodValue format: "2024"
		year, err := strconv.Atoi(periodValue)
		if err != nil {
			return nil, errors.New("invalid year format. Expected YYYY")
		}
		startTime = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	default:
		return nil, errors.New("invalid period. Must be 'month', 'quarter', or 'year'")
	}

	// Collect all tasks for the employee
	var allTasks []models.Task
	var projectTasksMap = make(map[string][]models.Task)    // projectID -> tasks
	var timePeriodTasksMap = make(map[string][]models.Task) // period -> tasks

	// Filter projects if projectID is provided
	projectsToProcess := projects
	if projectID != nil {
		filtered := []models.Project{}
		for _, p := range projects {
			if p.ID == *projectID {
				filtered = append(filtered, p)
				break
			}
		}
		projectsToProcess = filtered
	}

	// Fetch tasks for each project
	for _, project := range projectsToProcess {
		tasks, err := s.taskRepo.GetAll(ctx, project.ID)
		if err != nil {
			continue
		}

		for _, task := range tasks {
			// Check if user is assigned
			if task.AssignTo == nil || *task.AssignTo != userID {
				continue
			}

			// Filter by time period (check if task was created or updated in the period)
			taskTime := task.Created
			if task.Updated != nil {
				taskTime = *task.Updated
			}

			if taskTime.Before(startTime) || taskTime.After(endTime) {
				continue
			}

			allTasks = append(allTasks, task)
			projectTasksMap[project.ID] = append(projectTasksMap[project.ID], task)

			// Group by time period for breakdown
			var timeKey string
			switch strings.ToLower(period) {
			case "month":
				timeKey = taskTime.Format("2006-01")
			case "quarter":
				quarter := (int(taskTime.Month())-1)/3 + 1
				timeKey = fmt.Sprintf("%d-Q%d", taskTime.Year(), quarter)
			case "year":
				timeKey = strconv.Itoa(taskTime.Year())
			}
			timePeriodTasksMap[timeKey] = append(timePeriodTasksMap[timeKey], task)
		}
	}

	// Calculate overall stats
	overall := calculateTaskStats(allTasks, userID)

	// Calculate stats by project
	byProject := []ProjectTaskStats{}
	for projectID, tasks := range projectTasksMap {
		project := projectMap[projectID]
		if project == nil {
			continue
		}
		stats := calculateTaskStats(tasks, userID)
		byProject = append(byProject, ProjectTaskStats{
			ProjectID:   projectID,
			ProjectName: project.Title,
			Stats:       stats,
		})
	}

	// Calculate stats by time period
	byTime := []TimePeriodTaskStats{}
	for timeKey, tasks := range timePeriodTasksMap {
		var label string
		switch strings.ToLower(period) {
		case "month":
			t, _ := time.Parse("2006-01", timeKey)
			label = t.Format("January 2006")
		case "quarter":
			label = timeKey
		case "year":
			label = timeKey
		}
		stats := calculateTaskStats(tasks, userID)
		byTime = append(byTime, TimePeriodTaskStats{
			Period:      timeKey,
			PeriodLabel: label,
			Stats:       stats,
		})
	}

	// Calculate analysis
	analysis := calculateTaskAnalysis(allTasks, byProject, byTime, userID)

	return &EmployeeTaskStats{
		UserID:      user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Period:      period,
		PeriodValue: periodValue,
		Overall:     overall,
		ByProject:   byProject,
		ByTime:      byTime,
		Analysis:    analysis,
	}, nil
}

// calculateTaskStats calculates statistics for a set of tasks
func calculateTaskStats(tasks []models.Task, userID string) TaskStatsOverview {
	stats := TaskStatsOverview{
		TotalTasks: len(tasks),
	}

	totalTimeSpent := 0
	completedCount := 0

	for _, task := range tasks {
		normalizedStatus := constants.NormalizeTaskStatus(task.Status)

		// Count by master status
		switch normalizedStatus {
		case constants.GetTaskStatusString(constants.TaskStatusPending):
			stats.PendingTasks++
			stats.BacklogTasks++ // Keep backward compatibility
		case constants.GetTaskStatusString(constants.TaskStatusInProgress):
			stats.TasksInProgress++
		case constants.GetTaskStatusString(constants.TaskStatusInReview):
			stats.TasksInReview++
		case constants.GetTaskStatusString(constants.TaskStatusCompleted):
			completedCount++
		}

		// Count active tasks (not completed or rejected)
		if normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusCompleted) &&
			normalizedStatus != constants.GetTaskStatusString(constants.TaskStatusRejected) {
			stats.ActiveTasks++
		}

		// Calculate time spent for this user
		for _, timeSpent := range task.TimeSpent {
			if timeSpent.UserID == userID {
				totalTimeSpent += timeSpent.TimeSpent
			}
		}
	}

	stats.CompletedTasks = completedCount
	stats.TotalTimeSpent = totalTimeSpent

	if stats.TotalTasks > 0 {
		stats.AverageTimePerTask = float64(totalTimeSpent) / float64(stats.TotalTasks)
		stats.CompletionRate = (float64(completedCount) / float64(stats.TotalTasks)) * 100
	}

	return stats
}

// calculateTaskAnalysis calculates analysis metrics
func calculateTaskAnalysis(tasks []models.Task, byProject []ProjectTaskStats, byTime []TimePeriodTaskStats, userID string) TaskAnalysis {
	analysis := TaskAnalysis{
		TaskDistribution: make(map[string]int),
	}

	// Find most active project
	maxTasks := 0
	for _, projectStats := range byProject {
		if projectStats.Stats.TotalTasks > maxTasks {
			maxTasks = projectStats.Stats.TotalTasks
			analysis.MostActiveProject = projectStats.ProjectName
			analysis.MostActiveProjectID = projectStats.ProjectID
		}
	}

	// Calculate average completion time
	totalCompletionTime := 0
	completedTasksWithTime := 0
	for _, task := range tasks {
		normalizedStatus := constants.NormalizeTaskStatus(task.Status)
		if normalizedStatus == constants.GetTaskStatusString(constants.TaskStatusCompleted) && task.Updated != nil {
			duration := task.Updated.Sub(task.Created).Minutes()
			if duration > 0 {
				totalCompletionTime += int(duration)
				completedTasksWithTime++
			}
		}
	}
	if completedTasksWithTime > 0 {
		analysis.AverageCompletionTime = float64(totalCompletionTime) / float64(completedTasksWithTime)
	}

	// Find peak productivity month
	maxCompleted := 0
	for _, timeStats := range byTime {
		if timeStats.Stats.CompletedTasks > maxCompleted {
			maxCompleted = timeStats.Stats.CompletedTasks
			analysis.PeakProductivityMonth = timeStats.PeriodLabel
		}
	}

	// Calculate productivity trend (simplified: compare first half vs second half)
	if len(byTime) >= 2 {
		midPoint := len(byTime) / 2
		firstHalfCompleted := 0
		secondHalfCompleted := 0
		for i := 0; i < midPoint; i++ {
			firstHalfCompleted += byTime[i].Stats.CompletedTasks
		}
		for i := midPoint; i < len(byTime); i++ {
			secondHalfCompleted += byTime[i].Stats.CompletedTasks
		}
		if secondHalfCompleted > firstHalfCompleted {
			analysis.ProductivityTrend = "increasing"
		} else if secondHalfCompleted < firstHalfCompleted {
			analysis.ProductivityTrend = "decreasing"
		} else {
			analysis.ProductivityTrend = "stable"
		}
	} else {
		analysis.ProductivityTrend = "stable"
	}

	// Task distribution by status (using normalized master statuses)
	for _, task := range tasks {
		normalizedStatus := constants.NormalizeTaskStatus(task.Status)
		if normalizedStatus != "" {
			analysis.TaskDistribution[normalizedStatus]++
		}
	}

	return analysis
}
