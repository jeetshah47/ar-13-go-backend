package handlers

import (
	"context"
	"log"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/ar-13-go-backend/pkg/websocket"
	"github.com/gin-gonic/gin"
)

// ActivityLogReplyHandler handles activity log reply routes
type ActivityLogReplyHandler struct {
	replyService      *services.ActivityLogReplyService
	activityLogService *services.ActivityLogService
	projectService    *services.ProjectService
	websocketService *websocket.WebSocketService
}

// NewActivityLogReplyHandler creates a new activity log reply handler with dependency injection
func NewActivityLogReplyHandler(
	replyService *services.ActivityLogReplyService,
	activityLogService *services.ActivityLogService,
	projectService *services.ProjectService,
	websocketService *websocket.WebSocketService,
) *ActivityLogReplyHandler {
	return &ActivityLogReplyHandler{
		replyService:       replyService,
		activityLogService: activityLogService,
		projectService:     projectService,
		websocketService:   websocketService,
	}
}

// NewActivityLogReplyHandlerWithDefaults creates a new activity log reply handler with default dependencies
func NewActivityLogReplyHandlerWithDefaults(websocketService *websocket.WebSocketService) *ActivityLogReplyHandler {
	return NewActivityLogReplyHandler(
		services.NewActivityLogReplyServiceWithDefaults(),
		services.NewActivityLogServiceWithDefaults(),
		services.NewProjectServiceWithDefaults(),
		websocketService,
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

	// Send WebSocket event to all project members (same pattern as task updates)
	if projectID == "" {
		log.Printf("[ActivityLogReply] WARNING: Cannot send WebSocket event - projectID is empty (entityType: %s, entityId: %s)", activityLog.EntityType, activityLog.EntityID)
	} else if h.websocketService == nil {
		log.Printf("[ActivityLogReply] WARNING: Cannot send WebSocket event - WebSocket service is nil")
	} else {
		project, err := h.projectService.GetByID(c.Request.Context(), projectID)
		if err != nil {
			log.Printf("[ActivityLogReply] ERROR: Failed to get project %s for WebSocket broadcast: %v", projectID, err)
		} else if project == nil {
			log.Printf("[ActivityLogReply] ERROR: Project %s not found for WebSocket broadcast", projectID)
		} else {
			// Collect all project member IDs (owner + members) - same pattern as task updates
			userIDs := make(map[string]bool)
			userIDs[project.OwnerID] = true
			for _, memberID := range project.MembersIDs {
				userIDs[memberID] = true
			}

			log.Printf("[ActivityLogReply] Sending WebSocket event to project %s members (owner: %s, members: %v, total: %d)", projectID, project.OwnerID, project.MembersIDs, len(userIDs))

			wsData := map[string]interface{}{
				"activityLogId": request.ActivityLogID,
				"replyId":       reply.ID,
			}

			// Send to each project member individually (same pattern as task updates)
			sentCount := 0
			failedCount := 0
			for memberUserID := range userIDs {
				log.Printf("[ActivityLogReply] Sending activity-log:reply-added event via WebSocket to user: %s", memberUserID)
				if err := h.websocketService.SendToUser(memberUserID, "activity-log:reply-added", wsData); err != nil {
					log.Printf("[ActivityLogReply] ERROR: Failed to send activity-log:reply-added event via WebSocket to user %s: %v", memberUserID, err)
					failedCount++
				} else {
					log.Printf("[ActivityLogReply] SUCCESS: Sent activity-log:reply-added event via WebSocket - userId: %s", memberUserID)
					sentCount++
				}
			}

			log.Printf("[ActivityLogReply] Completed sending WebSocket events for activity log reply - projectId: %s, activityLogId: %s, replyId: %s, totalMembers: %d, sent: %d, failed: %d", projectID, request.ActivityLogID, reply.ID, len(userIDs), sentCount, failedCount)
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

// RegisterWebSocketHandlers registers WebSocket message handlers for activity log replies
func (h *ActivityLogReplyHandler) RegisterWebSocketHandlers(wsService *websocket.WebSocketService) {
	// Handler for creating a reply via WebSocket
	wsService.RegisterMessageHandler("activity-log:reply:create", func(client *websocket.Client, message websocket.Message) error {
		return h.handleCreateReplyWebSocket(client, message)
	})

	// Handler for requesting replies via WebSocket
	wsService.RegisterMessageHandler("activity-log:replies:request", func(client *websocket.Client, message websocket.Message) error {
		return h.handleGetRepliesWebSocket(client, message)
	})

	// Handler for typing indicator
	wsService.RegisterMessageHandler("activity-log:typing:start", func(client *websocket.Client, message websocket.Message) error {
		return h.handleTypingStartWebSocket(client, message)
	})

	// Handler for stopping typing indicator
	wsService.RegisterMessageHandler("activity-log:typing:stop", func(client *websocket.Client, message websocket.Message) error {
		return h.handleTypingStopWebSocket(client, message)
	})
}

// handleCreateReplyWebSocket handles creating a reply via WebSocket
func (h *ActivityLogReplyHandler) handleCreateReplyWebSocket(client *websocket.Client, message websocket.Message) error {
	userID := client.GetUserID()
	if userID == "" {
		return h.sendErrorResponse(client, "User not authenticated")
	}

	// Parse request data
	requestData, ok := message.Data.(map[string]interface{})
	if !ok {
		return h.sendErrorResponse(client, "Invalid request data")
	}

	activityLogID, ok := requestData["activityLogId"].(string)
	if !ok || activityLogID == "" {
		return h.sendErrorResponse(client, "activityLogId is required")
	}

	messageText, ok := requestData["message"].(string)
	if !ok || messageText == "" {
		return h.sendErrorResponse(client, "message is required and cannot be empty")
	}

	ctx := context.Background()

	// Get the activity log to determine the project
	log.Printf("[ActivityLogReply] WebSocket: Looking up activity log with ID: %s", activityLogID)
	activityLog, err := h.activityLogService.GetByID(ctx, activityLogID)
	if err != nil {
		log.Printf("[ActivityLogReply] WebSocket: Error getting activity log: %v", err)
		return h.sendErrorResponse(client, "Failed to retrieve activity log: "+err.Error())
	}
	if activityLog == nil {
		log.Printf("[ActivityLogReply] WebSocket: Activity log not found with ID: %s", activityLogID)
		return h.sendErrorResponse(client, "Activity log not found")
	}

	// Create the reply
	reply := &models.ActivityLogReply{
		ActivityLogID: activityLogID,
		Message:       messageText,
		CreatedBy:     userID,
	}

	// Save to database
	if err := h.replyService.Add(ctx, reply); err != nil {
		log.Printf("[ActivityLogReply] WebSocket: Error creating reply: %v", err)
		return h.sendErrorResponse(client, "Failed to create reply: "+err.Error())
	}

	// Get the reply with user details for response
	var replyResponse *models.ActivityLogReplyResponse
	replies, err := h.replyService.GetByActivityLogID(ctx, activityLogID)
	if err == nil {
		// Find the newly created reply
		for i := range replies {
			if replies[i].ID == reply.ID {
				replyResponse = &replies[i]
				break
			}
		}
	}
	
	// If we couldn't get the full reply, create a basic response
	if replyResponse == nil {
		replyResponse = &models.ActivityLogReplyResponse{
			ActivityLogReply: *reply,
		}
		// Try to populate user details manually
		if reply.CreatedBy != "" {
			userRepo := repos.NewUserRepo()
			user, _ := userRepo.GetByID(ctx, reply.CreatedBy)
			replyResponse.CreatedByUser = user
		}
	}

	// Determine project ID from activity log entity
	var projectID string
	if activityLog.EntityType == models.ActivityLogEntityTypeTask {
		taskRepo := repos.NewTaskRepo()
		task, err := taskRepo.GetByID(ctx, "", activityLog.EntityID)
		if err == nil && task != nil {
			projectID = task.ProjectID
		}
	} else if activityLog.EntityType == models.ActivityLogEntityTypeProject {
		projectID = activityLog.EntityID
	}

	// Get project members for broadcasting
	if projectID != "" && h.websocketService != nil {
		project, err := h.projectService.GetByID(ctx, projectID)
		if err == nil && project != nil {
			// Collect all project member IDs (owner + members)
			userIDs := make(map[string]bool)
			userIDs[project.OwnerID] = true
			for _, memberID := range project.MembersIDs {
				userIDs[memberID] = true
			}

			// Broadcast to all project members using the reply with user details
			wsData := map[string]interface{}{
				"activityLogId": activityLogID,
				"reply":         replyResponse,
			}

			for memberUserID := range userIDs {
				if err := h.websocketService.SendToUser(memberUserID, "activity-log:reply:created", wsData); err != nil {
					log.Printf("[ActivityLogReply] WebSocket: ERROR: Failed to send reply to user %s: %v", memberUserID, err)
				}
			}
			
			log.Printf("[ActivityLogReply] WebSocket: Successfully created and broadcasted reply - activityLogId: %s, replyId: %s, projectId: %s", activityLogID, reply.ID, projectID)
		}
	}

	return nil
}

// handleGetRepliesWebSocket handles requesting replies via WebSocket
func (h *ActivityLogReplyHandler) handleGetRepliesWebSocket(client *websocket.Client, message websocket.Message) error {
	// Parse request data
	requestData, ok := message.Data.(map[string]interface{})
	if !ok {
		return h.sendErrorResponse(client, "Invalid request data")
	}

	activityLogID, ok := requestData["activityLogId"].(string)
	if !ok || activityLogID == "" {
		return h.sendErrorResponse(client, "activityLogId is required")
	}

	ctx := context.Background()

	// Get replies from database
	replies, err := h.replyService.GetByActivityLogID(ctx, activityLogID)
	if err != nil {
		log.Printf("[ActivityLogReply] WebSocket: Error getting replies: %v", err)
		return h.sendErrorResponse(client, "Failed to get replies: "+err.Error())
	}

	// Send replies response
	responseData := map[string]interface{}{
		"activityLogId": activityLogID,
		"replies":      replies,
	}
	responseMessage := websocket.Message{
		Type: "activity-log:replies:response",
		Data: responseData,
	}

	select {
	case client.GetSendChannel() <- responseMessage:
		log.Printf("[ActivityLogReply] WebSocket: Sent %d replies to user %s for activityLogId %s", len(replies), client.GetUserID(), activityLogID)
	default:
		log.Printf("[ActivityLogReply] WebSocket: ERROR: Failed to send replies (channel full)")
		return h.sendErrorResponse(client, "Failed to send replies (channel full)")
	}

	return nil
}

// handleTypingStartWebSocket handles typing start events
func (h *ActivityLogReplyHandler) handleTypingStartWebSocket(client *websocket.Client, message websocket.Message) error {
	userID := client.GetUserID()
	if userID == "" {
		return nil // Silently ignore if user not authenticated
	}

	// Parse request data
	requestData, ok := message.Data.(map[string]interface{})
	if !ok {
		return nil // Silently ignore invalid data
	}

	activityLogID, ok := requestData["activityLogId"].(string)
	if !ok || activityLogID == "" {
		return nil // Silently ignore invalid activityLogId
	}

	ctx := context.Background()

	// Get the activity log to determine the project
	activityLog, err := h.activityLogService.GetByID(ctx, activityLogID)
	if err != nil || activityLog == nil {
		return nil // Silently ignore if activity log not found
	}

	// Determine project ID from activity log entity
	var projectID string
	if activityLog.EntityType == models.ActivityLogEntityTypeTask {
		taskRepo := repos.NewTaskRepo()
		task, err := taskRepo.GetByID(ctx, "", activityLog.EntityID)
		if err == nil && task != nil {
			projectID = task.ProjectID
		}
	} else if activityLog.EntityType == models.ActivityLogEntityTypeProject {
		projectID = activityLog.EntityID
	}

	// Get project members for broadcasting
	if projectID != "" && h.websocketService != nil {
		project, err := h.projectService.GetByID(ctx, projectID)
		if err == nil && project != nil {
			// Collect all project member IDs (owner + members)
			userIDs := make(map[string]bool)
			userIDs[project.OwnerID] = true
			for _, memberID := range project.MembersIDs {
				userIDs[memberID] = true
			}

			// Get user info for typing indicator
			userRepo := repos.NewUserRepo()
			user, _ := userRepo.GetByID(ctx, userID)

			// Broadcast typing indicator to all project members except the sender
			wsData := map[string]interface{}{
				"activityLogId": activityLogID,
				"userId":       userID,
				"userName":     "",
			}
			if user != nil {
				wsData["userName"] = user.Name
			}

			for memberUserID := range userIDs {
				if memberUserID != userID { // Don't send to the person typing
					if err := h.websocketService.SendToUser(memberUserID, "activity-log:typing:start", wsData); err != nil {
						log.Printf("[ActivityLogReply] WebSocket: ERROR: Failed to send typing indicator to user %s: %v", memberUserID, err)
					}
				}
			}
		}
	}

	return nil
}

// handleTypingStopWebSocket handles typing stop events
func (h *ActivityLogReplyHandler) handleTypingStopWebSocket(client *websocket.Client, message websocket.Message) error {
	userID := client.GetUserID()
	if userID == "" {
		return nil // Silently ignore if user not authenticated
	}

	// Parse request data
	requestData, ok := message.Data.(map[string]interface{})
	if !ok {
		return nil // Silently ignore invalid data
	}

	activityLogID, ok := requestData["activityLogId"].(string)
	if !ok || activityLogID == "" {
		return nil // Silently ignore invalid activityLogId
	}

	ctx := context.Background()

	// Get the activity log to determine the project
	activityLog, err := h.activityLogService.GetByID(ctx, activityLogID)
	if err != nil || activityLog == nil {
		return nil // Silently ignore if activity log not found
	}

	// Determine project ID from activity log entity
	var projectID string
	if activityLog.EntityType == models.ActivityLogEntityTypeTask {
		taskRepo := repos.NewTaskRepo()
		task, err := taskRepo.GetByID(ctx, "", activityLog.EntityID)
		if err == nil && task != nil {
			projectID = task.ProjectID
		}
	} else if activityLog.EntityType == models.ActivityLogEntityTypeProject {
		projectID = activityLog.EntityID
	}

	// Get project members for broadcasting
	if projectID != "" && h.websocketService != nil {
		project, err := h.projectService.GetByID(ctx, projectID)
		if err == nil && project != nil {
			// Collect all project member IDs (owner + members)
			userIDs := make(map[string]bool)
			userIDs[project.OwnerID] = true
			for _, memberID := range project.MembersIDs {
				userIDs[memberID] = true
			}

			// Broadcast typing stop to all project members except the sender
			wsData := map[string]interface{}{
				"activityLogId": activityLogID,
				"userId":       userID,
			}

			for memberUserID := range userIDs {
				if memberUserID != userID { // Don't send to the person who stopped typing
					if err := h.websocketService.SendToUser(memberUserID, "activity-log:typing:stop", wsData); err != nil {
						log.Printf("[ActivityLogReply] WebSocket: ERROR: Failed to send typing stop to user %s: %v", memberUserID, err)
					}
				}
			}
		}
	}

	return nil
}

// sendErrorResponse sends an error response to the client
func (h *ActivityLogReplyHandler) sendErrorResponse(client *websocket.Client, errorMsg string) error {
	errorMessage := websocket.Message{
		Type: "error",
		Data: map[string]interface{}{
			"error": errorMsg,
		},
	}
	select {
	case client.GetSendChannel() <- errorMessage:
		return nil
	default:
		log.Printf("[ActivityLogReply] WebSocket: ERROR: Failed to send error message (channel full)")
		return nil
	}
}

