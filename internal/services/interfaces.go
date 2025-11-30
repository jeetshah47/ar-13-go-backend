package services

import (
	"context"
	"io"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/email"
)

// CacheServiceInterface defines the interface for cache operations
type CacheServiceInterface interface {
	GetUser(ctx context.Context, userID string, dest interface{}) error
	SetUser(ctx context.Context, userID string, user interface{}) error
	GetTask(ctx context.Context, taskID string, dest interface{}) error
	SetTask(ctx context.Context, taskID string, task interface{}) error
	InvalidateTask(ctx context.Context, taskID string) error
	GetProjectStats(ctx context.Context) ([]interface{}, error)
	SetProjectStats(ctx context.Context, stats []interface{}) error
	InvalidateProjectStats(ctx context.Context) error
	GetDashboardStats(ctx context.Context, projectLimit, empLimit int) (map[string]interface{}, error)
	SetDashboardStats(ctx context.Context, projectLimit, empLimit int, stats map[string]interface{}) error
	InvalidateDashboardStats(ctx context.Context) error
	GetCalendarMonth(ctx context.Context, year, month int) ([]interface{}, error)
	SetCalendarMonth(ctx context.Context, year, month int, events []interface{}) error
	InvalidateCalendarMonth(ctx context.Context, year, month int) error
	GetActivityLogs(ctx context.Context, entityType string, limit int) ([]interface{}, error)
	SetActivityLogs(ctx context.Context, entityType string, limit int, logs []interface{}) error
	InvalidateActivityLogs(ctx context.Context, entityType string) error
}

// EmailClientInterface defines the interface for email operations
type EmailClientInterface interface {
	SendNotificationEmail(notification *models.Notification, userEmail string) error
	SendAlertEmail(to []string, title, message string, severity string) error
	SendSignupLinkEmail(data email.SignupEmailData) error
}

// ActivityLogServiceInterface defines the interface for activity log operations
type ActivityLogServiceInterface interface {
	Add(ctx context.Context, log *models.ActivityLogBase) error
	GetByEntity(ctx context.Context, entityType models.ActivityLogEntityType, entityID string) ([]models.ActivityLogResponse, error)
	GetByEntityType(ctx context.Context, entityType models.ActivityLogEntityType, limit *int) ([]models.ActivityLogResponse, error)
	GetByID(ctx context.Context, activityLogID string) (*models.ActivityLogBase, error)
}

// ActivityLogReplyServiceInterface defines the interface for activity log reply operations
type ActivityLogReplyServiceInterface interface {
	Add(ctx context.Context, reply *models.ActivityLogReply) error
	GetByActivityLogID(ctx context.Context, activityLogID string) ([]models.ActivityLogReplyResponse, error)
}

// WebSocketServiceInterface defines the interface for WebSocket operations
type WebSocketServiceInterface interface {
	SendToUser(userID string, eventType string, data interface{}) error
	BroadcastToProjectMembers(projectID string, eventType string, data interface{})
	BroadcastToProjectMembersWithProject(project *models.Project, eventType string, data interface{})
}

// StorageServiceInterface defines the interface for storage operations
type StorageServiceInterface interface {
	Initialize() error
	IsInitialized() bool
	ListObjects(ctx context.Context, prefix string) ([]StorageObject, error)
	UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
	DeleteObject(ctx context.Context, objectName string) error
	ObjectExists(ctx context.Context, objectName string) (bool, error)
}

// StorageObject represents a file or folder in storage
type StorageObject struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	IsFolder     bool      `json:"isFolder"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ContentType  string    `json:"contentType,omitempty"`
}

// Note: WebSocketServiceInterface is implemented by pkg/websocket.WebSocketService
// We can't add a compile-time check here due to import cycle prevention
// The interface methods match the WebSocketService implementation

// Verify that concrete types implement interfaces at compile time
var (
	_ CacheServiceInterface            = (*CacheService)(nil)
	_ EmailClientInterface             = (*email.Client)(nil)
	_ ActivityLogServiceInterface      = (*ActivityLogService)(nil)
	_ ActivityLogReplyServiceInterface = (*ActivityLogReplyService)(nil)
)
