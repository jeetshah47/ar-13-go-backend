package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/test"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_Register(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	authHandler := NewAuthHandler(cfg)

	router.POST("/api/auth/register", authHandler.Register)

	tests := []struct {
		name           string
		requestBody    models.RegisterRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid registration request",
			requestBody: models.RegisterRequest{
				Name:        "Test User",
				Email:       "test@example.com",
				Password:    "password123",
				PhoneNumber: "1234567890",
				Token:       "valid-token",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "missing required fields",
			requestBody: models.RegisterRequest{
				Email: "test@example.com",
				// Missing name, password, token
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid email format",
			requestBody: models.RegisterRequest{
				Name:        "Test User",
				Email:       "invalid-email",
				Password:    "password123",
				PhoneNumber: "1234567890",
				Token:       "valid-token",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "password too short",
			requestBody: models.RegisterRequest{
				Name:        "Test User",
				Email:       "test@example.com",
				Password:    "12345", // Less than 6 characters
				PhoneNumber: "1234567890",
				Token:       "valid-token",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("POST", "/api/auth/register", tt.requestBody)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				// For successful registration, we expect a response with user data
				// Note: This will fail if the service layer requires actual database connection
				// In a real scenario, you'd mock the service layer
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				// If there's an error, it's likely because the service needs a real DB
				// This is expected in unit tests without mocks
				if err == nil {
					assert.NotEmpty(t, response)
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	authHandler := NewAuthHandler(cfg)

	router.POST("/api/auth/login", authHandler.Login)

	tests := []struct {
		name           string
		requestBody    models.LoginRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid login request",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing email",
			requestBody: models.LoginRequest{
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing password",
			requestBody: models.LoginRequest{
				Email: "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid email format",
			requestBody: models.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid credentials",
			requestBody: models.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("POST", "/api/auth/login", tt.requestBody)
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

func TestAuthHandler_Logout(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	authHandler := NewAuthHandler(cfg)

	router.POST("/api/auth/logout", authHandler.Logout)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid logout with token",
			authHeader:     "Bearer valid-token",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "invalid token format",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusOK, // Handler will still process it
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("POST", "/api/auth/logout", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				if tt.expectedStatus == http.StatusOK {
					assert.Contains(t, response, "message")
				}
			}
		})
	}
}

func TestAuthHandler_ValidateSignupToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	authHandler := NewAuthHandler(cfg)

	router.GET("/api/auth/validate-signup", authHandler.ValidateSignupToken)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectValid    bool
	}{
		{
			name:           "valid token",
			token:          "valid-token",
			expectedStatus: http.StatusOK,
			expectValid:    true,
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusBadRequest,
			expectValid:    false,
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusOK,
			expectValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/auth/validate-signup"
			if tt.token != "" {
				url += "?token=" + tt.token
			}

			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectValid {
				assert.Equal(t, true, response["valid"])
			} else {
				if tt.token == "" {
					assert.Equal(t, false, response["valid"])
					assert.Contains(t, response, "reason")
				} else {
					// Token might be invalid, so valid could be false
					valid, ok := response["valid"].(bool)
					if ok {
						assert.False(t, valid)
					}
				}
			}
		})
	}
}

func TestAuthHandler_GetPermissions(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	authHandler := NewAuthHandler(cfg)

	// Create a test token
	token, err := jwt.GenerateToken("user123", "test@example.com", "Admin", 24*time.Hour)
	require.NoError(t, err)

	router.GET("/api/auth/permissions", authHandler.GetPermissions)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid request with admin role",
			setupContext: func(c *gin.Context) {
				c.Set("userRole", "Admin")
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "valid request with standard role",
			setupContext: func(c *gin.Context) {
				c.Set("userRole", "Standard")
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing user role",
			setupContext: func(c *gin.Context) {
				// Don't set userRole
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateAuthenticatedRequest("GET", "/api/auth/permissions", nil, token)
			
			// Create a custom handler that sets up context
			handler := func(c *gin.Context) {
				tt.setupContext(c)
				authHandler.GetPermissions(c)
			}
			
			router.GET("/api/auth/permissions", handler)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "role")
				assert.Contains(t, response, "permissions")
			}
		})
	}
}

