# WebSocket to SSE Migration Guide

## Overview

This project has been migrated from WebSocket to Server-Sent Events (SSE) for real-time communication. SSE provides better scalability, simpler infrastructure, and is well-suited for server-to-client event streaming.

## What Changed

### Backend Changes

1. **New SSE Service** (`pkg/sse/sse.go`)
   - Replaces `pkg/websocket/websocket.go`
   - Manages SSE connections and event broadcasting
   - Similar API to WebSocket service for easy migration

2. **New SSE Handler** (`internal/handlers/sse.go`)
   - Replaces `internal/handlers/websocket.go`
   - Handles SSE connection requests with JWT authentication

3. **Updated Task Handler** (`internal/handlers/task.go`)
   - `UpdateStatus` method now sends SSE events after status updates
   - Broadcasts events to all project members

4. **Updated Routes** (`cmd/server/main.go`)
   - Changed from `/ws` to `/api/events`
   - SSE endpoint: `GET /api/events?token=<jwt-token>`

5. **Updated Notification Handler** (`internal/handlers/notification.go`)
   - Now uses SSE service instead of WebSocket

### API Changes

#### Connection Endpoint
- **Old**: `ws://localhost:3000/ws?token=<jwt-token>`
- **New**: `GET /api/events?token=<jwt-token>`

#### Task Status Updates
- **Old**: Sent via WebSocket message `task:update-status`
- **New**: Use REST API `PUT /api/tasks/update-status/:projectId/:taskId`

## Frontend Migration Guide

### 1. Replace WebSocket with EventSource

**Before (WebSocket):**
```javascript
const ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  // Handle message
};

ws.send(JSON.stringify({
  type: 'task:update-status',
  data: { projectId, taskId, status }
}));
```

**After (SSE):**
```javascript
const eventSource = new EventSource(`/api/events?token=${token}`);

eventSource.addEventListener('authenticated', (event) => {
  const data = JSON.parse(event.data);
  console.log('Authenticated:', data.userId);
});

eventSource.addEventListener('task:status-updated', (event) => {
  const data = JSON.parse(event.data);
  // Handle task status update
});

eventSource.addEventListener('task:update-status:success', (event) => {
  const data = JSON.parse(event.data);
  // Handle success response
});

// For errors
eventSource.addEventListener('error', (event) => {
  const data = JSON.parse(event.data);
  console.error('Error:', data.error);
});
```

### 2. Move Task Updates to REST API

**Before (WebSocket):**
```javascript
ws.send(JSON.stringify({
  type: 'task:update-status',
  data: {
    projectId: 'project-123',
    taskId: 'task-456',
    status: 'in_progress'
  }
}));
```

**After (REST API):**
```javascript
const response = await fetch(`/api/tasks/update-status/${projectId}/${taskId}`, {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    status: 'in_progress',
    remark: 'Optional remark'
  })
});

const result = await response.json();
// The SSE event will be received separately via EventSource
```

### 3. Event Types

All event types remain the same:
- `authenticated` - Sent on connection
- `task:status-updated` - Broadcasted to all project members
- `task:update-status:success` - Sent to the user who made the update
- `error` - Error messages

### 4. React Hook Example

```typescript
import { useEffect, useRef, useState } from 'react';

interface SSEEvent {
  type: string;
  data: any;
}

export function useSSE(token: string) {
  const [isConnected, setIsConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<SSEEvent | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!token) return;

    const eventSource = new EventSource(`/api/events?token=${token}`);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      console.log('SSE connected');
      setIsConnected(true);
    };

    eventSource.addEventListener('authenticated', (event) => {
      const data = JSON.parse(event.data);
      setLastEvent({ type: 'authenticated', data });
    });

    eventSource.addEventListener('task:status-updated', (event) => {
      const data = JSON.parse(event.data);
      setLastEvent({ type: 'task:status-updated', data });
    });

    eventSource.addEventListener('task:update-status:success', (event) => {
      const data = JSON.parse(event.data);
      setLastEvent({ type: 'task:update-status:success', data });
    });

    eventSource.onerror = (error) => {
      console.error('SSE error:', error);
      setIsConnected(false);
      // EventSource automatically reconnects
    };

    return () => {
      eventSource.close();
    };
  }, [token]);

  const updateTaskStatus = async (projectId: string, taskId: string, status: string) => {
    const response = await fetch(`/api/tasks/update-status/${projectId}/${taskId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ status })
    });
    return response.json();
  };

  return { isConnected, lastEvent, updateTaskStatus };
}
```

## Benefits of SSE

1. **Simpler Infrastructure**: Standard HTTP, no special proxy configuration
2. **Better Scalability**: Stateless connections, easier horizontal scaling
3. **Automatic Reconnection**: EventSource handles reconnection automatically
4. **Lower Memory Usage**: ~50% less memory per connection
5. **Better HTTP Compatibility**: Works through all proxies and firewalls

## Breaking Changes

1. **Bidirectional Communication**: Client can no longer send messages through the connection
   - Solution: Use REST API endpoints for client actions

2. **Connection Endpoint**: Changed from `/ws` to `/api/events`
   - Update frontend connection URL

3. **Message Format**: SSE uses `event:` and `data:` format
   - EventSource API handles parsing automatically

## Testing

1. **Connection Test:**
   ```bash
   curl -N "http://localhost:3000/api/events?token=<your-token>"
   ```

2. **Update Task Status:**
   ```bash
   curl -X PUT "http://localhost:3000/api/tasks/update-status/project-id/task-id" \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"status": "in_progress"}'
   ```

## Migration Checklist

- [x] Backend SSE service implementation
- [x] SSE handler implementation
- [x] Task handler updated to send SSE events
- [x] Routes updated
- [x] Notification handler updated
- [ ] Frontend WebSocket client replaced with EventSource
- [ ] Frontend task updates moved to REST API
- [ ] Frontend event handlers updated
- [ ] Testing completed
- [ ] Documentation updated

## Notes

- The WebSocket code has been replaced but can be kept for reference if needed
- SSE connections are automatically reconnected by the browser
- Heartbeat comments are sent every 30 seconds to keep connections alive
- Connection timeout is set to 5 minutes

