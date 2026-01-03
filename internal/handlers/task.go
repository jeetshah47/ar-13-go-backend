package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/ar-13-go-backend/pkg/websocket"
	"github.com/gin-gonic/gin"
)

// TaskHandler handles task routes
type TaskHandler struct {
	taskService          *services.TaskService
	authorizationService *services.AuthorizationService
	websocketService     *websocket.WebSocketService
	notificationService  *services.NotificationService
	storageService       services.StorageServiceInterface
	cfg                  *config.Config
}

// NewTaskHandler creates a new task handler with dependency injection
func NewTaskHandler(
	taskService *services.TaskService,
	authorizationService *services.AuthorizationService,
) *TaskHandler {
	return &TaskHandler{
		taskService:          taskService,
		authorizationService: authorizationService,
	}
}

// NewTaskHandlerWithDefaults creates a new task handler with default implementations
// This is a convenience constructor for backward compatibility
func NewTaskHandlerWithDefaults(cfg *config.Config) *TaskHandler {
	handler := NewTaskHandler(
		services.NewTaskServiceWithDefaults(cfg),
		services.NewAuthorizationServiceWithDefaults(),
	)
	handler.cfg = cfg
	return handler
}

// SetWebSocketService sets the WebSocket service for broadcasting events
func (h *TaskHandler) SetWebSocketService(websocketService *websocket.WebSocketService) {
	h.websocketService = websocketService
}

// SetNotificationService sets the notification service
func (h *TaskHandler) SetNotificationService(notificationService *services.NotificationService) {
	h.notificationService = notificationService
}

// SetStorageService sets the storage service
func (h *TaskHandler) SetStorageService(storageService services.StorageServiceInterface) {
	h.storageService = storageService
}

// GetAll gets all tasks for a project with user details in assignTo field
func (h *TaskHandler) GetAll(c *gin.Context) {
	projectID := c.Param("projectId")
	tasks, err := h.taskService.GetAllWithUserDetails(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"tasks": tasks})
}

// GetAllTaskDetail gets all task details for a project with user details in assignTo field
func (h *TaskHandler) GetAllTaskDetail(c *gin.Context) {
	projectID := c.Param("projectId")
	tasks, err := h.taskService.GetAllWithUserDetails(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"tasks": tasks})
}

// convertFileAttachmentsToResponse converts FileAttachment slice to FileAttachmentResponse slice with token-based URLs
// It generates signed tokens and URLs pointing to filebrowser service's access endpoint
func (h *TaskHandler) convertFileAttachmentsToResponse(c *gin.Context, attachments []models.FileAttachment) []models.FileAttachmentResponse {
	responses := make([]models.FileAttachmentResponse, 0, len(attachments))

	for _, attachment := range attachments {
		response := models.FileAttachmentResponse{
			FileName:     attachment.FileName,
			OriginalName: attachment.OriginalName,
			FileSize:     attachment.FileSize,
			MimeType:     attachment.MimeType,
			UploadDate:   attachment.UploadDate,
			UploadedBy:   attachment.UploadedBy,
			FileURL:      attachment.FileURL, // Keep original relative path
		}

		// Generate direct URLs to filebrowser service if configured
		// Frontend will use JWT tokens to authenticate
		if h.storageService != nil && h.storageService.IsInitialized() && attachment.FileURL != "" {
			if h.cfg != nil && h.cfg.FileBrowserServiceURL != "" {
				// Build direct URL to filebrowser service
				filebrowserURL := strings.TrimSuffix(h.cfg.FileBrowserServiceURL, "/")
				previewURL := filebrowserURL + "/api/download?path=" + url.QueryEscape(attachment.FileURL)
				response.PreviewURL = &previewURL

				// Same for download URL
				downloadURL := filebrowserURL + "/api/download?path=" + url.QueryEscape(attachment.FileURL)
				response.DownloadURL = &downloadURL
			}
		}

		responses = append(responses, response)
	}

	return responses
}

// GetOneTaskDetail gets one task detail with user details in assignTo field
func (h *TaskHandler) GetOneTaskDetail(c *gin.Context) {
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")
	task, err := h.taskService.GetByIDWithUserDetails(c.Request.Context(), projectID, taskID)
	if err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if task == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": constants.MsgTaskNotFound})
		return
	}

	// Build response with converted file attachments that include pre-signed URLs
	response := gin.H{
		"task": task,
	}

	// Convert file attachments to include pre-signed URLs
	if task.FileAttachments != nil && len(task.FileAttachments) > 0 {
		fileAttachmentsResponse := h.convertFileAttachmentsToResponse(c, task.FileAttachments)
		// Create a new task map with converted file attachments
		taskMap := make(map[string]interface{})
		// Convert task to map (we'll use JSON marshaling/unmarshaling for simplicity)
		taskJSON, err := json.Marshal(task)
		if err == nil {
			if err := json.Unmarshal(taskJSON, &taskMap); err == nil {
				taskMap["fileAttachments"] = fileAttachmentsResponse
				response["task"] = taskMap
			} else {
				log.Printf("Warning: Failed to unmarshal task JSON: %v", err)
			}
		} else {
			log.Printf("Warning: Failed to marshal task to JSON: %v", err)
		}
	}

	c.JSON(constants.StatusOK, response)
}

// Add adds a new task
func (h *TaskHandler) Add(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize and validate status if provided
	if task.Status != "" {
		normalizedStatus := constants.NormalizeTaskStatus(task.Status)
		if normalizedStatus == "" {
			c.JSON(constants.StatusBadRequest, gin.H{"error": "Invalid task status. Valid statuses are: pending, in_progress, in_review, completed, accepted, rejected"})
			return
		}
		task.Status = normalizedStatus
	} else {
		// Default to pending if not provided
		task.Status = constants.GetTaskStatusString(constants.TaskStatusPending)
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

	// Normalize and validate statuses for all tasks
	for i := range req.Tasks {
		if req.Tasks[i].Status != "" {
			normalizedStatus := constants.NormalizeTaskStatus(req.Tasks[i].Status)
			if normalizedStatus == "" {
				c.JSON(constants.StatusBadRequest, gin.H{"error": "Invalid task status for task at index " + strconv.Itoa(i) + ". Valid statuses are: pending, in_progress, in_review, completed, accepted, rejected"})
				return
			}
			req.Tasks[i].Status = normalizedStatus
		} else {
			// Default to pending if not provided
			req.Tasks[i].Status = constants.GetTaskStatusString(constants.TaskStatusPending)
		}
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

	// Get projectId and taskId from URL parameters
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	if projectID == "" || taskID == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "projectId and taskId are required"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set projectId and taskId from URL parameters (ignore any values from request body)
	task.ProjectID = projectID
	task.ID = taskID

	// Normalize and validate status if provided
	if task.Status != "" {
		normalizedStatus := constants.NormalizeTaskStatus(task.Status)
		if normalizedStatus == "" {
			c.JSON(constants.StatusBadRequest, gin.H{"error": "Invalid task status. Valid statuses are: pending, in_progress, in_review, completed, accepted, rejected"})
			return
		}
		task.Status = normalizedStatus
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), projectID, taskID, userID); err != nil {
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
		Status      string  `json:"status" binding:"required"`
		Remark      *string `json:"remark,omitempty"`
		AdminBypass *bool   `json:"adminBypass,omitempty"` // Optional admin bypass flag
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize and validate status
	normalizedStatus := constants.NormalizeTaskStatus(req.Status)
	if normalizedStatus == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Invalid task status. Valid statuses are: pending, in_progress, in_review, completed, accepted, rejected"})
		return
	}

	// Check authorization: user must be assigned to task OR project owner/member
	if err := h.authorizationService.CanModifyTask(c.Request.Context(), projectID, taskID, userID); err != nil {
		c.JSON(constants.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Check if user is admin
	isAdmin, err := h.authorizationService.IsAdmin(c.Request.Context(), userID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to check admin status"})
		return
	}

	// Determine admin bypass - only allow if user is admin and flag is set
	adminBypass := false
	if isAdmin && req.AdminBypass != nil && *req.AdminBypass {
		adminBypass = true
	}

	// Get project before update (for notifications and broadcasting)
	projectService := services.NewProjectServiceWithDefaults()
	project, err := projectService.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Failed to get project: " + err.Error()})
		return
	}
	if project == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Get task before update to know old status
	existingTask, err := h.taskService.GetByID(c.Request.Context(), projectID, taskID)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Failed to get existing task: " + err.Error()})
		return
	}
	if existingTask == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	oldStatus := existingTask.Status

	// Update task status (with admin bypass check)
	if err := h.taskService.UpdateStatus(c.Request.Context(), projectID, taskID, normalizedStatus, req.Remark, adminBypass); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated task for notification with user details
	task, err := h.taskService.GetByIDWithUserDetails(c.Request.Context(), projectID, taskID)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Failed to get updated task: " + err.Error()})
		return
	}

	// Create notifications in database and send via WebSocket for all project members
	if h.notificationService != nil && h.websocketService != nil {
		log.Printf("[WebSocket-NOTIFICATION] Preparing to create notifications for task status update - projectId: %s, taskId: %s, oldStatus: %s, newStatus: %s, updatedBy: %s", projectID, taskID, oldStatus, normalizedStatus, userID)

		// Collect all project member IDs (owner + members)
		userIDs := make(map[string]bool)
		userIDs[project.OwnerID] = true
		for _, memberID := range project.MembersIDs {
			userIDs[memberID] = true
		}

		// Get updater name for notification message
		userRepo := repos.NewUserRepo()
		updater, _ := userRepo.GetByID(context.Background(), userID)
		updaterName := "Someone"
		if updater != nil {
			updaterName = updater.Name
		}

		// Create notification for each project member and send notifications-available event via WebSocket
		for memberUserID := range userIDs {
			// Skip if status didn't change (already handled in service)
			if oldStatus == normalizedStatus {
				continue
			}

			message := fmt.Sprintf("The status of task '%s' in project '%s' has been updated from '%s' to '%s' by %s.", task.Subject, project.Title, oldStatus, normalizedStatus, updaterName)

			notification := &models.Notification{
				Title:             fmt.Sprintf("Task Status Updated: %s", task.Subject),
				Message:           message,
				Type:              models.NotificationTypeTaskUpdated,
				UserID:            memberUserID,
				RelatedEntityID:   taskID,
				RelatedEntityType: models.RelatedEntityTypeTask,
				IsRead:            false,
			}

			// Store notification in database first
			if err := h.notificationService.CreateNotification(c.Request.Context(), notification); err != nil {
				log.Printf("[WebSocket-NOTIFICATION] ERROR: Failed to create task status update notification for user %s: %v", memberUserID, err)
				continue
			}
			log.Printf("[WebSocket-NOTIFICATION] Created task status update notification in database - notificationId: %s, taskId: %s, userId: %s, oldStatus: %s, newStatus: %s", notification.ID, taskID, memberUserID, oldStatus, normalizedStatus)

			// Send simple notifications-available event via WebSocket (client will fetch notifications via API)
			wsData := map[string]interface{}{
				"userId": memberUserID,
			}

			log.Printf("[WebSocket-NOTIFICATION] Sending notifications-available event via WebSocket to user: %s", memberUserID)
			if err := h.websocketService.SendToUser(memberUserID, "notifications-available", wsData); err != nil {
				log.Printf("[WebSocket-NOTIFICATION] ERROR: Failed to send notifications-available event via WebSocket to user %s: %v", memberUserID, err)
			} else {
				log.Printf("[WebSocket-NOTIFICATION] SUCCESS: Sent notifications-available event via WebSocket - userId: %s, type: notifications-available", memberUserID)
			}
		}

		log.Printf("[WebSocket-NOTIFICATION] Completed creating and sending notifications for task status update - projectId: %s, taskId: %s, totalMembers: %d", projectID, taskID, len(userIDs))
	} else {
		if h.notificationService == nil {
			log.Printf("[WebSocket-NOTIFICATION] WARNING: Notification service not available, skipping notifications for task status update")
		}
		if h.websocketService == nil {
			log.Printf("[WebSocket-NOTIFICATION] WARNING: WebSocket service not available, skipping WebSocket events for task status update")
		}
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskStatusUpdated, "task": task})
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

	var attachment models.FileAttachment

	// Option 1: Upload new file (multipart form)
	if file, err := c.FormFile("file"); err == nil {
		// Handle file upload
		path := c.DefaultPostForm("path", "")
		objectName := file.Filename
		if path != "" {
			objectName = path + "/" + file.Filename
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(constants.StatusInternalServerError, gin.H{"error": "failed to open file"})
			return
		}
		defer src.Close()

		// Upload to MinIO if storage service is available and initialized
		if h.storageService != nil && h.storageService.IsInitialized() {
			contentType := file.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			err = h.storageService.UploadFile(
				c.Request.Context(),
				objectName,
				src,
				file.Size,
				contentType,
			)
			if err != nil {
				c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		attachment = models.FileAttachment{
			FileName:     file.Filename,
			OriginalName: file.Filename,
			FileSize:     file.Size,
			MimeType:     file.Header.Get("Content-Type"),
			UploadDate:   time.Now(),
			UploadedBy:   userID,
			FileURL:      objectName, // Store the path, generate URL when needed
		}
	} else {
		// Option 2: Link existing file from NAS (JSON body with path)
		// Just save the path to database - no file validation needed
		// The file is assumed to exist on NAS
		if err := c.ShouldBindJSON(&attachment); err != nil {
			c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate that FileURL is provided
		if attachment.FileURL == "" {
			c.JSON(constants.StatusBadRequest, gin.H{"error": "FileURL is required when linking a file"})
			return
		}

		// Set upload info
		attachment.UploadDate = time.Now()
		attachment.UploadedBy = userID

		// Use LinkFileAttachment to ensure only one linked file per task
		if err := h.taskService.LinkFileAttachment(c.Request.Context(), projectID, taskID, attachment); err != nil {
			c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(constants.StatusOK, gin.H{"message": constants.MsgFileAttachmentAdded})
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

	// Convert to response format with pre-signed URLs
	responseAttachments := h.convertFileAttachmentsToResponse(c, attachments)
	c.JSON(constants.StatusOK, gin.H{"fileAttachments": responseAttachments})
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

// Transfer transfers a task to another user (admin only)
func (h *TaskHandler) Transfer(c *gin.Context) {
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	// Get user ID from context (set by auth middleware)
	adminUserID := middleware.GetUserID(c)
	if adminUserID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	// Parse request body to get target user ID
	var requestBody struct {
		UserId string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	// Transfer task using service
	if err := h.taskService.TransferTask(c.Request.Context(), projectID, taskID, requestBody.UserId, adminUserID); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": constants.MsgTaskTransferred})
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

// GetStatuses returns all available task statuses with their metadata
func (h *TaskHandler) GetStatuses(c *gin.Context) {
	statuses, err := h.taskService.GetStatuses(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"statuses": statuses,
		"total":    len(statuses),
	})
}
