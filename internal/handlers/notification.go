package handlers

import (
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// NotificationHandler handles notification routes
type NotificationHandler struct {
	notificationService *services.NotificationService
	sseHandler          *SSEHandler
}

// NewNotificationHandler creates a new notification handler with dependency injection
func NewNotificationHandler(
	notificationService *services.NotificationService,
	sseHandler *SSEHandler,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		sseHandler:          sseHandler,
	}
}

// NewNotificationHandlerWithDefaults creates a new notification handler with default dependencies
func NewNotificationHandlerWithDefaults(sseHandler *SSEHandler) *NotificationHandler {
	return NewNotificationHandler(
		services.NewNotificationServiceWithDefaults(),
		sseHandler,
	)
}

// GetAll gets all notifications for a user
func (h *NotificationHandler) GetAll(c *gin.Context) {
	userID := c.Param("userId")
	notifications, err := h.notificationService.GetAll(c.Request.Context(), userID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"notifications": notifications})
}

// GetUnread gets unread notifications for a user
func (h *NotificationHandler) GetUnread(c *gin.Context) {
	userID := c.Param("userId")
	notifications, err := h.notificationService.GetUnread(c.Request.Context(), userID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"notifications": notifications})
}

// GetCount gets notification count for a user
func (h *NotificationHandler) GetCount(c *gin.Context) {
	userID := c.Param("userId")
	count, err := h.notificationService.GetCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"count": count})
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	if err := h.notificationService.MarkAsRead(c.Request.Context(), id); err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead marks all notifications as read for a user
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.Param("userId")
	if err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// Delete deletes a notification
func (h *NotificationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.notificationService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// DeleteAllForUser deletes all notifications for a user
func (h *NotificationHandler) DeleteAllForUser(c *gin.Context) {
	userID := c.Param("userId")
	if err := h.notificationService.DeleteAllForUser(c.Request.Context(), userID); err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"message": "All notifications deleted successfully"})
}

// GetConnectionInfo gets SSE connection info
func (h *NotificationHandler) GetConnectionInfo(c *gin.Context) {
	sseService := h.sseHandler.GetSSEService()
	info := gin.H{
		"connectedUsers":   sseService.GetConnectedUsersCount(),
		"connectedUserIds": sseService.GetConnectedUserIDs(),
	}
	c.JSON(constants.StatusOK, info)
}
