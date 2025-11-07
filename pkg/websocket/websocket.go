package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// WriteWait is the time allowed to write a message to the peer
	WriteWait = 10 * time.Second
	// PongWait is the time allowed to read the next pong message from the peer
	PongWait = 60 * time.Second
	// PingPeriod is the interval for sending ping messages (must be less than PongWait)
	PingPeriod = (PongWait * 9) / 10
	// MaxMessageSize is the maximum message size allowed from peer
	MaxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebSocketService manages WebSocket connections
type WebSocketService struct {
	connectedUsers map[string]*websocket.Conn
	mu             sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		connectedUsers: make(map[string]*websocket.Conn),
	}
}

// HandleConnection handles a new WebSocket connection
func (ws *WebSocketService) HandleConnection(w http.ResponseWriter, r *http.Request, userID string) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	// Close existing connection if user is already connected
	ws.mu.Lock()
	if existingConn, exists := ws.connectedUsers[userID]; exists {
		existingConn.Close()
	}
	ws.connectedUsers[userID] = conn
	ws.mu.Unlock()

	log.Printf("User %s connected via WebSocket", userID)

	// Set up ping/pong handlers to keep connection alive
	conn.SetReadDeadline(time.Now().Add(PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})
	conn.SetReadLimit(MaxMessageSize)

	// Handle messages and disconnect in a goroutine
	go func() {
		defer func() {
			ws.mu.Lock()
			// Only delete if this is still the current connection
			if ws.connectedUsers[userID] == conn {
				delete(ws.connectedUsers, userID)
			}
			ws.mu.Unlock()
			conn.Close()
			log.Printf("User %s disconnected from WebSocket", userID)
		}()

		// Start ping ticker to keep connection alive
		pingTicker := time.NewTicker(PingPeriod)
		defer pingTicker.Stop()

		// Ping goroutine
		go func() {
			for range pingTicker.C {
				conn.SetWriteDeadline(time.Now().Add(WriteWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}()

		// Read messages
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				// Check if it's a close error
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error for user %s: %v", userID, err)
				}
				break
			}

			// Handle different message types
			switch messageType {
			case websocket.TextMessage, websocket.BinaryMessage:
				// Echo message back or process it
				log.Printf("Received message from user %s: %s", userID, string(message))
				// You can add message processing logic here
			case websocket.PingMessage:
				conn.SetWriteDeadline(time.Now().Add(WriteWait))
				conn.WriteMessage(websocket.PongMessage, nil)
			case websocket.CloseMessage:
				return
			}
		}
	}()

	return nil
}

// SendToUser sends a message to a specific user
func (ws *WebSocketService) SendToUser(userID string, message interface{}) error {
	ws.mu.RLock()
	conn, exists := ws.connectedUsers[userID]
	ws.mu.RUnlock()

	if !exists {
		return nil // User not connected
	}

	conn.SetWriteDeadline(time.Now().Add(WriteWait))
	return conn.WriteJSON(message)
}

// GetConnectedUsersCount returns the number of connected users
func (ws *WebSocketService) GetConnectedUsersCount() int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return len(ws.connectedUsers)
}

// GetConnectedUserIDs returns all connected user IDs
func (ws *WebSocketService) GetConnectedUserIDs() []string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	ids := make([]string, 0, len(ws.connectedUsers))
	for id := range ws.connectedUsers {
		ids = append(ids, id)
	}
	return ids
}
