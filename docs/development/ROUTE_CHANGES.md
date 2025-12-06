# Route Changes During Migration (Node.js → Go)

This document lists all route changes that occurred during the migration from Node.js/TypeScript to Go.

## 🔄 Major Route Changes

### 1. Activity Logs
**Old:** `/api/activity-logs/` (plural)  
**New:** `/api/activity-log/` (singular)

| Old Route | New Route | Notes |
|-----------|-----------|-------|
| `GET /api/activity-logs/task?limit=10` | `GET /api/activity-log/entity-type/task?limit=10` | Changed to use `entity-type` parameter |
| `GET /api/activity-logs/:entityType/:entityId` | `GET /api/activity-log/entity/:entityType/:entityId` | Added `entity` prefix |
| `GET /api/activity-logs/entity-types` | `GET /api/activity-log/entity-types` | Same functionality |

### 2. Calendar Routes
**Old:** `/api/calendar/:year/:month`  
**New:** `/api/calendar/month/:year/:month`

| Old Route | New Route | Notes |
|-----------|-----------|-------|
| `GET /api/calendar/:year/:month` | `GET /api/calendar/month/:year/:month` | Added `month` prefix |
| - | `GET /api/calendar/event/:id` | New route for getting single event |
| `POST /api/calendar/add` | `POST /api/calendar/add` | ✅ Same |
| `PUT /api/calendar/update/:id` | `PUT /api/calendar/update/:id` | ✅ Same |
| `DELETE /api/calendar/delete/:id` | `DELETE /api/calendar/delete/:id` | ✅ Same |

### 3. Vacation/Leave Request Routes
**Old:** Used `/request` and `/requests` suffixes  
**New:** Simplified paths

| Old Route | New Route | Notes |
|-----------|-----------|-------|
| `GET /api/vacation/my-requests` | `GET /api/vacation/my-requests` | ✅ Same |
| `GET /api/vacation/all-requests` | `GET /api/vacation/all` | Removed `-requests` suffix |
| `GET /api/vacation/pending-requests` | `GET /api/vacation/pending` | Removed `-requests` suffix |
| `GET /api/vacation/request/:requestId` | `GET /api/vacation/:requestId` | Removed `request/` prefix |
| `POST /api/vacation/create-request` | `POST /api/vacation/create` | Removed `-request` suffix |
| `PUT /api/vacation/update-status/:requestId` | `PUT /api/vacation/update-status/:requestId` | ✅ Same |
| `PUT /api/vacation/update-request` | `PUT /api/vacation/update` | Removed `-request` suffix |
| `DELETE /api/vacation/delete-request/:requestId` | `DELETE /api/vacation/delete/:requestId` | Removed `request/` prefix |
| `GET /api/vacation/summaries` | `GET /api/vacation/summaries` | ✅ Same |
| `GET /api/vacation/by-status/:status` | `GET /api/vacation/status/:status` | Simplified path |
| `GET /api/vacation/by-type/:type` | `GET /api/vacation/type/:type` | Simplified path |

### 4. Employee Routes
**Old:** `/api/employees/` (plural)  
**New:** `/api/employee/` (singular)

| Old Route | New Route | Notes |
|-----------|-----------|-------|
| `GET /api/employees/list` | `GET /api/employee/list` | Changed to singular |
| - | `GET /api/employee/task-counts/:userId` | New route added |

### 5. Project Details Routes
**Old:** Used `/add` and `/update` suffixes  
**New:** Simplified to use HTTP methods

| Old Route | New Route | Notes |
|-----------|-----------|-------|
| `GET /api/project-details/:projectId` | `GET /api/project-details/:projectId` | ✅ Same |
| `POST /api/project-details/:projectId/add` | `POST /api/project-details/:projectId` | Removed `/add` suffix |
| `PUT /api/project-details/:projectId/update` | `PUT /api/project-details/:projectId` | Removed `/update` suffix |
| `DELETE /api/project-details/:projectId/delete/:projectDetailsId` | `DELETE /api/project-details/:projectId/:projectDetailsId` | Simplified path |

### 6. Google Account Routes
**Old:** Used `/links` for getting all links  
**New:** Changed to `/all`

| Old Route | New Route | Notes |
|-----------|-----------|-------|
| `POST /api/google-account/link` | `POST /api/google-account/link` | ✅ Same |
| `POST /api/google-account/unlink` | `POST /api/google-account/unlink` | ✅ Same |
| `GET /api/google-account/status` | `GET /api/google-account/status` | ✅ Same |
| `GET /api/google-account/links` | `GET /api/google-account/all` | Changed to `/all` |
| `GET /api/google-account/auth/initiate` | `GET /api/google-account/auth/initiate` | ✅ Same |
| `GET /api/google-account/auth/callback` | `GET /api/google-account/auth/callback` | ✅ Same |

## ✅ Routes That Stayed the Same

### Auth Routes
- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/auth/validate-signup`

### User Routes
- `GET /api/users/all`
- `POST /api/users/invite`
- `PUT /api/users/update`
- `DELETE /api/users/delete/:id`
- `GET /api/users/profile/:id`

### Project Routes
- `GET /api/project/all`
- `GET /api/project/:id`
- `POST /api/project/add`
- `PUT /api/project/update`
- `DELETE /api/project/delete/:id`

### Task Routes
- All task routes remain the same (20+ endpoints)
- `GET /api/tasks/all/:projectId`
- `GET /api/tasks/all/details/:projectId`
- `GET /api/tasks/detail/:projectId/:taskId`
- `POST /api/tasks/add`
- `POST /api/tasks/add-multiple`
- `PUT /api/tasks/update`
- `PUT /api/tasks/update-duration/:projectId/:taskId`
- `PUT /api/tasks/update-description/:projectId/:taskId`
- `PUT /api/tasks/update-status/:projectId/:taskId`
- `POST /api/tasks/add-time-spent/:projectId/:taskId`
- `PUT /api/tasks/update-time-spent/:projectId/:taskId/:timeSpentIndex`
- `DELETE /api/tasks/remove-time-spent/:projectId/:taskId/:timeSpentIndex`
- `GET /api/tasks/time-spent/:projectId/:taskId`
- `POST /api/tasks/add-file-attachment/:projectId/:taskId`
- `DELETE /api/tasks/remove-file-attachment/:projectId/:taskId/:fileAttachmentIndex`
- `GET /api/tasks/file-attachments/:projectId/:taskId`
- `GET /api/tasks/activity-logs/:projectId/:taskId`
- `DELETE /api/tasks/delete/:projectId/:taskId`
- `PUT /api/tasks/assign/:taskId/:userId`
- `PUT /api/tasks/claim/:projectId/:taskId`
- `GET /api/tasks/assignable/:projectId`

### Dashboard Routes
- `GET /api/dashboard/stats`

### Notification Routes
- `GET /api/notifications/all/:userId`
- `GET /api/notifications/unread/:userId`
- `GET /api/notifications/count/:userId`
- `PUT /api/notifications/read/:id`
- `PUT /api/notifications/read-all/:userId`
- `DELETE /api/notifications/:id`
- `DELETE /api/notifications/user/:userId`
- `GET /api/notifications/connection-info`

### Info Portal Routes
- All info portal routes remain the same (13 endpoints)

## 🆕 New Routes Added

1. **Calendar Event by ID**
   - `GET /api/calendar/event/:id` - Get a single calendar event

2. **Employee Task Counts**
   - `GET /api/employee/task-counts/:userId` - Get task counts for an employee

3. **Activity Log Entity Types**
   - `GET /api/activity-log/entity-types` - Get all supported entity types

## 📝 Summary

- **Total Routes:** 88 endpoints
- **Changed Routes:** ~15 routes
- **New Routes:** 3 routes
- **Unchanged Routes:** ~70 routes

## 🔍 Quick Reference

| Category | Old Base Path | New Base Path |
|----------|---------------|---------------|
| Activity Logs | `/activity-logs` | `/activity-log` |
| Employees | `/employees` | `/employee` |
| Calendar | `/calendar/:year/:month` | `/calendar/month/:year/:month` |
| Vacation | Various with `/request` suffix | Simplified paths |
| Project Details | With `/add`, `/update` suffixes | HTTP method-based |
| Google Account | `/links` | `/all` |

