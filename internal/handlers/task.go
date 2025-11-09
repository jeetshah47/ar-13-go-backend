package handlers

import (
	"strconv"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// TaskHandler handles task routes
type TaskHandler struct {
	taskService          *services.TaskService
	authorizationService *services.AuthorizationService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(cfg *config.Config) *TaskHandler {
	return &TaskHandler{
		taskService:          services.NewTaskService(cfg),
		authorizationService: services.NewAuthorizationService(),
	}
}

// GetAll gets all tasks for a project
func (h *TaskHandler) GetAll(c *gin.Context) {
	projectID := c.Param("projectId")
	tasks, err := h.taskService.GetAll(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"tasks": tasks})
}

// GetAllTaskDetail gets all task details for a project
func (h *TaskHandler) GetAllTaskDetail(c *gin.Context) {
	projectID := c.Param("projectId")
	tasks, err := h.taskService.GetAll(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"tasks": tasks})
}

// GetOneTaskDetail gets one task detail
func (h *TaskHandler) GetOneTaskDetail(c *gin.Context) {
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")
	task, err := h.taskService.GetByID(c.Request.Context(), projectID, taskID)
	if err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if task == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": constants.MsgTaskNotFound})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"task": task})
}

// Add adds a new task
func (h *TaskHandler) Add(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.Add(c.Request.Context(), &task); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{"message": constants.MsgTaskCreated})
}

// AddMultiple adds multiple tasks
func (h *TaskHandler) AddMultiple(c *gin.Context) {
	var req struct {
		Tasks []models.Task `json:"tasks" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.AddMultiple(c.Request.Context(), req.Tasks); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{
		"message": "Tasks added successfully",
		"count":   len(req.Tasks),
	})
}

// Update updates a task
func (h *TaskHandler) Update(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), task.ProjectID, task.ID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.Update(c.Request.Context(), &task); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskUpdated})
}

// UpdateDeadline updates task deadline
func (h *TaskHandler) UpdateDeadline(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	var req struct {
		Deadline string `json:"deadline" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Parse deadline string to time.Time (RFC3339 format)
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": constants.MsgInvalidDeadlineFormat})
		return
	}

	if err := h.taskService.UpdateDeadline(c.Request.Context(), projectID, taskID, deadline); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskDeadlineUpdated})
}

// UpdateProgress updates task progress
func (h *TaskHandler) UpdateProgress(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	var req struct {
		Progress int `json:"progress" binding:"required,min=0,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": constants.MsgInvalidProgressValue})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.UpdateProgress(c.Request.Context(), projectID, taskID, req.Progress); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskProgressUpdated})
}

// UpdateDescription updates task description
func (h *TaskHandler) UpdateDescription(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	var req struct {
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.UpdateDescription(c.Request.Context(), projectID, taskID, req.Description); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskDescriptionUpdated})
}

// UpdateStatus updates task status
func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.UpdateStatus(c.Request.Context(), projectID, taskID, req.Status); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskStatusUpdated})
}

// AddTimeSpent adds time spent entry
func (h *TaskHandler) AddTimeSpent(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	var req struct {
		TimeSpent models.TimeSpent `json:"timeSpent" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTimeLog(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.AddTimeSpent(c.Request.Context(), projectID, taskID, req.TimeSpent); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTimeSpentAdded})
}

// UpdateTimeSpent updates time spent entry
func (h *TaskHandler) UpdateTimeSpent(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")
	indexStr := c.Param("timeSpentIndex")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": constants.MsgInvalidTimeSpentIndex})
		return
	}

	var req struct {
		TimeSpent models.TimeSpent `json:"timeSpent" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTimeLog(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.UpdateTimeSpent(c.Request.Context(), projectID, taskID, index, req.TimeSpent); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTimeSpentUpdated})
}

// RemoveTimeSpent removes time spent entry
func (h *TaskHandler) RemoveTimeSpent(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")
	indexStr := c.Param("timeSpentIndex")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": constants.MsgInvalidTimeSpentIndex})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTimeLog(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.RemoveTimeSpent(c.Request.Context(), projectID, taskID, index); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTimeSpentRemoved})
}

// GetTimeSpent gets time spent entries
func (h *TaskHandler) GetTimeSpent(c *gin.Context) {
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	timeSpent, err := h.taskService.GetTimeSpent(c.Request.Context(), projectID, taskID)
	if err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"timeSpent": timeSpent})
}

// AddFileAttachment adds file attachment
func (h *TaskHandler) AddFileAttachment(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// TODO: Handle file upload (multipart form)
	var attachment models.FileAttachment
	if err := c.ShouldBindJSON(&attachment); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.AddFileAttachment(c.Request.Context(), projectID, taskID, attachment); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgFileAttachmentAdded})
}

// RemoveFileAttachment removes file attachment
func (h *TaskHandler) RemoveFileAttachment(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")
	indexStr := c.Param("fileAttachmentIndex")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": constants.MsgInvalidFileAttachmentIndex})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.RemoveFileAttachment(c.Request.Context(), projectID, taskID, index); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgFileAttachmentRemoved})
}

// GetFileAttachments gets file attachments
func (h *TaskHandler) GetFileAttachments(c *gin.Context) {
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	attachments, err := h.taskService.GetFileAttachments(c.Request.Context(), projectID, taskID)
	if err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"fileAttachments": attachments})
}

// GetActivityLogs gets activity logs
func (h *TaskHandler) GetActivityLogs(c *gin.Context) {
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	task, err := h.taskService.GetByID(c.Request.Context(), projectID, taskID)
	if err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if task == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": constants.MsgTaskNotFound})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"activityLogs": task.ActivityLogs})
}

// Delete deletes a task
func (h *TaskHandler) Delete(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.Delete(c.Request.Context(), projectID, taskID); err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskDeleted})
}

// Assign assigns a task to a user
func (h *TaskHandler) Assign(c *gin.Context) {
	taskID := c.Param("taskId")
	userID := c.Param("userId")

	// Get projectID from query parameter or request body
	var projectID string
	if projectID = c.Query("projectId"); projectID == "" {
		var req struct {
			ProjectID string `json:"projectId"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			projectID = req.ProjectID
		}
	}

	if projectID == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": constants.MsgProjectIDRequired})
		return
	}

	if err := h.taskService.AssignTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskAssigned})
}

// Claim claims a task
func (h *TaskHandler) Claim(c *gin.Context) {
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	// Check authorization: user must be project member to claim tasks
	if err := h.authorizationService.CanClaimTask(c.Request.Context(), projectID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.ClaimTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskClaimed})
}

// GetAssignableUsers gets assignable users
func (h *TaskHandler) GetAssignableUsers(c *gin.Context) {
	users, err := h.taskService.GetAssignableUsers(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"users": users})
}
