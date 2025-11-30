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

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

func TestCalendarHandler_GetByMonth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	calendarHandler := NewCalendarHandlerWithDefaults(cfg)

	router.GET("/api/calendar/month/:year/:month", calendarHandler.GetByMonth)

	tests := []struct {
		name           string
		year           string
		month          string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid year and month",
			year:           "2024",
			month:          "1",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "valid year and month as string",
			year:           "2024",
			month:          "12",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid month",
			year:           "2024",
			month:          "13",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid year",
			year:           "invalid",
			month:          "1",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid month format",
			year:           "2024",
			month:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "zero month",
			year:           "2024",
			month:          "0",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/calendar/month/" + tt.year + "/" + tt.month
			req := test.CreateTestRequest("GET", url, nil)
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
					assert.Contains(t, response, "events")
				}
			}
		})
	}
}

func TestCalendarHandler_GetById(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	calendarHandler := NewCalendarHandlerWithDefaults(cfg)

	router.GET("/api/calendar/event/:id", calendarHandler.GetById)

	tests := []struct {
		name           string
		eventID        string
		expectedStatus int
	}{
		{
			name:           "get event with valid ID",
			eventID:        "event123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get event with empty ID",
			eventID:        "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent event",
			eventID:        "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/calendar/event/" + tt.eventID
			req := test.CreateTestRequest("GET", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestCalendarHandler_Add(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	calendarHandler := NewCalendarHandlerWithDefaults(cfg)

	router.POST("/api/calendar/add", calendarHandler.Add)

	validDate := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name           string
		requestBody    models.CalendarEvent
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid calendar event",
			requestBody: models.CalendarEvent{
				Title: "Test Event",
				Start: validDate,
				End:   validDate.Add(2 * time.Hour),
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "empty calendar event",
			requestBody:    models.CalendarEvent{},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "event with all fields",
			requestBody: models.CalendarEvent{
				Title:       "Complete Event",
				Description: stringPtr("Event description"),
				Start:       validDate,
				End:         validDate.Add(2 * time.Hour),
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := test.CreateTestRequest("POST", "/api/calendar/add", tt.requestBody)
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

func TestCalendarHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	calendarHandler := NewCalendarHandlerWithDefaults(cfg)

	router.PUT("/api/calendar/update/:id", calendarHandler.Update)

	validDate := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name           string
		eventID        string
		requestBody    models.CalendarEvent
		expectedStatus int
		expectError    bool
	}{
		{
			name:    "valid calendar event update",
			eventID:  "event123",
			requestBody: models.CalendarEvent{
				Title: "Updated Event",
				Start: validDate,
				End:   validDate.Add(2 * time.Hour),
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "update with empty ID",
			eventID:        "",
			requestBody:    models.CalendarEvent{},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/calendar/update/" + tt.eventID
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

func TestCalendarHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := test.SetupTestConfig()
	jwt.InitializeJWT(cfg.JWTSecret)
	calendarHandler := NewCalendarHandlerWithDefaults(cfg)

	router.DELETE("/api/calendar/delete/:id", calendarHandler.Delete)

	tests := []struct {
		name           string
		eventID        string
		expectedStatus int
	}{
		{
			name:           "delete event with valid ID",
			eventID:        "event123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete event with empty ID",
			eventID:        "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/calendar/delete/" + tt.eventID
			req := test.CreateTestRequest("DELETE", url, nil)
			recorder := test.ExecuteRequest(router, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

