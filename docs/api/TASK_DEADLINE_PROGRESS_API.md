# Task Deadline and Progress API Documentation

This document describes the Task Deadline and Progress API endpoints for updating task deadlines and tracking task completion progress.

## Base URL

All endpoints are prefixed with `/api/tasks`

## Authentication

All endpoints require authentication. Include the authentication token in the request headers:

```
Authorization: Bearer <your-token>
```

## Authorization

All endpoints require the user to have permission to modify the task. Users can update tasks if they:
- Are assigned to the task, OR
- Are a project owner or member of the project containing the task, OR
- Have admin role (admins have full access)

## Endpoints

### 1. Update Task Deadline

Updates the deadline (due date) for a specific task.

**Endpoint:** `PUT /api/tasks/update-deadline/:projectId/:taskId`

**Permissions Required:** 
- Authentication token
- Task access (user must be assigned to task OR project owner/member OR admin)
- `tasks:write` permission

**Path Parameters:**
- `projectId` (required): The unique identifier of the project containing the task
- `taskId` (required): The unique identifier of the task to update

**Request Headers:**
```
Authorization: Bearer <your-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "deadline": "2025-01-20T10:00:00Z"
}
```

**Request Fields:**
- `deadline` (required): The deadline date and time in RFC3339 format (ISO 8601)
  - Format: `YYYY-MM-DDTHH:MM:SSZ` or `YYYY-MM-DDTHH:MM:SS+00:00`
  - Example: `"2025-01-20T10:00:00Z"` (UTC timezone)
  - Example: `"2025-01-20T10:00:00+05:30"` (IST timezone)

**Request Example:**
```http
PUT /api/tasks/update-deadline/5CqqY56u8ZQ4oFBm9eAl/task123
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "deadline": "2025-01-20T10:00:00Z"
}
```

**Success Response (200 OK):**
```json
{
  "message": "Task deadline updated successfully"
}
```

**Error Responses:**

**400 Bad Request - Invalid Deadline Format:**
```json
{
  "error": "Invalid deadline format. Please use RFC3339 format (e.g., 2025-01-20T10:00:00Z)"
}
```

**400 Bad Request - Missing Required Field:**
```json
{
  "error": "Key: 'deadline' Error:Field validation for 'deadline' failed on the 'required' tag"
}
```

**401 Unauthorized - Missing or Invalid Token:**
```json
{
  "error": "User not authenticated"
}
```

**403 Forbidden - Insufficient Permissions:**
```json
{
  "error": "You do not have permission to modify this task. You must be assigned to this task or be a project member to perform this operation"
}
```

**404 Not Found - Task Not Found:**
```json
{
  "error": "task not found"
}
```

**cURL Example:**
```bash
curl -X PUT "https://api.example.com/api/tasks/update-deadline/5CqqY56u8ZQ4oFBm9eAl/task123" \
  -H "Authorization: Bearer your-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "deadline": "2025-01-20T10:00:00Z"
  }'
```

**JavaScript/Fetch Example:**
```javascript
const updateTaskDeadline = async (projectId, taskId, deadline) => {
  const response = await fetch(
    `https://api.example.com/api/tasks/update-deadline/${projectId}/${taskId}`,
    {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${yourToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        deadline: deadline // RFC3339 format string
      })
    }
  );

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to update deadline');
  }

  const result = await response.json();
  return result;
};

// Usage
updateTaskDeadline('5CqqY56u8ZQ4oFBm9eAl', 'task123', '2025-01-20T10:00:00Z')
  .then(result => console.log('Success:', result.message))
  .catch(error => console.error('Error:', error));
```

---

### 2. Update Task Progress

Updates the completion progress percentage for a specific task.

**Endpoint:** `PUT /api/tasks/update-progress/:projectId/:taskId`

**Permissions Required:** 
- Authentication token
- Task access (user must be assigned to task OR project owner/member OR admin)
- `tasks:write` permission

**Path Parameters:**
- `projectId` (required): The unique identifier of the project containing the task
- `taskId` (required): The unique identifier of the task to update

**Request Headers:**
```
Authorization: Bearer <your-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "progress": 75
}
```

**Request Fields:**
- `progress` (required): The completion percentage as an integer
  - Minimum value: `0` (0% complete)
  - Maximum value: `100` (100% complete)
  - Values outside this range will be automatically clamped to 0-100

**Request Example:**
```http
PUT /api/tasks/update-progress/5CqqY56u8ZQ4oFBm9eAl/task123
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "progress": 75
}
```

**Success Response (200 OK):**
```json
{
  "message": "Task progress updated successfully"
}
```

**Error Responses:**

**400 Bad Request - Invalid Progress Value:**
```json
{
  "error": "Invalid progress value. Progress must be between 0 and 100"
}
```

**400 Bad Request - Missing Required Field:**
```json
{
  "error": "Key: 'progress' Error:Field validation for 'progress' failed on the 'required' tag"
}
```

**401 Unauthorized - Missing or Invalid Token:**
```json
{
  "error": "User not authenticated"
}
```

**403 Forbidden - Insufficient Permissions:**
```json
{
  "error": "You do not have permission to modify this task. You must be assigned to this task or be a project member to perform this operation"
}
```

**404 Not Found - Task Not Found:**
```json
{
  "error": "task not found"
}
```

**cURL Example:**
```bash
curl -X PUT "https://api.example.com/api/tasks/update-progress/5CqqY56u8ZQ4oFBm9eAl/task123" \
  -H "Authorization: Bearer your-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "progress": 75
  }'
```

**JavaScript/Fetch Example:**
```javascript
const updateTaskProgress = async (projectId, taskId, progress) => {
  // Validate progress is between 0 and 100
  if (progress < 0 || progress > 100) {
    throw new Error('Progress must be between 0 and 100');
  }

  const response = await fetch(
    `https://api.example.com/api/tasks/update-progress/${projectId}/${taskId}`,
    {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${yourToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        progress: Math.round(progress) // Ensure integer value
      })
    }
  );

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to update progress');
  }

  const result = await response.json();
  return result;
};

// Usage
updateTaskProgress('5CqqY56u8ZQ4oFBm9eAl', 'task123', 75)
  .then(result => console.log('Success:', result.message))
  .catch(error => console.error('Error:', error));
```

---

## Task Model

The updated Task model includes the following fields related to deadline and progress:

```json
{
  "id": "task123",
  "subject": "Implement user authentication",
  "code": "TASK-001",
  "status": "in-progress",
  "deadline": "2025-01-20T10:00:00Z",
  "priority": "high",
  "progress": 75,
  "assignTo": "user456",
  "projectId": "5CqqY56u8ZQ4oFBm9eAl",
  "description": "Implement JWT-based authentication system",
  "timeSpent": [],
  "fileAttachments": [],
  "activityLogs": [],
  "created": "2025-01-10T08:00:00Z",
  "updated": "2025-01-15T14:30:00Z"
}
```

**Field Descriptions:**
- `deadline` (required): The task deadline/due date in RFC3339 format
- `progress` (optional): Completion percentage (0-100). Can be `null` if not set

---

## Date Format

All dates and times use the **RFC3339** format (ISO 8601):

- **UTC Format:** `YYYY-MM-DDTHH:MM:SSZ`
  - Example: `"2025-01-20T10:00:00Z"`
  
- **With Timezone:** `YYYY-MM-DDTHH:MM:SS±HH:MM`
  - Example: `"2025-01-20T10:00:00+05:30"` (IST)
  - Example: `"2025-01-20T10:00:00-05:00"` (EST)

**Important Notes:**
- Always use UTC or include timezone offset
- The `Z` suffix indicates UTC (Zulu time)
- Times are stored in UTC internally

---

## Progress Value Guidelines

The `progress` field represents task completion percentage:

- **0**: Task not started (0% complete)
- **1-99**: Task in progress (1-99% complete)
- **100**: Task fully completed (100% complete)

**Best Practices:**
- Update progress incrementally as work progresses
- Use meaningful milestones (e.g., 25%, 50%, 75%, 100%)
- Consider automatically setting progress to 100% when status changes to "completed"
- Consider automatically setting progress to 0% when status changes to "backlog" or "todo"

---

## Activity Logging

Both deadline and progress updates are automatically logged in the task's activity log:

**Deadline Update Activity:**
```json
{
  "id": "log123",
  "type": "deadline_updated",
  "timestamp": "2025-01-15T14:30:00Z",
  "userId": "user456",
  "userName": "John Doe",
  "description": "Task deadline updated to 2025-01-20T10:00:00Z",
  "metadata": {
    "oldDeadline": "2025-01-18T10:00:00Z",
    "newDeadline": "2025-01-20T10:00:00Z"
  }
}
```

**Progress Update Activity:**
```json
{
  "id": "log124",
  "type": "progress_updated",
  "timestamp": "2025-01-15T14:35:00Z",
  "userId": "user456",
  "userName": "John Doe",
  "description": "Task progress updated to 75%",
  "metadata": {
    "oldProgress": 50,
    "newProgress": 75
  }
}
```

---

## Migration Notes

### Backward Compatibility

The API maintains backward compatibility with existing data:

1. **Deadline Field Migration:**
   - Old field name: `duration` (deprecated)
   - New field name: `deadline`
   - The API automatically migrates old `duration` fields to `deadline` when reading tasks
   - New tasks should always use `deadline`

2. **Progress Field:**
   - The `progress` field is optional (can be `null`)
   - Existing tasks without progress will have `progress: null`
   - You can set progress for any task at any time

### Data Migration

If you have existing tasks with the old `duration` field, they will be automatically converted when:
- Tasks are retrieved from the database
- The migration script is run

No manual migration is required for deadline fields. Progress can be set incrementally as needed.

---

## Error Handling

### Common Error Scenarios

1. **Invalid Date Format:**
   - Ensure deadline is in RFC3339 format
   - Use a date parsing library to validate before sending

2. **Invalid Progress Value:**
   - Progress must be an integer between 0 and 100
   - Values outside this range are automatically clamped

3. **Authorization Failures:**
   - Ensure the user has access to the project
   - Ensure the user is assigned to the task (or is project owner/member/admin)

4. **Task Not Found:**
   - Verify the `projectId` and `taskId` are correct
   - Ensure the task exists in the database

### Error Response Format

All error responses follow this format:
```json
{
  "error": "Error message describing what went wrong"
}
```

---

## Rate Limiting

All endpoints are subject to rate limiting. Check response headers for rate limit information:
- `X-RateLimit-Limit`: Maximum number of requests allowed
- `X-RateLimit-Remaining`: Number of requests remaining
- `X-RateLimit-Reset`: Time when the rate limit resets

---

## Examples

### Complete Workflow: Creating and Updating a Task

```javascript
// 1. Create a task
const createTask = async () => {
  const response = await fetch('https://api.example.com/api/tasks/add', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      subject: 'Implement user authentication',
      code: 'TASK-001',
      status: 'todo',
      deadline: '2025-01-20T10:00:00Z',
      priority: 'high',
      projectId: '5CqqY56u8ZQ4oFBm9eAl',
      description: 'Implement JWT-based authentication system'
    })
  });
  return await response.json();
};

// 2. Update task deadline
const updateDeadline = async (projectId, taskId, newDeadline) => {
  const response = await fetch(
    `https://api.example.com/api/tasks/update-deadline/${projectId}/${taskId}`,
    {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ deadline: newDeadline })
    }
  );
  return await response.json();
};

// 3. Update task progress as work progresses
const updateProgress = async (projectId, taskId, progress) => {
  const response = await fetch(
    `https://api.example.com/api/tasks/update-progress/${projectId}/${taskId}`,
    {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ progress })
    }
  );
  return await response.json();
};

// Usage
(async () => {
  try {
    // Create task
    const task = await createTask();
    console.log('Task created:', task);

    // Update deadline
    await updateDeadline(task.projectId, task.id, '2025-01-25T10:00:00Z');
    console.log('Deadline updated');

    // Update progress incrementally
    await updateProgress(task.projectId, task.id, 25);
    console.log('Progress: 25%');

    await updateProgress(task.projectId, task.id, 50);
    console.log('Progress: 50%');

    await updateProgress(task.projectId, task.id, 75);
    console.log('Progress: 75%');

    await updateProgress(task.projectId, task.id, 100);
    console.log('Progress: 100% - Task completed!');
  } catch (error) {
    console.error('Error:', error);
  }
})();
```

---

## Support

For issues or questions regarding the Task Deadline and Progress API:
- Check the main API documentation
- Review the error messages for specific guidance
- Contact the API support team

---

## Changelog

### Version 1.0.0 (Current)
- Added `deadline` field to replace deprecated `duration` field
- Added `progress` field for tracking task completion percentage
- Added `PUT /api/tasks/update-deadline/:projectId/:taskId` endpoint
- Added `PUT /api/tasks/update-progress/:projectId/:taskId` endpoint
- Automatic backward compatibility for `duration` field migration
- Activity logging for deadline and progress updates

