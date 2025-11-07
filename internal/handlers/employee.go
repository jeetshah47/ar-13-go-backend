package handlers

import (
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// EmployeeHandler handles employee routes
type EmployeeHandler struct {
	employeeService *services.EmployeeService
}

// NewEmployeeHandler creates a new employee handler
func NewEmployeeHandler() *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: services.NewEmployeeService(),
	}
}

// GetEmployeeList gets employee list with task counts
func (h *EmployeeHandler) GetEmployeeList(c *gin.Context) {
	employees, err := h.employeeService.GetEmployeeList(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{
			"message": "Error fetching employee list",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"employees":      employees,
		"totalEmployees": len(employees),
	})
}

// GetEmployeeTaskCounts gets task counts for a specific employee
func (h *EmployeeHandler) GetEmployeeTaskCounts(c *gin.Context) {
	userID := c.Param("userId")
	employee, err := h.employeeService.GetEmployeeTaskCounts(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "employee not found" {
			c.JSON(constants.StatusNotFound, gin.H{
				"message": "Employee not found",
			})
			return
		}
		c.JSON(constants.StatusInternalServerError, gin.H{
			"message": "Error fetching employee task counts",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"employee": employee})
}

// GetEmployeeTaskStats gets detailed task statistics and analysis for an employee
func (h *EmployeeHandler) GetEmployeeTaskStats(c *gin.Context) {
	userID := c.Param("userId")
	period := c.Query("period")           // "month", "quarter", or "year"
	periodValue := c.Query("periodValue") // e.g., "2024-01", "2024-Q1", "2024"
	projectID := c.Query("projectId")     // optional: filter by project

	// Validate required parameters
	if period == "" {
		c.JSON(constants.StatusBadRequest, gin.H{
			"message": "period parameter is required (month, quarter, or year)",
		})
		return
	}
	if periodValue == "" {
		c.JSON(constants.StatusBadRequest, gin.H{
			"message": "periodValue parameter is required",
		})
		return
	}

	// Convert projectID to pointer if provided
	var projectIDPtr *string
	if projectID != "" {
		projectIDPtr = &projectID
	}

	stats, err := h.employeeService.GetEmployeeTaskStats(c.Request.Context(), userID, period, periodValue, projectIDPtr)
	if err != nil {
		if err.Error() == "employee not found" {
			c.JSON(constants.StatusNotFound, gin.H{
				"message": "Employee not found",
			})
			return
		}
		c.JSON(constants.StatusBadRequest, gin.H{
			"message": "Error fetching employee task stats",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"stats": stats})
}
