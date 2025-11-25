package middleware

import (
	"net/http"
	"strings"

	"github.com/ar-13-go-backend/services/filebrowser/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// ValidateJWT middleware validates JWT token from Authorization header
func ValidateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Verify token
		claims, err := jwt.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token: " + err.Error()})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("userId", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

