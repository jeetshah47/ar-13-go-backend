package constants

import "github.com/ar-13-go-backend/internal/models"

// Permission constants - mapped from Node.js backend
const (
	// Projects permissions
	PermissionProjectsRead   = "projects:read"
	PermissionProjectsWrite  = "projects:write"
	PermissionProjectsDelete = "projects:delete"

	// Tasks permissions
	PermissionTasksRead   = "tasks:read"
	PermissionTasksWrite  = "tasks:write"
	PermissionTasksDelete = "tasks:delete"
	PermissionTasksAssign = "tasks:assign"

	// User management permissions
	PermissionUsersRead   = "users:read"
	PermissionUsersWrite  = "users:write"
	PermissionUsersDelete = "users:delete"

	// Calendar permissions
	PermissionCalendarRead   = "calendar:read"
	PermissionCalendarWrite  = "calendar:write"
	PermissionCalendarDelete = "calendar:delete"

	// Vacation permissions
	PermissionVacationRead    = "vacation:read"
	PermissionVacationWrite   = "vacation:write"
	PermissionVacationDelete  = "vacation:delete"
	PermissionVacationApprove = "vacation:approve"

	// Notifications permissions
	PermissionNotificationsRead   = "notifications:read"
	PermissionNotificationsWrite  = "notifications:write"
	PermissionNotificationsDelete = "notifications:delete"

	// Activity Log permissions (note: plural as in Node.js backend)
	PermissionActivityLogsRead = "activityLogs:read"

	// Dashboard permissions
	PermissionDashboardRead = "dashboard:read"

	// Employees permissions (Go backend specific, not in Node.js RBAC)
	PermissionEmployeesRead = "employees:read"

	// Info Portal permissions (Go backend specific, not in Node.js RBAC)
	PermissionInfoPortalRead   = "infoPortal:read"
	PermissionInfoPortalWrite  = "infoPortal:write"
	PermissionInfoPortalDelete = "infoPortal:delete"

	// Google Account permissions
	PermissionGoogleAccountRead   = "googleAccount:read"
	PermissionGoogleAccountWrite  = "googleAccount:write"
	PermissionGoogleAccountLink   = "googleAccount:link"
	PermissionGoogleAccountUnlink = "googleAccount:unlink"

	// WebSocket permissions
	PermissionWebSocketConnect = "websocket:connect"

	// Auth permissions
	PermissionAuthRead = "auth:read"

	// User profile permissions
	PermissionUsersProfile = "users:profile"
	PermissionUsersInvite  = "users:invite"

	// Project Details permissions (uses projects permissions, but defined for clarity)
	PermissionProjectDetailsRead   = "projectDetails:read"
	PermissionProjectDetailsWrite  = "projectDetails:write"
	PermissionProjectDetailsDelete = "projectDetails:delete"

	// Backup permissions (Admin only)
	PermissionBackupRead  = "backup:read"
	PermissionBackupWrite = "backup:write"
)

// GetPermissionsByRole returns the list of permissions for a given role
func GetPermissionsByRole(role models.UserRole) []string {
	switch role {
	case models.UserRoleAdmin:
		return GetAdminPermissions()
	case models.UserRoleStandard:
		return GetStandardPermissions()
	default:
		return []string{}
	}
}

// GetAdminPermissions returns all permissions for admin users
// Mapped from Node.js backend RoleBasedMiddleware.ts
func GetAdminPermissions() []string {
	return []string{
		// Projects
		PermissionProjectsRead,
		PermissionProjectsWrite,
		PermissionProjectsDelete,
		// Tasks
		PermissionTasksRead,
		PermissionTasksWrite,
		PermissionTasksDelete,
		PermissionTasksAssign,
		// User management
		PermissionUsersRead,
		PermissionUsersWrite,
		PermissionUsersDelete,
		PermissionUsersProfile,
		PermissionUsersInvite,
		// Calendar
		PermissionCalendarRead,
		PermissionCalendarWrite,
		PermissionCalendarDelete,
		// Vacation
		PermissionVacationRead,
		PermissionVacationWrite,
		PermissionVacationDelete,
		PermissionVacationApprove,
		// Notifications
		PermissionNotificationsRead,
		PermissionNotificationsWrite,
		PermissionNotificationsDelete,
		// Activity Logs (plural as in Node.js backend)
		PermissionActivityLogsRead,
		// Dashboard
		PermissionDashboardRead,
		// Employees
		PermissionEmployeesRead,
		// Info Portal
		PermissionInfoPortalRead,
		PermissionInfoPortalWrite,
		PermissionInfoPortalDelete,
		// Google Account
		PermissionGoogleAccountRead,
		PermissionGoogleAccountWrite,
		PermissionGoogleAccountLink,
		PermissionGoogleAccountUnlink,
		// WebSocket
		PermissionWebSocketConnect,
		// Auth
		PermissionAuthRead,
		// Project Details
		PermissionProjectDetailsRead,
		PermissionProjectDetailsWrite,
		PermissionProjectDetailsDelete,
		// Backup (Admin only)
		PermissionBackupRead,
		PermissionBackupWrite,
	}
}

// GetStandardPermissions returns permissions for standard users
// Mapped from Node.js backend RoleBasedMiddleware.ts
func GetStandardPermissions() []string {
	return []string{
		// Projects (read only, write/delete based on project membership)
		PermissionProjectsRead,
		// Tasks (read, write, delete - access controlled by task assignment)
		PermissionTasksRead,
		PermissionTasksWrite,
		PermissionTasksDelete,
		// Calendar
		PermissionCalendarRead,
		PermissionCalendarWrite,
		PermissionCalendarDelete,
		// Vacation (read, write, delete - but not approve)
		PermissionVacationRead,
		PermissionVacationWrite,
		PermissionVacationDelete,
		// Notifications
		PermissionNotificationsRead,
		PermissionNotificationsWrite,
		PermissionNotificationsDelete,
		// Activity Logs (plural as in Node.js backend)
		PermissionActivityLogsRead,
		// Dashboard
		PermissionDashboardRead,
		// Employees
		PermissionEmployeesRead,
		// Info Portal
		PermissionInfoPortalRead,
		PermissionInfoPortalWrite,
		PermissionInfoPortalDelete,
		// Google Account (standard users can link/unlink their own accounts)
		PermissionGoogleAccountRead,
		PermissionGoogleAccountWrite,
		PermissionGoogleAccountLink,
		PermissionGoogleAccountUnlink,
		// WebSocket
		PermissionWebSocketConnect,
		// Auth (users can read their own permissions)
		PermissionAuthRead,
		// User profile (users can view their own profile)
		PermissionUsersProfile,
		// Project Details (read only, write/delete based on project membership)
		PermissionProjectDetailsRead,
	}
}

// HasPermission checks if a role has a specific permission
func HasPermission(role models.UserRole, permission string) bool {
	permissions := GetPermissionsByRole(role)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}
