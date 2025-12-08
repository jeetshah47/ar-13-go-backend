package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/gin-gonic/gin"
)

// AuditLogMiddleware creates middleware to log all API requests
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for:
		// - Health check endpoints
		// - Metrics endpoints
		// - WebSocket connections
		// - Static file serving
		path := c.Request.URL.Path
		if path == "/api/health" || 
		   strings.HasPrefix(path, "/api/metrics") ||
		   path == "/ws" ||
		   strings.HasPrefix(path, "/uploads") {
			c.Next()
			return
		}

		// Only log API routes (paths starting with /api)
		if !strings.HasPrefix(path, "/api") {
			c.Next()
			return
		}

		startTime := time.Now()

		// Capture request body for POST, PUT, PATCH requests
		var requestBody map[string]interface{}
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// Restore the request body for the handler
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Try to parse as JSON
				var body map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &body); err == nil {
					// Sanitize sensitive fields
					requestBody = sanitizeRequestBody(body)
				}
			}
		}

		// Capture query parameters
		queryParams := make(map[string]interface{})
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				if len(values) == 1 {
					queryParams[key] = values[0]
				} else {
					queryParams[key] = values
				}
			}
		}

		// Capture response
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:          &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime).Milliseconds()

		// Get user info if authenticated
		var userID *string
		var userEmail *string
		if uid, exists := c.Get("userId"); exists {
			if uidStr, ok := uid.(string); ok && uidStr != "" {
				userID = &uidStr
			}
		}
		if email, exists := c.Get("userEmail"); exists {
			if emailStr, ok := email.(string); ok && emailStr != "" {
				userEmail = &emailStr
			}
		}

		// Get IP address
		ipAddress := c.ClientIP()
		if forwardedIP := c.GetHeader("X-Forwarded-For"); forwardedIP != "" {
			ipAddress = forwardedIP
		} else if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
			ipAddress = realIP
		}

		// Get error if any
		var errorMsg *string
		if len(c.Errors) > 0 {
			errStr := c.Errors.String()
			errorMsg = &errStr
		}

		// Create audit log entry
		auditLog := &models.AuditLog{
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			UserID:      userID,
			UserEmail:   userEmail,
			IPAddress:   ipAddress,
			UserAgent:   c.Request.UserAgent(),
			StatusCode:  c.Writer.Status(),
			RequestTime: startTime,
			Duration:    duration,
			RequestBody: requestBody,
			QueryParams: queryParams,
			Error:       errorMsg,
			ResponseSize: int64(responseWriter.body.Len()),
		}

		// Save audit log asynchronously to avoid blocking the response
		go func() {
			repo := repos.NewAuditLogRepo()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := repo.Add(ctx, auditLog); err != nil {
				// Log error but don't fail the request
				// In production, you might want to use a proper logger here
				_ = err
			}
		}()
	}
}

// sanitizeRequestBody removes sensitive information from request body
func sanitizeRequestBody(body map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	sensitiveFields := []string{"password", "token", "secret", "accessToken", "refreshToken", "authorization"}

	for key, value := range body {
		isSensitive := false
		lowerKey := strings.ToLower(key)
		for _, sensitive := range sensitiveFields {
			if strings.Contains(lowerKey, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[key] = "***REDACTED***"
		} else {
			// Recursively sanitize nested maps
			if nestedMap, ok := value.(map[string]interface{}); ok {
				sanitized[key] = sanitizeRequestBody(nestedMap)
			} else {
				sanitized[key] = value
			}
		}
	}

	return sanitized
}

// responseBodyWriter is a custom ResponseWriter that captures the response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

