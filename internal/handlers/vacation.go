package handlers

import (
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// parseDateFlexible tries to parse a date string in multiple common formats
func parseDateFlexible(dateStr string) (time.Time, error) {
	// List of common date formats to try
	formats := []string{
		time.RFC3339,               // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,           // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05Z",     // ISO8601 with Z
		"2006-01-02T15:04:05",      // ISO8601 without timezone
		"2006-01-02 15:04:05",      // Space separated
		"2006-01-02",               // Date only
		"01/02/2006",               // US format
		"02/01/2006",               // European format
		"2006-01-02T15:04:05.000Z", // ISO8601 with milliseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s. Expected formats: RFC3339 (2006-01-02T15:04:05Z07:00), ISO8601 (2006-01-02T15:04:05Z), or date only (2006-01-02)", dateStr)
}

// VacationHandler handles vacation/leave request routes
type VacationHandler struct {
	vacationService *services.VacationService
}

// NewVacationHandler creates a new vacation handler
func NewVacationHandler() *VacationHandler {
	return &VacationHandler{
		vacationService: services.NewVacationService(),
	}
}

// GetMyRequests gets leave requests for the authenticated user
func (h *VacationHandler) GetMyRequests(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	requests, err := h.vacationService.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"requests": requests})
}

// GetAllRequests gets all leave requests
func (h *VacationHandler) GetAllRequests(c *gin.Context) {
	requests, err := h.vacationService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"requests": requests})
}

// GetPendingRequests gets pending leave requests
func (h *VacationHandler) GetPendingRequests(c *gin.Context) {
	requests, err := h.vacationService.GetPending(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"requests": requests})
}

// GetOneRequest gets a leave request by ID
func (h *VacationHandler) GetOneRequest(c *gin.Context) {
	requestID := c.Param("requestId")
	request, err := h.vacationService.GetByID(c.Request.Context(), requestID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if request == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"request": request})
}

// CreateRequest creates a new leave request
func (h *VacationHandler) CreateRequest(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		RequestType  string               `json:"requestType"`
		StartDate    string               `json:"startDate"`
		EndDate      *string              `json:"endDate,omitempty"`
		Duration     float64              `json:"duration"`
		DurationType string               `json:"durationType"`
		Comments     *string              `json:"comments,omitempty"`
		WorkingHours *models.WorkingHours `json:"workingHours,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := parseDateFlexible(req.StartDate)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := parseDateFlexible(*req.EndDate)
		if err != nil {
			c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		endDate = &parsed
	}

	request, err := h.vacationService.Create(
		c.Request.Context(),
		userID,
		models.LeaveRequestType(req.RequestType),
		startDate,
		endDate,
		req.Duration,
		models.DurationType(req.DurationType),
		req.Comments,
		req.WorkingHours,
	)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{
		"message": "Leave request created successfully",
		"request": request,
	})
}

// UpdateRequestStatus updates leave request status
func (h *VacationHandler) UpdateRequestStatus(c *gin.Context) {
	reviewedBy := middleware.GetUserID(c)
	if reviewedBy == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	requestID := c.Param("requestId")
	var req struct {
		Status         string  `json:"status"`
		ReviewComments *string `json:"reviewComments,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.vacationService.UpdateStatus(
		c.Request.Context(),
		requestID,
		models.LeaveRequestStatus(req.Status),
		reviewedBy,
		req.ReviewComments,
	); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message": "Leave request " + req.Status + " successfully",
	})
}

// UpdateRequest updates a leave request
func (h *VacationHandler) UpdateRequest(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request models.LeaveRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure user can only update their own requests
	if request.UserID != userID {
		c.JSON(constants.StatusForbidden, gin.H{"error": "You can only update your own leave requests"})
		return
	}

	// Check if request is pending
	existing, err := h.vacationService.GetByID(c.Request.Context(), request.ID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}
	if existing.Status != models.LeaveRequestStatusPending {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "You can only update pending leave requests"})
		return
	}

	if err := h.vacationService.Update(c.Request.Context(), &request); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Leave request updated successfully"})
}

// DeleteRequest deletes a leave request
func (h *VacationHandler) DeleteRequest(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	requestID := c.Param("requestId")
	request, err := h.vacationService.GetByID(c.Request.Context(), requestID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if request == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	if request.UserID != userID {
		c.JSON(constants.StatusForbidden, gin.H{"error": "You can only delete your own leave requests"})
		return
	}

	if err := h.vacationService.Delete(c.Request.Context(), requestID); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Leave request deleted successfully"})
}

// GetVacationSummaries gets vacation summaries
func (h *VacationHandler) GetVacationSummaries(c *gin.Context) {
	summaries, err := h.vacationService.GetVacationSummaries(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"summaries": summaries})
}

// GetRequestsByStatus gets leave requests by status
func (h *VacationHandler) GetRequestsByStatus(c *gin.Context) {
	status := c.Param("status")
	requests, err := h.vacationService.GetByStatus(c.Request.Context(), models.LeaveRequestStatus(status))
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"requests": requests})
}

// GetRequestsByType gets leave requests by type
func (h *VacationHandler) GetRequestsByType(c *gin.Context) {
	requestType := c.Param("type")
	requests, err := h.vacationService.GetByType(c.Request.Context(), models.LeaveRequestType(requestType))
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"requests": requests})
}
