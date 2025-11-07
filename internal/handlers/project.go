package handlers

import (
	"strconv"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ProjectHandler handles project routes
type ProjectHandler struct {
	projectService *services.ProjectService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{
		projectService: services.NewProjectService(),
	}
}

// GetAll gets all projects
func (h *ProjectHandler) GetAll(c *gin.Context) {
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
		c.JSON(constants.StatusNotFound, gin.H{"error": "Project not found"})
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

	c.JSON(constants.StatusCreated, gin.H{"message": "Project added successfully"})
}

// Update updates a project
func (h *ProjectHandler) Update(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectService.Update(c.Request.Context(), &project); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Project updated successfully"})
}

// Delete deletes a project
func (h *ProjectHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.projectService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"message": "Project deleted successfully"})
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
