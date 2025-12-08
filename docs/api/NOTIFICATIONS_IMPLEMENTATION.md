# Notifications Implementation for Create/Modify Operations

## Overview

This document describes the notification system implementation for create and modify operations across all modules in the application.

## Implementation Status

### ✅ Completed

#### 1. **Project Module**
- **Create**: Notifies project owner when a project is created
- **Update**: Notifies project owner and all project members when a project is updated
- **Notification Types**: `PROJECT_CREATED`, `PROJECT_UPDATED`
- **Service**: `ProjectService`
- **Files Modified**:
  - `internal/services/project_service.go`
  - `internal/handlers/handler.go` (service initialization)

#### 2. **Task Module**
- **Create**: Notifies project owner and assigned user (if different from owner) when a task is created
- **Update**: Notifies assigned user when a task is updated (already implemented)
- **Notification Types**: `TASK_CREATED`, `TASK_UPDATED`
- **Service**: `TaskService`
- **Files Modified**:
  - `internal/services/task_service.go`

#### 3. **Calendar Event Module**
- **Create**: Notifies event creator when a calendar event is created
- **Update**: Notifies event creator when a calendar event is updated
- **Notification Types**: `CALENDAR_EVENT_CREATED`, `CALENDAR_EVENT_UPDATED`
- **Service**: `CalendarEventService`
- **Files Modified**:
  - `internal/services/calendar_service.go`
  - `internal/handlers/calendar.go` (getter method)
  - `internal/handlers/handler.go` (service initialization)

#### 4. **User Module**
- **Create**: Notifies newly created user when their account is created
- **Update**: Notifies user when their profile is updated
- **Notification Types**: `USER_CREATED`, `USER_UPDATED`
- **Service**: `UserService`
- **Files Modified**:
  - `internal/services/user_service.go`

### 📋 Pending Implementation

#### 5. **Vacation/Leave Request Module**
- **Create**: Should notify admins when a leave request is created
- **Update**: Should notify requester when their leave request is updated
- **Notification Types**: `LEAVE_REQUEST_CREATED`, `LEAVE_REQUEST_UPDATED` (already defined)
- **Service**: `VacationService`
- **Files to Modify**:
  - `internal/services/vacation_service.go`
  - Add notification service fields and setters
  - Add notification creation in `Create()` and `Update()` methods

#### 6. **Project Details Module**
- **Create**: Should notify project owner and members when project details are created
- **Update**: Should notify project owner and members when project details are updated
- **Notification Types**: `PROJECT_DETAILS_CREATED`, `PROJECT_DETAILS_UPDATED` (already defined)
- **Service**: `ProjectDetailsService`
- **Files to Modify**:
  - `internal/services/project_details_service.go`
  - Add notification service fields and setters
  - Add notification creation in `Add()` and `Update()` methods

#### 7. **Info Portal Module**
- **Create**: Should notify users with `infoPortal:read` permission when folders/pages are created
- **Update**: Should notify users with `infoPortal:read` permission when folders/pages are updated
- **Notification Types**: `INFO_PORTAL_FOLDER_CREATED`, `INFO_PORTAL_FOLDER_UPDATED`, `INFO_PORTAL_PAGE_CREATED`, `INFO_PORTAL_PAGE_UPDATED` (already defined)
- **Service**: `InfoPortalService`
- **Files to Modify**:
  - `internal/services/info_portal_service.go`
  - Add notification service fields and setters
  - Add notification creation in folder/page create/update methods

#### 8. **Drawing List Module**
- **Create**: Should notify users with `drawingList:read` permission when categories/types are created
- **Update**: Should notify users with `drawingList:read` permission when categories/types are updated
- **Notification Types**: `DRAWING_LIST_CATEGORY_CREATED`, `DRAWING_LIST_CATEGORY_UPDATED`, `DRAWING_LIST_TYPE_CREATED`, `DRAWING_LIST_TYPE_UPDATED` (already defined)
- **Service**: `DrawingListService`
- **Files to Modify**:
  - `internal/services/drawing_list_service.go`
  - Add notification service fields and setters
  - Add notification creation in category/type create/update methods

## Notification Model

All notification types and related entity types have been added to `internal/models/notification.go`:

### Notification Types
- `PROJECT_CREATED`, `PROJECT_UPDATED`
- `TASK_CREATED`, `TASK_UPDATED`
- `CALENDAR_EVENT_CREATED`, `CALENDAR_EVENT_UPDATED`
- `USER_CREATED`, `USER_UPDATED`
- `LEAVE_REQUEST_CREATED`, `LEAVE_REQUEST_UPDATED`
- `PROJECT_DETAILS_CREATED`, `PROJECT_DETAILS_UPDATED`
- `INFO_PORTAL_FOLDER_CREATED`, `INFO_PORTAL_FOLDER_UPDATED`
- `INFO_PORTAL_PAGE_CREATED`, `INFO_PORTAL_PAGE_UPDATED`
- `DRAWING_LIST_CATEGORY_CREATED`, `DRAWING_LIST_CATEGORY_UPDATED`
- `DRAWING_LIST_TYPE_CREATED`, `DRAWING_LIST_TYPE_UPDATED`

### Related Entity Types
- `PROJECT`, `TASK`, `USER`, `LEAVE_REQUEST`
- `CALENDAR_EVENT`, `PROJECT_DETAILS`
- `INFO_PORTAL_FOLDER`, `INFO_PORTAL_PAGE`
- `DRAWING_LIST_CATEGORY`, `DRAWING_LIST_TYPE`

## Implementation Pattern

For each service that needs notifications:

1. **Add fields to service struct**:
   ```go
   notificationSvc  *NotificationService
   websocketService WebSocketServiceInterface
   ```

2. **Add setter methods**:
   ```go
   func (s *Service) SetNotificationService(notificationSvc *NotificationService) {
       s.notificationSvc = notificationSvc
   }
   
   func (s *Service) SetWebSocketService(websocketService WebSocketServiceInterface) {
       s.websocketService = websocketService
   }
   ```

3. **Add notification creation in Create/Update methods** (non-blocking):
   ```go
   if s.notificationSvc != nil {
       go func() {
           notification := &models.Notification{
               Title:             "Title",
               Message:           "Message",
               Type:              models.NotificationTypeXXX,
               UserID:            userID,
               RelatedEntityID:   entityID,
               RelatedEntityType: models.RelatedEntityTypeXXX,
               IsRead:            false,
           }
           if err := s.notificationSvc.CreateNotification(context.Background(), notification); err != nil {
               log.Printf("Failed to create notification: %v", err)
           }
           if s.websocketService != nil {
               wsData := map[string]interface{}{"userId": userID}
               _ = s.websocketService.SendToUser(userID, "notifications-available", wsData)
           }
       }()
   }
   ```

4. **Set services in handler initialization** (`internal/handlers/handler.go`):
   ```go
   service.SetNotificationService(notificationService)
   service.SetWebSocketService(websocketHandler.GetWebSocketService())
   ```

## Notification Recipients

- **Project Create/Update**: Project owner and all project members
- **Task Create**: Project owner and assigned user (if different)
- **Task Update**: Assigned user
- **Calendar Event Create/Update**: Event creator
- **User Create/Update**: The user themselves
- **Leave Request Create**: Admins (to be implemented)
- **Leave Request Update**: Requester (to be implemented)
- **Project Details Create/Update**: Project owner and members (to be implemented)
- **Info Portal Create/Update**: All users with `infoPortal:read` permission (to be implemented)
- **Drawing List Create/Update**: All users with `drawingList:read` permission (to be implemented)

## WebSocket Integration

All notifications are sent via WebSocket using the `notifications-available` event type. The client receives this event and fetches notifications via the API.

## Notes

- All notification creation is done asynchronously (in goroutines) to avoid blocking the main operation
- Notification failures are logged but don't fail the main operation
- WebSocket notifications are sent after database notification creation
- The notification service handles cache invalidation automatically

