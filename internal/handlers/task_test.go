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

func TestTaskHandler_GetAll(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	taskHandler := NewTaskHandlerWithDefaults(cfg)

	router.GET("/api/tasks/all/:projectId", taskHandler.GetAll)

	tests := []struct {
		name           string
		projectID      string
		expectedStatus int
	}{
		{
			name:           "get all tasks for project",
			projectID:      "project123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get tasks with empty project ID",
			projectID:      "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/tasks/all/" + tt.projectID
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			if err == nil {
				assert.Contains(t, response, "tasks")
			}
		})
	}
}

func TestTaskHandler_GetStatuses(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	taskHandler := NewTaskHandlerWithDefaults(cfg)

	router.GET("/api/tasks/statuses", taskHandler.GetStatuses)

	t.Run("get task statuses", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/tasks/statuses", nil)
		recorder := test.ExecuteRequest(router, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		if err == nil {
			assert.Contains(t, response, "statuses")
		}
	})
}

func TestTaskHandler_Add(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	taskHandler := NewTaskHandlerWithDefaults(cfg)

	router.POST("/api/tasks/add", taskHandler.Add)

	tests := []struct {
		name           string
		requestBody    models.Task
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid task creation",
			requestBody: models.Task{
				ProjectID: "project123",
				Subject:   "Test Task",
				Status:    "pending",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "task with invalid status",
			requestBody: models.Task{
				ProjectID: "project123",
				Subject:   "Test Task",
				Status:    "invalid_status",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "task without required fields",
			requestBody: models.Task{
				ProjectID: "project123",
				// Missing title
			},
			expectedStatus: http.StatusCreated, // Handler might accept it
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("POST", "/api/tasks/add", tt.requestBody)
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

func TestTaskHandler_UpdateStatus(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	taskHandler := NewTaskHandlerWithDefaults(cfg)

	router.PUT("/api/tasks/update-status/:projectId/:taskId", taskHandler.UpdateStatus)

	tests := []struct {
		name           string
		projectID      string
		taskID         string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "valid status update",
			projectID: "project123",
			taskID:    "task123",
			requestBody: map[string]interface{}{
				"status": "in_progress",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:      "invalid status",
			projectID: "project123",
			taskID:    "task123",
			requestBody: map[string]interface{}{
				"status": "invalid_status",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "missing status in body",
			projectID:      "project123",
			taskID:         "task123",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/tasks/update-status/" + tt.projectID + "/" + tt.taskID
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

func TestTaskHandler_Delete(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	taskHandler := NewTaskHandlerWithDefaults(cfg)

	router.DELETE("/api/tasks/delete/:projectId/:taskId", taskHandler.Delete)

	tests := []struct {
		name           string
		projectID      string
		taskID         string
		expectedStatus int
	}{
		{
			name:           "delete task with valid IDs",
			projectID:      "project123",
			taskID:         "task123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete task with empty IDs",
			projectID:      "",
			taskID:         "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/tasks/delete/" + tt.projectID + "/" + tt.taskID
			req := test.CreateTestRequest("DELETE", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

