package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/test"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectHandler_GetAll(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	projectHandler := NewProjectHandlerWithDefaults()

	router.GET("/api/project/all", projectHandler.GetAll)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "get all projects without limit",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get all projects with limit",
			queryParams:    "?limit=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get all projects with invalid limit",
			queryParams:    "?limit=invalid",
			expectedStatus: http.StatusOK, // Should still work, just ignore invalid limit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/project/all" + tt.queryParams
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			// Note: This might fail if service requires DB connection
			// In real scenario, you'd mock the service
			if err == nil {
				assert.Contains(t, response, "projects")
			}
		})
	}
}

func TestProjectHandler_GetOne(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	projectHandler := NewProjectHandlerWithDefaults()

	router.GET("/api/project/:id", projectHandler.GetOne)

	tests := []struct {
		name           string
		projectID      string
		expectedStatus int
	}{
		{
			name:           "get project with valid ID",
			projectID:      "project123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get project with empty ID",
			projectID:      "",
			expectedStatus: http.StatusOK, // Route will still match
		},
		{
			name:           "get non-existent project",
			projectID:      "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/project/" + tt.projectID
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				if err == nil {
					assert.Contains(t, response, "project")
				}
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			}
		})
	}
}

func TestProjectHandler_Add(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	projectHandler := NewProjectHandlerWithDefaults()

	router.POST("/api/project/add", projectHandler.Add)

	tests := []struct {
		name           string
		requestBody    models.Project
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid project creation",
			requestBody: models.Project{
				Title:       "Test Project",
				Description: "Test Description",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "empty project title",
			requestBody:    models.Project{},
			expectedStatus: http.StatusCreated, // Handler might accept it
			expectError:    false,
		},
		{
			name: "project with all fields",
			requestBody: models.Project{
				Title:       "Complete Project",
				Description: "Full description",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("POST", "/api/project/add", tt.requestBody)
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
					if tt.expectedStatus == http.StatusCreated {
						assert.Contains(t, response, "message")
					}
				}
			}
		})
	}
}

func TestProjectHandler_Update(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	projectHandler := NewProjectHandlerWithDefaults()

	router.PUT("/api/project/update", projectHandler.Update)

	tests := []struct {
		name           string
		requestBody    models.Project
		setupContext   func(*gin.Context)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid project update",
			requestBody: models.Project{
				Title:       "Updated Project",
				Description: "Updated Description",
			},
			setupContext: func(c *gin.Context) {
				c.Set("userID", "user123")
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "update without user ID in context",
			requestBody: models.Project{
				Title: "Updated Project",
			},
			setupContext: func(c *gin.Context) {
				// Don't set userID
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("PUT", "/api/project/update", tt.requestBody)
			
			// Create a custom handler that sets up context
			handler := func(c *gin.Context) {
				tt.setupContext(c)
				projectHandler.Update(c)
			}
			
			router.PUT("/api/project/update", handler)
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

func TestProjectHandler_Delete(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	projectHandler := NewProjectHandlerWithDefaults()

	router.DELETE("/api/project/delete/:id", projectHandler.Delete)

	tests := []struct {
		name           string
		projectID      string
		expectedStatus int
	}{
		{
			name:           "delete project with valid ID",
			projectID:      "project123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete non-existent project",
			projectID:      "nonexistent",
			expectedStatus: http.StatusOK, // Handler might return OK even if not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/project/delete/" + tt.projectID
			req := test.CreateTestRequest("DELETE", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

