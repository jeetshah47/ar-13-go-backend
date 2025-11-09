# WebSocket API Documentation

This document describes how to connect and use the WebSocket API from a client application.

## Endpoint

```
ws://localhost:3000/ws
```

For production with HTTPS:
```
wss://your-domain.com/ws
```

**Note:** The default port is 3000, but you can change it by setting the `PORT` environment variable.

## Authentication

The WebSocket connection requires authentication via JWT token. The token must be included in the connection request.

### Connection with Token

You can pass the token as a query parameter:

```javascript
const token = 'your-jwt-token-here';
const ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);
```

Or via Authorization header (if your WebSocket library supports it):

```javascript
// Note: Native WebSocket API doesn't support custom headers in browsers
// Use query parameter instead
const ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);
```

**Note:** The server extracts the user ID from the JWT token during connection. The token must be valid and include a valid user ID. If authentication fails, the connection will be rejected with HTTP 401.

## Connection Lifecycle

### 1. Connection Setup

The WebSocket connection is established through an HTTP upgrade request. The server automatically:
- Validates your JWT token
- Extracts your user ID
- Establishes the WebSocket connection
- Closes any existing connection for the same user (only one active connection per user)
- Sends an `authenticated` message with your user ID

### 2. Keep-Alive

The server automatically sends ping messages every 54 seconds (pongWait * 9/10). The client must respond with pong messages. The native WebSocket API handles this automatically.

### 3. Disconnection

The connection will close if:
- Client explicitly disconnects
- Network error occurs
- Server shuts down
- Authentication token is invalid or expired
- No pong response received within 60 seconds

## Message Format

All messages are JSON objects with the following structure:

```typescript
interface Message {
  type: string;
  data: any;
}
```

### Example Messages

**Server to Client:**
```json
{
  "type": "authenticated",
  "data": {
    "userId": "30EpwymEH1RMMNjk58zOfwqLgyE2"
  }
}
```

**Client to Server:**
```json
{
  "type": "task:update-status",
  "data": {
    "projectId": "project-123",
    "taskId": "task-456",
    "status": "In Progress"
  }
}
```

## Client Implementation Examples

### JavaScript/TypeScript (Browser)

```javascript
const token = 'your-jwt-token-here';
const ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

// Connection opened
ws.onopen = () => {
  console.log('WebSocket connected');
};

// Listen for messages
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received message:', message);
  
  switch (message.type) {
    case 'authenticated':
      console.log('Authenticated as:', message.data.userId);
      break;
    case 'task:status-updated':
      console.log('Task status updated:', message.data);
      // Update UI
      break;
    case 'error':
      console.error('Error:', message.data.error);
      break;
    default:
      console.log('Unknown message type:', message.type);
  }
};

// Connection closed
ws.onclose = (event) => {
  console.log('WebSocket disconnected:', event.code, event.reason);
  // Implement reconnection logic if needed
};

// Connection error
ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

// Send a message
function sendMessage(type, data) {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type, data }));
  } else {
    console.error('WebSocket is not open');
  }
}

// Example: Update task status
sendMessage('task:update-status', {
  projectId: 'project-123',
  taskId: 'task-456',
  status: 'In Progress'
});
```

### React Hook Example

```typescript
import { useEffect, useRef, useState } from 'react';

interface Message {
  type: string;
  data: any;
}

export function useWebSocket(token: string) {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<Message | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!token) return;

    const ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      const message: Message = JSON.parse(event.data);
      setLastMessage(message);
      
      if (message.type === 'authenticated') {
        console.log('Authenticated as:', message.data.userId);
      }
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
      // Optional: Implement reconnection logic
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    return () => {
      ws.close();
    };
  }, [token]);

  const sendMessage = (type: string, data: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, data }));
    }
  };

  return { isConnected, lastMessage, sendMessage };
}

// Usage in component
function TaskBoard() {
  const token = 'your-jwt-token';
  const { isConnected, lastMessage, sendMessage } = useWebSocket(token);

  useEffect(() => {
    if (lastMessage?.type === 'task:status-updated') {
      // Update task in UI
      console.log('Task updated:', lastMessage.data);
    }
  }, [lastMessage]);

  const handleStatusUpdate = (taskId: string, projectId: string, status: string) => {
    sendMessage('task:update-status', {
      projectId,
      taskId,
      status,
    });
  };

  return (
    <div>
      <p>Status: {isConnected ? 'Connected' : 'Disconnected'}</p>
      {/* Your task board UI */}
    </div>
  );
}
```

### Node.js Example

```javascript
const WebSocket = require('ws');

const token = 'your-jwt-token-here';
const ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

ws.on('open', () => {
  console.log('WebSocket connected');
});

ws.on('message', (data) => {
  const message = JSON.parse(data.toString());
  console.log('Received:', message);
  
  if (message.type === 'authenticated') {
    console.log('Authenticated as:', message.data.userId);
  }
});

ws.on('close', () => {
  console.log('WebSocket disconnected');
});

ws.on('error', (error) => {
  console.error('WebSocket error:', error);
});

// Send a message
function sendMessage(type, data) {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type, data }));
  }
}

// Example: Update task status
sendMessage('task:update-status', {
  projectId: 'project-123',
  taskId: 'task-456',
  status: 'In Progress'
});
```

## Available Events

### Client-Sent Events

#### `task:update-status`

Updates the status of a task (used for drag-and-drop functionality).

**Request:**
```json
{
  "type": "task:update-status",
  "data": {
    "projectId": "string (required)",
    "taskId": "string (required)",
    "status": "string (required)"
  }
}
```

**Success Response:**
```json
{
  "type": "task:update-status:success",
  "data": {
    "projectId": "project-123",
    "taskId": "task-456",
    "status": "In Progress",
    "updatedBy": "user-id",
    "task": { /* full task object */ }
  }
```

**Error Response:**
```json
{
  "type": "error",
  "data": {
    "error": "Error message"
  }
}
```

**Possible Errors:**
- `"User not authenticated"` - User is not authenticated
- `"Missing required fields: projectId, taskId, status"` - Missing required fields
- `"Project not found"` - Project doesn't exist
- `"Access denied to project"` - User doesn't have access to the project
- `"task not found"` - Task doesn't exist
- Other database errors

### Server-Sent Events

#### `authenticated`

Sent immediately after successful connection and authentication.

```json
{
  "type": "authenticated",
  "data": {
    "userId": "user-id"
  }
}
```

#### `task:status-updated`

Broadcasted to all connected project members when a task status is updated.

```json
{
  "type": "task:status-updated",
  "data": {
    "projectId": "project-123",
    "taskId": "task-456",
    "status": "In Progress",
    "updatedBy": "user-id",
    "task": { /* full task object */ }
  }
}
```

#### `error`

Sent when an error occurs.

```json
{
  "type": "error",
  "data": {
    "error": "Error message"
  }
}
```

## Reconnection Strategy

The native WebSocket API doesn't have built-in reconnection. You should implement your own reconnection logic:

```javascript
class WebSocketClient {
  constructor(url, token) {
    this.url = url;
    this.token = token;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000; // Start with 1 second
  }

  connect() {
    this.ws = new WebSocket(`${this.url}?token=${this.token}`);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.reconnectDelay = 1000;
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.attemptReconnect();
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      
      setTimeout(() => {
        this.connect();
      }, this.reconnectDelay);
      
      // Exponential backoff
      this.reconnectDelay *= 2;
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  handleMessage(message) {
    // Handle messages
    console.log('Received:', message);
  }

  send(type, data) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }));
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

// Usage
const client = new WebSocketClient('ws://localhost:3000/ws', 'your-token');
client.connect();
```

## Best Practices

1. **Token Management**: Always use a valid JWT token. Refresh the token before it expires.

2. **Error Handling**: Always handle connection errors and implement reconnection logic.

3. **Message Validation**: Validate message structure before processing.

4. **Connection State**: Check `ws.readyState === WebSocket.OPEN` before sending messages.

5. **Cleanup**: Always close WebSocket connections when components unmount or pages are closed.

6. **Rate Limiting**: Be mindful of message frequency to avoid overwhelming the server.

7. **Security**: Never expose JWT tokens in client-side code that's publicly accessible. Use environment variables or secure storage.

## Troubleshooting

### Connection Fails Immediately

- **Check token**: Ensure the JWT token is valid and not expired
- **Check URL**: Verify the WebSocket URL is correct (`ws://localhost:3000/ws`)
- **Check CORS**: Ensure CORS is properly configured on the server
- **Check server logs**: Look for authentication errors in server logs

### Connection Closes Unexpectedly

- **Check network**: Verify network stability
- **Check token expiration**: Token might have expired during connection
- **Check server logs**: Look for errors in server logs
- **Check pong responses**: Ensure client is responding to ping messages (handled automatically by browser)

### Messages Not Received

- **Check connection state**: Ensure WebSocket is in `OPEN` state
- **Check message format**: Verify messages are valid JSON
- **Check server logs**: Look for errors in message handling
- **Check event handlers**: Ensure `onmessage` handler is properly set up

### Authentication Errors

- **401 Unauthorized**: Token is missing, invalid, or expired
- **Check token format**: Ensure token is a valid JWT
- **Check token claims**: Verify token contains required user ID

## Server Configuration

The WebSocket server is configured with:
- **Write timeout**: 10 seconds
- **Read timeout (pong wait)**: 60 seconds
- **Ping interval**: 54 seconds
- **Max message size**: 512 KB

These settings ensure reliable connections and prevent resource exhaustion.
