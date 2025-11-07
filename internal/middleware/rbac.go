package middleware

import (
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/gin-gonic/gin"
)

// RequireAdmin middleware checks if user is admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement admin check from user data
		// For now, this is a placeholder
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// TODO: Check user role from database
		// This should query the user's role and verify it's "Admin"
		c.Next()
	}
}

// RequirePermission middleware checks if user has required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// TODO: Implement permission checking
		// This should check user's role and permissions
		c.Next()
	}
}

// RequireProjectAccess middleware checks if user has access to project
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

		// TODO: Implement project access check
		// This should verify user has access to the project
		c.Next()
	}
}

// RequireTaskAccess middleware checks if user has access to task
func RequireTaskAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		_ = c.Param("projectId") // TODO: Use projectID when implementing task access check
		_ = c.Param("taskId")    // TODO: Use taskID when implementing task access check

		// TODO: Implement task access check
		// This should verify user has access to the task
		c.Next()
	}
}
