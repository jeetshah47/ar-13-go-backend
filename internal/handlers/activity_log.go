package handlers

import (
	"strconv"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ActivityLogHandler handles activity log routes
type ActivityLogHandler struct {
	activityLogService *services.ActivityLogService
}

// NewActivityLogHandler creates a new activity log handler with dependency injection
func NewActivityLogHandler(activityLogService *services.ActivityLogService) *ActivityLogHandler {
	return &ActivityLogHandler{
		activityLogService: activityLogService,
	}
}

// NewActivityLogHandlerWithDefaults creates a new activity log handler with default dependencies
func NewActivityLogHandlerWithDefaults() *ActivityLogHandler {
	return NewActivityLogHandler(
		services.NewActivityLogServiceWithDefaults(),
	)
}

// GetByEntity gets activity logs for a specific entity
func (h *ActivityLogHandler) GetByEntity(c *gin.Context) {
	entityType := models.ActivityLogEntityType(c.Param("entityType"))
	entityID := c.Param("entityId")

	logs, err := h.activityLogService.GetByEntity(c.Request.Context(), entityType, entityID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"activityLogs": logs})
}

// GetByEntityType gets activity logs by entity type
func (h *ActivityLogHandler) GetByEntityType(c *gin.Context) {
	entityType := models.ActivityLogEntityType(c.Param("entityType"))
	limitStr := c.Query("limit")

	var limit *int
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = &parsedLimit
		}
	}

	logs, err := h.activityLogService.GetByEntityType(c.Request.Context(), entityType, limit)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"activityLogs": logs})
}

// GetEntityTypes gets supported entity types
func (h *ActivityLogHandler) GetEntityTypes(c *gin.Context) {
	entityTypes := []models.ActivityLogEntityType{
		models.ActivityLogEntityTypeTask,
		models.ActivityLogEntityTypeProject,
		models.ActivityLogEntityTypeUser,
		models.ActivityLogEntityTypeCalendarEvent,
	}

	c.JSON(constants.StatusOK, gin.H{"entityTypes": entityTypes})
}
