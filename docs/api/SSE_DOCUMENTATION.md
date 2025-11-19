# SSE (Server-Sent Events) Documentation

Complete documentation for Server-Sent Events API and notification system integration.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Connection Setup](#connection-setup)
4. [Event Types](#event-types)
5. [Notification Flow](#notification-flow)
6. [Connection Lifecycle](#connection-lifecycle)
7. [Notification Types](#notification-types)
8. [Notification Data Structure](#notification-data-structure)
9. [Integration with REST API](#integration-with-rest-api)
10. [Implementation Guide](#implementation-guide)
11. [Best Practices](#best-practices)
12. [Security Considerations](#security-considerations)
13. [Troubleshooting](#troubleshooting)

## Overview

The application uses Server-Sent Events (SSE) to notify clients when new notifications are available. When task-related events occur (task assigned, updated, status changed, time logged), notifications are automatically stored in the database and a simple notification event is sent via SSE to inform clients that new notifications are available. Clients then fetch the actual notifications via REST API.

## Architecture

The notification system follows a two-step process:

1. **Database Storage**: All notifications are stored in the database notification collection first
2. **SSE Notification**: A simple `notifications-available` event is sent via SSE to inform clients
3. **Client Fetch**: Clients receive the event and fetch notifications via REST API

This approach ensures:
- Notifications persist in the database even if clients are disconnected
- Clients can retrieve missed notifications via REST API
- SSE events are lightweight and efficient
- Notification data is always up-to-date when fetched from the database

### Why This Approach?

- **Persistence**: Notifications are stored in the database, so they persist even if clients are disconnected
- **Efficiency**: SSE events are lightweight, containing only a signal that notifications are available
- **Reliability**: Clients can fetch missed notifications via REST API
- **Data Accuracy**: Notification data is always up-to-date when fetched from the database

## Connection Setup

### Endpoint

```
GET /api/events?token=<jwt-token>
```

For production:
```
GET https://your-domain.com/api/events?token=<jwt-token>
```

**Note:** The default port is 3000, but you can change it by setting the `PORT` environment variable.

### Authentication

The SSE connection requires authentication via JWT token. The token can be provided either:
- As a query parameter: `token=<jwt-token>`
- In the Authorization header: `Authorization: Bearer <jwt-token>`

**Note:** The server extracts the user ID from the JWT token during connection. The token must be valid and include a valid user ID. If authentication fails, the connection will be rejected with HTTP 401.

### Connection Headers

The server will respond with SSE-specific headers:
- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`
- `X-Accel-Buffering: no` (prevents proxy buffering)
- `Access-Control-Allow-Origin: *`

### Connection Request

**Query Parameter:**
```
GET /api/events?token=<jwt-token>
```

**Authorization Header:**
```
GET /api/events
Authorization: Bearer <jwt-token>
```

## Event Format

All events follow the SSE format:

```
event: <event-type>
data: <json-data>

```

Multiple events can be sent in sequence, each separated by a blank line. The EventSource API automatically parses this format and provides events through the `addEventListener` method.

## Event Types

### `authenticated`

Sent immediately after successful connection and authentication.

**Event Type:** `authenticated`

**Data Structure:**
- `userId`: The authenticated user's ID

**Purpose:** Confirms successful connection and identifies the authenticated user.

### `notifications-available`

Sent when new notifications are available in the database for the user. This is a lightweight signal event that informs the client to fetch notifications via REST API.

**Event Type:** `notifications-available`

**When Sent:**
- Task assigned to a user
- Task updated (any field change)
- Task status updated
- Time logged on a task
- Any other event that creates a notification

**Data Structure:**
- `userId`: The user ID for whom notifications are available

**Purpose:** Signals to the client that new notifications exist in the database and should be fetched via REST API.

**Client Action Required:** Upon receiving this event, the client must call the REST API to fetch notifications (typically unread notifications) and update the UI accordingly.

### `error`

Sent when an error occurs.

**Event Type:** `error`

**Data Structure:**
- `error`: Error message describing what went wrong

## Notification Flow

### Server-Side Process

1. **Event Occurs**: A task-related event occurs (task assigned, updated, status changed, time logged)
2. **Create Notification**: Server creates a notification object with all relevant details
3. **Store in Database**: Notification is stored in the database notification collection
4. **Send SSE Event**: Server sends a `notifications-available` event via SSE to the affected user(s)
5. **Client Response**: Client receives the event and fetches notifications via REST API

### Client-Side Process

1. **Establish Connection**: Client connects to SSE endpoint with authentication token
2. **Receive Authentication**: Client receives `authenticated` event confirming connection
3. **Listen for Events**: Client listens for `notifications-available` events
4. **Receive Notification Signal**: When `notifications-available` event is received
5. **Fetch Notifications**: Client calls REST API to fetch new notifications
6. **Update UI**: Client displays notifications and updates notification count/badge

## Connection Lifecycle

### 1. Connection Setup

The SSE connection is established through a standard HTTP GET request. The server automatically:
- Validates your JWT token
- Extracts your user ID
- Establishes the SSE connection
- Closes any existing connection for the same user (only one active connection per user)
- Sends an `authenticated` event with your user ID

### 2. Keep-Alive

The server automatically sends heartbeat comments every 30 seconds to keep the connection alive. The heartbeat format is:

```
: heartbeat

```

These are automatically handled by the EventSource API and don't trigger events. Client should handle these silently.

### 3. Disconnection

The connection will close if:
- Client explicitly disconnects
- Network error occurs
- Server shuts down
- Authentication token is invalid or expired
- Connection timeout (5 minutes of inactivity)

### Connection Limits

- **One connection per user**: If a user connects with an existing connection, the old connection is closed
- **Connection timeout**: 5 minutes of inactivity
- **Heartbeat interval**: 30 seconds

### Establishing Connection

1. Create SSE connection to the endpoint with authentication token
2. Wait for `authenticated` event to confirm successful connection
3. Begin listening for `notifications-available` events

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

## Notification Types

The notification system supports various notification types stored in the database:

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

## Notification Data Structure

When fetching notifications via REST API, each notification contains:

- `id`: Unique notification identifier
- `title`: Notification title
- `message`: Detailed notification message
- `type`: Notification type (enum from list above)
- `userId`: ID of the user who should receive this notification
- `relatedEntityId`: ID of the related entity (task ID, project ID, etc.)
- `relatedEntityType`: Type of related entity (TASK, PROJECT, LEAVE_REQUEST, USER)
- `isRead`: Boolean indicating if notification has been read
- `createdAt`: Timestamp when notification was created

## Integration with REST API

### Complementary Endpoints

SSE notifications work alongside REST API endpoints:

- **Fetch All Notifications:** GET `/api/notifications/:userId`
- **Fetch Unread Notifications:** GET `/api/notifications/:userId/unread`
- **Get Notification Count:** GET `/api/notifications/:userId/count`
- **Mark as Read:** PUT `/api/notifications/:id/read`
- **Mark All as Read:** PUT `/api/notifications/:userId/read-all`
- **Delete Notification:** DELETE `/api/notifications/:id`

When a `notifications-available` event is received, clients should call the appropriate REST API endpoint to fetch the actual notification data.

### Synchronization Strategy

1. **Initial Load:**
   - Fetch existing notifications via REST API on app start
   - Establish SSE connection
   - Display fetched notifications in UI

2. **Ongoing Sync:**
   - Use SSE `notifications-available` events for real-time updates
   - When event received, fetch unread notifications via REST API
   - Merge new notifications with existing ones
   - Update notification count and UI

3. **Offline Handling:**
   - Cache notifications locally when offline
   - When connection restored, fetch missed notifications via REST API
   - Handle duplicate detection using notification IDs

## Implementation Guide

### Implementation Steps

#### 1. Establish Connection

Create an SSE connection with your JWT token. Wait for the `authenticated` event to confirm successful connection.

#### 2. Listen for Events

Set up event listeners for:
- `authenticated` - Confirm connection
- `notifications-available` - New notifications available
- `error` - Handle errors

#### 3. Handle Notifications-Available Event

When `notifications-available` event is received:
- Extract the `userId` from the event data
- Verify the `userId` matches the current user
- Call REST API to fetch unread notifications
- Update notification count/badge
- Display new notifications in UI

#### 4. Fetch Notifications

Use the REST API to fetch notifications. Typically, you'll want to fetch unread notifications to get the latest updates.

#### 5. Update UI

After fetching notifications:
- Display new notifications to the user
- Update notification count/badge
- Mark notifications as read when user views them
- Handle notification interactions (navigate to related task/project)

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

- **Event Processing:**
  - When `notifications-available` event is received, immediately fetch notifications via REST API
  - Fetch unread notifications to get the latest updates
  - Update notification count/badge after fetching
  - Don't delay fetching as notifications are already in the database

- **Deduplication:**
  - Check notification ID to avoid processing duplicates
  - Compare with locally cached notifications
  - Use notification ID as unique identifier

- **Rate Limiting:**
  - Handle burst of notifications gracefully
  - Queue notification fetches if multiple events arrive quickly
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
  - Validate notification structure when fetching from API
  - Handle missing or malformed fields gracefully
  - Skip invalid notifications with error logging

- **REST API Errors:**
  - Handle REST API errors when fetching notifications
  - Retry failed API calls with exponential backoff
  - Log errors for debugging

### Performance Considerations

- **Memory Management:**
  - Limit number of notifications kept in memory
  - Implement notification cleanup/archival
  - Use pagination for notification history

- **Network Efficiency:**
  - SSE connection is persistent, no polling needed
  - Only one connection per user (new connection closes old one)
  - Server handles connection management automatically
  - Fetch notifications only when `notifications-available` event is received
  - Fetch only unread notifications when `notifications-available` event is received
  - Use pagination for notification history
  - Cache notifications locally for offline access

### Notification State Management

- **Read/Unread Status:**
  - Notifications start as `isRead: false` when created
  - Client should mark notifications as read when user views/interacts
  - Use REST API to update read status on server
  - Sync read status across multiple client instances

- **Notification Count:**
  - Server maintains unread notification count
  - Fetch count via REST API endpoint
  - Update count after fetching notifications when `notifications-available` event is received
  - Update count when marking notifications as read

## Security Considerations

### Authentication

- Always use valid JWT tokens
- Refresh tokens before expiration
- Handle token expiration gracefully
- Re-authenticate if connection fails with 401
- Validate JWT tokens and handle expiration

### Data Privacy

- Notifications contain task and project information
- Ensure user only receives notifications intended for them
- Server validates user permissions before creating and sending notifications
- Client should validate `userId` in `notifications-available` event matches current user

### Connection Security

- Use HTTPS in production
- Validate server identity
- Handle certificate errors appropriately
- Protect against connection hijacking
- Implement proper error handling for authentication failures
- Ensure user permissions are validated before sending events

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
- Ensure SSE connection is active and receiving events
- Verify SSE connection is active
- Check if `notifications-available` events are being received
- Verify REST API calls are being made after receiving events

**Notifications Not Appearing After Event:**
- Verify REST API call is made after receiving `notifications-available` event
- Check REST API endpoint is correct
- Verify authentication token is valid for REST API calls
- Check if notifications exist in database via API
- Verify REST API endpoint is correct
- Check if notifications exist in database
- Verify authentication token is valid for REST API calls

**Duplicate Notifications:**
- Implement deduplication by notification ID
- Check if multiple SSE connections are active
- Verify notification creation logic on server
- Use notification ID as unique identifier when merging

**Connection Drops Frequently:**
- Check network stability
- Verify server heartbeat is working
- Check for proxy/firewall interference
- Review connection timeout settings

## Event Flow

### Server-Side Flow

1. Task event occurs (task assigned, updated, status changed, time logged)
2. Server creates notification object with all details
3. Server stores notification in database
4. Server sends `notifications-available` SSE event to affected user(s)
5. Server continues processing

### Client-Side Flow

1. Client establishes SSE connection
2. Client receives `authenticated` event
3. Client listens for `notifications-available` events
4. Client receives `notifications-available` event
5. Client calls REST API to fetch notifications
6. Client updates UI with new notifications

## Summary

The notification system provides real-time awareness of new notifications through SSE events, while ensuring reliability through database persistence. The system ensures:

- **Real-time awareness** via SSE `notifications-available` events for connected clients
- **Persistence** via database storage for offline access and missed notifications
- **Reliability** through automatic reconnection and error handling
- **Efficiency** through lightweight SSE events and on-demand fetching
- **Scalability** through efficient connection management
- **Security** through authentication and permission validation

Clients should implement robust connection management, event handling, and REST API integration to provide a complete notification experience. Always fetch notifications from the database via REST API when `notifications-available` events are received to ensure data accuracy and completeness.

The SSE API provides a lightweight, efficient way to notify clients when new notifications are available. Clients receive simple `notifications-available` events and fetch the actual notification data via REST API, ensuring data persistence, accuracy, and reliability.

