package handlers

import (
	"strconv"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// DashboardHandler handles dashboard routes
type DashboardHandler struct {
	dashboardService *services.DashboardService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{
		dashboardService: services.NewDashboardService(),
	}
}

// GetAllStats gets all dashboard statistics
func (h *DashboardHandler) GetAllStats(c *gin.Context) {
	var projectLimit, empLimit *int

	if projectLimitStr := c.Query("project_limit"); projectLimitStr != "" {
		if limit, err := strconv.Atoi(projectLimitStr); err == nil {
			projectLimit = &limit
		}
	}

	if empLimitStr := c.Query("emp_limit"); empLimitStr != "" {
		if limit, err := strconv.Atoi(empLimitStr); err == nil {
			empLimit = &limit
		}
	}

	stats, err := h.dashboardService.GetAllStats(c.Request.Context(), projectLimit, empLimit)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, stats)
}
