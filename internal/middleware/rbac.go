package middleware

import (
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/gin-gonic/gin"
)

// RequireAdmin middleware checks if user is admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check user role from context (set by AuthenticateUser middleware)
		role := GetUserRole(c)
		if role != string(models.UserRoleAdmin) {
			c.JSON(constants.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware checks if user has required permission
func RequirePermission(permission string) gin.HandlerFunc {
	rolePermissionRepo := repos.NewRolePermissionRepo()
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Get user role from context
		roleStr := GetUserRole(c)
		if roleStr == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// Check if user has the required permission
		role := models.UserRole(roleStr)

		// Try database first, fallback to constants if database is empty
		hasPermission, err := rolePermissionRepo.HasPermission(c.Request.Context(), role, permission)
		if err == nil && hasPermission {
			c.Next()
			return
		}

		// Fallback to constants (for backward compatibility and initial setup)
		if constants.HasPermission(role, permission) {
			c.Next()
			return
		}

		c.JSON(constants.StatusForbidden, gin.H{
			"error": "Permission denied: " + permission,
		})
		c.Abort()
	}
}

// RequireProjectAccess middleware checks if user has access to project
// Admin users can access all projects
// Standard users can view all projects (GET requests)
// Standard users can only modify/delete projects where they are owner or member (PUT/DELETE requests)
func RequireProjectAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		projectID := c.Param("id")
		if projectID == "" {
			projectID = c.Param("projectId")
		}

		if projectID == "" {
			c.JSON(constants.StatusBadRequest, gin.H{"error": "Project ID is required"})
			c.Abort()
			return
		}

		// Check if user is admin - admins have full access
		role := GetUserRole(c)
		if role == string(models.UserRoleAdmin) {
			c.Next()
			return
		}

		// For GET requests, Standard users can view all projects
		if c.Request.Method == "GET" {
			// Verify project exists
			projectRepo := repos.NewProjectRepo()
			project, err := projectRepo.GetByID(c.Request.Context(), projectID)
			if err != nil {
				c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to check project access"})
				c.Abort()
				return
			}
			if project == nil {
				c.JSON(constants.StatusNotFound, gin.H{"error": constants.MsgProjectNotFound})
				c.Abort()
				return
			}
			// Allow access for viewing
			c.Next()
			return
		}

		// For PUT/DELETE requests, Standard users must be project owner or member
		projectRepo := repos.NewProjectRepo()
		project, err := projectRepo.GetByID(c.Request.Context(), projectID)
		if err != nil {
			c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to check project access"})
			c.Abort()
			return
		}
		if project == nil {
			c.JSON(constants.StatusNotFound, gin.H{"error": constants.MsgProjectNotFound})
			c.Abort()
			return
		}

		// Check if user is the owner
		if project.OwnerID == userID {
			c.Next()
			return
		}

		// Check if user is a member
		for _, memberID := range project.MembersIDs {
			if memberID == userID {
				c.Next()
				return
			}
		}

		c.JSON(constants.StatusForbidden, gin.H{
			"error": "Access denied: You must be a project owner or member to modify this project",
		})
		c.Abort()
	}
}

// RequireTaskAccess middleware checks if user has access to task
// Admin users can access all tasks
// Standard users can view all tasks (GET requests)
// Standard users can only operate on tasks if:
//   - They are assigned to the task (can perform all operations)
//   - They are project owner or member (can modify tasks in their projects)
func RequireTaskAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		projectID := c.Param("projectId")
		taskID := c.Param("taskId")

		if projectID == "" {
			c.JSON(constants.StatusBadRequest, gin.H{"error": "Project ID is required"})
			c.Abort()
			return
		}

		if taskID == "" {
			c.JSON(constants.StatusBadRequest, gin.H{"error": "Task ID is required"})
			c.Abort()
			return
		}

		// Check if user is admin - admins have full access
		role := GetUserRole(c)
		if role == string(models.UserRoleAdmin) {
			c.Next()
			return
		}

		// Verify task exists
		taskRepo := repos.NewTaskRepo()
		task, err := taskRepo.GetByID(c.Request.Context(), projectID, taskID)
		if err != nil {
			c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to check task access"})
			c.Abort()
			return
		}
		if task == nil {
			c.JSON(constants.StatusNotFound, gin.H{"error": constants.MsgTaskNotFound})
			c.Abort()
			return
		}

		// For GET requests, Standard users can view all tasks
		if c.Request.Method == "GET" {
			// Allow access for viewing
			c.Next()
			return
		}

		// For PUT/DELETE/POST requests, Standard users must be assigned or project member
		// Check if user is assigned to the task
		if task.AssignTo != nil && *task.AssignTo == userID {
			c.Next()
			return
		}

		// Check if user is project owner or member
		projectRepo := repos.NewProjectRepo()
		project, err := projectRepo.GetByID(c.Request.Context(), projectID)
		if err != nil {
			c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to check project access"})
			c.Abort()
			return
		}
		if project == nil {
			c.JSON(constants.StatusNotFound, gin.H{"error": constants.MsgProjectNotFound})
			c.Abort()
			return
		}

		// Check if user is the owner
		if project.OwnerID == userID {
			c.Next()
			return
		}

		// Check if user is a member
		for _, memberID := range project.MembersIDs {
			if memberID == userID {
				c.Next()
				return
			}
		}

		c.JSON(constants.StatusForbidden, gin.H{
			"error": "Access denied: You must be assigned to this task or be a project member to modify it",
		})
		c.Abort()
	}
}
