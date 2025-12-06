package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/ar-13-go-backend/services/filebrowser/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// ValidateJWT middleware validates JWT token from Authorization header
func ValidateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log incoming request
		log.Printf("[JWT Middleware] Request: %s %s from %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("[JWT Middleware] ERROR: Authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		log.Printf("[JWT Middleware] Authorization header present (length: %d)", len(authHeader))

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("[JWT Middleware] ERROR: Invalid authorization header format. Parts count: %d, First part: '%s'", len(parts), parts[0])
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		
		// Debug logging (only log token length and first/last chars for security)
		if len(tokenString) > 0 {
			log.Printf("[JWT Middleware] Token extracted - Length: %d, First 15 chars: %s...", len(tokenString), tokenString[:min(15, len(tokenString))])
		} else {
			log.Printf("[JWT Middleware] ERROR: Token string is empty after extraction")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is empty"})
			c.Abort()
			return
		}

		// Verify token
		log.Printf("[JWT Middleware] Verifying token...")
		claims, err := jwt.VerifyToken(tokenString)
		if err != nil {
			log.Printf("[JWT Middleware] ERROR: Token verification failed - %v", err)
			log.Printf("[JWT Middleware] Error type: %T", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token: " + err.Error()})
			c.Abort()
			return
		}

		// Log successful authentication
		log.Printf("[JWT Middleware] SUCCESS: Token verified - UserID: %s, Email: %s, Role: %s", claims.UserID, claims.Email, claims.Role)

		// Store user info in context
		c.Set("userId", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

