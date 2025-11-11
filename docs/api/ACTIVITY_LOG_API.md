# Activity Log API Documentation

This document describes the Activity Log API endpoints for retrieving activity logs for tasks, projects, users, and calendar events.

## Overview

The Activity Log system provides comprehensive tracking of all operations performed on entities in the application. All activity logs are stored in a centralized `activity_logs` DynamoDB table and can be queried by entity type and entity ID.

## Base URL

All endpoints are prefixed with `/api/activity-log`

## Authentication

All endpoints require authentication. Include the authentication token in the request headers:

```
Authorization: Bearer <your-token>
```

## Authorization

Most endpoints require the `activityLog:read` permission. The `GetEntityTypes` endpoint is publicly accessible (no permission required).

## Supported Entity Types

- `task` - Task activity logs
- `project` - Project activity logs
- `user` - User activity logs
- `calendarEvent` - Calendar event activity logs

## Supported Action Types

### Task Actions
- `created` - Task was created
- `updated` - Task was updated
- `deleted` - Task was deleted
- `assigned` - Task was assigned to a user
- `unassigned` - Task was unassigned from a user
- `status_changed` - Task status changed
- `priority_changed` - Task priority changed
- `description_updated` - Task description updated
- `deadline_updated` - Task deadline updated
- `progress_updated` - Task progress updated
- `file_uploaded` - File attached to task
- `file_removed` - File removed from task
- `time_spent_added` - Time entry added
- `time_spent_updated` - Time entry updated
- `time_spent_removed` - Time entry removed

### Project Actions
- `created` - Project was created
- `updated` - Project was updated
- `deleted` - Project was deleted
- `member_added` - Member added to project
- `member_removed` - Member removed from project
- `owner_changed` - Project owner changed
- `deadline_updated` - Project deadline updated

### User Actions
- `created` - User was created
- `updated` - User was updated
- `deleted` - User was deleted
- `profile_updated` - User profile updated
- `role_changed` - User role changed
- `password_changed` - User password changed

### Calendar Event Actions
- `event_created` - Calendar event was created
- `event_updated` - Calendar event was updated
- `event_deleted` - Calendar event was deleted
- `event_cancelled` - Calendar event was cancelled

## Endpoints

### 1. Get Supported Entity Types

Returns a list of all supported entity types for activity logs.

**Endpoint:** `GET /api/activity-log/entity-types`

**Permissions Required:** None (publicly accessible)

**Request Headers:**
```
Authorization: Bearer <your-token>
```

**Request Example:**
```http
GET /api/activity-log/entity-types
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response (200 OK):**
```json
{
  "entityTypes": [
    "task",
    "project",
    "user",
    "calendarEvent"
  ]
}
```

---

### 2. Get Activity Logs for a Specific Entity

Retrieves all activity logs for a specific entity (task, project, user, or calendar event).

**Endpoint:** `GET /api/activity-log/entity/:entityType/:entityId`

**Permissions Required:** 
- Authentication token
- `activityLog:read` permission

**Path Parameters:**
- `entityType` (required): The type of entity (`task`, `project`, `user`, or `calendarEvent`)
- `entityId` (required): The unique identifier of the entity

**Request Headers:**
```
Authorization: Bearer <your-token>
```

**Request Examples:**

```http
# Get activity logs for a specific task
GET /api/activity-log/entity/task/task123
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Get activity logs for a specific project
GET /api/activity-log/entity/project/proj456
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Get activity logs for a specific user
GET /api/activity-log/entity/user/user789
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response (200 OK):**

**Example Response for Task:**
```json
{
  "activityLogs": [
    {
      "id": "task-task123-1705320000",
      "entityType": "task",
      "entityId": "task123",
      "action": "status_changed",
      "createdAt": "2025-01-15T14:30:00Z",
      "created": "2025-01-15T14:30:00Z",
      "updated": null,
      "createdBy": "user456",
      "createdByUser": {
        "id": "user456",
        "name": "John Doe",
        "email": "john.doe@example.com",
        "phoneNumber": "+1234567890",
        "role": "Standard",
        "designation": "Developer",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "oldStatus": "Open",
        "newStatus": "In Progress"
      },
      "description": "Status changed from \"Open\" to \"In Progress\"",
      "metadata": null
    },
    {
      "id": "task-task123-1705310000",
      "entityType": "task",
      "entityId": "task123",
      "action": "assigned",
      "createdAt": "2025-01-15T10:15:00Z",
      "created": "2025-01-15T10:15:00Z",
      "updated": null,
      "createdBy": "user789",
      "createdByUser": {
        "id": "user789",
        "name": "Jane Smith",
        "email": "jane.smith@example.com",
        "phoneNumber": "+1234567891",
        "role": "Admin",
        "designation": "Project Manager",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "assignedTo": "user456"
      },
      "description": "Task 'Implement user authentication' was assigned to user",
      "metadata": null
    },
    {
      "id": "task-task123-1705300000",
      "entityType": "task",
      "entityId": "task123",
      "action": "time_spent_added",
      "createdAt": "2025-01-14T16:45:00Z",
      "created": "2025-01-14T16:45:00Z",
      "updated": null,
      "createdBy": "user456",
      "createdByUser": {
        "id": "user456",
        "name": "John Doe",
        "email": "john.doe@example.com",
        "phoneNumber": "+1234567890",
        "role": "Standard",
        "designation": "Developer",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "timeSpent": 120,
        "date": "2025-01-14",
        "userId": "user456",
        "description": "Initial development work"
      },
      "description": "Time log entry added for task 'Implement user authentication'",
      "metadata": null
    },
    {
      "id": "task-task123-1705290000",
      "entityType": "task",
      "entityId": "task123",
      "action": "file_uploaded",
      "createdAt": "2025-01-14T09:20:00Z",
      "created": "2025-01-14T09:20:00Z",
      "updated": null,
      "createdBy": "user456",
      "createdByUser": {
        "id": "user456",
        "name": "John Doe",
        "email": "john.doe@example.com",
        "phoneNumber": "+1234567890",
        "role": "Standard",
        "designation": "Developer",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "fileName": "document.pdf",
        "originalName": "document.pdf",
        "fileSize": 524288,
        "mimeType": "application/pdf",
        "uploadedBy": "user456"
      },
      "description": "File 'document.pdf' was uploaded to task 'Implement user authentication'",
      "metadata": null
    },
    {
      "id": "task-task123-1705280000",
      "entityType": "task",
      "entityId": "task123",
      "action": "created",
      "createdAt": "2025-01-13T08:00:00Z",
      "created": "2025-01-13T08:00:00Z",
      "updated": null,
      "createdBy": "user789",
      "createdByUser": {
        "id": "user789",
        "name": "Jane Smith",
        "email": "jane.smith@example.com",
        "phoneNumber": "+1234567891",
        "role": "Admin",
        "designation": "Project Manager",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "subject": "Implement user authentication",
        "status": "Open",
        "priority": "High",
        "projectId": "proj456"
      },
      "description": "Task 'Implement user authentication' was created",
      "metadata": null
    }
  ]
}
```

**Example Response for Project:**
```json
{
  "activityLogs": [
    {
      "id": "project-proj456-1705320000",
      "entityType": "project",
      "entityId": "proj456",
      "action": "member_added",
      "createdAt": "2025-01-15T11:30:00Z",
      "created": "2025-01-15T11:30:00Z",
      "updated": null,
      "createdBy": "user789",
      "createdByUser": {
        "id": "user789",
        "name": "Jane Smith",
        "email": "jane.smith@example.com",
        "phoneNumber": "+1234567891",
        "role": "Admin",
        "designation": "Project Manager",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "memberId": "user456",
        "memberName": "John Doe"
      },
      "description": "Added member John Doe",
      "metadata": null
    },
    {
      "id": "project-proj456-1705310000",
      "entityType": "project",
      "entityId": "proj456",
      "action": "created",
      "createdAt": "2025-01-10T09:00:00Z",
      "created": "2025-01-10T09:00:00Z",
      "updated": null,
      "createdBy": "user789",
      "createdByUser": {
        "id": "user789",
        "name": "Jane Smith",
        "email": "jane.smith@example.com",
        "phoneNumber": "+1234567891",
        "role": "Admin",
        "designation": "Project Manager",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "title": "Mobile App Development"
      },
      "description": "Created project \"Mobile App Development\"",
      "metadata": null
    }
  ]
}
```

**Empty Response (No Logs Found):**
```json
{
  "activityLogs": []
}
```

**Error Responses:**

**401 Unauthorized - Missing or Invalid Token:**
```json
{
  "error": "User not authenticated"
}
```

**403 Forbidden - Missing Permission:**
```json
{
  "error": "Permission denied: activityLog:read"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Internal server error"
}
```

---

### 3. Get Activity Logs by Entity Type

Retrieves recent activity logs for all entities of a specific type. Results are ordered by creation date (most recent first).

**Endpoint:** `GET /api/activity-log/entity-type/:entityType?limit=<n>`

**Permissions Required:** 
- Authentication token
- `activityLog:read` permission

**Path Parameters:**
- `entityType` (required): The type of entity (`task`, `project`, `user`, or `calendarEvent`)

**Query Parameters:**
- `limit` (optional): Maximum number of logs to return. If not specified, returns all logs (use with caution for large datasets)

**Request Headers:**
```
Authorization: Bearer <your-token>
```

**Request Examples:**

```http
# Get recent task logs (limited to 20)
GET /api/activity-log/entity-type/task?limit=20
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Get recent project logs (limited to 10)
GET /api/activity-log/entity-type/project?limit=10
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Get all recent user logs (no limit - use with caution)
GET /api/activity-log/entity-type/user
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response (200 OK):**

**Example Response for Task (limit=3):**
```json
{
  "activityLogs": [
    {
      "id": "task-task456-1705330000",
      "entityType": "task",
      "entityId": "task456",
      "action": "status_changed",
      "createdAt": "2025-01-15T18:00:00Z",
      "created": "2025-01-15T18:00:00Z",
      "updated": null,
      "createdBy": "user123",
      "createdByUser": {
        "id": "user123",
        "name": "Alice Johnson",
        "email": "alice.johnson@example.com",
        "phoneNumber": "+1234567892",
        "role": "Standard",
        "designation": "Designer",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "oldStatus": "In Progress",
        "newStatus": "Completed"
      },
      "description": "Status changed from \"In Progress\" to \"Completed\"",
      "metadata": null
    },
    {
      "id": "task-task789-1705325000",
      "entityType": "task",
      "entityId": "task789",
      "action": "assigned",
      "createdAt": "2025-01-15T17:30:00Z",
      "created": "2025-01-15T17:30:00Z",
      "updated": null,
      "createdBy": "user456",
      "createdByUser": {
        "id": "user456",
        "name": "John Doe",
        "email": "john.doe@example.com",
        "phoneNumber": "+1234567890",
        "role": "Standard",
        "designation": "Developer",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "assignedTo": "user123"
      },
      "description": "Task 'Design user interface' was assigned to user",
      "metadata": null
    },
    {
      "id": "task-task123-1705320000",
      "entityType": "task",
      "entityId": "task123",
      "action": "time_spent_added",
      "createdAt": "2025-01-15T17:00:00Z",
      "created": "2025-01-15T17:00:00Z",
      "updated": null,
      "createdBy": "user789",
      "createdByUser": {
        "id": "user789",
        "name": "Jane Smith",
        "email": "jane.smith@example.com",
        "phoneNumber": "+1234567891",
        "role": "Admin",
        "designation": "Project Manager",
        "created": "2025-01-01T00:00:00Z",
        "updated": null
      },
      "fields": {
        "timeSpent": 90,
        "date": "2025-01-15",
        "userId": "user789",
        "description": "Bug fixes and testing"
      },
      "description": "Time log entry added for task 'Fix authentication bug'",
      "metadata": null
    }
  ]
}
```

**Empty Response (No Logs Found):**
```json
{
  "activityLogs": []
}
```

**Error Responses:**

**401 Unauthorized - Missing or Invalid Token:**
```json
{
  "error": "User not authenticated"
}
```

**403 Forbidden - Missing Permission:**
```json
{
  "error": "Permission denied: activityLog:read"
}
```

**400 Bad Request - Invalid Entity Type:**
```json
{
  "error": "Invalid entity type"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Internal server error"
}
```

---

## Response Object Structure

### ActivityLogResponse

All activity log responses return an array of `ActivityLogResponse` objects with the following structure:

```typescript
{
  id: string;                    // Unique identifier for the activity log
  entityType: string;            // Type of entity: "task" | "project" | "user" | "calendarEvent"
  entityId: string;              // ID of the entity this log refers to
  action: string;                // Action performed (see Action Types above)
  createdAt: string;             // ISO 8601 timestamp when the action occurred
  created: string;               // ISO 8601 timestamp (same as createdAt)
  updated: string | null;        // ISO 8601 timestamp if log was updated (usually null)
  createdBy: string;            // User ID who performed the action
  createdByUser: User | null;   // Full user object (populated) or null if user not found
  fields: object | null;        // Additional fields with old/new values or metadata
  description: string | null;   // Human-readable description of the activity
  metadata: object | null;      // Additional metadata (usually null)
}
```

### User Object (in createdByUser)

```typescript
{
  id: string;
  name: string;
  email: string;
  phoneNumber: string;
  role: string;                 // "Admin" | "Standard"
  designation: string;
  created: string;              // ISO 8601 timestamp
  updated: string | null;        // ISO 8601 timestamp or null
}
```

### Fields Object

The `fields` object contains relevant information about the change, typically including:
- For status changes: `oldStatus`, `newStatus`
- For assignments: `assignedTo`, `oldAssignedTo`, `newAssignedTo`
- For updates: `old` and `new` values for changed fields
- For file uploads: `fileName`, `fileSize`, `mimeType`, `uploadedBy`
- For time entries: `timeSpent`, `date`, `userId`, `description`

---

## Caching

The `GetByEntityType` endpoint uses Redis caching to improve performance:
- Cache TTL: 1 minute
- Cache key format: `activity:logs:{entityType}:limit:{limit}`
- Cache is automatically invalidated when new activity logs are created

---

## Performance Considerations

1. **Use Limits**: Always specify a `limit` parameter when querying by entity type to avoid fetching large datasets
2. **Caching**: The `GetByEntityType` endpoint is cached for 1 minute, so subsequent requests within that time will be faster
3. **Pagination**: For large result sets, consider implementing pagination on the client side
4. **Indexing**: The `activity_logs` table has a Global Secondary Index on `entityId` for efficient queries

---

## Example Usage

### Get all activity logs for a task

```bash
curl -X GET \
  -H "Authorization: Bearer <your-token>" \
  http://localhost:8080/api/activity-log/entity/task/task123
```

### Get recent task activity logs (last 10)

```bash
curl -X GET \
  -H "Authorization: Bearer <your-token>" \
  "http://localhost:8080/api/activity-log/entity-type/task?limit=10"
```

### Get all activity logs for a project

```bash
curl -X GET \
  -H "Authorization: Bearer <your-token>" \
  http://localhost:8080/api/activity-log/entity/project/proj456
```

### Get supported entity types

```bash
curl -X GET \
  -H "Authorization: Bearer <your-token>" \
  http://localhost:8080/api/activity-log/entity-types
```

---

## Integration Notes

### Automatic Activity Log Creation

Activity logs are automatically created for all task operations:
- Task creation, updates, and deletion
- Status, priority, deadline, progress, and description changes
- Time log entries (add, update, remove)
- File attachments (upload, remove)
- Task assignments

All activity logs include:
- The user who performed the action (`createdBy`)
- Timestamp of the action (`createdAt`)
- Human-readable description
- Relevant metadata in the `fields` object

### Task-Specific Endpoint

There is also a task-specific endpoint that returns embedded activity logs from the task document:

```
GET /api/tasks/activity-logs/:projectId/:taskId
```

**Note:** This endpoint returns embedded logs from the task document, not from the centralized `activity_logs` table. For comprehensive activity tracking, use the centralized endpoints described above.

---

## Error Handling

All endpoints follow standard HTTP status codes:
- `200 OK` - Request successful
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication token
- `403 Forbidden` - User lacks required permissions
- `500 Internal Server Error` - Server error occurred

Error responses include an `error` field with a descriptive message.

---

## Support

For questions or issues with the Activity Log API:
- Check the implementation in `internal/handlers/activity_log.go`
- Review the service layer in `internal/services/activity_log_service.go`
- Check the repository in `internal/repos/activity_log_repo.go`
- Review the models in `internal/models/activity_log.go`

