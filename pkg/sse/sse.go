package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
)

const (
	// Heartbeat interval to keep connections alive
	heartbeatInterval = 30 * time.Second
	// Connection timeout
	connectionTimeout = 5 * time.Minute
)

// Event represents an SSE event
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Client represents an SSE client connection
type Client struct {
	userID      string
	send        chan Event
	service     *SSEService
	connectedAt time.Time
	mu          sync.Mutex
	closeOnce   sync.Once
}

// SSEService manages SSE connections
type SSEService struct {
	clients         map[*Client]bool
	userConnections map[string]map[*Client]bool // Map userID to multiple clients (for multiple tabs)
	taskService     *services.TaskService
	projectService  *services.ProjectService
	register        chan *Client
	unregister      chan *Client
	mu              sync.RWMutex
}

// NewSSEService creates a new SSE service
func NewSSEService(taskService *services.TaskService, projectService *services.ProjectService) *SSEService {
	return &SSEService{
		clients:         make(map[*Client]bool),
		userConnections: make(map[string]map[*Client]bool),
		taskService:     taskService,
		projectService:  projectService,
		register:        make(chan *Client),
		unregister:      make(chan *Client),
	}
}

// HandleConnection handles a new SSE connection
func (s *SSEService) HandleConnection(w http.ResponseWriter, r *http.Request, userID string) error {
	log.Printf("[SSE] Connection attempt from user %s, remote: %s", userID, r.RemoteAddr)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Create a flusher to ensure data is sent immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("[SSE] ERROR: Stream flushing not supported for user %s", userID)
		return fmt.Errorf("streaming not supported")
	}

	// Allow multiple connections per user (multiple tabs)
	// No need to close existing connections

	// Create new client
	client := &Client{
		userID:      userID,
		send:        make(chan Event, 256),
		service:     s,
		connectedAt: time.Now(),
	}

	log.Printf("[SSE] Registering client for user %s", userID)
	s.register <- client

	// Send initial authentication event
	authEvent := Event{
		Type: "authenticated",
		Data: map[string]string{"userId": userID},
	}
	client.sendEvent(w, flusher, authEvent)
	log.Printf("[SSE] Sent authentication event to user %s", userID)

	log.Printf("[SSE] SUCCESS: Client connected - user: %s, remote: %s", userID, r.RemoteAddr)

	// Use request context for connection lifecycle
	ctx := r.Context()

	// Handle client disconnect in background
	go func() {
		<-ctx.Done()
		log.Printf("[SSE] Client disconnected - user: %s", userID)
		s.unregister <- client
	}()

	// Main event loop
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	timeoutTicker := time.NewTicker(connectionTimeout)
	defer timeoutTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[SSE] Context cancelled for user %s", userID)
			return nil

		case event, ok := <-client.send:
			if !ok {
				log.Printf("[SSE] Send channel closed for user %s", userID)
				return nil
			}
			if err := client.sendEvent(w, flusher, event); err != nil {
				log.Printf("[SSE] ERROR: Failed to send event to user %s: %v", userID, err)
				return err
			}

		case <-ticker.C:
			// Send heartbeat comment to keep connection alive
			if _, err := fmt.Fprintf(w, ": heartbeat\n\n"); err != nil {
				log.Printf("[SSE] ERROR: Failed to send heartbeat to user %s: %v", userID, err)
				return err
			}
			flusher.Flush()

		case <-timeoutTicker.C:
			log.Printf("[SSE] Connection timeout for user %s", userID)
			return nil
		}
	}
}

// sendEvent sends an SSE event to the client
func (c *Client) sendEvent(w http.ResponseWriter, flusher http.Flusher, event Event) error {
	// Marshal event data to JSON
	eventData, err := json.Marshal(event.Data)
	if err != nil {
		log.Printf("[SSE] ERROR: Failed to marshal event data for user %s: %v", c.userID, err)
		return err
	}

	// Write SSE formatted message
	// Format: "event: <type>\ndata: <json>\n\n"
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(eventData)); err != nil {
		return err
	}

	flusher.Flush()
	log.Printf("[SSE] Event sent to user %s - type: %s, size: %d bytes", c.userID, event.Type, len(eventData))
	return nil
}

// Run starts the SSE service
func (s *SSEService) Run() {
	log.Printf("[SSE] Service started, waiting for connections...")
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			// Add client to user's connection map
			if s.userConnections[client.userID] == nil {
				s.userConnections[client.userID] = make(map[*Client]bool)
			}
			s.userConnections[client.userID][client] = true
			totalClients := len(s.clients)
			userConnectionsCount := len(s.userConnections[client.userID])
			s.mu.Unlock()
			log.Printf("[SSE] Client registered - user: %s, total clients: %d, user connections: %d", client.userID, totalClients, userConnectionsCount)

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				// Remove client from user's connection map
				if userClients, exists := s.userConnections[client.userID]; exists {
					delete(userClients, client)
					// If no more connections for this user, remove the map entry
					if len(userClients) == 0 {
						delete(s.userConnections, client.userID)
					}
				}
				// Close channel safely using sync.Once
				client.Close()
				totalClients := len(s.clients)
				remainingUserConnections := 0
				if userClients, exists := s.userConnections[client.userID]; exists {
					remainingUserConnections = len(userClients)
				}
				s.mu.Unlock()
				log.Printf("[SSE] Client unregistered - user: %s, remaining clients: %d, user connections: %d", client.userID, totalClients, remainingUserConnections)
			} else {
				s.mu.Unlock()
				log.Printf("[SSE] WARNING: Attempted to unregister unknown client for user %s", client.userID)
			}
		}
	}
}

// SendToUser sends an event to a specific user (all their tabs/connections)
func (s *SSEService) SendToUser(userID string, eventType string, data interface{}) error {
	log.Printf("[SSE] Sending event to user %s - type: %s", userID, eventType)

	s.mu.RLock()
	userClients, exists := s.userConnections[userID]
	s.mu.RUnlock()

	if !exists || len(userClients) == 0 {
		log.Printf("[SSE] WARNING: User %s not connected, event not sent (type: %s)", userID, eventType)
		return nil // User not connected
	}

	event := Event{
		Type: eventType,
		Data: data,
	}

	// Send to all connections for this user (multiple tabs)
	sentCount := 0
	failedCount := 0
	s.mu.RLock()
	for client := range userClients {
		select {
		case client.send <- event:
			sentCount++
			log.Printf("[SSE] Event queued for user %s (connection) - type: %s", userID, eventType)
		default:
			failedCount++
			log.Printf("[SSE] ERROR: Send channel blocked for user %s, removing connection", userID)
			// Close channel safely using sync.Once
			client.Close()
			// Remove from clients map (will be cleaned up in unregister)
			s.mu.RUnlock()
			s.mu.Lock()
			delete(s.clients, client)
			if clients, exists := s.userConnections[userID]; exists {
				delete(clients, client)
				if len(clients) == 0 {
					delete(s.userConnections, userID)
				}
			}
			s.mu.Unlock()
			s.mu.RLock()
		}
	}
	s.mu.RUnlock()

	log.Printf("[SSE] Sent event to user %s - type: %s, connections: %d, sent: %d, failed: %d", userID, eventType, len(userClients), sentCount, failedCount)
	return nil
}

// BroadcastToProjectMembers broadcasts an event to all connected members of a project
func (s *SSEService) BroadcastToProjectMembers(projectID string, eventType string, data interface{}) {
	log.Printf("[SSE] Starting broadcast - project: %s, type: %s", projectID, eventType)
	ctx := context.Background()

	// Get project to find members
	project, err := s.projectService.GetByID(ctx, projectID)
	if err != nil || project == nil {
		log.Printf("[SSE] ERROR: Failed to get project %s for broadcast: %v", projectID, err)
		return
	}

	s.BroadcastToProjectMembersWithProject(project, eventType, data)
}

// BroadcastToProjectMembersWithProject broadcasts an event to all connected members of a project
// This version accepts a project object to avoid redundant database reads
func (s *SSEService) BroadcastToProjectMembersWithProject(project *models.Project, eventType string, data interface{}) {
	projectID := project.ID
	log.Printf("[SSE] Starting broadcast - project: %s, type: %s", projectID, eventType)

	// Collect all user IDs (owner + members)
	userIDs := make(map[string]bool)
	userIDs[project.OwnerID] = true
	for _, memberID := range project.MembersIDs {
		userIDs[memberID] = true
	}

	log.Printf("[SSE] Project %s has %d total members (owner + %d members)",
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
	log.Printf("[SSE] Currently connected users: %v (connection counts: %v)", connectedUserIDs, connectionCounts)

	event := Event{
		Type: eventType,
		Data: data,
	}

	// Send to all connected project members (all their tabs/connections)
	s.mu.RLock()
	targetCount := 0
	sentCount := 0
	failedCount := 0
	for userID := range userIDs {
		if userClients, exists := s.userConnections[userID]; exists && len(userClients) > 0 {
			targetCount++
			// Send to all connections for this user (multiple tabs)
			userSentCount := 0
			userFailedCount := 0
			for client := range userClients {
				select {
				case client.send <- event:
					userSentCount++
					sentCount++
					log.Printf("[SSE] Broadcast sent to user %s (connection) (project: %s)", userID, projectID)
				default:
					userFailedCount++
					failedCount++
					log.Printf("[SSE] WARNING: Failed to send broadcast to user %s (channel blocked)", userID)
					// Close and remove blocked connection
					client.Close()
					s.mu.RUnlock()
					s.mu.Lock()
					delete(s.clients, client)
					if clients, exists := s.userConnections[userID]; exists {
						delete(clients, client)
						if len(clients) == 0 {
							delete(s.userConnections, userID)
						}
					}
					s.mu.Unlock()
					s.mu.RLock()
				}
			}
			log.Printf("[SSE] User %s received broadcast on %d/%d connections", userID, userSentCount, userSentCount+userFailedCount)
		} else {
			log.Printf("[SSE] User %s (project member) not connected, skipping", userID)
		}
	}
	s.mu.RUnlock()

	log.Printf("[SSE] Broadcast complete - project: %s, type: %s, target members: %d, connected: %d, sent: %d, failed: %d",
		projectID, eventType, len(userIDs), targetCount, sentCount, failedCount)
}

// GetConnectedUsersCount returns the number of connected users
func (s *SSEService) GetConnectedUsersCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := len(s.userConnections)
	log.Printf("[SSE] Connected users count: %d", count)
	return count
}

// GetConnectedUserIDs returns all connected user IDs
func (s *SSEService) GetConnectedUserIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.userConnections))
	for id := range s.userConnections {
		ids = append(ids, id)
	}
	return ids
}

// Close closes the SSE client connection
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.send != nil {
			close(c.send)
			c.send = nil
		}
		log.Printf("[SSE] Client closed for user %s", c.userID)
	})
}
