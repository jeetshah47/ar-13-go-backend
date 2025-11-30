package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ar-13-go-backend/internal/test"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationHandler_GetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	
	// Create a mock websocket handler for notification handler
	websocketHandler := NewWebSocketHandler(nil, nil)
	notificationHandler := NewNotificationHandlerWithDefaults(websocketHandler)

	router.GET("/api/notifications/all/:userId", notificationHandler.GetAll)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "get all notifications with valid user ID",
			userID:         "user123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get all notifications with empty user ID",
			userID:         "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/notifications/all/" + tt.userID
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			if err == nil {
				assert.Contains(t, response, "notifications")
			}
		})
	}
}

func TestNotificationHandler_GetUnread(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	
	websocketHandler := NewWebSocketHandler(nil, nil)
	notificationHandler := NewNotificationHandlerWithDefaults(websocketHandler)

	router.GET("/api/notifications/unread/:userId", notificationHandler.GetUnread)

	t.Run("get unread notifications", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/notifications/unread/user123", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		if err == nil {
			assert.Contains(t, response, "notifications")
		}
	})
}

func TestNotificationHandler_GetCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	
	websocketHandler := NewWebSocketHandler(nil, nil)
	notificationHandler := NewNotificationHandlerWithDefaults(websocketHandler)

	router.GET("/api/notifications/count/:userId", notificationHandler.GetCount)

	t.Run("get notification count", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/notifications/count/user123", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		if err == nil {
			assert.Contains(t, response, "count")
		}
	})
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	
	websocketHandler := NewWebSocketHandler(nil, nil)
	notificationHandler := NewNotificationHandlerWithDefaults(websocketHandler)

	router.PUT("/api/notifications/read/:id", notificationHandler.MarkAsRead)

	tests := []struct {
		name           string
		notificationID string
		expectedStatus int
	}{
		{
			name:            "mark notification as read with valid ID",
			notificationID:  "notification123",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "mark notification as read with empty ID",
			notificationID:  "",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "mark non-existent notification as read",
			notificationID:  "nonexistent",
			expectedStatus:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/notifications/read/" + tt.notificationID
			req := test.CreateTestRequest("PUT", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestNotificationHandler_MarkAllAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	
	websocketHandler := NewWebSocketHandler(nil, nil)
	notificationHandler := NewNotificationHandlerWithDefaults(websocketHandler)

	router.PUT("/api/notifications/read-all/:userId", notificationHandler.MarkAllAsRead)

	t.Run("mark all notifications as read", func(t *testing.T) {
		req := test.CreateTestRequest("PUT", "/api/notifications/read-all/user123", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		if err == nil {
			assert.Contains(t, response, "message")
		}
	})
}

func TestNotificationHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	
	websocketHandler := NewWebSocketHandler(nil, nil)
	notificationHandler := NewNotificationHandlerWithDefaults(websocketHandler)

	router.DELETE("/api/notifications/:id", notificationHandler.Delete)

	tests := []struct {
		name           string
		notificationID string
		expectedStatus int
	}{
		{
			name:            "delete notification with valid ID",
			notificationID:  "notification123",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "delete notification with empty ID",
			notificationID:  "",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "delete non-existent notification",
			notificationID:  "nonexistent",
			expectedStatus:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/notifications/" + tt.notificationID
			req := test.CreateTestRequest("DELETE", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestNotificationHandler_GetConnectionInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	
	websocketHandler := NewWebSocketHandler(nil, nil)
	notificationHandler := NewNotificationHandlerWithDefaults(websocketHandler)

	router.GET("/api/notifications/connection-info", notificationHandler.GetConnectionInfo)

	t.Run("get connection info", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/notifications/connection-info", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "connectedUsers")
		assert.Contains(t, response, "connectedUserIds")
	})
}

