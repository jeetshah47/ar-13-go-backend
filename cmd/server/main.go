package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/handlers"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/pkg/cache"
	"github.com/ar-13-go-backend/pkg/dynamodb"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize JWT
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	jwt.InitializeJWT(cfg.JWTSecret)

	// Initialize DynamoDB
	_, err = dynamodb.InitializeDynamoDB(cfg.AWSRegion)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB: %v", err)
	}
	log.Println("DynamoDB initialized successfully")

	// Initialize Redis (optional - will continue if Redis is unavailable)
	_, err = cache.InitializeRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis: %v. Caching will be disabled.", err)
	} else {
		log.Println("Redis initialized successfully")
	}

	// Create router
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Serve static files
	router.Static("/uploads", "./upload")

	// Initialize handlers
	handler := handlers.NewHandler(cfg)

	// Setup routes
	setupRoutes(router, handler, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRoutes(router *gin.Engine, handler *handlers.Handler, cfg *config.Config) {
	// WebSocket endpoint
	router.GET("/ws", handler.WebSocket.HandleConnection)

	api := router.Group("/api")

	// Auth routes (no auth required)
	auth := api.Group("/auth")
	{
		auth.POST("/register", handler.Auth.Register)
		auth.POST("/login", handler.Auth.Login)
		auth.POST("/logout", middleware.ValidateToken(), handler.Auth.Logout)
		auth.GET("/validate-signup", handler.Auth.ValidateSignupToken)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthenticateUser())
	{
		// User routes
		users := protected.Group("/users")
		{
			users.GET("/all", middleware.RequireAdmin(), handler.User.GetAll)
			users.POST("/invite", middleware.RequireAdmin(), handler.User.CreateInvitation)
			users.PUT("/update", middleware.RequireAdmin(), handler.User.Update)
			users.DELETE("/delete/:id", middleware.RequireAdmin(), handler.User.Delete)
			users.GET("/profile/:id", handler.User.GetProfile)
		}

		// Project routes
		projects := protected.Group("/project")
		{
			projects.GET("/all", handler.Project.GetAll)
			projects.GET("/all/statistics", handler.Project.GetAllWithStatistics)
			projects.GET("/:id", middleware.RequireProjectAccess(), handler.Project.GetOne)
			projects.POST("/add", middleware.RequirePermission("projects:write"), handler.Project.Add)
			projects.PUT("/update", middleware.RequireProjectAccess(), middleware.RequirePermission("projects:write"), handler.Project.Update)
			projects.DELETE("/delete/:id", middleware.RequireProjectAccess(), middleware.RequirePermission("projects:delete"), handler.Project.Delete)
		}

		// Task routes
		tasks := protected.Group("/tasks")
		{
			tasks.GET("/all/:projectId", handler.Task.GetAll)
			tasks.GET("/all/details/:projectId", handler.Task.GetAllTaskDetail)
			tasks.GET("/detail/:projectId/:taskId", middleware.RequireTaskAccess(), handler.Task.GetOneTaskDetail)
			tasks.POST("/add", middleware.RequirePermission("tasks:write"), handler.Task.Add)
			tasks.POST("/add-multiple", middleware.RequirePermission("tasks:write"), handler.Task.AddMultiple)
			tasks.PUT("/update", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.Update)
			tasks.PUT("/update-deadline/:projectId/:taskId", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.UpdateDeadline)
			tasks.PUT("/update-progress/:projectId/:taskId", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.UpdateProgress)
			tasks.PUT("/update-description/:projectId/:taskId", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.UpdateDescription)
			tasks.PUT("/update-status/:projectId/:taskId", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.UpdateStatus)
			tasks.POST("/add-time-spent/:projectId/:taskId", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.AddTimeSpent)
			tasks.PUT("/update-time-spent/:projectId/:taskId/:timeSpentIndex", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.UpdateTimeSpent)
			tasks.DELETE("/remove-time-spent/:projectId/:taskId/:timeSpentIndex", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.RemoveTimeSpent)
			tasks.GET("/time-spent/:projectId/:taskId", middleware.RequireTaskAccess(), handler.Task.GetTimeSpent)
			tasks.POST("/add-file-attachment/:projectId/:taskId", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.AddFileAttachment)
			tasks.DELETE("/remove-file-attachment/:projectId/:taskId/:fileAttachmentIndex", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:write"), handler.Task.RemoveFileAttachment)
			tasks.GET("/file-attachments/:projectId/:taskId", middleware.RequireTaskAccess(), handler.Task.GetFileAttachments)
			tasks.GET("/activity-logs/:projectId/:taskId", middleware.RequireTaskAccess(), handler.Task.GetActivityLogs)
			tasks.DELETE("/delete/:projectId/:taskId", middleware.RequireTaskAccess(), middleware.RequirePermission("tasks:delete"), handler.Task.Delete)
			tasks.PUT("/assign/:taskId/:userId", middleware.RequirePermission("tasks:assign"), handler.Task.Assign)
			tasks.PUT("/claim/:projectId/:taskId", middleware.RequireTaskAccess(), handler.Task.Claim)
			tasks.GET("/assignable/:projectId", middleware.RequirePermission("tasks:assign"), handler.Task.GetAssignableUsers)
		}

		// Dashboard routes
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/stats", middleware.RequirePermission("dashboard:read"), handler.Dashboard.GetAllStats)
		}

		// Calendar routes
		calendar := protected.Group("/calendar")
		{
			calendar.GET("/month/:year/:month", middleware.RequirePermission("calendar:read"), handler.Calendar.GetByMonth)
			calendar.GET("/event/:id", middleware.RequirePermission("calendar:read"), handler.Calendar.GetById)
			calendar.POST("/add", middleware.RequirePermission("calendar:write"), handler.Calendar.Add)
			calendar.PUT("/update/:id", middleware.RequirePermission("calendar:write"), handler.Calendar.Update)
			calendar.DELETE("/delete/:id", middleware.RequirePermission("calendar:delete"), handler.Calendar.Delete)
		}

		// Notification routes
		notifications := protected.Group("/notifications")
		{
			notifications.GET("/all/:userId", middleware.RequirePermission("notifications:read"), handler.Notification.GetAll)
			notifications.GET("/unread/:userId", middleware.RequirePermission("notifications:read"), handler.Notification.GetUnread)
			notifications.GET("/count/:userId", middleware.RequirePermission("notifications:read"), handler.Notification.GetCount)
			notifications.PUT("/read/:id", middleware.RequirePermission("notifications:write"), handler.Notification.MarkAsRead)
			notifications.PUT("/read-all/:userId", middleware.RequirePermission("notifications:write"), handler.Notification.MarkAllAsRead)
			notifications.DELETE("/:id", middleware.RequirePermission("notifications:delete"), handler.Notification.Delete)
			notifications.DELETE("/user/:userId", middleware.RequirePermission("notifications:delete"), handler.Notification.DeleteAllForUser)
			notifications.GET("/connection-info", middleware.RequirePermission("notifications:read"), handler.Notification.GetConnectionInfo)
		}

		// Vacation routes
		vacation := protected.Group("/vacation")
		{
			vacation.GET("/my-requests", handler.Vacation.GetMyRequests)
			vacation.GET("/all", middleware.RequireAdmin(), handler.Vacation.GetAllRequests)
			vacation.GET("/pending", middleware.RequireAdmin(), handler.Vacation.GetPendingRequests)
			vacation.GET("/:requestId", handler.Vacation.GetOneRequest)
			vacation.POST("/create", handler.Vacation.CreateRequest)
			vacation.PUT("/update-status/:requestId", middleware.RequireAdmin(), handler.Vacation.UpdateRequestStatus)
			vacation.PUT("/update", handler.Vacation.UpdateRequest)
			vacation.DELETE("/delete/:requestId", handler.Vacation.DeleteRequest)
			vacation.GET("/summaries", middleware.RequireAdmin(), handler.Vacation.GetVacationSummaries)
			vacation.GET("/status/:status", middleware.RequireAdmin(), handler.Vacation.GetRequestsByStatus)
			vacation.GET("/type/:type", middleware.RequireAdmin(), handler.Vacation.GetRequestsByType)
		}

		// Employee routes
		employee := protected.Group("/employee")
		{
			employee.GET("/list", middleware.RequirePermission("employees:read"), handler.Employee.GetEmployeeList)
			employee.GET("/task-counts/:userId", middleware.RequirePermission("employees:read"), handler.Employee.GetEmployeeTaskCounts)
			employee.GET("/stats/:userId", middleware.RequirePermission("employees:read"), handler.Employee.GetEmployeeTaskStats)
		}

		// Info Portal routes
		infoPortal := protected.Group("/info-portal")
		{
			// Folders
			infoPortal.GET("/folders", middleware.RequirePermission("infoPortal:read"), handler.InfoPortal.GetAllFolders)
			infoPortal.GET("/folders/:folderId", middleware.RequirePermission("infoPortal:read"), handler.InfoPortal.GetFolderById)
			infoPortal.POST("/folders", middleware.RequirePermission("infoPortal:write"), handler.InfoPortal.CreateFolder)
			infoPortal.PUT("/folders/:folderId", middleware.RequirePermission("infoPortal:write"), handler.InfoPortal.UpdateFolder)
			infoPortal.DELETE("/folders/:folderId", middleware.RequirePermission("infoPortal:delete"), handler.InfoPortal.DeleteFolder)

			// Pages
			infoPortal.GET("/pages/:pageId", middleware.RequirePermission("infoPortal:read"), handler.InfoPortal.GetPageById)
			infoPortal.POST("/folders/:folderId/pages", middleware.RequirePermission("infoPortal:write"), handler.InfoPortal.CreatePage)
			infoPortal.PUT("/pages/:pageId", middleware.RequirePermission("infoPortal:write"), handler.InfoPortal.UpdatePage)
			infoPortal.DELETE("/pages/:pageId", middleware.RequirePermission("infoPortal:delete"), handler.InfoPortal.DeletePage)
			infoPortal.PUT("/pages/:pageId/sections", middleware.RequirePermission("infoPortal:write"), handler.InfoPortal.UpdatePageSections)

			// Attachments
			infoPortal.POST("/pages/:pageId/attachments", middleware.RequirePermission("infoPortal:write"), handler.InfoPortal.UploadAttachment)
			infoPortal.DELETE("/attachments/:attachmentId", middleware.RequirePermission("infoPortal:delete"), handler.InfoPortal.DeleteAttachment)

			// Statistics
			infoPortal.GET("/statistics", middleware.RequirePermission("infoPortal:read"), handler.InfoPortal.GetStatistics)
		}

		// Project Details routes
		projectDetails := protected.Group("/project-details")
		{
			projectDetails.GET("/:projectId", middleware.RequireProjectAccess(), handler.ProjectDetails.Get)
			projectDetails.POST("/:projectId", middleware.RequireProjectAccess(), middleware.RequirePermission("projects:write"), handler.ProjectDetails.Add)
			projectDetails.PUT("/:projectId", middleware.RequireProjectAccess(), middleware.RequirePermission("projects:write"), handler.ProjectDetails.Update)
			projectDetails.DELETE("/:projectId/:projectDetailsId", middleware.RequireProjectAccess(), middleware.RequirePermission("projects:delete"), handler.ProjectDetails.Delete)
		}

		// Activity Log routes
		activityLog := protected.Group("/activity-log")
		{
			activityLog.GET("/entity/:entityType/:entityId", middleware.RequirePermission("activityLog:read"), handler.ActivityLog.GetByEntity)
			activityLog.GET("/entity-type/:entityType", middleware.RequirePermission("activityLog:read"), handler.ActivityLog.GetByEntityType)
			activityLog.GET("/entity-types", handler.ActivityLog.GetEntityTypes)
		}

		// Google Account routes
		googleAccount := protected.Group("/google-account")
		{
			googleAccount.POST("/link", handler.GoogleAccount.LinkGoogleAccount)
			googleAccount.POST("/unlink", handler.GoogleAccount.UnlinkGoogleAccount)
			googleAccount.GET("/status", handler.GoogleAccount.GetGoogleAccountStatus)
			googleAccount.GET("/all", handler.GoogleAccount.GetAllLinkedAccounts)
			googleAccount.GET("/auth/initiate", handler.GoogleAccount.InitiateGoogleOAuth)
			googleAccount.GET("/calendar/events", handler.GoogleAccount.GetGoogleCalendarEvents)
		}

		// Backup routes (admin only)
		backup := protected.Group("/backup")
		{
			backup.POST("/all", middleware.RequireAdmin(), handler.Backup.BackupAllCollections)
		}
	}

	// Google OAuth callback (public - called by Google, not by authenticated user)
	// This must be outside the protected group because Google redirects here without auth headers
	api.GET("/google-account/auth/callback", handler.GoogleAccount.HandleGoogleOAuthCallback)
}
