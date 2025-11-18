package handlers

import (
	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// CalendarHandler handles calendar routes
type CalendarHandler struct {
	calendarService *services.CalendarEventService
}

// NewCalendarHandler creates a new calendar handler with dependency injection
func NewCalendarHandler(calendarService *services.CalendarEventService) *CalendarHandler {
	return &CalendarHandler{
		calendarService: calendarService,
	}
}

// NewCalendarHandlerWithDefaults creates a new calendar handler with default dependencies
func NewCalendarHandlerWithDefaults(cfg *config.Config) *CalendarHandler {
	return NewCalendarHandler(
		services.NewCalendarEventServiceWithDefaults(cfg),
	)
}

// GetByMonth gets calendar events by month
func (h *CalendarHandler) GetByMonth(c *gin.Context) {
	monthStr := c.Param("month")
	yearStr := c.Param("year")

	month, year, err := services.ParseMonthYear(monthStr, yearStr)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Invalid month or year"})
		return
	}

	events, err := h.calendarService.GetByMonth(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"events": events})
}

// GetById gets calendar event by ID
func (h *CalendarHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	event, err := h.calendarService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"event": event})
}

// Add adds a calendar event
func (h *CalendarHandler) Add(c *gin.Context) {
	var event models.CalendarEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.calendarService.Add(c.Request.Context(), &event); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{
		"message": "Calendar event added successfully",
		"event":   event,
	})
}

// Update updates a calendar event
func (h *CalendarHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var event models.CalendarEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.ID = id
	if err := h.calendarService.Update(c.Request.Context(), &event); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message": "Calendar event updated successfully",
		"event":   event,
	})
}

// Delete deletes a calendar event
func (h *CalendarHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.calendarService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Calendar event deleted successfully"})
}
