package handlers

import (
	"log"
	"strings"

	"github.com/ar-13-go-backend/internal/services"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/ar-13-go-backend/pkg/sse"
	"github.com/gin-gonic/gin"
)

// SSEHandler handles SSE connection requests
type SSEHandler struct {
	sseService *sse.SSEService
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(taskService *services.TaskService, projectService *services.ProjectService) *SSEHandler {
	sseService := sse.NewSSEService(taskService, projectService)

	// Start the SSE service in a goroutine
	go sseService.Run()

	return &SSEHandler{
		sseService: sseService,
	}
}

// HandleConnection handles SSE connection requests
func (h *SSEHandler) HandleConnection(c *gin.Context) {
	// Extract token from query parameter or Authorization header
	token := c.Query("token")

	if token == "" {
		// Try Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
	}

	if token == "" {
		log.Printf("SSE connection rejected: no token provided")
		c.JSON(401, gin.H{"error": "Authentication required"})
		return
	}

	// Verify JWT token
	claims, err := jwt.VerifyToken(token)
	if err != nil {
		log.Printf("SSE connection rejected: invalid token: %v", err)
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims.UserID
	log.Printf("SSE connection attempt from user: %s", userID)

	// Handle the SSE connection
	err = h.sseService.HandleConnection(c.Writer, c.Request, userID)
	if err != nil {
		log.Printf("SSE connection failed for user %s: %v", userID, err)
		c.JSON(500, gin.H{"error": "Failed to establish SSE connection: " + err.Error()})
		return
	}

	// Abort to prevent Gin from writing further responses
	// The connection is now an SSE stream
	c.Abort()
}

// GetSSEService returns the SSE service instance
func (h *SSEHandler) GetSSEService() *sse.SSEService {
	return h.sseService
}

