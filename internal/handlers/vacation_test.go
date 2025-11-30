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

func TestVacationHandler_GetMyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	vacationHandler := NewVacationHandlerWithDefaults()

	router.GET("/api/vacation/my-requests", vacationHandler.GetMyRequests)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "get my requests with user ID in context",
			setupContext: func(c *gin.Context) {
				c.Set("userID", "user123")
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "get my requests without user ID",
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(c *gin.Context) {
				tt.setupContext(c)
				vacationHandler.GetMyRequests(c)
			}
			router.GET("/api/vacation/my-requests", handler)

			req := test.CreateTestRequest("GET", "/api/vacation/my-requests", nil)
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
				if err == nil {
					assert.Contains(t, response, "requests")
				}
			}
		})
	}
}

func TestVacationHandler_CreateRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	vacationHandler := NewVacationHandlerWithDefaults()

	router.POST("/api/vacation/create", vacationHandler.CreateRequest)

	startDate := time.Now().Add(24 * time.Hour)
	endDate := startDate.Add(7 * 24 * time.Hour)

	tests := []struct {
		name           string
		requestBody    models.VacationRequest
		setupContext   func(*gin.Context)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid vacation request",
			requestBody: models.VacationRequest{
				Type:      "annual",
				StartDate: startDate,
				EndDate:   endDate,
			},
			setupContext: func(c *gin.Context) {
				c.Set("userID", "user123")
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "vacation request without user ID",
			requestBody: models.VacationRequest{
				Type:      "annual",
				StartDate: startDate,
				EndDate:   endDate,
			},
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "empty vacation request",
			requestBody:    models.VacationRequest{},
			setupContext:   func(c *gin.Context) { c.Set("userID", "user123") },
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(c *gin.Context) {
				tt.setupContext(c)
				vacationHandler.CreateRequest(c)
			}
			router.POST("/api/vacation/create", handler)

			req := test.CreateTestRequest("POST", "/api/vacation/create", tt.requestBody)
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

func TestVacationHandler_GetOneRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	vacationHandler := NewVacationHandlerWithDefaults()

	router.GET("/api/vacation/:requestId", vacationHandler.GetOneRequest)

	tests := []struct {
		name           string
		requestID      string
		expectedStatus int
	}{
		{
			name:           "get request with valid ID",
			requestID:      "request123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get request with empty ID",
			requestID:      "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent request",
			requestID:      "nonexistent",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/vacation/" + tt.requestID
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestVacationHandler_UpdateRequestStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	vacationHandler := NewVacationHandlerWithDefaults()

	router.PUT("/api/vacation/update-status/:requestId", vacationHandler.UpdateRequestStatus)

	tests := []struct {
		name           string
		requestID      string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "update status to approved",
			requestID: "request123",
			requestBody: map[string]interface{}{
				"status": "approved",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:      "update status to rejected",
			requestID: "request123",
			requestBody: map[string]interface{}{
				"status": "rejected",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "update status with empty request ID",
			requestID:      "",
			requestBody:    map[string]interface{}{"status": "approved"},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/vacation/update-status/" + tt.requestID
			req := test.CreateTestRequest("PUT", url, tt.requestBody)
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

func TestVacationHandler_DeleteRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	vacationHandler := NewVacationHandlerWithDefaults()

	router.DELETE("/api/vacation/delete/:requestId", vacationHandler.DeleteRequest)

	tests := []struct {
		name           string
		requestID      string
		expectedStatus int
	}{
		{
			name:           "delete request with valid ID",
			requestID:      "request123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete request with empty ID",
			requestID:      "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/vacation/delete/" + tt.requestID
			req := test.CreateTestRequest("DELETE", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

