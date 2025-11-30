package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// SetupTestRouter creates a test router with handlers
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// SetupTestConfig creates a test configuration
func SetupTestConfig() *config.Config {
	return &config.Config{
		JWTSecret:         "test-secret-key-for-testing-only",
		JWTExpiration:     24,
		RefreshExpiration: 7,
		MongoDBURI:        "mongodb://localhost:27017",
		MongoDBDatabase:   "test_db",
		Port:              8080,
		NodeEnv:           "test",
	}
}

// SetupTestHandler creates a test handler with test config
// Note: This function is commented out to avoid import cycle
// Use handlers.NewHandler(cfg) directly in test files if needed
// func SetupTestHandler() *handlers.Handler {
// 	cfg := SetupTestConfig()
// 	jwt.InitializeJWT(cfg.JWTSecret)
// 	return handlers.NewHandler(cfg)
// }

// CreateTestRequest creates an HTTP request for testing
func CreateTestRequest(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// CreateAuthenticatedRequest creates an authenticated HTTP request for testing
func CreateAuthenticatedRequest(method, url string, body interface{}, token string) *http.Request {
	req := CreateTestRequest(method, url, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

// ExecuteRequest executes an HTTP request and returns the response
func ExecuteRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

// AssertJSONResponse asserts that the response matches expected JSON
func AssertJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, expectedBody interface{}) {
	assert.Equal(t, expectedStatus, recorder.Code)

	if expectedBody != nil {
		var expectedJSON, actualJSON interface{}
		expectedBytes, _ := json.Marshal(expectedBody)
		json.Unmarshal(expectedBytes, &expectedJSON)
		json.Unmarshal(recorder.Body.Bytes(), &actualJSON)
		assert.Equal(t, expectedJSON, actualJSON)
	}
}

// AssertErrorResponse asserts that the response contains an error
func AssertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(t, expectedStatus, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

// GenerateTestToken generates a test JWT token for a user
func GenerateTestToken(userID, email, role string) (string, error) {
	cfg := SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	// Use 24 hours expiration for test tokens
	return jwt.GenerateToken(userID, email, role, 24*time.Hour)
}

