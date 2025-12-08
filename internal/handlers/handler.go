package handlers

import (
	"log"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/repos"
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

	// Initialize storage service (FileBrowser Service, FileBrowser, or MinIO)
	var storageService services.StorageServiceInterface

	// Priority: FileBrowser Service > FileBrowser > MinIO
	if cfg.FileBrowserServiceURL != "" {
		// Use new filebrowser service (with JWT authentication)
		storageService = services.NewFileBrowserServiceStorage(cfg.FileBrowserServiceURL)
		if err := storageService.Initialize(); err != nil {
			log.Printf("Warning: Failed to initialize filebrowser service: %v. Trying fallback...", err)
			log.Printf("FileBrowser Service config - URL: %s", cfg.FileBrowserServiceURL)
			// Fall through to next option
			storageService = nil
		} else {
			log.Printf("FileBrowser service initialized successfully at %s", cfg.FileBrowserServiceURL)
		}
	}

	// Fallback to old FileBrowser if new service failed or not configured
	if storageService == nil && cfg.FileBrowserEnabled {
		// Use old FileBrowser if enabled
		if cfg.FileBrowserToken == "" {
			log.Printf("Warning: FILEBROWSER_ENABLED is true but FILEBROWSER_TOKEN is not set. Trying MinIO...")
		} else {
			storageService = services.NewFileBrowserStorageService(cfg.FileBrowserURL, cfg.FileBrowserToken)
			if err := storageService.Initialize(); err != nil {
				log.Printf("Warning: Failed to initialize FileBrowser storage service: %v. Trying MinIO...", err)
				log.Printf("FileBrowser config - URL: %s", cfg.FileBrowserURL)
				storageService = nil
			} else {
				log.Printf("FileBrowser storage service initialized successfully at %s", cfg.FileBrowserURL)
			}
		}
	}

	// Fallback to MinIO if filebrowser services failed or not configured
	if storageService == nil {
		// Use MinIO (default)
		storageService = services.NewStorageService(cfg)
		if err := storageService.Initialize(); err != nil {
			log.Printf("Warning: Failed to initialize MinIO storage service: %v. File storage features will be disabled.", err)
			log.Printf("MinIO config - Endpoint: %s, Bucket: %s, UseSSL: %v", cfg.MinIOEndpoint, cfg.MinIOBucket, cfg.MinIOUseSSL)
		} else {
			log.Printf("MinIO storage service initialized successfully - Endpoint: %s, Bucket: %s", cfg.MinIOEndpoint, cfg.MinIOBucket)
		}
	}

	// Initialize time tracking service
	timeTrackingService := services.NewTimeTrackingService(
		repos.NewTimeTrackingRepo(),
		repos.NewTaskRepo(),
	)

	// Set time tracking service on task service
	taskService.SetTimeTrackingService(timeTrackingService)

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
	taskHandler.SetTimeTrackingService(timeTrackingService)

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
	authHandler.SetTimeTrackingService(timeTrackingService)

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
