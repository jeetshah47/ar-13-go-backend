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

func TestEmployeeHandler_GetEmployeeList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	employeeHandler := NewEmployeeHandlerWithDefaults()

	router.GET("/api/employee/list", employeeHandler.GetEmployeeList)

	t.Run("get employee list", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/employee/list", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		if err == nil {
			assert.Contains(t, response, "employees")
			assert.Contains(t, response, "totalEmployees")
		}
	})
}

func TestEmployeeHandler_GetEmployeeTaskCounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	employeeHandler := NewEmployeeHandlerWithDefaults()

	router.GET("/api/employee/task-counts/:userId", employeeHandler.GetEmployeeTaskCounts)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "get task counts with valid user ID",
			userID:         "user123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get task counts with empty user ID",
			userID:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get task counts for non-existent user",
			userID:         "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/employee/task-counts/" + tt.userID
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				if err == nil {
					assert.Contains(t, response, "employee")
				}
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
			}
		})
	}
}

func TestEmployeeHandler_GetEmployeeTaskStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	employeeHandler := NewEmployeeHandlerWithDefaults()

	router.GET("/api/employee/stats/:userId", employeeHandler.GetEmployeeTaskStats)

	tests := []struct {
		name           string
		userID         string
		queryParams    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "get stats with valid parameters",
			userID:         "user123",
			queryParams:    "?period=month&periodValue=2024-01",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "get stats with quarter period",
			userID:         "user123",
			queryParams:    "?period=quarter&periodValue=2024-Q1",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "get stats with year period",
			userID:         "user123",
			queryParams:    "?period=year&periodValue=2024",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "get stats with project filter",
			userID:         "user123",
			queryParams:    "?period=month&periodValue=2024-01&projectId=project123",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing period parameter",
			userID:         "user123",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "missing periodValue parameter",
			userID:         "user123",
			queryParams:    "?period=month",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid period value",
			userID:         "user123",
			queryParams:    "?period=invalid&periodValue=2024-01",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/employee/stats/" + tt.userID + tt.queryParams
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
			}
		})
	}
}

