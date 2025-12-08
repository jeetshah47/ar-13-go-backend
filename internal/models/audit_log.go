package models

import "time"

// AuditLog represents an API request audit log entry
type AuditLog struct {
	Model
	Method      string                 `json:"method" bson:"method"`           // HTTP method (GET, POST, PUT, DELETE, etc.)
	Path        string                 `json:"path" bson:"path"`               // Request path/endpoint
	UserID      *string                `json:"userId,omitempty" bson:"userId,omitempty"` // User ID if authenticated
	UserEmail   *string                `json:"userEmail,omitempty" bson:"userEmail,omitempty"` // User email if authenticated
	IPAddress   string                 `json:"ipAddress" bson:"ipAddress"`     // Client IP address
	UserAgent   string                 `json:"userAgent" bson:"userAgent"`     // User agent string
	StatusCode  int                    `json:"statusCode" bson:"statusCode"`   // HTTP response status code
	RequestTime time.Time              `json:"requestTime" bson:"requestTime"` // Timestamp when request was made
	Duration    int64                  `json:"duration" bson:"duration"`       // Request duration in milliseconds
	RequestBody map[string]interface{} `json:"requestBody,omitempty" bson:"requestBody,omitempty"` // Request body (sanitized)
	QueryParams map[string]interface{} `json:"queryParams,omitempty" bson:"queryParams,omitempty"` // Query parameters
	Error       *string                `json:"error,omitempty" bson:"error,omitempty"` // Error message if any
	ResponseSize int64                 `json:"responseSize" bson:"responseSize"` // Response body size in bytes
}

