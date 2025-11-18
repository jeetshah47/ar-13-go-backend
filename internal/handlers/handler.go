package handlers

import (
	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/services"
)

// Handler contains all route handlers
type Handler struct {
	SSE            *SSEHandler
	Auth           *AuthHandler
	User           *UserHandler
	Project        *ProjectHandler
	Task           *TaskHandler
	Dashboard      *DashboardHandler
	Calendar       *CalendarHandler
	Notification   *NotificationHandler
	Vacation       *VacationHandler
	Employee       *EmployeeHandler
	InfoPortal     *InfoPortalHandler
	ProjectDetails *ProjectDetailsHandler
	ActivityLog    *ActivityLogHandler
	GoogleAccount  *GoogleAccountHandler
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config) *Handler {
	// Create services for SSE
	taskService := services.NewTaskServiceWithDefaults(cfg)
	projectService := services.NewProjectServiceWithDefaults()
	
	sseHandler := NewSSEHandler(taskService, projectService)
	
	// Create handlers with dependency injection
	projectHandler := NewProjectHandlerWithDefaults()
	taskHandler := NewTaskHandlerWithDefaults(cfg)
	// Pass SSE service to task handler for event broadcasting
	taskHandler.SetSSEService(sseHandler.GetSSEService())
	
	return &Handler{
		SSE:            sseHandler,
		Auth:           NewAuthHandler(cfg),
		User:           NewUserHandlerWithDefaults(cfg),
		Project:        projectHandler,
		Task:           taskHandler,
		Dashboard:      NewDashboardHandlerWithDefaults(),
		Calendar:       NewCalendarHandlerWithDefaults(cfg),
		Notification:   NewNotificationHandlerWithDefaults(sseHandler),
		Vacation:       NewVacationHandlerWithDefaults(),
		Employee:       NewEmployeeHandlerWithDefaults(),
		InfoPortal:     NewInfoPortalHandlerWithDefaults(),
		ProjectDetails: NewProjectDetailsHandlerWithDefaults(),
		ActivityLog:    NewActivityLogHandlerWithDefaults(),
		GoogleAccount:  NewGoogleAccountHandlerWithDefaults(),
	}
}
