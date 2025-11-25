package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS middleware for allowing cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Allowed origins
		allowedOrigins := []string{
			"http://localhost:5173",
			"https://localhost:5173",
		}
		
		// Check if the origin is in the allowed list
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin != "" {
			// For other origins, allow but log (for development)
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// No origin header, allow all (for direct requests)
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

