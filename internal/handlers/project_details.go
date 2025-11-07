package handlers

import (
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ProjectDetailsHandler handles project details routes
type ProjectDetailsHandler struct {
	projectDetailsService *services.ProjectDetailsService
}

// NewProjectDetailsHandler creates a new project details handler
func NewProjectDetailsHandler() *ProjectDetailsHandler {
	return &ProjectDetailsHandler{
		projectDetailsService: services.NewProjectDetailsService(),
	}
}

// Get gets project details
func (h *ProjectDetailsHandler) Get(c *gin.Context) {
	projectID := c.Param("projectId")
	details, err := h.projectDetailsService.Get(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"projectDetails": details})
}

// Add adds project details
func (h *ProjectDetailsHandler) Add(c *gin.Context) {
	projectID := c.Param("projectId")
	var details models.ProjectDetails
	if err := c.ShouldBindJSON(&details); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectDetailsService.Add(c.Request.Context(), projectID, &details); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{"message": "Project details added successfully"})
}

// Update updates project details
func (h *ProjectDetailsHandler) Update(c *gin.Context) {
	projectID := c.Param("projectId")
	var details models.ProjectDetails
	if err := c.ShouldBindJSON(&details); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectDetailsService.Update(c.Request.Context(), projectID, &details); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Project details updated successfully"})
}

// Delete deletes project details
func (h *ProjectDetailsHandler) Delete(c *gin.Context) {
	projectID := c.Param("projectId")
	projectDetailsID := c.Param("projectDetailsId")
	if err := h.projectDetailsService.Delete(c.Request.Context(), projectID, projectDetailsID); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Project details deleted successfully"})
}
