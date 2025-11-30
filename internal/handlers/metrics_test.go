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

func TestMetricsHandler_GetAllMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	metricsHandler := NewMetricsHandler()

	router.GET("/api/metrics/all", metricsHandler.GetAllMetrics)

	t.Run("get all metrics", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/metrics/all", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "endpoints")
		assert.Contains(t, response, "total_endpoints")
	})
}

func TestMetricsHandler_GetMetricsByService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	metricsHandler := NewMetricsHandler()

	router.GET("/api/metrics/by-service", metricsHandler.GetMetricsByService)

	t.Run("get metrics by service", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/metrics/by-service", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "services")
		assert.Contains(t, response, "total_services")
	})
}

func TestMetricsHandler_GetTopServices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	metricsHandler := NewMetricsHandler()

	router.GET("/api/metrics/top", metricsHandler.GetTopServices)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "get top services with default parameters",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get top services sorted by requests",
			queryParams:    "?sort_by=requests&limit=5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get top services sorted by duration",
			queryParams:    "?sort_by=duration&limit=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get top services with custom limit",
			queryParams:    "?limit=20",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get top services with invalid limit",
			queryParams:    "?limit=invalid",
			expectedStatus: http.StatusOK, // Should default to 10
		},
		{
			name:           "get top services with negative limit",
			queryParams:    "?limit=-5",
			expectedStatus: http.StatusOK, // Should default to 10
		},
		{
			name:           "get top services with zero limit",
			queryParams:    "?limit=0",
			expectedStatus: http.StatusOK, // Should default to 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/metrics/top" + tt.queryParams
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			if err == nil {
				assert.Contains(t, response, "services")
			}
		})
	}
}

func TestMetricsHandler_ResetMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	metricsHandler := NewMetricsHandler()

	router.POST("/api/metrics/reset", metricsHandler.ResetMetrics)

	t.Run("reset metrics", func(t *testing.T) {
		req := test.CreateTestRequest("POST", "/api/metrics/reset", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		if err == nil {
			assert.Contains(t, response, "message")
		}
	})
}

