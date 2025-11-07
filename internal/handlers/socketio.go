package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	wsocket "github.com/googollee/go-socket.io/engineio/transport/websocket"
)

// SocketIOHandler handles Socket.IO connections
type SocketIOHandler struct {
	server *socketio.Server
}

// NewSocketIOHandler creates a new Socket.IO handler
func NewSocketIOHandler() *SocketIOHandler {
	// Create Socket.IO server with custom transports
	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true // Allow all origins in development
				},
			},
			&wsocket.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true // Allow all origins in development
				},
			},
		},
	})

	// Handle connection events
	server.OnConnect("/", func(s socketio.Conn) error {
		log.Printf("Socket.IO client connected: %s", s.ID())
		return nil
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Printf("Socket.IO client disconnected: %s, reason: %s", s.ID(), reason)
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Printf("Socket.IO error for client %s: %v", s.ID(), e)
	})

	return &SocketIOHandler{
		server: server,
	}
}

// HandleConnection handles Socket.IO connection requests
func (h *SocketIOHandler) HandleConnection(c *gin.Context) {
	h.server.ServeHTTP(c.Writer, c.Request)
}

// GetServer returns the Socket.IO server instance
func (h *SocketIOHandler) GetServer() *socketio.Server {
	return h.server
}
