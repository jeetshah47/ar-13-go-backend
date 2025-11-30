package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gorilla/websocket"
)

const (
	// Heartbeat interval to keep connections alive
	heartbeatInterval = 30 * time.Second
	// Connection timeout
	connectionTimeout = 5 * time.Minute
	// Write timeout for WebSocket
	writeTimeout = 10 * time.Second
	// Pong wait time
	pongWait = 60 * time.Second
	// Ping period (should be less than pongWait)
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins - adjust in production if needed
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Message represents a WebSocket message
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Client represents a WebSocket client connection
type Client struct {
	userID      string
	conn        *websocket.Conn
	service     *WebSocketService
	send        chan Message
	connectedAt time.Time
	mu          sync.Mutex
	closeOnce   sync.Once
	done        chan struct{} // Signal channel to coordinate goroutine shutdown
	doneOnce    sync.Once     // Ensure done channel is only closed once
}

// MessageHandler handles incoming WebSocket messages
type MessageHandler func(client *Client, message Message) error

// WebSocketService manages WebSocket connections
type WebSocketService struct {
	clients         map[*Client]bool
	userConnections map[string]map[*Client]bool // Map userID to multiple clients (for multiple tabs)
	onlineUsers     map[string]bool             // Track which users are online
	taskService     *services.TaskService
	projectService  *services.ProjectService
	register        chan *Client
	unregister      chan *Client
	messageHandlers map[string]MessageHandler // Map message type to handler
	mu              sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(taskService *services.TaskService, projectService *services.ProjectService) *WebSocketService {
	return &WebSocketService{
		clients:         make(map[*Client]bool),
		userConnections: make(map[string]map[*Client]bool),
		onlineUsers:     make(map[string]bool),
		taskService:     taskService,
		projectService:  projectService,
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		messageHandlers: make(map[string]MessageHandler),
	}
}

// RegisterMessageHandler registers a handler for a specific message type
func (s *WebSocketService) RegisterMessageHandler(messageType string, handler MessageHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messageHandlers[messageType] = handler
}

// HandleConnection handles a new WebSocket connection
func (s *WebSocketService) HandleConnection(w http.ResponseWriter, r *http.Request, userID string) error {
	log.Printf("[WebSocket] Connection attempt from user %s, remote: %s", userID, r.RemoteAddr)

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] ERROR: Failed to upgrade connection for user %s: %v", userID, err)
		return err
	}

	// Create new client
	client := &Client{
		userID:      userID,
		conn:        conn,
		send:        make(chan Message, 256),
		service:     s,
		connectedAt: time.Now(),
		done:        make(chan struct{}),
	}

	log.Printf("[WebSocket] Registering client for user %s", userID)
	s.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()

	// Send initial authentication message
	authMessage := Message{
		Type: "authenticated",
		Data: map[string]string{"userId": userID},
	}
	client.send <- authMessage
	log.Printf("[WebSocket] Sent authentication message to user %s", userID)

	log.Printf("[WebSocket] SUCCESS: Client connected - user: %s, remote: %s", userID, r.RemoteAddr)
	return nil
}

// readPump pumps messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		// Signal writePump to exit (safely close done channel using sync.Once)
		c.doneOnce.Do(func() {
			close(c.done)
		})
		
		// Close the connection safely
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		c.mu.Unlock()
		
		// Unregister the client
		c.service.unregister <- c
	}()

	// Get connection reference
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	
	if conn == nil {
		return
	}

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()
		if conn != nil {
			conn.SetReadDeadline(time.Now().Add(pongWait))
		}
		return nil
	})

	for {
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()
		
		if conn == nil {
			break
		}
		
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] ERROR: Unexpected close error for user %s: %v", c.userID, err)
			} else {
				log.Printf("[WebSocket] Client disconnected - user: %s, error: %v", c.userID, err)
			}
			break
		}

		// Parse incoming message
		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("[WebSocket] ERROR: Failed to parse message from user %s: %v", c.userID, err)
			// Send error response
			errorMessage := Message{
				Type: "error",
				Data: map[string]interface{}{
					"error": "Invalid message format",
				},
			}
			select {
			case c.send <- errorMessage:
			default:
				log.Printf("[WebSocket] ERROR: Failed to send error message to user %s (channel full)", c.userID)
			}
			continue
		}

		// Ignore ping messages (handled by pong handler)
		if message.Type == "ping" {
			pongMessage := Message{
				Type: "pong",
				Data: map[string]interface{}{},
			}
			select {
			case c.send <- pongMessage:
			default:
			}
			continue
		}

		// Route message to appropriate handler
		c.service.mu.RLock()
		handler, exists := c.service.messageHandlers[message.Type]
		c.service.mu.RUnlock()

		if exists {
			// Handle message in a goroutine to avoid blocking readPump
			go func() {
				if err := handler(c, message); err != nil {
					log.Printf("[WebSocket] ERROR: Handler error for message type %s from user %s: %v", message.Type, c.userID, err)
					errorMessage := Message{
						Type: "error",
						Data: map[string]interface{}{
							"error":     err.Error(),
							"messageId":  message.Type,
						},
					}
					select {
					case c.send <- errorMessage:
					default:
						log.Printf("[WebSocket] ERROR: Failed to send error message to user %s (channel full)", c.userID)
					}
				}
			}()
		} else {
			log.Printf("[WebSocket] WARNING: No handler registered for message type %s from user %s", message.Type, c.userID)
			errorMessage := Message{
				Type: "error",
				Data: map[string]interface{}{
					"error":    "Unknown message type",
					"messageId": message.Type,
				},
			}
			select {
			case c.send <- errorMessage:
			default:
				log.Printf("[WebSocket] ERROR: Failed to send error message to user %s (channel full)", c.userID)
			}
		}
	}
}

// writePump pumps messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
		}
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.done:
			// readPump signaled shutdown
			c.mu.Lock()
			if c.conn != nil {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}
			c.mu.Unlock()
			return

		case message, ok := <-c.send:
			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()
			
			if conn == nil {
				return
			}
			
			conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				// Channel closed
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Marshal message to JSON
			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("[WebSocket] ERROR: Failed to marshal message for user %s: %v", c.userID, err)
				return
			}

			// Write message
			if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				log.Printf("[WebSocket] ERROR: Failed to write message to user %s: %v", c.userID, err)
				return
			}

			log.Printf("[WebSocket] Message sent to user %s - type: %s, size: %d bytes", c.userID, message.Type, len(messageBytes))

		case <-ticker.C:
			// Send ping to keep connection alive
			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()
			
			if conn == nil {
				return
			}
			
			conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[WebSocket] ERROR: Failed to send ping to user %s: %v", c.userID, err)
				return
			}
		}
	}
}

// Run starts the WebSocket service
func (s *WebSocketService) Run() {
	log.Printf("[WebSocket] Service started, waiting for connections...")
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			// Add client to user's connection map
			wasOffline := len(s.userConnections[client.userID]) == 0
			if s.userConnections[client.userID] == nil {
				s.userConnections[client.userID] = make(map[*Client]bool)
			}
			s.userConnections[client.userID][client] = true
			// Mark user as online if this is their first connection
			if wasOffline {
				s.onlineUsers[client.userID] = true
			}
			totalClients := len(s.clients)
			userConnectionsCount := len(s.userConnections[client.userID])
			s.mu.Unlock()
			log.Printf("[WebSocket] Client registered - user: %s, total clients: %d, user connections: %d", client.userID, totalClients, userConnectionsCount)
			
			// Broadcast user online status if this is their first connection
			if wasOffline {
				s.broadcastUserPresence(client.userID, true)
			}

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				userID := client.userID
				// Remove client from user's connection map
				if userClients, exists := s.userConnections[userID]; exists {
					delete(userClients, client)
					// If no more connections for this user, mark as offline
					if len(userClients) == 0 {
						delete(s.userConnections, userID)
						s.onlineUsers[userID] = false
					}
				}
				// Close connection safely using sync.Once
				client.Close()
				totalClients := len(s.clients)
				remainingUserConnections := 0
				isNowOffline := false
				if userClients, exists := s.userConnections[userID]; exists {
					remainingUserConnections = len(userClients)
				} else {
					isNowOffline = true
				}
				s.mu.Unlock()
				log.Printf("[WebSocket] Client unregistered - user: %s, remaining clients: %d, user connections: %d", userID, totalClients, remainingUserConnections)
				
				// Broadcast user offline status if this was their last connection
				if isNowOffline {
					s.broadcastUserPresence(userID, false)
				}
			} else {
				s.mu.Unlock()
				log.Printf("[WebSocket] WARNING: Attempted to unregister unknown client for user %s", client.userID)
			}
		}
	}
}

// SendToUser sends an event to a specific user (all their tabs/connections)
func (s *WebSocketService) SendToUser(userID string, eventType string, data interface{}) error {
	log.Printf("[WebSocket] Sending event to user %s - type: %s", userID, eventType)

	s.mu.RLock()
	userClients, exists := s.userConnections[userID]
	s.mu.RUnlock()

	if !exists || len(userClients) == 0 {
		log.Printf("[WebSocket] WARNING: User %s not connected, event not sent (type: %s)", userID, eventType)
		return nil // User not connected
	}

	message := Message{
		Type: eventType,
		Data: data,
	}

	// Send to all connections for this user (multiple tabs)
	sentCount := 0
	failedCount := 0
	clientsToRemove := make([]*Client, 0)
	
	s.mu.RLock()
	for client := range userClients {
		select {
		case client.send <- message:
			sentCount++
			log.Printf("[WebSocket] Event queued for user %s (connection) - type: %s", userID, eventType)
		default:
			failedCount++
			log.Printf("[WebSocket] ERROR: Send channel blocked for user %s, marking for removal", userID)
			clientsToRemove = append(clientsToRemove, client)
		}
	}
	s.mu.RUnlock()

	// Remove blocked clients outside of read lock to avoid deadlock
	if len(clientsToRemove) > 0 {
		s.mu.Lock()
		for _, client := range clientsToRemove {
			// Close connection safely using sync.Once
			client.Close()
			delete(s.clients, client)
			if clients, exists := s.userConnections[userID]; exists {
				delete(clients, client)
				if len(clients) == 0 {
					delete(s.userConnections, userID)
				}
			}
		}
		s.mu.Unlock()
	}

	log.Printf("[WebSocket] Sent event to user %s - type: %s, connections: %d, sent: %d, failed: %d", userID, eventType, len(userClients), sentCount, failedCount)
	return nil
}

// broadcastUserPresence broadcasts user online/offline status to all connected users
func (s *WebSocketService) broadcastUserPresence(userID string, isOnline bool) {
	s.mu.RLock()
	allClients := make([]*Client, 0, len(s.clients))
	for client := range s.clients {
		allClients = append(allClients, client)
	}
	s.mu.RUnlock()

	status := "offline"
	if isOnline {
		status = "online"
	}

	message := Message{
		Type: "user:presence",
		Data: map[string]interface{}{
			"userId": userID,
			"status": status,
		},
	}

	sentCount := 0
	for _, client := range allClients {
		// Don't send to the user themselves
		if client.userID == userID {
			continue
		}

		select {
		case client.send <- message:
			sentCount++
		default:
			log.Printf("[WebSocket] WARNING: Failed to send presence update to user %s (channel full)", client.userID)
		}
	}

	log.Printf("[WebSocket] Broadcasted user presence - user: %s, status: %s, sent to: %d users", userID, status, sentCount)
}

// GetOnlineUsers returns a map of online user IDs
func (s *WebSocketService) GetOnlineUsers() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	onlineUsers := make(map[string]bool)
	for userID, clients := range s.userConnections {
		if len(clients) > 0 {
			onlineUsers[userID] = true
		}
	}
	return onlineUsers
}

// BroadcastToProjectMembers broadcasts an event to all connected members of a project
func (s *WebSocketService) BroadcastToProjectMembers(projectID string, eventType string, data interface{}) {
	log.Printf("[WebSocket] Starting broadcast - project: %s, type: %s", projectID, eventType)
	ctx := context.Background()

	// Get project to find members
	project, err := s.projectService.GetByID(ctx, projectID)
	if err != nil || project == nil {
		log.Printf("[WebSocket] ERROR: Failed to get project %s for broadcast: %v", projectID, err)
		return
	}

	s.BroadcastToProjectMembersWithProject(project, eventType, data)
}

// BroadcastToProjectMembersWithProject broadcasts an event to all connected members of a project
// This version accepts a project object to avoid redundant database reads
func (s *WebSocketService) BroadcastToProjectMembersWithProject(project *models.Project, eventType string, data interface{}) {
	projectID := project.ID
	log.Printf("[WebSocket] Starting broadcast - project: %s, type: %s", projectID, eventType)

	// Collect all user IDs (owner + members)
	userIDs := make(map[string]bool)
	userIDs[project.OwnerID] = true
	for _, memberID := range project.MembersIDs {
		userIDs[memberID] = true
	}

	log.Printf("[WebSocket] Project %s has %d total members (owner + %d members)",
		projectID, len(userIDs), len(project.MembersIDs))

	// Log all connected users for debugging
	s.mu.RLock()
	connectedUserIDs := make([]string, 0, len(s.userConnections))
	connectionCounts := make(map[string]int)
	for uid, clients := range s.userConnections {
		connectedUserIDs = append(connectedUserIDs, uid)
		connectionCounts[uid] = len(clients)
	}
	s.mu.RUnlock()
	log.Printf("[WebSocket] Currently connected users: %v (connection counts: %v)", connectedUserIDs, connectionCounts)

	message := Message{
		Type: eventType,
		Data: data,
	}

	// Send to all connected project members (all their tabs/connections)
	s.mu.RLock()
	targetCount := 0
	sentCount := 0
	failedCount := 0
	clientsToRemove := make([]*Client, 0)
	
	for userID := range userIDs {
		if userClients, exists := s.userConnections[userID]; exists && len(userClients) > 0 {
			targetCount++
			// Send to all connections for this user (multiple tabs)
			userSentCount := 0
			userFailedCount := 0
			for client := range userClients {
				select {
				case client.send <- message:
					userSentCount++
					sentCount++
					log.Printf("[WebSocket] Broadcast sent to user %s (connection) (project: %s)", userID, projectID)
				default:
					userFailedCount++
					failedCount++
					log.Printf("[WebSocket] WARNING: Failed to send broadcast to user %s (channel blocked)", userID)
					clientsToRemove = append(clientsToRemove, client)
				}
			}
			log.Printf("[WebSocket] User %s received broadcast on %d/%d connections", userID, userSentCount, userSentCount+userFailedCount)
		} else {
			log.Printf("[WebSocket] User %s (project member) not connected, skipping", userID)
		}
	}
	s.mu.RUnlock()

	// Remove blocked clients outside of read lock to avoid deadlock
	if len(clientsToRemove) > 0 {
		s.mu.Lock()
		for _, client := range clientsToRemove {
			// Close connection safely using sync.Once
			client.Close()
			delete(s.clients, client)
			if clients, exists := s.userConnections[client.userID]; exists {
				delete(clients, client)
				if len(clients) == 0 {
					delete(s.userConnections, client.userID)
				}
			}
		}
		s.mu.Unlock()
	}

	log.Printf("[WebSocket] Broadcast complete - project: %s, type: %s, target members: %d, connected: %d, sent: %d, failed: %d",
		projectID, eventType, len(userIDs), targetCount, sentCount, failedCount)
}

// GetConnectedUsersCount returns the number of connected users
func (s *WebSocketService) GetConnectedUsersCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := len(s.userConnections)
	log.Printf("[WebSocket] Connected users count: %d", count)
	return count
}

// GetConnectedUserIDs returns all connected user IDs
func (s *WebSocketService) GetConnectedUserIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.userConnections))
	for id := range s.userConnections {
		ids = append(ids, id)
	}
	return ids
}

// GetUserID returns the user ID for this client
func (c *Client) GetUserID() string {
	return c.userID
}

// GetSendChannel returns the send channel for this client
func (c *Client) GetSendChannel() chan Message {
	return c.send
}

// Close closes the WebSocket client connection
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		
		// Signal readPump to exit (which will signal writePump via done channel)
		// Use sync.Once to ensure done channel is only closed once
		c.doneOnce.Do(func() {
			close(c.done)
		})
		
		// Close send channel to signal writePump (if not already closed)
		if c.send != nil {
			select {
			case <-c.send:
				// Channel already closed
			default:
				close(c.send)
			}
			c.send = nil
		}
		
		// Close WebSocket connection
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		
		log.Printf("[WebSocket] Client closed for user %s", c.userID)
	})
}

