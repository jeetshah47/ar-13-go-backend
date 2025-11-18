package services

import (
	"context"

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
}

// Verify that concrete types implement interfaces at compile time
var (
	_ CacheServiceInterface       = (*CacheService)(nil)
	_ EmailClientInterface        = (*email.Client)(nil)
	_ ActivityLogServiceInterface = (*ActivityLogService)(nil)
)
