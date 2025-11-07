package handlers

import (
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/pkg/websocket"
	"github.com/gin-gonic/gin"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	wsService *websocket.WebSocketService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(wsService *websocket.WebSocketService) *WebSocketHandler {
	return &WebSocketHandler{
		wsService: wsService,
	}
}

// HandleConnection handles WebSocket connection
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}

	// Upgrade connection to WebSocket
	// Note: After upgrade, the connection is a WebSocket connection, not HTTP
	// So we don't return JSON responses after this point
	err := h.wsService.HandleConnection(c.Writer, c.Request, userID)
	if err != nil {
		// Only send JSON error if upgrade failed (connection is still HTTP)
		c.JSON(500, gin.H{"error": "Failed to upgrade connection: " + err.Error()})
		return
	}
	// Connection upgraded successfully - handler will block until connection closes
}
