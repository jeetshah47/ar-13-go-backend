package handlers

import (
	"log"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/ar-13-go-backend/pkg/sse"
	"github.com/gin-gonic/gin"
)

// ActivityLogReplyHandler handles activity log reply routes
type ActivityLogReplyHandler struct {
	replyService      *services.ActivityLogReplyService
	activityLogService *services.ActivityLogService
	projectService    *services.ProjectService
	sseService        *sse.SSEService
}

// NewActivityLogReplyHandler creates a new activity log reply handler with dependency injection
func NewActivityLogReplyHandler(
	replyService *services.ActivityLogReplyService,
	activityLogService *services.ActivityLogService,
	projectService *services.ProjectService,
	sseService *sse.SSEService,
) *ActivityLogReplyHandler {
	return &ActivityLogReplyHandler{
		replyService:       replyService,
		activityLogService: activityLogService,
		projectService:     projectService,
		sseService:         sseService,
	}
}

// NewActivityLogReplyHandlerWithDefaults creates a new activity log reply handler with default dependencies
func NewActivityLogReplyHandlerWithDefaults(sseService *sse.SSEService) *ActivityLogReplyHandler {
	return NewActivityLogReplyHandler(
		services.NewActivityLogReplyServiceWithDefaults(),
		services.NewActivityLogServiceWithDefaults(),
		services.NewProjectServiceWithDefaults(),
		sseService,
	)
}

// GetReplies gets all replies for a specific activity log
func (h *ActivityLogReplyHandler) GetReplies(c *gin.Context) {
	var request struct {
		ActivityLogID string `json:"activityLogId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "activityLogId is required"})
		return
	}

	replies, err := h.replyService.GetByActivityLogID(c.Request.Context(), request.ActivityLogID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"replies": replies})
}

// CreateReply creates a new reply to an activity log
func (h *ActivityLogReplyHandler) CreateReply(c *gin.Context) {
	userID := c.GetString("userId")

	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request struct {
		ActivityLogID string `json:"activityLogId" binding:"required"`
		Message       string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate message is not empty
	if request.Message == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Message cannot be empty"})
		return
	}

	// Get the activity log to determine the project
	log.Printf("[ActivityLogReply] Looking up activity log with ID: %s", request.ActivityLogID)
	activityLog, err := h.activityLogService.GetByID(c.Request.Context(), request.ActivityLogID)
	if err != nil {
		log.Printf("[ActivityLogReply] Error getting activity log: %v", err)
		c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to retrieve activity log", "details": err.Error()})
		return
	}
	if activityLog == nil {
		log.Printf("[ActivityLogReply] Activity log not found with ID: %s", request.ActivityLogID)
		// Try to get all activity logs to see what IDs exist (for debugging)
		log.Printf("[ActivityLogReply] Attempting to list recent activity logs to debug ID format...")
		c.JSON(constants.StatusNotFound, gin.H{
			"error": "Activity log not found",
			"activityLogId": request.ActivityLogID,
		})
		return
	}
	log.Printf("[ActivityLogReply] Found activity log - entityType: %s, entityId: %s, id: %s", activityLog.EntityType, activityLog.EntityID, activityLog.ID)

	// Create the reply
	reply := &models.ActivityLogReply{
		ActivityLogID: request.ActivityLogID,
		Message:       request.Message,
		CreatedBy:     userID,
	}

	if err := h.replyService.Add(c.Request.Context(), reply); err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get project members for broadcasting
	// Determine project ID from activity log entity
	var projectID string
	if activityLog.EntityType == models.ActivityLogEntityTypeTask {
		// For task activity logs, get the task to find project
		// Note: GetByID searches by taskID, projectID parameter is not used in the filter
		taskRepo := repos.NewTaskRepo()
		task, err := taskRepo.GetByID(c.Request.Context(), "", activityLog.EntityID)
		if err != nil {
			log.Printf("[ActivityLogReply] Error getting task for project lookup: %v", err)
		}
		if task != nil {
			projectID = task.ProjectID
			log.Printf("[ActivityLogReply] Found projectID from task: %s", projectID)
		} else {
			log.Printf("[ActivityLogReply] Task not found for entityId: %s", activityLog.EntityID)
		}
	} else if activityLog.EntityType == models.ActivityLogEntityTypeProject {
		projectID = activityLog.EntityID
		log.Printf("[ActivityLogReply] Using entityId as projectID: %s", projectID)
	}

	// Send SSE event to all project members (same pattern as task updates)
	if projectID == "" {
		log.Printf("[ActivityLogReply] WARNING: Cannot send SSE event - projectID is empty (entityType: %s, entityId: %s)", activityLog.EntityType, activityLog.EntityID)
	} else if h.sseService == nil {
		log.Printf("[ActivityLogReply] WARNING: Cannot send SSE event - SSE service is nil")
	} else {
		project, err := h.projectService.GetByID(c.Request.Context(), projectID)
		if err != nil {
			log.Printf("[ActivityLogReply] ERROR: Failed to get project %s for SSE broadcast: %v", projectID, err)
		} else if project == nil {
			log.Printf("[ActivityLogReply] ERROR: Project %s not found for SSE broadcast", projectID)
		} else {
			// Collect all project member IDs (owner + members) - same pattern as task updates
			userIDs := make(map[string]bool)
			userIDs[project.OwnerID] = true
			for _, memberID := range project.MembersIDs {
				userIDs[memberID] = true
			}

			log.Printf("[ActivityLogReply] Sending SSE event to project %s members (owner: %s, members: %v, total: %d)", projectID, project.OwnerID, project.MembersIDs, len(userIDs))

			sseData := map[string]interface{}{
				"activityLogId": request.ActivityLogID,
				"replyId":       reply.ID,
			}

			// Send to each project member individually (same pattern as task updates)
			sentCount := 0
			failedCount := 0
			for memberUserID := range userIDs {
				log.Printf("[ActivityLogReply] Sending activity-log:reply-added event via SSE to user: %s", memberUserID)
				if err := h.sseService.SendToUser(memberUserID, "activity-log:reply-added", sseData); err != nil {
					log.Printf("[ActivityLogReply] ERROR: Failed to send activity-log:reply-added event via SSE to user %s: %v", memberUserID, err)
					failedCount++
				} else {
					log.Printf("[ActivityLogReply] SUCCESS: Sent activity-log:reply-added event via SSE - userId: %s", memberUserID)
					sentCount++
				}
			}

			log.Printf("[ActivityLogReply] Completed sending SSE events for activity log reply - projectId: %s, activityLogId: %s, replyId: %s, totalMembers: %d, sent: %d, failed: %d", projectID, request.ActivityLogID, reply.ID, len(userIDs), sentCount, failedCount)
		}
	}

	// Get the reply with user details for response
	replies, err := h.replyService.GetByActivityLogID(c.Request.Context(), request.ActivityLogID)
	if err != nil {
		// If we can't get the reply with user details, return the basic reply
		c.JSON(constants.StatusOK, gin.H{"reply": reply})
		return
	}

	// Find the newly created reply
	var createdReply *models.ActivityLogReplyResponse
	for i := range replies {
		if replies[i].ID == reply.ID {
			createdReply = &replies[i]
			break
		}
	}

	if createdReply != nil {
		c.JSON(constants.StatusOK, gin.H{"reply": createdReply})
	} else {
		c.JSON(constants.StatusOK, gin.H{"reply": reply})
	}
}

