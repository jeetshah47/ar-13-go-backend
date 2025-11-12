# Permissions API Documentation

This document describes the Permissions API for retrieving user roles and permissions, implementing role-based access control (RBAC) in your application.

## Base URL

All endpoints are prefixed with `/api/auth`

## Authentication

All endpoints require authentication. Include the authentication token in the request headers:

```
Authorization: Bearer <your-token>
```

## Overview

The Permissions API provides a comprehensive role-based access control system with two main roles:

- **Admin**: Full access to all features and operations
- **Standard**: Limited access based on project membership and task assignment

## Endpoints

### 1. Get User Permissions

Retrieves the current user's role and all available permissions based on their role.

**Endpoint:** `GET /api/auth/permissions`

**Permissions Required:** `auth:read` (automatically granted to authenticated users)

**Request:**
```http
GET /api/auth/permissions
Authorization: Bearer <token>
Content-Type: application/json
```

**Response (200 OK):**
```json
{
  "role": "Admin",
  "permissions": [
    "projects:read",
    "projects:write",
    "projects:delete",
    "tasks:read",
    "tasks:write",
    "tasks:delete",
    "tasks:assign",
    "users:read",
    "users:write",
    "users:delete",
    "users:profile",
    "users:invite",
    "calendar:read",
    "calendar:write",
    "calendar:delete",
    "vacation:read",
    "vacation:write",
    "vacation:delete",
    "vacation:approve",
    "notifications:read",
    "notifications:write",
    "notifications:delete",
    "activityLogs:read",
    "dashboard:read",
    "employees:read",
    "infoPortal:read",
    "infoPortal:write",
    "infoPortal:delete",
    "googleAccount:read",
    "googleAccount:write",
    "googleAccount:link",
    "googleAccount:unlink",
    "websocket:connect",
    "auth:read",
    "projectDetails:read",
    "projectDetails:write",
    "projectDetails:delete"
  ]
}
```

**Response for Standard User:**
```json
{
  "role": "Standard",
  "permissions": [
    "projects:read",
    "tasks:read",
    "tasks:write",
    "tasks:delete",
    "calendar:read",
    "calendar:write",
    "calendar:delete",
    "vacation:read",
    "vacation:write",
    "vacation:delete",
    "notifications:read",
    "notifications:write",
    "notifications:delete",
    "activityLogs:read",
    "dashboard:read",
    "employees:read",
    "infoPortal:read",
    "infoPortal:write",
    "infoPortal:delete",
    "googleAccount:read",
    "googleAccount:write",
    "googleAccount:link",
    "googleAccount:unlink",
    "websocket:connect",
    "auth:read",
    "users:profile",
    "projectDetails:read"
  ]
}
```

**Response Fields:**
- `role`: User's role (`Admin` or `Standard`)
- `permissions`: Array of permission strings available to the user

**Example:**
```bash
curl -X GET "https://api.example.com/api/auth/permissions" \
  -H "Authorization: Bearer your-token-here" \
  -H "Content-Type: application/json"
```

**Error Responses:**

```json
// 401 Unauthorized - Missing or invalid token
{
  "error": "Invalid token"
}

// 401 Unauthorized - User role not found
{
  "error": "User role not found"
}
```

---

## Permission Constants

### Projects Permissions
- `projects:read` - View projects
- `projects:write` - Create and update projects
- `projects:delete` - Delete projects

### Tasks Permissions
- `tasks:read` - View tasks
- `tasks:write` - Create, update, and modify tasks
- `tasks:delete` - Delete tasks
- `tasks:assign` - Assign tasks to users

### User Management Permissions
- `users:read` - View user information
- `users:write` - Create and update users
- `users:delete` - Delete users
- `users:profile` - View user profiles
- `users:invite` - Invite new users

### Calendar Permissions
- `calendar:read` - View calendar events
- `calendar:write` - Create and update calendar events
- `calendar:delete` - Delete calendar events

### Vacation Permissions
- `vacation:read` - View vacation requests
- `vacation:write` - Create and update vacation requests
- `vacation:delete` - Delete vacation requests
- `vacation:approve` - Approve/reject vacation requests (Admin only)

### Notifications Permissions
- `notifications:read` - View notifications
- `notifications:write` - Mark notifications as read
- `notifications:delete` - Delete notifications

### Activity Log Permissions
- `activityLogs:read` - View activity logs

### Dashboard Permissions
- `dashboard:read` - View dashboard statistics

### Employees Permissions
- `employees:read` - View employee information and statistics

### Info Portal Permissions
- `infoPortal:read` - View info portal folders and pages
- `infoPortal:write` - Create and update info portal content
- `infoPortal:delete` - Delete info portal content

### Google Account Permissions
- `googleAccount:read` - View Google account status and linked accounts
- `googleAccount:write` - Update Google account settings
- `googleAccount:link` - Link Google account
- `googleAccount:unlink` - Unlink Google account

### WebSocket Permissions
- `websocket:connect` - Connect to WebSocket for real-time updates

### Auth Permissions
- `auth:read` - Read authentication and permission information

### Project Details Permissions
- `projectDetails:read` - View project details
- `projectDetails:write` - Create and update project details
- `projectDetails:delete` - Delete project details

---

## Permission Matrix

| Permission | Admin | Standard | Notes |
|------------|-------|----------|-------|
| **Projects** |
| `projects:read` | ✅ | ✅ | Standard users can view all projects |
| `projects:write` | ✅ | ❌ | Standard users can only modify projects they're members of (enforced by middleware) |
| `projects:delete` | ✅ | ❌ | Standard users cannot delete projects |
| **Tasks** |
| `tasks:read` | ✅ | ✅ | Standard users can view all tasks in all projects |
| `tasks:write` | ✅ | ✅ | Standard users can only modify assigned tasks or tasks in their projects |
| `tasks:delete` | ✅ | ✅ | Standard users can only delete assigned tasks or tasks in their projects |
| `tasks:assign` | ✅ | ❌ | Only admins and project members can assign tasks |
| **Users** |
| `users:read` | ✅ | ❌ | Admin only |
| `users:write` | ✅ | ❌ | Admin only |
| `users:delete` | ✅ | ❌ | Admin only |
| `users:profile` | ✅ | ✅ | All users can view profiles |
| `users:invite` | ✅ | ❌ | Admin only |
| **Calendar** |
| `calendar:read` | ✅ | ✅ | All users |
| `calendar:write` | ✅ | ✅ | All users |
| `calendar:delete` | ✅ | ✅ | All users |
| **Vacation** |
| `vacation:read` | ✅ | ✅ | All users |
| `vacation:write` | ✅ | ✅ | All users |
| `vacation:delete` | ✅ | ✅ | All users |
| `vacation:approve` | ✅ | ❌ | Admin only |
| **Notifications** |
| `notifications:read` | ✅ | ✅ | All users |
| `notifications:write` | ✅ | ✅ | All users |
| `notifications:delete` | ✅ | ✅ | All users |
| **Activity Logs** |
| `activityLogs:read` | ✅ | ✅ | All users |
| **Dashboard** |
| `dashboard:read` | ✅ | ✅ | All users |
| **Employees** |
| `employees:read` | ✅ | ✅ | All users |
| **Info Portal** |
| `infoPortal:read` | ✅ | ✅ | All users |
| `infoPortal:write` | ✅ | ✅ | All users |
| `infoPortal:delete` | ✅ | ✅ | All users |
| **Google Account** |
| `googleAccount:read` | ✅ | ✅ | All users |
| `googleAccount:write` | ✅ | ✅ | All users |
| `googleAccount:link` | ✅ | ✅ | All users (their own accounts) |
| `googleAccount:unlink` | ✅ | ✅ | All users (their own accounts) |
| **WebSocket** |
| `websocket:connect` | ✅ | ✅ | All authenticated users |
| **Auth** |
| `auth:read` | ✅ | ✅ | All authenticated users |
| **Project Details** |
| `projectDetails:read` | ✅ | ✅ | Standard users can view details for all projects |
| `projectDetails:write` | ✅ | ❌ | Standard users can only modify details for projects they're members of (enforced by middleware) |
| `projectDetails:delete` | ✅ | ❌ | Standard users cannot delete project details |

---

## Usage Examples

### JavaScript/TypeScript

```javascript
// Fetch user permissions
async function getUserPermissions() {
  const token = localStorage.getItem('authToken');
  
  const response = await fetch('http://localhost:3000/api/auth/permissions', {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    throw new Error('Failed to fetch permissions');
  }

  return await response.json();
}

// Check if user has specific permission
async function hasPermission(permission) {
  const { permissions } = await getUserPermissions();
  return permissions.includes(permission);
}

// Usage
const canCreateProject = await hasPermission('projects:write');
if (canCreateProject) {
  // Show create project button
}
```

### React Hook Example

```jsx
import { useState, useEffect } from 'react';

function usePermissions() {
  const [permissions, setPermissions] = useState([]);
  const [role, setRole] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPermissions = async () => {
      try {
        const token = localStorage.getItem('authToken');
        const response = await fetch('/api/auth/permissions', {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        });
        const data = await response.json();
        setPermissions(data.permissions);
        setRole(data.role);
      } catch (error) {
        console.error('Error fetching permissions:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchPermissions();
  }, []);

  const hasPermission = (permission) => {
    return permissions.includes(permission);
  };

  const isAdmin = () => {
    return role === 'Admin';
  };

  return { permissions, role, loading, hasPermission, isAdmin };
}

// Usage in component
function ProjectActions() {
  const { hasPermission, isAdmin } = usePermissions();

  return (
    <div>
      {hasPermission('projects:write') && (
        <button>Create Project</button>
      )}
      {isAdmin() && (
        <button>Admin Panel</button>
      )}
    </div>
  );
}
```

### Python Example

```python
import requests

def get_user_permissions(token):
    """Fetch user permissions from API"""
    url = "http://localhost:3000/api/auth/permissions"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    response = requests.get(url, headers=headers)
    response.raise_for_status()
    return response.json()

def has_permission(token, permission):
    """Check if user has specific permission"""
    data = get_user_permissions(token)
    return permission in data['permissions']

# Usage
token = "your-jwt-token"
if has_permission(token, 'projects:write'):
    print("User can create projects")
```

---

## Resource-Level Access Control

### Projects

**Viewing (GET requests):**
- **Admin users**: Can view all projects
- **Standard users**: Can view all projects (no restrictions)

**Modifying/Deleting (PUT/DELETE requests):**
- **Admin users**: Can modify/delete all projects
- **Standard users**: Can only modify/delete projects where they are:
  - **Owner** of the project (`OwnerID` matches user ID)
  - **Member** of the project (`MembersIDs` contains user ID)

### Tasks

**Viewing (GET requests):**
- **Admin users**: Can view all tasks
- **Standard users**: Can view all tasks in all projects (no restrictions)

**Modifying/Deleting (PUT/DELETE/POST requests):**
- **Admin users**: Can modify/delete all tasks
- **Standard users**: Can only modify/delete tasks if they are:
  - **Assigned** to the task (`AssignTo` matches user ID) - Full access
  - **Project owner** (`OwnerID` matches user ID) - Can modify tasks in their projects
  - **Project member** (`MembersIDs` contains user ID) - Can modify tasks in their projects

### Project Details

**Viewing (GET requests):**
- **Admin users**: Can view all project details
- **Standard users**: Can view project details for all projects (no restrictions)

**Modifying/Deleting (PUT/DELETE requests):**
- **Admin users**: Can modify/delete all project details
- **Standard users**: Can only modify/delete project details for projects where they are:
  - **Project owner or member** - Can modify details for their projects

---

## Error Handling

### Common Errors

**401 Unauthorized - Invalid Token**
```json
{
  "error": "Invalid token"
}
```
**Solution:** Ensure the JWT token is valid and not expired. Re-authenticate if necessary.

**401 Unauthorized - User Role Not Found**
```json
{
  "error": "User role not found"
}
```
**Solution:** This typically occurs when the user's role is not set in the JWT token. Contact support if this persists.

**403 Forbidden - Permission Denied**
```json
{
  "error": "Permission denied: projects:write"
}
```
**Solution:** The user does not have the required permission. Check the permission matrix above to understand what permissions are available for each role.

---

## Best Practices

### 1. Cache Permissions

Cache permissions in localStorage or sessionStorage to reduce API calls:

```javascript
const CACHE_KEY = 'user_permissions';
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

async function getCachedPermissions() {
  const cached = localStorage.getItem(CACHE_KEY);
  if (cached) {
    const { data, timestamp } = JSON.parse(cached);
    if (Date.now() - timestamp < CACHE_DURATION) {
      return data;
    }
  }

  const data = await getUserPermissions();
  localStorage.setItem(CACHE_KEY, JSON.stringify({
    data,
    timestamp: Date.now()
  }));
  
  return data;
}
```

### 2. Refresh on Login/Logout

Always refresh permissions when user logs in or out:

```javascript
function handleLogin(token) {
  localStorage.setItem('authToken', token);
  // Clear cached permissions
  localStorage.removeItem('user_permissions');
  // Fetch fresh permissions
  getUserPermissions();
}

function handleLogout() {
  localStorage.removeItem('authToken');
  localStorage.removeItem('user_permissions');
}
```

### 3. Use Permission Constants

Define permission constants to avoid typos:

```javascript
const PERMISSIONS = {
  PROJECTS: {
    READ: 'projects:read',
    WRITE: 'projects:write',
    DELETE: 'projects:delete'
  },
  TASKS: {
    READ: 'tasks:read',
    WRITE: 'tasks:write',
    DELETE: 'tasks:delete',
    ASSIGN: 'tasks:assign'
  }
  // ... more permissions
};

// Usage
if (hasPermission(PERMISSIONS.PROJECTS.WRITE)) {
  // Create project
}
```

### 4. Check Permissions Before API Calls

Always validate permissions before making API calls:

```javascript
async function createProject(projectData) {
  if (!hasPermission('projects:write')) {
    throw new Error('You do not have permission to create projects');
  }
  
  // Proceed with API call
  const response = await fetch('/api/project/add', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(projectData)
  });
  
  return response.json();
}
```

---

## Integration with Other APIs

### Using Permissions with Project API

```javascript
// Check permission before calling project API
const { permissions } = await getUserPermissions();

if (permissions.includes('projects:write')) {
  // User can create projects
  const response = await fetch('/api/project/add', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(projectData)
  });
} else {
  alert('You do not have permission to create projects');
}
```

### Using Permissions with Task API

```javascript
// Check permission before assigning tasks
const { permissions } = await getUserPermissions();

if (permissions.includes('tasks:assign')) {
  // User can assign tasks
  const response = await fetch(`/api/tasks/assign/${taskId}/${userId}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
} else {
  alert('You do not have permission to assign tasks');
}
```

---

## Testing

### Test with Admin User

```bash
# Get admin token (from login)
ADMIN_TOKEN="your-admin-token"

# Fetch permissions
curl -X GET "http://localhost:3000/api/auth/permissions" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json"
```

### Test with Standard User

```bash
# Get standard user token (from login)
USER_TOKEN="your-user-token"

# Fetch permissions
curl -X GET "http://localhost:3000/api/auth/permissions" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json"
```

---

## Rate Limiting

The Permissions API endpoint is rate-limited to prevent abuse:
- **Limit:** 100 requests per minute per user
- **Response:** 429 Too Many Requests if limit exceeded

---

## Changelog

### Version 1.0 (2025-01-XX)
- Initial release
- Added GetPermissions endpoint
- Added all permission constants
- Added permission matrix for Admin and Standard roles

---

**Last Updated:** 2025-01-XX  
**API Version:** 1.0

