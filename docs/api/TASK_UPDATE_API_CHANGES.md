# Task Update API Changes

## Breaking Changes

### Endpoint Path Change

**Old Endpoint:**
```
PUT /api/tasks/update
```

**New Endpoint:**
```
PUT /api/tasks/update/:projectId/:taskId
```

### URL Parameters (New - Required)

- `projectId` (string, required) - The project ID containing the task
- `taskId` (string, required) - The task ID to update

### Request Body Changes

**Before:**
- `projectId` and `id` could be included in the request body
- These values were used for authorization and identification

**After:**
- `projectId` and `id` in the request body are **ignored**
- These values are now taken from URL parameters only
- Request body should only contain fields to be updated (subject, status, priority, deadline, description, progress, etc.)

### Security Improvements

1. **Project ID Validation**: The `projectId` in the URL must match the task's actual project ID. Attempting to update a task with a different project ID will result in an error.

2. **Project ID Immutability**: The `projectId` field cannot be changed through this endpoint. Tasks cannot be moved between projects via the update API.

### Error Responses

**New Error:**
- `400 Bad Request`: "cannot change task projectId - task belongs to a different project" - Returned when the URL `projectId` doesn't match the task's actual project ID

### Migration Notes

- Update all client code that calls `PUT /api/tasks/update` to include `projectId` and `taskId` in the URL path
- Remove `projectId` and `id` from request body if they were being sent
- Ensure the `projectId` in the URL matches the task's actual project ID











