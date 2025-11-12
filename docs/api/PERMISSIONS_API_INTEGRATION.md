# Permissions API Integration Guide

This guide provides step-by-step instructions for integrating the Permissions API into your UI application to implement role-based access control (RBAC).

## Table of Contents

1. [Overview](#overview)
2. [API Endpoint](#api-endpoint)
3. [Quick Start](#quick-start)
4. [React Integration](#react-integration)
5. [Vue.js Integration](#vuejs-integration)
6. [Vanilla JavaScript](#vanilla-javascript)
7. [TypeScript Types](#typescript-types)
8. [Permission Checking Utilities](#permission-checking-utilities)
9. [Common Use Cases](#common-use-cases)
10. [View vs Modify Permissions](#view-vs-modify-permissions)
11. [Best Practices](#best-practices)
12. [Troubleshooting](#troubleshooting)

## Overview

The Permissions API allows you to:
- Get the current user's role and available permissions
- Implement UI-level access control
- Show/hide features based on user permissions
- Prevent unauthorized actions before API calls

### Permission System

The system uses role-based permissions with two main roles:

- **Admin**: Full access to all features
- **Standard**: 
  - Can view all projects and tasks (read-only access)
  - Can only modify/delete resources they're assigned to or are members of
  - Write/delete operations are restricted to project members or assigned tasks

## API Endpoint

### Get User Permissions

**Endpoint:** `GET /api/auth/permissions`

**Authentication:** Required (Bearer Token)

**Request Headers:**
```
Authorization: Bearer <your-jwt-token>
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
    "infoPortal:delete"
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
    "infoPortal:delete"
  ]
}
```

**Error Responses:**

```json
// 401 Unauthorized
{
  "error": "User role not found"
}

// 401 Unauthorized (Invalid Token)
{
  "error": "Invalid token"
}
```

## Quick Start

### Basic Example

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

  const data = await response.json();
  return data;
}

// Usage
getUserPermissions()
  .then(data => {
    console.log('User Role:', data.role);
    console.log('Permissions:', data.permissions);
    
    // Check if user can create projects
    if (data.permissions.includes('projects:write')) {
      // Show create project button
    }
  })
  .catch(error => {
    console.error('Error:', error);
  });
```

## React Integration

### 1. Create Permissions Context

```jsx
// contexts/PermissionsContext.jsx
import React, { createContext, useContext, useState, useEffect } from 'react';

const PermissionsContext = createContext();

export const usePermissions = () => {
  const context = useContext(PermissionsContext);
  if (!context) {
    throw new Error('usePermissions must be used within PermissionsProvider');
  }
  return context;
};

export const PermissionsProvider = ({ children }) => {
  const [permissions, setPermissions] = useState([]);
  const [role, setRole] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchPermissions();
  }, []);

  const fetchPermissions = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('authToken');
      
      const response = await fetch('http://localhost:3000/api/auth/permissions', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch permissions');
      }

      const data = await response.json();
      setPermissions(data.permissions);
      setRole(data.role);
      setError(null);
    } catch (err) {
      setError(err.message);
      console.error('Error fetching permissions:', err);
    } finally {
      setLoading(false);
    }
  };

  const hasPermission = (permission) => {
    return permissions.includes(permission);
  };

  const hasAnyPermission = (permissionList) => {
    return permissionList.some(perm => permissions.includes(perm));
  };

  const hasAllPermissions = (permissionList) => {
    return permissionList.every(perm => permissions.includes(perm));
  };

  const isAdmin = () => {
    return role === 'Admin';
  };

  const value = {
    permissions,
    role,
    loading,
    error,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    isAdmin,
    refreshPermissions: fetchPermissions
  };

  return (
    <PermissionsContext.Provider value={value}>
      {children}
    </PermissionsContext.Provider>
  );
};
```

### 2. Wrap Your App

```jsx
// App.jsx
import { PermissionsProvider } from './contexts/PermissionsContext';

function App() {
  return (
    <PermissionsProvider>
      {/* Your app components */}
    </PermissionsProvider>
  );
}
```

### 3. Use Permissions in Components

```jsx
// components/ProjectActions.jsx
import { usePermissions } from '../contexts/PermissionsContext';

function ProjectActions({ projectId }) {
  const { hasPermission, isAdmin } = usePermissions();

  return (
    <div>
      {hasPermission('projects:write') && (
        <button onClick={handleEdit}>Edit Project</button>
      )}
      
      {hasPermission('projects:delete') && (
        <button onClick={handleDelete}>Delete Project</button>
      )}
      
      {hasPermission('tasks:assign') && (
        <button onClick={handleAssignTask}>Assign Task</button>
      )}
    </div>
  );
}
```

### 4. Permission-Based Route Protection

```jsx
// components/ProtectedRoute.jsx
import { Navigate } from 'react-router-dom';
import { usePermissions } from '../contexts/PermissionsContext';

function ProtectedRoute({ children, requiredPermission, requiredRole }) {
  const { hasPermission, role, loading } = usePermissions();

  if (loading) {
    return <div>Loading...</div>;
  }

  if (requiredRole && role !== requiredRole) {
    return <Navigate to="/unauthorized" />;
  }

  if (requiredPermission && !hasPermission(requiredPermission)) {
    return <Navigate to="/unauthorized" />;
  }

  return children;
}

// Usage
<Route
  path="/admin/users"
  element={
    <ProtectedRoute requiredRole="Admin">
      <UserManagement />
    </ProtectedRoute>
  }
/>
```

### 5. Conditional Rendering Hook

```jsx
// hooks/usePermissionCheck.js
import { usePermissions } from '../contexts/PermissionsContext';

export const usePermissionCheck = () => {
  const { hasPermission, hasAnyPermission, hasAllPermissions, isAdmin } = usePermissions();

  return {
    canCreateProject: hasPermission('projects:write'),
    canDeleteProject: hasPermission('projects:delete'),
    canAssignTask: hasPermission('tasks:assign'),
    canApproveVacation: hasPermission('vacation:approve'),
    canManageUsers: isAdmin() || hasPermission('users:write'),
    canViewDashboard: hasPermission('dashboard:read'),
    isAdmin: isAdmin()
  };
};
```

## Vue.js Integration

### 1. Create Permissions Plugin

```javascript
// plugins/permissions.js
export default {
  install(app) {
    let permissions = [];
    let role = null;

    const fetchPermissions = async () => {
      try {
        const token = localStorage.getItem('authToken');
        const response = await fetch('http://localhost:3000/api/auth/permissions', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });

        if (!response.ok) {
          throw new Error('Failed to fetch permissions');
        }

        const data = await response.json();
        permissions = data.permissions;
        role = data.role;
      } catch (error) {
        console.error('Error fetching permissions:', error);
      }
    };

    const hasPermission = (permission) => {
      return permissions.includes(permission);
    };

    const isAdmin = () => {
      return role === 'Admin';
    };

    app.config.globalProperties.$permissions = {
      list: permissions,
      role: role,
      hasPermission,
      isAdmin,
      refresh: fetchPermissions
    };

    // Fetch permissions on app initialization
    fetchPermissions();
  }
};
```

### 2. Register Plugin

```javascript
// main.js
import { createApp } from 'vue';
import App from './App.vue';
import permissionsPlugin from './plugins/permissions';

const app = createApp(App);
app.use(permissionsPlugin);
app.mount('#app');
```

### 3. Use in Components

```vue
<template>
  <div>
    <button v-if="$permissions.hasPermission('projects:write')" @click="editProject">
      Edit Project
    </button>
    
    <button v-if="$permissions.hasPermission('projects:delete')" @click="deleteProject">
      Delete Project
    </button>
    
    <button v-if="$permissions.isAdmin()" @click="manageUsers">
      Manage Users
    </button>
  </div>
</template>

<script>
export default {
  methods: {
    editProject() {
      // Edit logic
    },
    deleteProject() {
      // Delete logic
    },
    manageUsers() {
      // User management logic
    }
  }
}
</script>
```

### 4. Composition API

```vue
<script setup>
import { ref, onMounted } from 'vue';

const permissions = ref([]);
const role = ref(null);

const fetchPermissions = async () => {
  const token = localStorage.getItem('authToken');
  const response = await fetch('http://localhost:3000/api/auth/permissions', {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  const data = await response.json();
  permissions.value = data.permissions;
  role.value = data.role;
};

const hasPermission = (permission) => {
  return permissions.value.includes(permission);
};

onMounted(() => {
  fetchPermissions();
});
</script>

<template>
  <div>
    <button v-if="hasPermission('projects:write')">Edit</button>
  </div>
</template>
```

## Vanilla JavaScript

### Permission Manager Class

```javascript
// PermissionManager.js
class PermissionManager {
  constructor(apiBaseUrl) {
    this.apiBaseUrl = apiBaseUrl;
    this.permissions = [];
    this.role = null;
    this.listeners = [];
  }

  async fetchPermissions() {
    try {
      const token = localStorage.getItem('authToken');
      const response = await fetch(`${this.apiBaseUrl}/api/auth/permissions`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch permissions');
      }

      const data = await response.json();
      this.permissions = data.permissions;
      this.role = data.role;
      
      // Notify listeners
      this.listeners.forEach(listener => listener(this.permissions, this.role));
      
      return data;
    } catch (error) {
      console.error('Error fetching permissions:', error);
      throw error;
    }
  }

  hasPermission(permission) {
    return this.permissions.includes(permission);
  }

  hasAnyPermission(permissionList) {
    return permissionList.some(perm => this.hasPermission(perm));
  }

  hasAllPermissions(permissionList) {
    return permissionList.every(perm => this.hasPermission(perm));
  }

  isAdmin() {
    return this.role === 'Admin';
  }

  onUpdate(callback) {
    this.listeners.push(callback);
  }

  checkAndShow(elementId, permission) {
    const element = document.getElementById(elementId);
    if (element) {
      element.style.display = this.hasPermission(permission) ? 'block' : 'none';
    }
  }

  checkAndEnable(elementId, permission) {
    const element = document.getElementById(elementId);
    if (element) {
      element.disabled = !this.hasPermission(permission);
    }
  }
}

// Usage
const permissionManager = new PermissionManager('http://localhost:3000');

// Fetch permissions on page load
permissionManager.fetchPermissions().then(() => {
  // Show/hide elements based on permissions
  permissionManager.checkAndShow('createProjectBtn', 'projects:write');
  permissionManager.checkAndShow('deleteProjectBtn', 'projects:delete');
  permissionManager.checkAndEnable('assignTaskBtn', 'tasks:assign');
});

// Use in event handlers
document.getElementById('createProjectBtn')?.addEventListener('click', () => {
  if (permissionManager.hasPermission('projects:write')) {
    // Create project logic
  } else {
    alert('You do not have permission to create projects');
  }
});
```

## TypeScript Types

```typescript
// types/permissions.ts
export type UserRole = 'Admin' | 'Standard';

export type Permission =
  | 'projects:read'
  | 'projects:write'
  | 'projects:delete'
  | 'tasks:read'
  | 'tasks:write'
  | 'tasks:delete'
  | 'tasks:assign'
  | 'users:read'
  | 'users:write'
  | 'users:delete'
  | 'calendar:read'
  | 'calendar:write'
  | 'calendar:delete'
  | 'vacation:read'
  | 'vacation:write'
  | 'vacation:delete'
  | 'vacation:approve'
  | 'notifications:read'
  | 'notifications:write'
  | 'notifications:delete'
  | 'activityLogs:read'
  | 'dashboard:read'
  | 'employees:read'
  | 'infoPortal:read'
  | 'infoPortal:write'
  | 'infoPortal:delete';

export interface PermissionsResponse {
  role: UserRole;
  permissions: Permission[];
}

export interface PermissionManager {
  permissions: Permission[];
  role: UserRole | null;
  loading: boolean;
  error: string | null;
  hasPermission(permission: Permission): boolean;
  hasAnyPermission(permissions: Permission[]): boolean;
  hasAllPermissions(permissions: Permission[]): boolean;
  isAdmin(): boolean;
  refreshPermissions(): Promise<void>;
}
```

## Permission Checking Utilities

### Utility Functions

```javascript
// utils/permissions.js

/**
 * Check if user has a specific permission
 */
export const hasPermission = (permissions, permission) => {
  return permissions.includes(permission);
};

/**
 * Check if user has any of the specified permissions
 */
export const hasAnyPermission = (permissions, permissionList) => {
  return permissionList.some(perm => permissions.includes(perm));
};

/**
 * Check if user has all of the specified permissions
 */
export const hasAllPermissions = (permissions, permissionList) => {
  return permissionList.every(perm => permissions.includes(perm));
};

/**
 * Check if user is admin
 */
export const isAdmin = (role) => {
  return role === 'Admin';
};

/**
 * Get permission display name
 */
export const getPermissionDisplayName = (permission) => {
  const displayNames = {
    'projects:read': 'View Projects',
    'projects:write': 'Create/Edit Projects',
    'projects:delete': 'Delete Projects',
    'tasks:read': 'View Tasks',
    'tasks:write': 'Create/Edit Tasks',
    'tasks:delete': 'Delete Tasks',
    'tasks:assign': 'Assign Tasks',
    'users:read': 'View Users',
    'users:write': 'Create/Edit Users',
    'users:delete': 'Delete Users',
    'vacation:approve': 'Approve Vacation Requests'
  };
  
  return displayNames[permission] || permission;
};
```

## Common Use Cases

### 1. Show/Hide UI Elements

```jsx
// React Example
function ProjectToolbar() {
  const { hasPermission } = usePermissions();

  return (
    <div className="toolbar">
      {hasPermission('projects:write') && (
        <button onClick={createProject}>Create Project</button>
      )}
      {hasPermission('projects:delete') && (
        <button onClick={deleteProject}>Delete Project</button>
      )}
    </div>
  );
}
```

### 1.1. Handling View vs Modify Permissions

Since Standard users can view all projects and tasks but can only modify ones they have access to, you need to check both permissions and resource-level access:

```jsx
// React Example - Project List with Conditional Actions
function ProjectList({ projects, currentUserId }) {
  const { hasPermission, isAdmin } = usePermissions();

  const canModifyProject = (project) => {
    // Admins can modify all projects
    if (isAdmin()) return true;
    
    // Standard users can only modify if they're owner or member
    return project.ownerId === currentUserId || 
           project.membersIds?.includes(currentUserId);
  };

  return (
    <div>
      {projects.map(project => (
        <div key={project.id} className="project-card">
          <h3>{project.name}</h3>
          <p>{project.description}</p>
          
          {/* All users can view project details */}
          <button onClick={() => viewProject(project.id)}>
            View Details
          </button>
          
          {/* Only show edit/delete if user has permission AND access */}
          {hasPermission('projects:write') && canModifyProject(project) && (
            <button onClick={() => editProject(project.id)}>
              Edit Project
            </button>
          )}
          
          {hasPermission('projects:delete') && canModifyProject(project) && (
            <button onClick={() => deleteProject(project.id)}>
              Delete Project
            </button>
          )}
        </div>
      ))}
    </div>
  );
}
```

```jsx
// React Example - Task List with Conditional Actions
function TaskList({ tasks, project, currentUserId }) {
  const { hasPermission, isAdmin } = usePermissions();

  const canModifyTask = (task) => {
    // Admins can modify all tasks
    if (isAdmin()) return true;
    
    // Standard users can modify if:
    // 1. They're assigned to the task
    // 2. They're project owner or member
    const isAssigned = task.assignTo === currentUserId;
    const isProjectOwner = project.ownerId === currentUserId;
    const isProjectMember = project.membersIds?.includes(currentUserId);
    
    return isAssigned || isProjectOwner || isProjectMember;
  };

  return (
    <div>
      {tasks.map(task => (
        <div key={task.id} className="task-card">
          <h4>{task.title}</h4>
          <p>Assigned to: {task.assignTo}</p>
          
          {/* All users can view task details */}
          <button onClick={() => viewTask(task.id)}>
            View Details
          </button>
          
          {/* Only show edit/delete if user has permission AND access */}
          {hasPermission('tasks:write') && canModifyTask(task) && (
            <button onClick={() => editTask(task.id)}>
              Edit Task
            </button>
          )}
          
          {hasPermission('tasks:delete') && canModifyTask(task) && (
            <button onClick={() => deleteTask(task.id)}>
              Delete Task
            </button>
          )}
          
          {/* Show disabled state if user can view but not modify */}
          {hasPermission('tasks:write') && !canModifyTask(task) && (
            <button disabled title="You can only modify tasks you're assigned to or tasks in your projects">
              Edit Task (Restricted)
            </button>
          )}
        </div>
      ))}
    </div>
  );
}
```

### 2. Disable Actions

```jsx
function TaskActions({ task, project, currentUserId }) {
  const { hasPermission, isAdmin } = usePermissions();

  // Check if user can modify this specific task
  const canModify = () => {
    if (isAdmin()) return true;
    if (task.assignTo === currentUserId) return true;
    if (project.ownerId === currentUserId) return true;
    if (project.membersIds?.includes(currentUserId)) return true;
    return false;
  };

  return (
    <div>
      <button 
        disabled={!hasPermission('tasks:write') || !canModify()}
        onClick={updateTask}
        title={!canModify() ? "You can only modify tasks you're assigned to or tasks in your projects" : ""}
      >
        Update Task
      </button>
    </div>
  );
}
```

### 3. Conditional Navigation

```jsx
function Navigation() {
  const { hasPermission, isAdmin } = usePermissions();

  return (
    <nav>
      <Link to="/projects">Projects</Link>
      {hasPermission('dashboard:read') && (
        <Link to="/dashboard">Dashboard</Link>
      )}
      {isAdmin() && (
        <Link to="/admin">Admin Panel</Link>
      )}
    </nav>
  );
}
```

### 4. Pre-API Call Validation

```javascript
async function createProject(projectData) {
  const { hasPermission } = usePermissions();
  
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

// Update project with resource-level access check
async function updateProject(projectId, projectData, project) {
  const { hasPermission, isAdmin } = usePermissions();
  
  if (!hasPermission('projects:write')) {
    throw new Error('You do not have permission to update projects');
  }

  // Check resource-level access for Standard users
  if (!isAdmin()) {
    const isOwner = project.ownerId === currentUserId;
    const isMember = project.membersIds?.includes(currentUserId);
    
    if (!isOwner && !isMember) {
      throw new Error('You can only update projects you are a member of');
    }
  }

  // Proceed with API call
  const response = await fetch('/api/project/update', {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ ...projectData, id: projectId })
  });

  return response.json();
}
```

### 5. Permission-Based Filtering with Resource Access

```javascript
function filterAvailableActions(actions, permissions, resource, currentUserId, isAdmin) {
  return actions.filter(action => {
    if (!action.requiredPermission) return true;
    
    // Check permission first
    if (!permissions.includes(action.requiredPermission)) {
      return false;
    }
    
    // For write/delete actions, check resource-level access
    if (action.requiresResourceAccess && resource) {
      if (isAdmin) return true;
      
      // For projects
      if (resource.ownerId === currentUserId) return true;
      if (resource.membersIds?.includes(currentUserId)) return true;
      
      return false;
    }
    
    return true;
  });
}

const allActions = [
  { id: 'view', label: 'View', requiredPermission: 'projects:read' },
  { id: 'edit', label: 'Edit', requiredPermission: 'projects:write', requiresResourceAccess: true },
  { id: 'delete', label: 'Delete', requiredPermission: 'projects:delete', requiresResourceAccess: true }
];

// Usage
const availableActions = filterAvailableActions(
  allActions, 
  userPermissions, 
  project, 
  currentUserId,
  isAdmin
);
```

## View vs Modify Permissions

### Understanding the Distinction

The permission system distinguishes between **viewing** and **modifying** resources:

- **Viewing (GET requests)**: Standard users can view ALL projects and tasks
- **Modifying (PUT/DELETE/POST requests)**: Standard users can only modify resources they have access to

### Implementation Strategy

#### 1. Always Show All Resources in Lists

```jsx
// ✅ CORRECT: Show all projects/tasks in lists
function ProjectList() {
  const { hasPermission } = usePermissions();
  
  // Fetch all projects - Standard users can see all
  const [projects, setProjects] = useState([]);
  
  useEffect(() => {
    fetch('/api/project/all', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
      .then(res => res.json())
      .then(data => setProjects(data.projects));
  }, []);
  
  return (
    <div>
      {projects.map(project => (
        <ProjectCard 
          key={project.id} 
          project={project}
          // Pass project data to check modify access
        />
      ))}
    </div>
  );
}
```

#### 2. Check Resource-Level Access for Actions

```jsx
// ✅ CORRECT: Check both permission AND resource access
function ProjectCard({ project, currentUserId }) {
  const { hasPermission, isAdmin } = usePermissions();
  
  // Check if user can modify THIS specific project
  const canModify = isAdmin() || 
                    project.ownerId === currentUserId || 
                    project.membersIds?.includes(currentUserId);
  
  const hasWritePermission = hasPermission('projects:write');
  const canEdit = hasWritePermission && canModify;
  
  return (
    <div>
      <h3>{project.name}</h3>
      
      {/* Always show view button if user has read permission */}
      {hasPermission('projects:read') && (
        <button onClick={() => viewProject(project.id)}>
          View
        </button>
      )}
      
      {/* Only show edit if user has permission AND can modify this project */}
      {canEdit && (
        <button onClick={() => editProject(project.id)}>
          Edit
        </button>
      )}
      
      {/* Show disabled state if user can view but not modify */}
      {hasWritePermission && !canModify && (
        <button disabled title="You can only edit projects you're a member of">
          Edit (Restricted)
        </button>
      )}
    </div>
  );
}
```

#### 3. Utility Functions for Access Checking

```javascript
// utils/accessControl.js

/**
 * Check if user can modify a project
 */
export const canModifyProject = (project, currentUserId, isAdmin) => {
  if (isAdmin) return true;
  if (project.ownerId === currentUserId) return true;
  if (project.membersIds?.includes(currentUserId)) return true;
  return false;
};

/**
 * Check if user can modify a task
 */
export const canModifyTask = (task, project, currentUserId, isAdmin) => {
  if (isAdmin) return true;
  if (task.assignTo === currentUserId) return true;
  if (project.ownerId === currentUserId) return true;
  if (project.membersIds?.includes(currentUserId)) return true;
  return false;
};

/**
 * Get available actions for a project
 */
export const getProjectActions = (project, permissions, currentUserId, isAdmin) => {
  const actions = [];
  
  if (permissions.includes('projects:read')) {
    actions.push({ id: 'view', label: 'View', enabled: true });
  }
  
  const canModify = canModifyProject(project, currentUserId, isAdmin);
  
  if (permissions.includes('projects:write') && canModify) {
    actions.push({ id: 'edit', label: 'Edit', enabled: true });
  }
  
  if (permissions.includes('projects:delete') && canModify) {
    actions.push({ id: 'delete', label: 'Delete', enabled: true });
  }
  
  return actions;
};
```

#### 4. React Hook for Resource Access

```jsx
// hooks/useResourceAccess.js
import { usePermissions } from './usePermissions';
import { canModifyProject, canModifyTask } from '../utils/accessControl';

export const useResourceAccess = () => {
  const { permissions, role, hasPermission, isAdmin } = usePermissions();
  
  const checkProjectAccess = (project, currentUserId) => {
    return {
      canView: hasPermission('projects:read'),
      canModify: canModifyProject(project, currentUserId, isAdmin()),
      canDelete: hasPermission('projects:delete') && 
                 canModifyProject(project, currentUserId, isAdmin())
    };
  };
  
  const checkTaskAccess = (task, project, currentUserId) => {
    return {
      canView: hasPermission('tasks:read'),
      canModify: canModifyTask(task, project, currentUserId, isAdmin()),
      canDelete: hasPermission('tasks:delete') && 
                 canModifyTask(task, project, currentUserId, isAdmin())
    };
  };
  
  return {
    checkProjectAccess,
    checkTaskAccess
  };
};

// Usage
function ProjectCard({ project, currentUserId }) {
  const { checkProjectAccess } = useResourceAccess();
  const access = checkProjectAccess(project, currentUserId);
  
  return (
    <div>
      <h3>{project.name}</h3>
      {access.canView && <button>View</button>}
      {access.canModify && <button>Edit</button>}
      {access.canDelete && <button>Delete</button>}
    </div>
  );
}
```

#### 5. Visual Indicators

```jsx
// Show visual indicators for read-only items
function ProjectCard({ project, currentUserId }) {
  const { hasPermission, isAdmin } = usePermissions();
  const canModify = canModifyProject(project, currentUserId, isAdmin());
  
  return (
    <div className={`project-card ${!canModify ? 'read-only' : ''}`}>
      <div className="project-header">
        <h3>{project.name}</h3>
        {!canModify && (
          <span className="badge badge-info" title="You can view but not modify this project">
            View Only
          </span>
        )}
      </div>
      
      {/* Show lock icon for restricted actions */}
      <div className="project-actions">
        <button onClick={() => viewProject(project.id)}>
          View
        </button>
        
        {hasPermission('projects:write') && (
          <button 
            onClick={() => editProject(project.id)}
            disabled={!canModify}
            className={!canModify ? 'disabled' : ''}
            title={!canModify ? "You can only edit projects you're a member of" : ""}
          >
            {!canModify && <span className="icon-lock">🔒</span>}
            Edit
          </button>
        )}
      </div>
    </div>
  );
}
```

#### 6. Error Handling for Resource Access

```javascript
// Handle 403 errors gracefully
async function handleProjectUpdate(projectId, projectData) {
  try {
    const response = await fetch('/api/project/update', {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ ...projectData, id: projectId })
    });

    if (response.status === 403) {
      const error = await response.json();
      // Show user-friendly error
      showNotification({
        type: 'error',
        message: error.error || 'You do not have permission to modify this project'
      });
      return;
    }

    if (!response.ok) {
      throw new Error('Failed to update project');
    }

    return await response.json();
  } catch (error) {
    console.error('Error:', error);
    showNotification({
      type: 'error',
      message: 'An error occurred while updating the project'
    });
  }
}
```

## Best Practices

### 1. Cache Permissions

```javascript
// Cache permissions in localStorage or sessionStorage
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

  const data = await fetchPermissions();
  localStorage.setItem(CACHE_KEY, JSON.stringify({
    data,
    timestamp: Date.now()
  }));
  
  return data;
}
```

### 2. Refresh on Login/Logout

```javascript
// Refresh permissions when user logs in
function handleLogin(token) {
  localStorage.setItem('authToken', token);
  permissionManager.fetchPermissions();
}

// Clear permissions on logout
function handleLogout() {
  localStorage.removeItem('authToken');
  localStorage.removeItem(CACHE_KEY);
  permissionManager.permissions = [];
  permissionManager.role = null;
}
```

### 3. Error Handling

```javascript
async function fetchPermissionsWithRetry(maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fetchPermissions();
    } catch (error) {
      if (i === maxRetries - 1) {
        // Use cached permissions as fallback
        const cached = getCachedPermissions();
        if (cached) return cached;
        throw error;
      }
      await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)));
    }
  }
}
```

### 4. Permission Constants

```javascript
// constants/permissions.js
export const PERMISSIONS = {
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
  },
  USERS: {
    READ: 'users:read',
    WRITE: 'users:write',
    DELETE: 'users:delete'
  },
  VACATION: {
    READ: 'vacation:read',
    WRITE: 'vacation:write',
    DELETE: 'vacation:delete',
    APPROVE: 'vacation:approve'
  }
};

// Usage
if (hasPermission(PERMISSIONS.PROJECTS.WRITE)) {
  // Create project
}
```

### 5. Resource Access Helpers

```javascript
// utils/resourceAccess.js

/**
 * Check if user can modify a project
 * Standard users can view all projects but only modify ones they're members of
 */
export const canModifyProject = (project, currentUserId, isAdmin) => {
  if (isAdmin) return true;
  if (!project) return false;
  if (project.ownerId === currentUserId) return true;
  if (project.membersIds?.includes(currentUserId)) return true;
  return false;
};

/**
 * Check if user can modify a task
 * Standard users can view all tasks but only modify assigned tasks or tasks in their projects
 */
export const canModifyTask = (task, project, currentUserId, isAdmin) => {
  if (isAdmin) return true;
  if (!task || !project) return false;
  
  // Check if assigned to task
  if (task.assignTo === currentUserId) return true;
  
  // Check if project owner or member
  if (project.ownerId === currentUserId) return true;
  if (project.membersIds?.includes(currentUserId)) return true;
  
  return false;
};

/**
 * Get action availability for a project
 */
export const getProjectActions = (project, permissions, currentUserId, isAdmin) => {
  const canModify = canModifyProject(project, currentUserId, isAdmin);
  
  return {
    canView: permissions.includes('projects:read'),
    canEdit: permissions.includes('projects:write') && canModify,
    canDelete: permissions.includes('projects:delete') && canModify,
    isReadOnly: permissions.includes('projects:read') && !canModify
  };
};

/**
 * Get action availability for a task
 */
export const getTaskActions = (task, project, permissions, currentUserId, isAdmin) => {
  const canModify = canModifyTask(task, project, currentUserId, isAdmin);
  
  return {
    canView: permissions.includes('tasks:read'),
    canEdit: permissions.includes('tasks:write') && canModify,
    canDelete: permissions.includes('tasks:delete') && canModify,
    canAssign: permissions.includes('tasks:assign') && canModifyProject(project, currentUserId, isAdmin),
    isReadOnly: permissions.includes('tasks:read') && !canModify
  };
};
```

## Troubleshooting

### Issue: Permissions not loading

**Solution:**
- Check if the JWT token is valid and not expired
- Verify the API endpoint URL is correct
- Check browser console for CORS errors
- Ensure the token is included in the Authorization header

### Issue: Permissions are cached incorrectly

**Solution:**
- Clear localStorage cache
- Refresh permissions after role changes
- Implement cache invalidation on logout

### Issue: UI shows wrong permissions

**Solution:**
- Verify the user's role in the database
- Check if permissions are being refreshed after role changes
- Ensure the API response matches expected format

### Issue: CORS errors

**Solution:**
- Ensure the backend CORS middleware allows your frontend origin
- Check if credentials are being sent correctly

## Example: Complete Integration

```javascript
// Complete example with error handling and caching
class PermissionService {
  constructor(apiBaseUrl) {
    this.apiBaseUrl = apiBaseUrl;
    this.permissions = [];
    this.role = null;
    this.cacheKey = 'user_permissions_cache';
    this.cacheDuration = 5 * 60 * 1000; // 5 minutes
  }

  async initialize() {
    try {
      // Try to load from cache first
      const cached = this.getCachedPermissions();
      if (cached) {
        this.permissions = cached.permissions;
        this.role = cached.role;
      }

      // Always fetch fresh permissions
      await this.fetchPermissions();
    } catch (error) {
      console.error('Failed to initialize permissions:', error);
      // Use cached if available
      if (this.permissions.length === 0) {
        const cached = this.getCachedPermissions();
        if (cached) {
          this.permissions = cached.permissions;
          this.role = cached.role;
        }
      }
    }
  }

  async fetchPermissions() {
    const token = localStorage.getItem('authToken');
    if (!token) {
      throw new Error('No authentication token found');
    }

    const response = await fetch(`${this.apiBaseUrl}/api/auth/permissions`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      if (response.status === 401) {
        // Token expired or invalid
        localStorage.removeItem('authToken');
        window.location.href = '/login';
        throw new Error('Authentication required');
      }
      throw new Error(`Failed to fetch permissions: ${response.statusText}`);
    }

    const data = await response.json();
    this.permissions = data.permissions;
    this.role = data.role;

    // Cache the permissions
    this.cachePermissions(data);

    return data;
  }

  getCachedPermissions() {
    try {
      const cached = localStorage.getItem(this.cacheKey);
      if (!cached) return null;

      const { data, timestamp } = JSON.parse(cached);
      if (Date.now() - timestamp > this.cacheDuration) {
        localStorage.removeItem(this.cacheKey);
        return null;
      }

      return data;
    } catch (error) {
      return null;
    }
  }

  cachePermissions(data) {
    localStorage.setItem(this.cacheKey, JSON.stringify({
      data,
      timestamp: Date.now()
    }));
  }

  hasPermission(permission) {
    return this.permissions.includes(permission);
  }

  isAdmin() {
    return this.role === 'Admin';
  }

  clearCache() {
    localStorage.removeItem(this.cacheKey);
    this.permissions = [];
    this.role = null;
  }
}

// Initialize on app load
const permissionService = new PermissionService('http://localhost:3000');
permissionService.initialize();
```

---

**Last Updated:** 2025-01-XX  
**API Version:** 1.0

