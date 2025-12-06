package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ar-13-go-backend/services/filebrowser/handlers"
	"github.com/ar-13-go-backend/services/filebrowser/middleware"
	"github.com/ar-13-go-backend/services/filebrowser/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	defaultPort     = "8082"
	defaultDataRoot = "/data"
)

func main() {
	// Get configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dataRoot := os.Getenv("DATA_ROOT")
	if dataRoot == "" {
		dataRoot = defaultDataRoot
	}

	// Verify data root exists
	if _, err := os.Stat(dataRoot); os.IsNotExist(err) {
		log.Fatalf("Data root directory does not exist: %s", dataRoot)
	}

	log.Printf("Starting File Browser Service")
	log.Printf("Port: %s", port)
	log.Printf("Data Root: %s", dataRoot)

	// Initialize JWT with secret key from environment
	log.Printf("[Startup] Loading JWT_SECRET from environment...")
	jwtSecretRaw := os.Getenv("JWT_SECRET")
	log.Printf("[Startup] JWT_SECRET raw value length: %d", len(jwtSecretRaw))

	jwtSecret := strings.TrimSpace(jwtSecretRaw)
	log.Printf("[Startup] JWT_SECRET after trim length: %d", len(jwtSecret))

	if jwtSecret == "" {
		log.Fatal("[Startup] ERROR: JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) == 0 {
		log.Fatal("[Startup] ERROR: JWT_SECRET environment variable is empty after trimming")
	}

	log.Printf("[Startup] Initializing JWT with secret (length: %d)...", len(jwtSecret))
	jwt.InitializeJWT(jwtSecret)
	log.Printf("[Startup] JWT authentication initialized successfully")

	// Setup Gin router
	router := gin.Default()

	// Add request logging middleware
	router.Use(func(c *gin.Context) {
		log.Printf("[Request] %s %s from %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Next()
	})

	// Apply CORS middleware
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "file-browser",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes with JWT authentication
	api := router.Group("/api")
	api.Use(middleware.ValidateJWT())
	{
		api.GET("/browse", handlers.BrowseHandler(dataRoot))
		api.GET("/file-info", handlers.FileInfoHandler(dataRoot))
		api.GET("/download", handlers.DownloadHandler(dataRoot))
		api.POST("/create-folder", handlers.CreateFolderHandler(dataRoot))
		api.POST("/upload", handlers.UploadHandler(dataRoot))
		api.DELETE("/delete", handlers.DeleteHandler(dataRoot))
		api.PUT("/rename", handlers.RenameHandler(dataRoot))
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("File Browser Service is running on port %s", port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down File Browser Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("File Browser Service exited")
}
