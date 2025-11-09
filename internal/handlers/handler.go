package handlers

import (
	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/services"
)

// Handler contains all route handlers
type Handler struct {
	WebSocket      *WebSocketHandler
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
	Backup         *BackupHandler
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config) *Handler {
	taskHandler := NewTaskHandler(cfg)
	projectHandler := NewProjectHandler()
	
	// Create services for WebSocket
	taskService := services.NewTaskService(cfg)
	projectService := services.NewProjectService()
	
	webSocketHandler := NewWebSocketHandler(taskService, projectService)
	return &Handler{
		WebSocket:      webSocketHandler,
		Auth:           NewAuthHandler(cfg),
		User:           NewUserHandler(cfg),
		Project:        projectHandler,
		Task:           taskHandler,
		Dashboard:      NewDashboardHandler(),
		Calendar:       NewCalendarHandler(cfg),
		Notification:   NewNotificationHandler(webSocketHandler),
		Vacation:       NewVacationHandler(),
		Employee:       NewEmployeeHandler(),
		InfoPortal:     NewInfoPortalHandler(),
		ProjectDetails: NewProjectDetailsHandler(),
		ActivityLog:    NewActivityLogHandler(),
		GoogleAccount:  NewGoogleAccountHandler(),
		Backup:         NewBackupHandler(),
	}
}
