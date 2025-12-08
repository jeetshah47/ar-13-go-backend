package handlers

import (
	"strconv"
	"time"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// AuditLogHandler handles audit log routes
type AuditLogHandler struct {
	auditLogService *services.AuditLogService
}

// NewAuditLogHandler creates a new audit log handler with dependency injection
func NewAuditLogHandler(auditLogService *services.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{
		auditLogService: auditLogService,
	}
}

// NewAuditLogHandlerWithDefaults creates a new audit log handler with default dependencies
func NewAuditLogHandlerWithDefaults() *AuditLogHandler {
	return NewAuditLogHandler(
		services.NewAuditLogServiceWithDefaults(),
	)
}

// GetRecent gets recent audit logs with optional filters
// GET /api/audit-logs?limit=100&userId=xxx&method=POST&path=/api/users&statusCode=200&startDate=2024-01-01&endDate=2024-01-31
func (h *AuditLogHandler) GetRecent(c *gin.Context) {
	limit := 100 // default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Build filters
	filters := make(map[string]interface{})

	if userId := c.Query("userId"); userId != "" {
		filters["userId"] = userId
	}

	if method := c.Query("method"); method != "" {
		filters["method"] = method
	}

	if path := c.Query("path"); path != "" {
		filters["path"] = path
	}

	if statusCodeStr := c.Query("statusCode"); statusCodeStr != "" {
		if statusCode, err := strconv.Atoi(statusCodeStr); err == nil {
			filters["statusCode"] = statusCode
		}
	}

	// Date range filters
	var startDate, endDate time.Time
	if startDateStr := c.Query("startDate"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr := c.Query("endDate"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
			// Set to end of day
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	// If date range is provided, use date range query
	if !startDate.IsZero() || !endDate.IsZero() {
		if startDate.IsZero() {
			startDate = time.Now().AddDate(0, 0, -30) // Default to 30 days ago
		}
		if endDate.IsZero() {
			endDate = time.Now()
		}
		logs, err := h.auditLogService.GetByDateRange(c.Request.Context(), startDate, endDate, &limit)
		if err != nil {
			c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(constants.StatusOK, gin.H{"auditLogs": logs, "total": len(logs)})
		return
	}

	// Otherwise use recent query with filters
	logs, err := h.auditLogService.GetRecent(c.Request.Context(), limit, filters)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"auditLogs": logs, "total": len(logs)})
}

// GetByUserID gets audit logs for a specific user
// GET /api/audit-logs/user/:userId?limit=100
func (h *AuditLogHandler) GetByUserID(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var limit *int
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = &l
		}
	}

	logs, err := h.auditLogService.GetByUserID(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"auditLogs": logs, "total": len(logs)})
}

// GetByPath gets audit logs for a specific path
// GET /api/audit-logs/path?path=/api/users/all&limit=100
func (h *AuditLogHandler) GetByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	var limit *int
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = &l
		}
	}

	logs, err := h.auditLogService.GetByPath(c.Request.Context(), path, limit)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"auditLogs": logs, "total": len(logs)})
}

// GetByMethod gets audit logs for a specific HTTP method
// GET /api/audit-logs/method?method=POST&limit=100
func (h *AuditLogHandler) GetByMethod(c *gin.Context) {
	method := c.Query("method")
	if method == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Method is required"})
		return
	}

	var limit *int
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = &l
		}
	}

	logs, err := h.auditLogService.GetByMethod(c.Request.Context(), method, limit)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"auditLogs": logs, "total": len(logs)})
}

// GetByStatusCode gets audit logs for a specific status code
// GET /api/audit-logs/status/:statusCode?limit=100
func (h *AuditLogHandler) GetByStatusCode(c *gin.Context) {
	statusCodeStr := c.Param("statusCode")
	statusCode, err := strconv.Atoi(statusCodeStr)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Invalid status code"})
		return
	}

	var limit *int
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = &l
		}
	}

	logs, err := h.auditLogService.GetByStatusCode(c.Request.Context(), statusCode, limit)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"auditLogs": logs, "total": len(logs)})
}

