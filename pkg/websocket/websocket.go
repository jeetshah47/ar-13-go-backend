package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ar-13-go-backend/internal/services"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// In production, you should check the origin
		return true
	},
}

// Client represents a WebSocket client connection
type Client struct {
	conn     *websocket.Conn
	userID   string
	send     chan []byte
	service  *WebSocketService
	mu       sync.Mutex
}

// Message represents a WebSocket message
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WebSocketService manages WebSocket connections
type WebSocketService struct {
	clients          map[*Client]bool
	userConnections  map[string]*Client // Map userID to client
	taskService      *services.TaskService
	projectService   *services.ProjectService
	register         chan *Client
	unregister       chan *Client
	broadcast        chan []byte
	mu               sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(taskService *services.TaskService, projectService *services.ProjectService) *WebSocketService {
	return &WebSocketService{
		clients:         make(map[*Client]bool),
		userConnections: make(map[string]*Client),
		taskService:     taskService,
		projectService:  projectService,
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		broadcast:       make(chan []byte, 256),
	}
}

// HandleConnection handles a new WebSocket connection
func (ws *WebSocketService) HandleConnection(w http.ResponseWriter, r *http.Request, userID string) error {
	log.Printf("[WebSocket] Connection attempt from user %s, remote: %s", userID, r.RemoteAddr)
	
	// Check if this is a WebSocket upgrade request
	if !websocket.IsWebSocketUpgrade(r) {
		log.Printf("[WebSocket] ERROR: Invalid upgrade request for user %s (not a WebSocket upgrade)", userID)
		return http.ErrNotSupported
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] ERROR: Upgrade failed for user %s: %v", userID, err)
		return err
	}

	// Close existing connection if user is already connected
	ws.mu.Lock()
	if existingClient, exists := ws.userConnections[userID]; exists {
		log.Printf("[WebSocket] WARNING: User %s already connected, closing existing connection", userID)
		ws.mu.Unlock()
		existingClient.Close()
		ws.mu.Lock()
	}
	ws.mu.Unlock()

	client := &Client{
		conn:    conn,
		userID:  userID,
		send:    make(chan []byte, 256),
		service: ws,
	}

	log.Printf("[WebSocket] Registering client for user %s", userID)
	ws.register <- client

	// Send authentication success message
	authMsg := Message{
		Type: "authenticated",
		Data: map[string]string{"userId": userID},
	}
	authBytes, err := json.Marshal(authMsg)
	if err != nil {
		log.Printf("[WebSocket] ERROR: Failed to marshal auth message for user %s: %v", userID, err)
	} else {
		client.send <- authBytes
		log.Printf("[WebSocket] Sent authentication message to user %s", userID)
	}

	log.Printf("[WebSocket] SUCCESS: Client connected - user: %s, remote: %s", userID, conn.RemoteAddr())

	// Start client goroutines
	go client.writePump()
	go client.readPump()

	return nil
}

// Run starts the WebSocket service
func (ws *WebSocketService) Run() {
	log.Printf("[WebSocket] Service started, waiting for connections...")
	for {
		select {
		case client := <-ws.register:
			ws.mu.Lock()
			ws.clients[client] = true
			ws.userConnections[client.userID] = client
			totalClients := len(ws.clients)
			ws.mu.Unlock()
			log.Printf("[WebSocket] Client registered - user: %s, total clients: %d", client.userID, totalClients)

		case client := <-ws.unregister:
			ws.mu.Lock()
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				delete(ws.userConnections, client.userID)
				close(client.send)
				totalClients := len(ws.clients)
				ws.mu.Unlock()
				log.Printf("[WebSocket] Client unregistered - user: %s, remaining clients: %d", client.userID, totalClients)
			} else {
				ws.mu.Unlock()
				log.Printf("[WebSocket] WARNING: Attempted to unregister unknown client for user %s", client.userID)
			}

		case message := <-ws.broadcast:
			ws.mu.RLock()
			clientCount := len(ws.clients)
			sentCount := 0
			failedCount := 0
			for client := range ws.clients {
				select {
				case client.send <- message:
					sentCount++
				default:
					failedCount++
					close(client.send)
					delete(ws.clients, client)
					delete(ws.userConnections, client.userID)
					log.Printf("[WebSocket] WARNING: Client %s send channel blocked, removed from connections", client.userID)
				}
			}
			ws.mu.RUnlock()
			log.Printf("[WebSocket] Broadcast complete - total clients: %d, sent: %d, failed: %d, message size: %d bytes", 
				clientCount, sentCount, failedCount, len(message))
		}
	}
}

// SendToUser sends a message to a specific user
func (ws *WebSocketService) SendToUser(userID string, messageType string, data interface{}) error {
	log.Printf("[WebSocket] Sending message to user %s - type: %s", userID, messageType)
	
	ws.mu.RLock()
	client, exists := ws.userConnections[userID]
	ws.mu.RUnlock()

	if !exists {
		log.Printf("[WebSocket] WARNING: User %s not connected, message not sent (type: %s)", userID, messageType)
		return nil // User not connected
	}

	msg := Message{
		Type: messageType,
		Data: data,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WebSocket] ERROR: Failed to marshal message for user %s (type: %s): %v", userID, messageType, err)
		return err
	}

	select {
	case client.send <- msgBytes:
		log.Printf("[WebSocket] Message sent to user %s - type: %s, size: %d bytes", userID, messageType, len(msgBytes))
	default:
		log.Printf("[WebSocket] ERROR: Send channel blocked for user %s, removing connection", userID)
		close(client.send)
		ws.mu.Lock()
		delete(ws.clients, client)
		delete(ws.userConnections, userID)
		ws.mu.Unlock()
	}

	return nil
}

// BroadcastToProjectMembers broadcasts a message to all connected members of a project
func (ws *WebSocketService) BroadcastToProjectMembers(projectID string, messageType string, data interface{}) {
	log.Printf("[WebSocket] Starting broadcast - project: %s, type: %s", projectID, messageType)
	ctx := context.Background()

	// Get project to find members
	project, err := ws.projectService.GetByID(ctx, projectID)
	if err != nil || project == nil {
		log.Printf("[WebSocket] ERROR: Failed to get project %s for broadcast: %v", projectID, err)
		return
	}

	// Collect all user IDs (owner + members)
	userIDs := make(map[string]bool)
	userIDs[project.OwnerID] = true
	for _, memberID := range project.MembersIDs {
		userIDs[memberID] = true
	}
	
	log.Printf("[WebSocket] Project %s has %d total members (owner + %d members)", 
		projectID, len(userIDs), len(project.MembersIDs))

	msg := Message{
		Type: messageType,
		Data: data,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WebSocket] ERROR: Failed to marshal broadcast message for project %s: %v", projectID, err)
		return
	}

	// Send to all connected project members
	ws.mu.RLock()
	targetCount := 0
	sentCount := 0
	failedCount := 0
	for userID := range userIDs {
		if client, exists := ws.userConnections[userID]; exists {
			targetCount++
			select {
			case client.send <- msgBytes:
				sentCount++
				log.Printf("[WebSocket] Broadcast sent to user %s (project: %s)", userID, projectID)
			default:
				failedCount++
				log.Printf("[WebSocket] WARNING: Failed to send broadcast to user %s (channel blocked)", userID)
			}
		} else {
			log.Printf("[WebSocket] User %s (project member) not connected, skipping", userID)
		}
	}
	ws.mu.RUnlock()

	log.Printf("[WebSocket] Broadcast complete - project: %s, type: %s, target members: %d, connected: %d, sent: %d, failed: %d, message size: %d bytes",
		projectID, messageType, len(userIDs), targetCount, sentCount, failedCount, len(msgBytes))
}

// GetConnectedUsersCount returns the number of connected users
func (ws *WebSocketService) GetConnectedUsersCount() int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	count := len(ws.userConnections)
	log.Printf("[WebSocket] Connected users count: %d", count)
	return count
}

// GetConnectedUserIDs returns all connected user IDs
func (ws *WebSocketService) GetConnectedUserIDs() []string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	ids := make([]string, 0, len(ws.userConnections))
	for id := range ws.userConnections {
		ids = append(ids, id)
	}
	return ids
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	log.Printf("[WebSocket] Read pump started for user %s", c.userID)
	defer func() {
		log.Printf("[WebSocket] Read pump stopped for user %s", c.userID)
		c.service.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		log.Printf("[WebSocket] Pong received from user %s", c.userID)
		return nil
	})

	messageCount := 0
	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] ERROR: Unexpected close error for user %s: %v", c.userID, err)
			} else {
				log.Printf("[WebSocket] Connection closed for user %s: %v", c.userID, err)
			}
			break
		}

		messageCount++
		log.Printf("[WebSocket] Message received from user %s - size: %d bytes, total messages: %d", 
			c.userID, len(messageBytes), messageCount)

		// Parse message
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("[WebSocket] ERROR: Failed to parse message from user %s: %v, raw: %s", 
				c.userID, err, string(messageBytes))
			c.sendError("Invalid message format")
			continue
		}

		log.Printf("[WebSocket] Processing message from user %s - type: %s", c.userID, msg.Type)
		// Handle message based on type
		c.handleMessage(&msg)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	log.Printf("[WebSocket] Write pump started for user %s", c.userID)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Printf("[WebSocket] Write pump stopped for user %s", c.userID)
		ticker.Stop()
		c.conn.Close()
	}()

	messageCount := 0
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Printf("[WebSocket] Send channel closed for user %s, sending close message", c.userID)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("[WebSocket] ERROR: Failed to get writer for user %s: %v", c.userID, err)
				return
			}
			w.Write(message)
			messageCount++

			// Add queued messages to the current websocket message
			n := len(c.send)
			if n > 0 {
				log.Printf("[WebSocket] Batching %d queued messages for user %s", n, c.userID)
			}
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
				messageCount++
			}

			if err := w.Close(); err != nil {
				log.Printf("[WebSocket] ERROR: Failed to close writer for user %s: %v", c.userID, err)
				return
			}
			
			log.Printf("[WebSocket] Message sent to user %s - size: %d bytes, total sent: %d", 
				c.userID, len(message), messageCount)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[WebSocket] ERROR: Failed to send ping to user %s: %v", c.userID, err)
				return
			}
			log.Printf("[WebSocket] Ping sent to user %s", c.userID)
		}
	}
}

// handleMessage handles incoming messages from the client
func (c *Client) handleMessage(msg *Message) {
	log.Printf("[WebSocket] Handling message from user %s - type: %s", c.userID, msg.Type)
	
	switch msg.Type {
	case "task:update-status":
		c.handleTaskUpdateStatus(msg.Data)
	default:
		log.Printf("[WebSocket] WARNING: Unknown message type from user %s: %s", c.userID, msg.Type)
		c.sendError("Unknown message type: " + msg.Type)
	}
}

// handleTaskUpdateStatus handles task status update requests
func (c *Client) handleTaskUpdateStatus(data interface{}) {
	log.Printf("[WebSocket] Task status update request from user %s", c.userID)
	
	// Convert data to map
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("[WebSocket] ERROR: Invalid data format for task:update-status from user %s", c.userID)
		c.sendError("Invalid data format for task:update-status")
		return
	}

	projectID, _ := dataMap["projectId"].(string)
	taskID, _ := dataMap["taskId"].(string)
	status, _ := dataMap["status"].(string)

	if projectID == "" || taskID == "" || status == "" {
		log.Printf("[WebSocket] ERROR: Missing required fields from user %s - projectId: %s, taskId: %s, status: %s", 
			c.userID, projectID, taskID, status)
		c.sendError("Missing required fields: projectId, taskId, status")
		return
	}

	log.Printf("[WebSocket] Processing task status update - user: %s, project: %s, task: %s, new status: %s", 
		c.userID, projectID, taskID, status)

	ctx := context.Background()

	// Verify user has access to the project
	project, err := c.service.projectService.GetByID(ctx, projectID)
	if err != nil {
		log.Printf("[WebSocket] ERROR: Failed to get project %s for user %s: %v", projectID, c.userID, err)
		c.sendError("Failed to verify project access")
		return
	}
	if project == nil {
		log.Printf("[WebSocket] ERROR: Project %s not found for user %s", projectID, c.userID)
		c.sendError("Project not found")
		return
	}

	// Check if user is owner or member
	hasAccess := project.OwnerID == c.userID
	if !hasAccess {
		for _, memberID := range project.MembersIDs {
			if memberID == c.userID {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		log.Printf("[WebSocket] ERROR: Access denied - user %s does not have access to project %s", c.userID, projectID)
		c.sendError("Access denied to project")
		return
	}

	log.Printf("[WebSocket] Access verified - user %s has access to project %s, updating task %s", 
		c.userID, projectID, taskID)

	// Update task status
	err = c.service.taskService.UpdateStatus(ctx, projectID, taskID, status)
	if err != nil {
		log.Printf("[WebSocket] ERROR: Failed to update task status - user: %s, project: %s, task: %s, status: %s, error: %v", 
			c.userID, projectID, taskID, status, err)
		c.sendError(err.Error())
		return
	}

	log.Printf("[WebSocket] Task status updated successfully - user: %s, project: %s, task: %s, status: %s", 
		c.userID, projectID, taskID, status)

	// Get updated task
	task, err := c.service.taskService.GetByID(ctx, projectID, taskID)
	if err != nil {
		log.Printf("[WebSocket] WARNING: Failed to get updated task - project: %s, task: %s, error: %v", 
			projectID, taskID, err)
		// Still send success, task was updated
	}

	// Prepare response
	response := map[string]interface{}{
		"projectId": projectID,
		"taskId":    taskID,
		"status":    status,
		"updatedBy": c.userID,
	}
	if task != nil {
		response["task"] = task
	}

	// Send success to the sender
	log.Printf("[WebSocket] Sending success response to user %s for task update", c.userID)
	c.sendMessage("task:update-status:success", response)

	// Broadcast update to all project members who are connected
	log.Printf("[WebSocket] Broadcasting task status update to project %s members", projectID)
	c.service.BroadcastToProjectMembers(projectID, "task:status-updated", response)
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(messageType string, data interface{}) {
	msg := Message{
		Type: messageType,
		Data: data,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WebSocket] ERROR: Failed to marshal message for user %s (type: %s): %v", 
			c.userID, messageType, err)
		return
	}

	select {
	case c.send <- msgBytes:
		log.Printf("[WebSocket] Message queued for user %s - type: %s, size: %d bytes", 
			c.userID, messageType, len(msgBytes))
	default:
		log.Printf("[WebSocket] ERROR: Send channel blocked for user %s, closing connection", c.userID)
		close(c.send)
		c.service.mu.Lock()
		delete(c.service.clients, c)
		delete(c.service.userConnections, c.userID)
		c.service.mu.Unlock()
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errorMsg string) {
	log.Printf("[WebSocket] Sending error to user %s: %s", c.userID, errorMsg)
	c.sendMessage("error", map[string]string{"error": errorMsg})
}

// Close closes the WebSocket connection
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		log.Printf("[WebSocket] Closing connection for user %s", c.userID)
		c.conn.Close()
	}
}

