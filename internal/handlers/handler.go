package handlers

import (
	"log"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/services"
)

// Handler contains all route handlers
type Handler struct {
	WebSocket        *WebSocketHandler
	Auth             *AuthHandler
	User             *UserHandler
	Project          *ProjectHandler
	Task             *TaskHandler
	Dashboard        *DashboardHandler
	Calendar         *CalendarHandler
	Notification     *NotificationHandler
	Vacation         *VacationHandler
	Employee         *EmployeeHandler
	InfoPortal       *InfoPortalHandler
	ProjectDetails   *ProjectDetailsHandler
	ActivityLog      *ActivityLogHandler
	ActivityLogReply *ActivityLogReplyHandler
	GoogleAccount    *GoogleAccountHandler
	DrawingList      *DrawingListHandler
	Storage          *StorageHandler
	NAS              *NASHandler
	AuditLog         *AuditLogHandler
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config) *Handler {
	// Create services for WebSocket
	taskService := services.NewTaskServiceWithDefaults(cfg)
	projectService := services.NewProjectServiceWithDefaults()

	websocketHandler := NewWebSocketHandler(taskService, projectService)

	// Set WebSocket and Notification services on TaskService
	notificationService := services.NewNotificationServiceWithDefaults()
	taskService.SetWebSocketService(websocketHandler.GetWebSocketService())
	taskService.SetNotificationService(notificationService)

	// Initialize storage service
	// Priority: NAS Direct Filesystem (primary) > MinIO (fallback only if explicitly configured)
	var storageService services.StorageServiceInterface

	// Primary: NAS Direct Filesystem (when backend is deployed on NAS)
	if cfg.NASBasePath != "" {
		storageService = services.NewNASDirectFilesystemStorage(cfg)
		if err := storageService.Initialize(); err != nil {
			log.Printf("Error: Failed to initialize NAS direct filesystem storage: %v", err)
			log.Printf("NAS Direct Filesystem config - Base Path: %s", cfg.NASBasePath)
			log.Printf("Please ensure NAS_BASE_PATH is set correctly and the path exists and is accessible")
			// Don't fall through - fail if NAS_BASE_PATH is set but initialization fails
			storageService = nil
		} else {
			log.Printf("NAS direct filesystem storage initialized successfully - Base Path: %s", cfg.NASBasePath)
		}
	}

	// Fallback to MinIO only if NAS_BASE_PATH is not set AND MinIO is explicitly configured
	if storageService == nil {
		if cfg.MinIOEndpoint != "" && cfg.MinIOBucket != "" {
			// Use MinIO only if explicitly configured
			storageService = services.NewStorageService(cfg)
			if err := storageService.Initialize(); err != nil {
				log.Printf("Error: Failed to initialize MinIO storage service: %v. File storage features will be disabled.", err)
				log.Printf("MinIO config - Endpoint: %s, Bucket: %s, UseSSL: %v", cfg.MinIOEndpoint, cfg.MinIOBucket, cfg.MinIOUseSSL)
			} else {
				log.Printf("MinIO storage service initialized successfully - Endpoint: %s, Bucket: %s", cfg.MinIOEndpoint, cfg.MinIOBucket)
			}
		} else {
			// Neither NAS_BASE_PATH nor MinIO is configured
			log.Printf("Warning: No storage service configured. Please set NAS_BASE_PATH for direct filesystem access.")
			log.Printf("File storage features will be disabled.")
		}
	}

	// Create handlers with dependency injection
	projectHandler := NewProjectHandlerWithDefaults()
	// Set notification and websocket services for project service
	if projectService := projectHandler.GetProjectService(); projectService != nil {
		projectService.SetNotificationService(notificationService)
		projectService.SetWebSocketService(websocketHandler.GetWebSocketService())
	}

	taskHandler := NewTaskHandlerWithDefaults(cfg)
	// Pass WebSocket service and notification service to task handler for event broadcasting
	taskHandler.SetWebSocketService(websocketHandler.GetWebSocketService())
	taskHandler.SetNotificationService(notificationService)
	taskHandler.SetStorageService(storageService)

	// Set notification and websocket services for calendar service
	calendarHandler := NewCalendarHandlerWithDefaults(cfg)
	calendarService := calendarHandler.GetCalendarService()
	if calendarService != nil {
		calendarService.SetNotificationService(notificationService)
		calendarService.SetWebSocketService(websocketHandler.GetWebSocketService())
	}

	storageHandler := NewStorageHandler(storageService, cfg)

	activityLogReplyHandler := NewActivityLogReplyHandlerWithDefaults(websocketHandler.GetWebSocketService())

	// Register WebSocket message handlers for activity log replies
	activityLogReplyHandler.RegisterWebSocketHandlers(websocketHandler.GetWebSocketService())

	authHandler := NewAuthHandler(cfg)

	// Set notification and websocket services for user service
	userHandler := NewUserHandlerWithDefaults(cfg)
	if userService := userHandler.GetUserService(); userService != nil {
		userService.SetNotificationService(notificationService)
		userService.SetWebSocketService(websocketHandler.GetWebSocketService())
	}

	return &Handler{
		WebSocket:        websocketHandler,
		Auth:             authHandler,
		User:             userHandler,
		Project:          projectHandler,
		Task:             taskHandler,
		Dashboard:        NewDashboardHandlerWithDefaults(),
		Calendar:         calendarHandler,
		Notification:     NewNotificationHandlerWithDefaults(websocketHandler),
		Vacation:         NewVacationHandlerWithDefaults(),
		Employee:         NewEmployeeHandlerWithDefaults(),
		InfoPortal:       NewInfoPortalHandlerWithDefaults(),
		ProjectDetails:   NewProjectDetailsHandlerWithDefaults(),
		ActivityLog:      NewActivityLogHandlerWithDefaults(),
		ActivityLogReply: activityLogReplyHandler,
		GoogleAccount:    NewGoogleAccountHandlerWithDefaults(),
		DrawingList:      NewDrawingListHandlerWithDefaults(),
		Storage:          storageHandler,
		NAS:              NewNASHandler(cfg),
		AuditLog:         NewAuditLogHandlerWithDefaults(),
	}
}
