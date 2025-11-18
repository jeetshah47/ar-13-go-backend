# User Management API Documentation

This document describes the User Management API for viewing and managing user profiles and permissions.

## Base URL

All endpoints are prefixed with `/api/users`

## Authentication

All endpoints require authentication. Include the authentication token in the request headers:

```
Authorization: Bearer <your-token>
```

## Overview

The User Management API provides endpoints for:
- Viewing user profiles with permissions
- Getting user permissions by user ID
- Managing user accounts (create, update, delete)
- Inviting new users

## Endpoints

### 1. Get All Users

Retrieves a list of all users in the system.

**Endpoint:** `GET /api/users/all`

**Permissions Required:** None (all authenticated users can view user list)

**Request:**
```http
GET /api/users/all
Authorization: Bearer <token>
Content-Type: application/json
```

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": "user-id-1",
      "name": "John Doe",
      "email": "john@example.com",
      "phoneNumber": "+1234567890",
      "role": "Admin",
      "designation": "Software Engineer",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    },
    {
      "id": "user-id-2",
      "name": "Jane Smith",
      "email": "jane@example.com",
      "phoneNumber": "+1234567891",
      "role": "Standard",
      "designation": "Developer",
      "createdAt": "2025-01-02T00:00:00Z",
      "updatedAt": "2025-01-02T00:00:00Z"
    }
  ]
}
```

**Response Fields:**
- `users`: Array of user objects
  - `id`: Unique user identifier
  - `name`: User's full name
  - `email`: User's email address
  - `phoneNumber`: User's phone number
  - `role`: User's role (`Admin` or `Standard`)
  - `designation`: User's job designation (optional)
  - `createdAt`: User creation timestamp
  - `updatedAt`: User last update timestamp

**Error Responses:**

```json
// 401 Unauthorized - Missing or invalid token
{
  "error": "Invalid token"
}

// 500 Internal Server Error
{
  "error": "Failed to retrieve users"
}
```

---

### 2. Get User Profile with Permissions

Retrieves a user's profile information along with their role and permissions list.

**Endpoint:** `GET /api/users/profile/:id`

**Permissions Required:** None (all authenticated users can view profiles)

**Path Parameters:**
- `id` (string, required): The user ID

**Request:**
```http
GET /api/users/profile/user-id-123
Authorization: Bearer <token>
Content-Type: application/json
```

**Response (200 OK):**
```json
{
  "user": {
    "id": "user-id-123",
    "name": "John Doe",
    "email": "john@example.com",
    "phoneNumber": "+1234567890",
    "role": "Admin",
    "designation": "Software Engineer",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  },
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
    "projectDetails:delete",
    "backup:read",
    "backup:write"
  ]
}
```

**Response Fields:**
- `user`: Complete user object with all user information
- `role`: User's role (`Admin` or `Standard`)
- `permissions`: Array of permission strings available to the user based on their role

**Error Responses:**

```json
// 404 Not Found - User not found
{
  "error": "User not found"
}

// 500 Internal Server Error - Failed to get permissions
{
  "error": "Failed to get permissions"
}

// 401 Unauthorized - Missing or invalid token
{
  "error": "Invalid token"
}
```

---

### 3. Get User Permissions

Retrieves the permissions list for a specific user by their user ID. This endpoint is useful for admin pages that need to display and manage user permissions.

**Endpoint:** `GET /api/users/permissions/:id`

**Permissions Required:** `users:read` (Admin users have this permission)

**Path Parameters:**
- `id` (string, required): The user ID

**Request:**
```http
GET /api/users/permissions/user-id-123
Authorization: Bearer <token>
Content-Type: application/json
```

**Response (200 OK):**
```json
{
  "userId": "user-id-123",
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
    "projectDetails:delete",
    "backup:read",
    "backup:write"
  ]
}
```

**Response for Standard User:**
```json
{
  "userId": "user-id-456",
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
- `userId`: The user's unique identifier
- `role`: User's role (`Admin` or `Standard`)
- `permissions`: Array of permission strings available to the user based on their role

**Error Responses:**

```json
// 404 Not Found - User not found
{
  "error": "User not found"
}

// 403 Forbidden - Permission denied
{
  "error": "Permission denied: users:read"
}

// 500 Internal Server Error - Failed to get permissions
{
  "error": "Failed to get permissions"
}

// 401 Unauthorized - Missing or invalid token
{
  "error": "Invalid token"
}
```

---

### 4. Update User

Updates a user's information including their role. When the role is updated, the user's permissions automatically change based on the new role.

**Endpoint:** `PUT /api/users/update`

**Permissions Required:** Admin only (`RequireAdmin()` middleware)

**Request:**
```http
PUT /api/users/update
Authorization: Bearer <admin-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "id": "user-id-123",
  "name": "John Doe Updated",
  "email": "john.updated@example.com",
  "phoneNumber": "+1234567890",
  "role": "Standard",
  "designation": "Senior Developer"
}
```

**Request Fields:**
- `id` (string, required): The user ID to update
- `name` (string, optional): User's full name
- `email` (string, optional): User's email address
- `phoneNumber` (string, optional): User's phone number
- `role` (string, optional): User's role (`Admin` or `Standard`)
- `designation` (string, optional): User's job designation

**Response (200 OK):**
```json
{
  "message": "User updated successfully"
}
```

**Error Responses:**

```json
// 400 Bad Request - Invalid request body
{
  "error": "Invalid request format"
}

// 400 Bad Request - User not found
{
  "error": "User not found"
}

// 403 Forbidden - Admin access required
{
  "error": "Admin access required"
}

// 401 Unauthorized - Missing or invalid token
{
  "error": "Invalid token"
}
```

**Note:** When updating a user's role, their permissions will automatically be updated to match the permissions for the new role. For example:
- Changing role from `Standard` to `Admin` will grant all admin permissions
- Changing role from `Admin` to `Standard` will restrict permissions to standard user permissions

---

### 5. Delete User

Deletes a user from the system.

**Endpoint:** `DELETE /api/users/delete/:id`

**Permissions Required:** Admin only (`RequireAdmin()` middleware)

**Path Parameters:**
- `id` (string, required): The user ID to delete

**Request:**
```http
DELETE /api/users/delete/user-id-123
Authorization: Bearer <admin-token>
Content-Type: application/json
```

**Response (200 OK):**
```json
{
  "message": "User deleted successfully"
}
```

**Error Responses:**

```json
// 404 Not Found - User not found
{
  "error": "User not found"
}

// 403 Forbidden - Admin access required
{
  "error": "Admin access required"
}

// 401 Unauthorized - Missing or invalid token
{
  "error": "Invalid token"
}
```

---

### 6. Create User Invitation

Creates a signup invitation and sends it to the specified email address.

**Endpoint:** `POST /api/users/invite`

**Permissions Required:** Admin only (`RequireAdmin()` middleware)

**Request:**
```http
POST /api/users/invite
Authorization: Bearer <admin-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "email": "newuser@example.com"
}
```

**Request Fields:**
- `email` (string, required): Email address to send the invitation to

**Response (201 Created):**
```json
{
  "message": "Signup invitation sent successfully"
}
```

**Error Responses:**

```json
// 400 Bad Request - Invalid email format
{
  "error": "Invalid email format"
}

// 400 Bad Request - Email already taken
{
  "error": "email already taken"
}

// 400 Bad Request - Invitation already sent
{
  "error": "invitation already sent to this email"
}

// 403 Forbidden - Admin access required
{
  "error": "Admin access required"
}

// 401 Unauthorized - Missing or invalid token
{
  "error": "Invalid token"
}
```

---

## Permission-Based Access Control

### Role-Based Permissions

User permissions are automatically determined by their role:

- **Admin Role**: Has all permissions including user management, project management, task assignment, and vacation approval
- **Standard Role**: Has limited permissions focused on viewing and managing their own work (projects they're members of, tasks they're assigned to)

### Permission Updates

When a user's role is updated via the `PUT /api/users/update` endpoint:
1. The user's role field is updated in the database
2. Their permissions are automatically recalculated based on the new role
3. The new permissions take effect immediately

### Viewing Permissions

There are two ways to view user permissions:

1. **Get User Profile** (`GET /api/users/profile/:id`): Returns user profile with permissions (accessible to all authenticated users)
2. **Get User Permissions** (`GET /api/users/permissions/:id`): Returns only permissions (requires `users:read` permission, typically Admin only)

---

## Use Cases

### Admin Viewing User Profile and Permissions

An admin can view a user's complete profile including their permissions:

1. Call `GET /api/users/profile/:id` to get user profile with permissions
2. Display the user information and permissions list in the admin UI
3. Optionally call `GET /api/users/permissions/:id` if only permissions are needed

### Admin Updating User Role

An admin can update a user's role to change their permissions:

1. Call `PUT /api/users/update` with the user ID and new role
2. The user's permissions will automatically update based on the new role
3. Verify the change by calling `GET /api/users/permissions/:id`

### Standard User Viewing Their Own Profile

A standard user can view their own profile and permissions:

1. Call `GET /api/users/profile/:id` with their own user ID
2. View their role and available permissions

---

## Response Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 OK | Request successful |
| 201 Created | Resource created successfully |
| 400 Bad Request | Invalid request format or data |
| 401 Unauthorized | Missing or invalid authentication token |
| 403 Forbidden | User does not have required permissions |
| 404 Not Found | User not found |
| 500 Internal Server Error | Server error occurred |

---

## Notes

1. **Password Field**: The password field is never returned in API responses for security reasons
2. **Role Values**: Only `Admin` and `Standard` are valid role values
3. **Permission Inheritance**: Permissions are role-based and cannot be individually assigned. To change permissions, update the user's role
4. **Email Uniqueness**: Email addresses must be unique across all users
5. **Invitation Expiry**: Signup invitations expire after 7 days

---

**Last Updated:** 2025-01-XX  
**API Version:** 1.0

