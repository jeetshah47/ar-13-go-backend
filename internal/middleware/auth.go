package middleware

import (
	"context"
	"strings"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// UserKey is the key for storing user info in context
type UserKey string

const UserIDKey UserKey = "userId"
const UserEmailKey UserKey = "userEmail"

// AuthenticateUser middleware validates JWT token
func AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "Token required"})
			c.Abort()
			return
		}

		// Verify JWT token
		claims, err := jwt.VerifyToken(token)
		if err != nil {
			c.JSON(constants.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store user info in context
		userID := claims.UserID
		email := claims.Email

		ctx := context.WithValue(c.Request.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserEmailKey, email)
		c.Request = c.Request.WithContext(ctx)

		c.Set("userId", userID)
		c.Set("userEmail", email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// ValidateToken is a simpler version for routes that just need token validation
func ValidateToken() gin.HandlerFunc {
	return AuthenticateUser()
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("userId"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUserEmail extracts user email from context
func GetUserEmail(c *gin.Context) string {
	if email, exists := c.Get("userEmail"); exists {
		if e, ok := email.(string); ok {
			return e
		}
	}
	return ""
}
