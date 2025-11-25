package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	// SecretKeyHeader is the header name for the secret key
	SecretKeyHeader = "X-API-Key"
	// SecretKeyEnvVar is the environment variable name for the secret key
	SecretKeyEnvVar = "SECRET_KEY"
)

// ValidateSecretKey middleware validates the secret key from header
func ValidateSecretKey() gin.HandlerFunc {
	expectedSecret := os.Getenv(SecretKeyEnvVar)
	if expectedSecret == "" {
		// If no secret is configured, allow all requests (for development)
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Get secret key from header
		providedSecret := c.GetHeader(SecretKeyHeader)
		if providedSecret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Secret key required"})
			c.Abort()
			return
		}

		// Compare secrets using constant-time comparison to prevent timing attacks
		if !secureCompare(providedSecret, expectedSecret) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid secret key"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// secureCompare performs a constant-time comparison of two strings
// This helps prevent timing attacks
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i]) ^ int(b[i])
	}

	return result == 0
}

