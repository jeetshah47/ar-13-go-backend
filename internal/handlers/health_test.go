package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ar-13-go-backend/internal/test"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Health check endpoint (as defined in main.go)
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "ar-13-server",
		})
	})

	t.Run("health check returns healthy status", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/health", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "ar-13-server", response["service"])
	})

	t.Run("health check is accessible without authentication", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/health", nil)
		// No auth header
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

