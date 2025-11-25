package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ar-13-go-backend/services/filebrowser/handlers"
	"github.com/ar-13-go-backend/services/filebrowser/middleware"
	// "github.com/ar-13-go-backend/services/filebrowser/pkg/jwt" // AUTHENTICATION COMMENTED OUT
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

	// Setup Gin router
	router := gin.Default()

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

	// AUTHENTICATION COMMENTED OUT - All APIs are now publicly accessible
	// Get JWT secret key (should match main backend's JWT_SECRET)
	// jwtSecret := os.Getenv("JWT_SECRET")
	// if jwtSecret == "" {
	// 	log.Fatal("JWT_SECRET environment variable is required")
	// }

	// Initialize JWT
	// jwt.InitializeJWT(jwtSecret)

	// API routes (NO AUTHENTICATION - publicly accessible)
	api := router.Group("/api")
	// api.Use(middleware.ValidateJWT()) // AUTHENTICATION COMMENTED OUT
	{
		api.GET("/browse", handlers.BrowseHandler(dataRoot))
		api.GET("/file-info", handlers.FileInfoHandler(dataRoot))
		api.GET("/download", handlers.DownloadHandler(dataRoot))
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
