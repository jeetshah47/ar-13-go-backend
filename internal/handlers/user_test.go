package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/test"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_GetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	userHandler := NewUserHandlerWithDefaults(cfg)

	router.GET("/api/users/all", userHandler.GetAll)

	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "get all users",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("GET", "/api/users/all", nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			if err == nil {
				assert.Contains(t, response, "users")
			}
		})
	}
}

func TestUserHandler_CreateInvitation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	userHandler := NewUserHandlerWithDefaults(cfg)

	router.POST("/api/users/invite", userHandler.CreateInvitation)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid invitation request",
			requestBody: map[string]interface{}{
				"email": "newuser@example.com",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "missing email",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email": "invalid-email",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("POST", "/api/users/invite", tt.requestBody)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			}
		})
	}
}

func TestUserHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	userHandler := NewUserHandlerWithDefaults(cfg)

	router.PUT("/api/users/update", userHandler.Update)

	tests := []struct {
		name           string
		requestBody    models.User
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid user update",
			requestBody: models.User{
				ID:    "user123",
				Name:  "Updated Name",
				Email: "updated@example.com",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty user",
			requestBody:    models.User{},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("PUT", "/api/users/update", tt.requestBody)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			}
		})
	}
}

func TestUserHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	userHandler := NewUserHandlerWithDefaults(cfg)

	router.DELETE("/api/users/delete/:id", userHandler.Delete)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "delete user with valid ID",
			userID:         "user123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete user with empty ID",
			userID:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete non-existent user",
			userID:         "nonexistent",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/users/delete/" + tt.userID
			req := test.CreateTestRequest("DELETE", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestUserHandler_GetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	userHandler := NewUserHandlerWithDefaults(cfg)

	router.GET("/api/users/profile/:id", userHandler.GetProfile)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "get profile with valid ID",
			userID:         "user123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get profile with empty ID",
			userID:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent user profile",
			userID:         "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/users/profile/" + tt.userID
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestUserHandler_GetUserPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	userHandler := NewUserHandlerWithDefaults(cfg)

	token, err := jwt.GenerateToken("user123", "test@example.com", "Admin", 24*time.Hour)
	require.NoError(t, err)

	router.GET("/api/users/permissions/:id", userHandler.GetUserPermissions)

	tests := []struct {
		name           string
		userID         string
		useAuth        bool
		expectedStatus int
	}{
		{
			name:           "get permissions with valid ID and auth",
			userID:         "user123",
			useAuth:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get permissions without auth",
			userID:         "user123",
			useAuth:        false,
			expectedStatus: http.StatusOK, // Handler might not check auth directly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/users/permissions/" + tt.userID
			var req *http.Request
			if tt.useAuth {
				req = test.CreateAuthenticatedRequest("GET", url, nil, token)
			} else {
				req = test.CreateTestRequest("GET", url, nil)
			}
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

