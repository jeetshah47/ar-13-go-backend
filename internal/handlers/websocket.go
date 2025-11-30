package handlers

import (
	"log"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/services"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/ar-13-go-backend/pkg/websocket"
	"github.com/gin-gonic/gin"
)

// WebSocketHandler handles WebSocket connection requests
type WebSocketHandler struct {
	websocketService *websocket.WebSocketService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(taskService *services.TaskService, projectService *services.ProjectService) *WebSocketHandler {
	websocketService := websocket.NewWebSocketService(taskService, projectService)

	// Start the WebSocket service in a goroutine
	go websocketService.Run()

	return &WebSocketHandler{
		websocketService: websocketService,
	}
}

// HandleConnection handles WebSocket connection requests
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
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
		log.Printf("WebSocket connection rejected: no token provided")
		c.JSON(401, gin.H{"error": "Authentication required"})
		return
	}

	// Verify JWT token
	claims, err := jwt.VerifyToken(token)
	if err != nil {
		log.Printf("WebSocket connection rejected: invalid token: %v", err)
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims.UserID
	log.Printf("WebSocket connection attempt from user: %s", userID)

	// Handle the WebSocket connection
	err = h.websocketService.HandleConnection(c.Writer, c.Request, userID)
	if err != nil {
		log.Printf("WebSocket connection failed for user %s: %v", userID, err)
		c.JSON(500, gin.H{"error": "Failed to establish WebSocket connection: " + err.Error()})
		return
	}

	// Send initial list of online users to the newly connected client
	// This is done after connection is established
	go func() {
		// Small delay to ensure connection is fully established
		time.Sleep(100 * time.Millisecond)
		onlineUsers := h.websocketService.GetOnlineUsers()
		onlineUsersList := make([]string, 0, len(onlineUsers))
		for userID := range onlineUsers {
			onlineUsersList = append(onlineUsersList, userID)
		}
		
		// Send to the newly connected user
		h.websocketService.SendToUser(userID, "users:online:list", map[string]interface{}{
			"users": onlineUsersList,
		})
	}()

	// Abort to prevent Gin from writing further responses
	// The connection is now a WebSocket connection
	c.Abort()
}

// GetWebSocketService returns the WebSocket service instance
func (h *WebSocketHandler) GetWebSocketService() *websocket.WebSocketService {
	return h.websocketService
}

