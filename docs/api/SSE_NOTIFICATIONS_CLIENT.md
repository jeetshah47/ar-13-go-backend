# SSE Notifications Client Documentation

## Overview

The application uses Server-Sent Events (SSE) to deliver real-time notifications to clients. When task-related events occur (task assigned, updated, status changed, time logged), notifications are automatically sent to relevant users via SSE and stored in the database for later retrieval.

## Connection Setup

### Endpoint

Connect to the SSE endpoint using the authenticated endpoint with a valid JWT token.

### Authentication

The SSE connection requires authentication via JWT token. The token can be provided either:
- As a query parameter: `token=<jwt-token>`
- In the Authorization header: `Authorization: Bearer <jwt-token>`

### Connection Headers

The server will respond with SSE-specific headers:
- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`
- `X-Accel-Buffering: no` (prevents proxy buffering)

## Event Types

### Authentication Event

Immediately after successful connection, the server sends an authentication confirmation event.

**Event Type:** `authenticated`

**Data Structure:**
- `userId`: The authenticated user's ID

**Purpose:** Confirms successful connection and identifies the authenticated user.

### Notification Event

All task-related notifications are delivered through this event type.

**Event Type:** `notification`

**When Received:**
- Task assigned to a user
- Task updated (any field change)
- Task status updated
- Time logged on a task

## Notification Data Structure

### Common Fields

All notification events contain the following structure:

**Root Level:**
- `notification`: Complete notification object
- `taskId`: ID of the related task
- `projectId`: ID of the related project
- `projectTitle`: Title of the related project

**Notification Object:**
- `id`: Unique notification identifier
- `title`: Notification title
- `message`: Detailed notification message
- `type`: Notification type (enum)
- `userId`: ID of the user who should receive this notification
- `relatedEntityId`: ID of the related entity (task ID)
- `relatedEntityType`: Type of related entity (always "TASK" for task notifications)
- `isRead`: Boolean indicating if notification has been read
- `createdAt`: Timestamp when notification was created

### Notification Types

The `type` field in the notification object indicates the kind of event:

- `TASK_ASSIGNED`: Task has been assigned to a user
- `TASK_UPDATED`: Task has been updated (any field changed)
- `TASK_CREATED`: New task has been created
- `PROJECT_CREATED`: New project has been created
- `PROJECT_UPDATED`: Project has been updated
- `LEAVE_REQUEST_CREATED`: Leave request has been created
- `LEAVE_REQUEST_APPROVED`: Leave request has been approved
- `LEAVE_REQUEST_REJECTED`: Leave request has been rejected
- `USER_LOGIN`: User has logged in
- `USER_LOGOUT`: User has logged out

### Event-Specific Additional Data

Depending on the notification type, additional context may be included:

**Task Updated:**
- `updaterName`: Name of the user who made the update

**Task Status Updated:**
- `oldStatus`: Previous task status
- `newStatus`: New task status

**Time Logged:**
- `memberName`: Name of the member who logged time
- `hours`: Number of hours logged (as float)
- `date`: Date when time was logged (formatted string)
- `timeDescription`: Optional description of the time log entry

## Connection Lifecycle

### Establishing Connection

1. Create SSE connection to the endpoint with authentication token
2. Wait for `authenticated` event to confirm successful connection
3. Begin listening for `notification` events

### Connection Maintenance

**Heartbeat:**
- Server sends heartbeat comments every 30 seconds to keep connection alive
- Heartbeat format: `: heartbeat\n\n`
- Client should handle these silently

**Connection Timeout:**
- Connections automatically timeout after 5 minutes of inactivity
- Client should implement reconnection logic

**Disconnection Handling:**
- Monitor connection state
- Implement automatic reconnection on disconnect
- Handle network interruptions gracefully

## Event Processing Flow

### Receiving Notifications

1. **Listen for Events:**
   - Monitor SSE stream for `notification` event type
   - Parse JSON data from event payload

2. **Validate Notification:**
   - Check `userId` matches current user
   - Verify notification structure is complete

3. **Process Notification:**
   - Extract notification object and metadata
   - Determine notification type from `type` field
   - Extract event-specific additional data if present

4. **Update UI:**
   - Display notification to user
   - Update notification count/badge
   - Show appropriate visual indicator
   - Optionally navigate to related task/project

5. **Store Locally (Optional):**
   - Cache notification for offline access
   - Mark as unread until user interacts

### Notification Persistence

**Important:** Notifications received via SSE are also stored in the database. This means:
- Notifications persist even if client disconnects
- Client can fetch missed notifications via REST API
- Notification read/unread state is maintained server-side
- Client should sync local state with server state

## Best Practices

### Connection Management

- **Reconnection Strategy:**
  - Implement exponential backoff for reconnection attempts
  - Retry connection on network errors
  - Maintain connection state across page navigation when possible

- **Connection Cleanup:**
  - Close SSE connection when user logs out
  - Close connection when navigating away from application
  - Handle browser tab/window close events

### Notification Handling

- **Deduplication:**
  - Check notification ID to avoid processing duplicates
  - Compare with locally cached notifications

- **Rate Limiting:**
  - Handle burst of notifications gracefully
  - Queue notifications if UI update is slow
  - Batch UI updates when possible

- **User Experience:**
  - Show notification count badge
  - Display toast/alert for important notifications
  - Provide notification history view
  - Allow user to mark notifications as read
  - Support notification filtering by type

### Error Handling

- **Connection Errors:**
  - Handle authentication failures (401)
  - Handle server errors (500)
  - Handle network timeouts
  - Log errors for debugging

- **Data Errors:**
  - Validate notification structure before processing
  - Handle missing or malformed fields gracefully
  - Skip invalid notifications with error logging

### Performance Considerations

- **Memory Management:**
  - Limit number of notifications kept in memory
  - Implement notification cleanup/archival
  - Use pagination for notification history

- **Network Efficiency:**
  - SSE connection is persistent, no polling needed
  - Only one connection per user (new connection closes old one)
  - Server handles connection management automatically

## Notification State Management

### Read/Unread Status

- Notifications start as `isRead: false`
- Client should mark notifications as read when user views/interacts
- Use REST API to update read status on server
- Sync read status across multiple client instances

### Notification Count

- Server maintains unread notification count
- Fetch count via REST API endpoint
- Update count when receiving new notifications via SSE
- Update count when marking notifications as read

## Integration with REST API

### Complementary Endpoints

SSE notifications work alongside REST API endpoints:

- **Fetch All Notifications:** GET `/api/notifications`
- **Fetch Unread Notifications:** GET `/api/notifications/unread`
- **Get Notification Count:** GET `/api/notifications/count`
- **Mark as Read:** PUT `/api/notifications/:id/read`
- **Mark All as Read:** PUT `/api/notifications/read-all`
- **Delete Notification:** DELETE `/api/notifications/:id`

### Synchronization Strategy

1. **Initial Load:**
   - Fetch existing notifications via REST API on app start
   - Establish SSE connection
   - Merge SSE notifications with fetched notifications

2. **Ongoing Sync:**
   - Use SSE for real-time updates
   - Periodically sync with REST API to catch missed notifications
   - Handle conflicts (same notification from both sources)

3. **Offline Handling:**
   - Cache notifications locally when offline
   - Sync with server when connection restored
   - Handle duplicate detection

## Security Considerations

### Authentication

- Always use valid JWT tokens
- Refresh tokens before expiration
- Handle token expiration gracefully
- Re-authenticate if connection fails with 401

### Data Privacy

- Notifications contain task and project information
- Ensure user only receives notifications intended for them
- Server validates user permissions before sending
- Client should validate `userId` matches current user

### Connection Security

- Use HTTPS in production
- Validate server identity
- Handle certificate errors appropriately
- Protect against connection hijacking

## Troubleshooting

### Common Issues

**Connection Not Establishing:**
- Verify JWT token is valid and not expired
- Check network connectivity
- Verify endpoint URL is correct
- Check browser console for errors

**Notifications Not Received:**
- Verify user has permission for related task/project
- Check if user is assigned to the task (for task notifications)
- Verify notification service is running on server
- Check server logs for errors

**Duplicate Notifications:**
- Implement deduplication by notification ID
- Check if multiple SSE connections are active
- Verify notification creation logic on server

**Connection Drops Frequently:**
- Check network stability
- Verify server heartbeat is working
- Check for proxy/firewall interference
- Review connection timeout settings

## Event Flow Diagram

```
Client                    Server
  |                         |
  |-- Connect (with token)-->|
  |                         |
  |<-- authenticated -------|
  |                         |
  |                         | (Task event occurs)
  |                         |
  |                         | (Create notification in DB)
  |                         |
  |                         | (Send via SSE)
  |<-- notification --------|
  |                         |
  | (Process & Display)     |
  |                         |
```

## Summary

SSE notifications provide real-time delivery of task-related events. The system ensures:
- **Real-time delivery** via SSE for connected clients
- **Persistence** via database storage for offline access
- **Reliability** through automatic reconnection and error handling
- **Scalability** through efficient connection management
- **Security** through authentication and permission validation

Clients should implement robust connection management, notification processing, and synchronization with REST API to provide a complete notification experience.

