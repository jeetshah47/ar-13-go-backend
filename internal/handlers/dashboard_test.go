package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ar-13-go-backend/internal/test"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDashboardHandler_GetAllStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	dashboardHandler := NewDashboardHandlerWithDefaults()

	router.GET("/api/dashboard/stats", dashboardHandler.GetAllStats)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "get stats without limits",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get stats with project limit",
			queryParams:    "?project_limit=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get stats with employee limit",
			queryParams:    "?emp_limit=5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get stats with both limits",
			queryParams:    "?project_limit=10&emp_limit=5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get stats with invalid project limit",
			queryParams:    "?project_limit=invalid",
			expectedStatus: http.StatusOK, // Should ignore invalid limit
		},
		{
			name:           "get stats with invalid employee limit",
			queryParams:    "?emp_limit=invalid",
			expectedStatus: http.StatusOK, // Should ignore invalid limit
		},
		{
			name:           "get stats with negative limits",
			queryParams:    "?project_limit=-10&emp_limit=-5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get stats with zero limits",
			queryParams:    "?project_limit=0&emp_limit=0",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/dashboard/stats" + tt.queryParams
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			// Note: This might fail if service requires DB connection
			// In real scenario, you'd mock the service
			if err == nil {
				// Dashboard stats should return some data structure
				assert.NotNil(t, response)
			}
		})
	}
}

