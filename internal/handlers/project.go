package handlers

import (
	"strconv"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ProjectHandler handles project routes
type ProjectHandler struct {
	projectService      *services.ProjectService
	authorizationService *services.AuthorizationService
}

// NewProjectHandler creates a new project handler with dependency injection
func NewProjectHandler(
	projectService *services.ProjectService,
	authorizationService *services.AuthorizationService,
) *ProjectHandler {
	return &ProjectHandler{
		projectService:      projectService,
		authorizationService: authorizationService,
	}
}

// NewProjectHandlerWithDefaults creates a new project handler with default dependencies
func NewProjectHandlerWithDefaults() *ProjectHandler {
	return NewProjectHandler(
		services.NewProjectServiceWithDefaults(),
		services.NewAuthorizationServiceWithDefaults(),
	)
}

// GetAll gets all projects
func (h *ProjectHandler) GetAll(c *gin.Context) {
	// This endpoint is accessible to all authenticated users
	// No additional permission checks required
	
	var limit *int
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = &l
		}
	}

	projects, err := h.projectService.GetAll(c.Request.Context(), limit)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"projects": projects})
}

// GetOne gets one project by ID
func (h *ProjectHandler) GetOne(c *gin.Context) {
	id := c.Param("id")
	project, err := h.projectService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if project == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": constants.MsgProjectNotFound})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"project": project})
}

// Add adds a new project
func (h *ProjectHandler) Add(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectService.Add(c.Request.Context(), &project); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{"message": constants.MsgProjectCreated})
}

// Update updates a project
func (h *ProjectHandler) Update(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": constants.MsgInvalidRequest})
		return
	}

	// Check authorization: user must be project owner or member
	if err := h.authorizationService.CanModifyProject(c.Request.Context(), project.ID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectService.Update(c.Request.Context(), &project); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgProjectUpdated})
}

// Delete deletes a project
func (h *ProjectHandler) Delete(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	id := c.Param("id")

	// Check authorization: user must be project owner or member
	if err := h.authorizationService.CanModifyProject(c.Request.Context(), id, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgProjectDeleted})
}

// GetAllWithStatistics gets all projects with their task statistics
func (h *ProjectHandler) GetAllWithStatistics(c *gin.Context) {
	var limit *int
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = &l
		}
	}

	projects, err := h.projectService.GetAllWithStatistics(c.Request.Context(), limit)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{
		"projects":      projects,
		"totalProjects": len(projects),
	})
}

// GetStatistics gets task statistics for a single project
func (h *ProjectHandler) GetStatistics(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "project ID is required"})
		return
	}

	statistics, err := h.projectService.GetStatistics(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"statistics": statistics})
}

// UpdateAgencyContact updates the agency contact for a project
func (h *ProjectHandler) UpdateAgencyContact(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "project ID is required"})
		return
	}

	// Check authorization: user must be project owner or member
	if err := h.authorizationService.CanModifyProject(c.Request.Context(), projectID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	var agencyContact models.AgencyContact
	if err := c.ShouldBindJSON(&agencyContact); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": constants.MsgInvalidRequest})
		return
	}

	if err := h.projectService.UpdateAgencyContact(c.Request.Context(), projectID, &agencyContact); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Agency contact updated successfully"})
}