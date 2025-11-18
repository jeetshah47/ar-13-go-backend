# SSE (Server-Sent Events) Client Integration Guide

This document describes how to connect and use the SSE API from a client application.

## Endpoint

```
GET /api/events?token=<jwt-token>
```

For production:
```
GET https://your-domain.com/api/events?token=<jwt-token>
```

**Note:** The default port is 3000, but you can change it by setting the `PORT` environment variable.

## Authentication

The SSE connection requires authentication via JWT token. The token must be included as a query parameter.

### Connection with Token

```javascript
const token = 'your-jwt-token-here';
const eventSource = new EventSource(`/api/events?token=${token}`);
```

**Note:** The server extracts the user ID from the JWT token during connection. The token must be valid and include a valid user ID. If authentication fails, the connection will be rejected with HTTP 401.

## Connection Lifecycle

### 1. Connection Setup

The SSE connection is established through a standard HTTP GET request. The server automatically:
- Validates your JWT token
- Extracts your user ID
- Establishes the SSE connection
- Closes any existing connection for the same user (only one active connection per user)
- Sends an `authenticated` event with your user ID

### 2. Keep-Alive

The server automatically sends heartbeat comments every 30 seconds to keep the connection alive. The EventSource API handles this automatically.

### 3. Disconnection

The connection will close if:
- Client explicitly disconnects
- Network error occurs
- Server shuts down
- Authentication token is invalid or expired
- Connection timeout (5 minutes)

## Event Format

All events follow the SSE format:
```
event: <event-type>
data: <json-data>

```

The EventSource API automatically parses this format and provides events through the `addEventListener` method.

## Client Implementation Examples

### JavaScript/TypeScript (Browser)

#### Basic Example

```javascript
const token = 'your-jwt-token-here';
const eventSource = new EventSource(`/api/events?token=${token}`);

// Connection opened
eventSource.onopen = () => {
  console.log('SSE connected');
};

// Listen for authenticated event
eventSource.addEventListener('authenticated', (event) => {
  const data = JSON.parse(event.data);
  console.log('Authenticated as:', data.userId);
});

// Listen for task status updates
eventSource.addEventListener('task:status-updated', (event) => {
  const data = JSON.parse(event.data);
  console.log('Task status updated:', data);
  // Update UI with new task status
  updateTaskInUI(data.task);
});

// Listen for task update success
eventSource.addEventListener('task:update-status:success', (event) => {
  const data = JSON.parse(event.data);
  console.log('Task update successful:', data);
  // Show success notification
});

// Listen for errors
eventSource.addEventListener('error', (event) => {
  const data = JSON.parse(event.data);
  console.error('Error:', data.error);
});

// Connection error
eventSource.onerror = (error) => {
  console.error('SSE connection error:', error);
  // EventSource automatically attempts to reconnect
};

// Close connection
function disconnect() {
  eventSource.close();
}
```

#### Complete Example with Reconnection

```javascript
class SSEClient {
  constructor(baseURL, token) {
    this.baseURL = baseURL;
    this.token = token;
    this.eventSource = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000; // Start with 1 second
    this.listeners = new Map();
  }

  connect() {
    if (this.eventSource && this.eventSource.readyState !== EventSource.CLOSED) {
      console.log('Already connected');
      return;
    }

    const url = `${this.baseURL}/api/events?token=${this.token}`;
    this.eventSource = new EventSource(url);

    this.eventSource.onopen = () => {
      console.log('SSE connected');
      this.reconnectAttempts = 0;
      this.reconnectDelay = 1000;
    };

    this.eventSource.onerror = (error) => {
      console.error('SSE connection error:', error);
      
      if (this.eventSource.readyState === EventSource.CLOSED) {
        this.attemptReconnect();
      }
    };

    // Register all listeners
    this.listeners.forEach((callback, eventType) => {
      this.eventSource.addEventListener(eventType, (event) => {
        try {
          const data = JSON.parse(event.data);
          callback(data, event);
        } catch (err) {
          console.error('Failed to parse event data:', err);
        }
      });
    });
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

  on(eventType, callback) {
    this.listeners.set(eventType, callback);
    
    if (this.eventSource && this.eventSource.readyState !== EventSource.CLOSED) {
      this.eventSource.addEventListener(eventType, (event) => {
        try {
          const data = JSON.parse(event.data);
          callback(data, event);
        } catch (err) {
          console.error('Failed to parse event data:', err);
        }
      });
    }
  }

  disconnect() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    this.listeners.clear();
  }
}

// Usage
const client = new SSEClient('http://localhost:3000', 'your-token');
client.connect();

client.on('authenticated', (data) => {
  console.log('Authenticated as:', data.userId);
});

client.on('task:status-updated', (data) => {
  console.log('Task updated:', data);
  // Update UI
});
```

### React Hook Example

```typescript
import { useEffect, useRef, useState, useCallback } from 'react';

interface SSEEvent {
  type: string;
  data: any;
}

interface UseSSEOptions {
  token: string;
  baseURL?: string;
  autoReconnect?: boolean;
  maxReconnectAttempts?: number;
}

export function useSSE(options: UseSSEOptions) {
  const { token, baseURL = '', autoReconnect = true, maxReconnectAttempts = 5 } = options;
  const [isConnected, setIsConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<SSEEvent | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectDelayRef = useRef(1000);

  const connect = useCallback(() => {
    if (!token) return;

    // Close existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const url = `${baseURL}/api/events?token=${token}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      console.log('SSE connected');
      setIsConnected(true);
      reconnectAttemptsRef.current = 0;
      reconnectDelayRef.current = 1000;
    };

    eventSource.addEventListener('authenticated', (event) => {
      const data = JSON.parse(event.data);
      setLastEvent({ type: 'authenticated', data });
      console.log('Authenticated as:', data.userId);
    });

    eventSource.addEventListener('task:status-updated', (event) => {
      const data = JSON.parse(event.data);
      setLastEvent({ type: 'task:status-updated', data });
    });

    eventSource.addEventListener('task:update-status:success', (event) => {
      const data = JSON.parse(event.data);
      setLastEvent({ type: 'task:update-status:success', data });
    });

    eventSource.addEventListener('error', (event) => {
      const data = JSON.parse(event.data);
      setLastEvent({ type: 'error', data });
      console.error('SSE error:', data.error);
    });

    eventSource.onerror = (error) => {
      console.error('SSE connection error:', error);
      setIsConnected(false);

      if (autoReconnect && eventSource.readyState === EventSource.CLOSED) {
        if (reconnectAttemptsRef.current < maxReconnectAttempts) {
          reconnectAttemptsRef.current++;
          const delay = reconnectDelayRef.current;
          reconnectDelayRef.current *= 2; // Exponential backoff

          setTimeout(() => {
            connect();
          }, delay);
        } else {
          console.error('Max reconnection attempts reached');
        }
      }
    };
  }, [token, baseURL, autoReconnect, maxReconnectAttempts]);

  useEffect(() => {
    connect();

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, [connect]);

  const updateTaskStatus = useCallback(async (
    projectId: string,
    taskId: string,
    status: string,
    remark?: string
  ) => {
    const response = await fetch(`${baseURL}/api/tasks/update-status/${projectId}/${taskId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        status,
        ...(remark && { remark })
      })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to update task status');
    }

    return response.json();
  }, [baseURL, token]);

  return {
    isConnected,
    lastEvent,
    updateTaskStatus,
    reconnect: connect
  };
}

// Usage in component
function TaskBoard() {
  const token = 'your-jwt-token';
  const { isConnected, lastEvent, updateTaskStatus } = useSSE({
    token,
    baseURL: 'http://localhost:3000',
    autoReconnect: true
  });

  useEffect(() => {
    if (lastEvent?.type === 'task:status-updated') {
      // Update task in UI
      console.log('Task updated:', lastEvent.data);
      // Update your state/UI here
    }
  }, [lastEvent]);

  const handleStatusUpdate = async (taskId: string, projectId: string, status: string) => {
    try {
      await updateTaskStatus(projectId, taskId, status);
      // The SSE event will be received separately
    } catch (error) {
      console.error('Failed to update task status:', error);
    }
  };

  return (
    <div>
      <p>Status: {isConnected ? 'Connected' : 'Disconnected'}</p>
      {/* Your task board UI */}
    </div>
  );
}
```

### React Context Example

```typescript
// SSEContext.tsx
import React, { createContext, useContext, useEffect, useRef, useState } from 'react';

interface SSEContextType {
  isConnected: boolean;
  lastEvent: any;
  updateTaskStatus: (projectId: string, taskId: string, status: string) => Promise<void>;
}

const SSEContext = createContext<SSEContextType | null>(null);

export function SSEProvider({ 
  children, 
  token, 
  baseURL = '' 
}: { 
  children: React.ReactNode; 
  token: string;
  baseURL?: string;
}) {
  const [isConnected, setIsConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<any>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!token) return;

    const url = `${baseURL}/api/events?token=${token}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      setIsConnected(true);
    };

    eventSource.addEventListener('authenticated', (event) => {
      setLastEvent({ type: 'authenticated', data: JSON.parse(event.data) });
    });

    eventSource.addEventListener('task:status-updated', (event) => {
      setLastEvent({ type: 'task:status-updated', data: JSON.parse(event.data) });
    });

    eventSource.addEventListener('task:update-status:success', (event) => {
      setLastEvent({ type: 'task:update-status:success', data: JSON.parse(event.data) });
    });

    eventSource.onerror = () => {
      setIsConnected(false);
    };

    return () => {
      eventSource.close();
    };
  }, [token, baseURL]);

  const updateTaskStatus = async (projectId: string, taskId: string, status: string) => {
    const response = await fetch(`${baseURL}/api/tasks/update-status/${projectId}/${taskId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ status })
    });

    if (!response.ok) {
      throw new Error('Failed to update task status');
    }

    return response.json();
  };

  return (
    <SSEContext.Provider value={{ isConnected, lastEvent, updateTaskStatus }}>
      {children}
    </SSEContext.Provider>
  );
}

export function useSSEContext() {
  const context = useContext(SSEContext);
  if (!context) {
    throw new Error('useSSEContext must be used within SSEProvider');
  }
  return context;
}

// Usage in App.tsx
function App() {
  return (
    <SSEProvider token={getToken()} baseURL="http://localhost:3000">
      <TaskBoard />
    </SSEProvider>
  );
}
```

## Available Events

### Server-Sent Events

#### `authenticated`

Sent immediately after successful connection and authentication.

```typescript
{
  type: 'authenticated',
  data: {
    userId: 'user-id'
  }
}
```

#### `task:status-updated`

Broadcasted to all connected project members when a task status is updated.

```typescript
{
  type: 'task:status-updated',
  data: {
    projectId: 'project-123',
    taskId: 'task-456',
    status: 'in_progress',
    updatedBy: 'user-id',
    task: { /* full task object */ }
  }
}
```

#### `task:update-status:success`

Sent to the user who made the update after successful status change.

```typescript
{
  type: 'task:update-status:success',
  data: {
    projectId: 'project-123',
    taskId: 'task-456',
    status: 'in_progress',
    updatedBy: 'user-id',
    task: { /* full task object */ }
  }
}
```

#### `error`

Sent when an error occurs.

```typescript
{
  type: 'error',
  data: {
    error: 'Error message'
  }
}
```

## Task Status Updates

Since SSE is unidirectional (server to client), task status updates must be done via REST API.

### REST API Endpoint

```
PUT /api/tasks/update-status/:projectId/:taskId
```

**Headers:**
```
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "status": "in_progress",
  "remark": "Optional remark"
}
```

**Response:**
```json
{
  "message": "Task status updated successfully",
  "task": { /* full task object */ }
}
```

**Example:**
```javascript
async function updateTaskStatus(projectId, taskId, status, remark) {
  const response = await fetch(`/api/tasks/update-status/${projectId}/${taskId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      status,
      ...(remark && { remark })
    })
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to update task status');
  }

  const result = await response.json();
  
  // The SSE event will be received separately via EventSource
  // You don't need to manually update the UI here
  
  return result;
}
```

## Valid Task Statuses

- `pending` - Task is pending
- `in_progress` - Task is in progress
- `in_review` - Task is under review
- `completed` - Task is completed
- `accepted` - Task is accepted
- `rejected` - Task is rejected

## Error Handling

### Connection Errors

The EventSource API automatically attempts to reconnect on connection errors. You can handle reconnection manually:

```javascript
eventSource.onerror = (error) => {
  if (eventSource.readyState === EventSource.CLOSED) {
    // Connection closed, attempt manual reconnection
    setTimeout(() => {
      eventSource = new EventSource(`/api/events?token=${token}`);
    }, 5000);
  }
};
```

### Authentication Errors

If authentication fails, the server will return HTTP 401. The EventSource will close the connection.

```javascript
eventSource.onerror = (error) => {
  if (eventSource.readyState === EventSource.CLOSED) {
    // Check if it's an authentication error
    // You may need to refresh the token and reconnect
    refreshToken().then(newToken => {
      eventSource = new EventSource(`/api/events?token=${newToken}`);
    });
  }
};
```

## Best Practices

1. **Token Management**: Always use a valid JWT token. Refresh the token before it expires.

2. **Error Handling**: Always handle connection errors and implement reconnection logic if needed.

3. **Event Validation**: Validate event structure before processing.

4. **Connection State**: Check `eventSource.readyState` before relying on the connection.

5. **Cleanup**: Always close EventSource connections when components unmount or pages are closed.

6. **Rate Limiting**: Be mindful of update frequency to avoid overwhelming the server.

7. **Security**: Never expose JWT tokens in client-side code that's publicly accessible. Use environment variables or secure storage.

8. **Reconnection**: EventSource automatically reconnects, but you may want to implement custom reconnection logic with exponential backoff.

## Troubleshooting

### Connection Fails Immediately

- **Check token**: Ensure the JWT token is valid and not expired
- **Check URL**: Verify the SSE URL is correct (`/api/events?token=<token>`)
- **Check CORS**: Ensure CORS is properly configured on the server
- **Check server logs**: Look for authentication errors in server logs

### Connection Closes Unexpectedly

- **Check network**: Verify network stability
- **Check token expiration**: Token might have expired during connection
- **Check server logs**: Look for errors in server logs
- **Check timeout**: Connection times out after 5 minutes of inactivity

### Events Not Received

- **Check connection state**: Ensure EventSource is in `OPEN` state
- **Check event listeners**: Ensure event listeners are registered before connection
- **Check server logs**: Look for errors in event broadcasting
- **Check event types**: Verify you're listening for the correct event types

### Authentication Errors

- **401 Unauthorized**: Token is missing, invalid, or expired
- **Check token format**: Ensure token is a valid JWT
- **Check token claims**: Verify token contains required user ID

## Comparison with WebSocket

| Feature | SSE | WebSocket |
|---------|-----|-----------|
| Direction | Server → Client only | Bidirectional |
| Protocol | HTTP | WS/WSS |
| Reconnection | Automatic | Manual |
| Complexity | Simple | More complex |
| Browser Support | Excellent | Excellent |
| Proxy Support | Better | Requires special config |
| Scaling | Easier (stateless) | Requires sticky sessions |

## Migration from WebSocket

If you're migrating from WebSocket:

1. Replace `WebSocket` with `EventSource`
2. Replace `ws.send()` with REST API calls
3. Replace `ws.onmessage` with `eventSource.addEventListener()`
4. Update event parsing (SSE uses `event.data` instead of `event.data`)
5. Remove manual reconnection logic (EventSource handles it automatically)

## Server Configuration

The SSE server is configured with:
- **Heartbeat interval**: 30 seconds
- **Connection timeout**: 5 minutes
- **Content-Type**: `text/event-stream`
- **CORS**: Enabled for all origins (configure for production)

These settings ensure reliable connections and prevent resource exhaustion.

