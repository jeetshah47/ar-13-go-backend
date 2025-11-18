# SSE API Documentation

This document describes the Server-Sent Events (SSE) API for real-time event streaming.

## Endpoint

```
GET /api/events?token=<jwt-token>
```

## Authentication

The SSE connection requires authentication via JWT token passed as a query parameter.

### Connection Request

```http
GET /api/events?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response Headers:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Accel-Buffering: no
Access-Control-Allow-Origin: *
```

## Event Format

All events follow the SSE format:

```
event: <event-type>
data: <json-data>

```

Multiple events can be sent in sequence, each separated by a blank line.

## Available Events

### `authenticated`

Sent immediately after successful connection and authentication.

**Event:**
```
event: authenticated
data: {"userId":"user-123"}

```

**Data Structure:**
```json
{
  "userId": "user-123"
}
```

### `task:status-updated`

Broadcasted to all connected project members when a task status is updated.

**Event:**
```
event: task:status-updated
data: {"projectId":"project-123","taskId":"task-456","status":"in_progress","updatedBy":"user-123","task":{...}}

```

**Data Structure:**
```json
{
  "projectId": "project-123",
  "taskId": "task-456",
  "status": "in_progress",
  "updatedBy": "user-123",
  "task": {
    "id": "task-456",
    "subject": "Task title",
    "status": "in_progress",
    ...
  }
}
```

### `task:update-status:success`

Sent to the user who made the update after successful status change.

**Event:**
```
event: task:update-status:success
data: {"projectId":"project-123","taskId":"task-456","status":"in_progress","updatedBy":"user-123","task":{...}}

```

**Data Structure:**
Same as `task:status-updated`

### `error`

Sent when an error occurs.

**Event:**
```
event: error
data: {"error":"Error message"}

```

**Data Structure:**
```json
{
  "error": "Error message"
}
```

## Heartbeat

The server sends heartbeat comments every 30 seconds to keep the connection alive:

```
: heartbeat

```

These are automatically handled by the EventSource API and don't trigger events.

## Connection Lifecycle

1. **Connection**: Client sends GET request with JWT token
2. **Authentication**: Server validates token and extracts user ID
3. **Initial Event**: Server sends `authenticated` event
4. **Keep-Alive**: Server sends heartbeat comments every 30 seconds
5. **Events**: Server sends events as they occur
6. **Disconnection**: Connection closes on timeout (5 min), error, or client disconnect

## Task Status Updates

Since SSE is unidirectional, task status updates must be done via REST API.

### Endpoint

```
PUT /api/tasks/update-status/:projectId/:taskId
```

### Request

**Headers:**
```
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

**Body:**
```json
{
  "status": "in_progress",
  "remark": "Optional remark"
}
```

### Response

**Success (200 OK):**
```json
{
  "message": "Task status updated successfully",
  "task": {
    "id": "task-456",
    "subject": "Task title",
    "status": "in_progress",
    ...
  }
}
```

**Error (400 Bad Request):**
```json
{
  "error": "Invalid task status"
}
```

**Error (401 Unauthorized):**
```json
{
  "error": "Authentication required"
}
```

**Error (403 Forbidden):**
```json
{
  "error": "Access denied to project"
}
```

## Valid Task Statuses

- `pending` - Task is pending
- `in_progress` - Task is in progress
- `in_review` - Task is under review
- `completed` - Task is completed
- `accepted` - Task is accepted
- `rejected` - Task is rejected

## Error Responses

### 401 Unauthorized

Returned when:
- Token is missing
- Token is invalid
- Token is expired

**Response:**
```json
{
  "error": "Authentication required"
}
```

### 500 Internal Server Error

Returned when:
- Server fails to establish SSE connection
- Streaming not supported

**Response:**
```json
{
  "error": "Failed to establish SSE connection: <error message>"
}
```

## Connection Limits

- **One connection per user**: If a user connects with an existing connection, the old connection is closed
- **Connection timeout**: 5 minutes of inactivity
- **Heartbeat interval**: 30 seconds

## Testing

### Using curl

```bash
curl -N "http://localhost:3000/api/events?token=your-jwt-token"
```

The `-N` flag disables buffering, allowing you to see events in real-time.

### Using JavaScript

```javascript
const eventSource = new EventSource('/api/events?token=your-jwt-token');

eventSource.addEventListener('authenticated', (event) => {
  console.log('Authenticated:', JSON.parse(event.data));
});

eventSource.addEventListener('task:status-updated', (event) => {
  console.log('Task updated:', JSON.parse(event.data));
});
```

## Best Practices

1. **Token Refresh**: Refresh tokens before they expire to maintain connection
2. **Error Handling**: Always handle connection errors and implement reconnection
3. **Event Validation**: Validate event data before processing
4. **Cleanup**: Close connections when no longer needed
5. **Rate Limiting**: Be mindful of update frequency

## Migration from WebSocket

If migrating from WebSocket:

1. Change endpoint from `/ws` to `/api/events`
2. Replace `WebSocket` with `EventSource`
3. Move client actions to REST API endpoints
4. Update event listeners to use `addEventListener` instead of `onmessage`

