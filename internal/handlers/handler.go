package handlers

import (
	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/pkg/websocket"
)

// Handler contains all route handlers
type Handler struct {
	WebSocket      *WebSocketHandler
	SocketIO       *SocketIOHandler
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
func NewHandler(cfg *config.Config, wsService *websocket.WebSocketService) *Handler {
	return &Handler{
		WebSocket:      NewWebSocketHandler(wsService),
		SocketIO:       NewSocketIOHandler(),
		Auth:           NewAuthHandler(cfg),
		User:           NewUserHandler(cfg),
		Project:        NewProjectHandler(),
		Task:           NewTaskHandler(cfg),
		Dashboard:      NewDashboardHandler(),
		Calendar:       NewCalendarHandler(cfg),
		Notification:   NewNotificationHandler(wsService),
		Vacation:       NewVacationHandler(),
		Employee:       NewEmployeeHandler(),
		InfoPortal:     NewInfoPortalHandler(),
		ProjectDetails: NewProjectDetailsHandler(),
		ActivityLog:    NewActivityLogHandler(),
		GoogleAccount:  NewGoogleAccountHandler(),
		Backup:         NewBackupHandler(),
	}
}
