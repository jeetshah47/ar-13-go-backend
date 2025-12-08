package services

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"go.mongodb.org/mongo-driver/bson"
)

// AuditLogService handles audit log business logic
type AuditLogService struct {
	auditLogRepo *repos.AuditLogRepo
}

// NewAuditLogService creates a new audit log service with dependency injection
func NewAuditLogService(auditLogRepo *repos.AuditLogRepo) *AuditLogService {
	return &AuditLogService{
		auditLogRepo: auditLogRepo,
	}
}

// NewAuditLogServiceWithDefaults creates a new audit log service with default dependencies
func NewAuditLogServiceWithDefaults() *AuditLogService {
	return NewAuditLogService(
		repos.NewAuditLogRepo(),
	)
}

// GetRecent gets recent audit logs with optional filters
func (s *AuditLogService) GetRecent(ctx context.Context, limit int, filters map[string]interface{}) ([]models.AuditLog, error) {
	// Convert filters to bson.M for MongoDB query
	bsonFilters := bson.M{}
	
	if userId, ok := filters["userId"].(string); ok && userId != "" {
		bsonFilters["userId"] = userId
	}
	
	if path, ok := filters["path"].(string); ok && path != "" {
		bsonFilters["path"] = path
	}
	
	if method, ok := filters["method"].(string); ok && method != "" {
		bsonFilters["method"] = method
	}
	
	if statusCode, ok := filters["statusCode"].(int); ok && statusCode > 0 {
		bsonFilters["statusCode"] = statusCode
	}
	
	if minStatusCode, ok := filters["minStatusCode"].(int); ok && minStatusCode > 0 {
		if maxStatusCode, ok := filters["maxStatusCode"].(int); ok && maxStatusCode > 0 {
			bsonFilters["statusCode"] = bson.M{
				"$gte": minStatusCode,
				"$lte": maxStatusCode,
			}
		} else {
			bsonFilters["statusCode"] = bson.M{
				"$gte": minStatusCode,
			}
		}
	}

	return s.auditLogRepo.GetRecent(ctx, limit, bsonFilters)
}

// GetByUserID gets audit logs for a specific user
func (s *AuditLogService) GetByUserID(ctx context.Context, userID string, limit *int) ([]models.AuditLog, error) {
	return s.auditLogRepo.GetByUserID(ctx, userID, limit)
}

// GetByPath gets audit logs for a specific path
func (s *AuditLogService) GetByPath(ctx context.Context, path string, limit *int) ([]models.AuditLog, error) {
	return s.auditLogRepo.GetByPath(ctx, path, limit)
}

// GetByMethod gets audit logs for a specific HTTP method
func (s *AuditLogService) GetByMethod(ctx context.Context, method string, limit *int) ([]models.AuditLog, error) {
	return s.auditLogRepo.GetByMethod(ctx, method, limit)
}

// GetByStatusCode gets audit logs for a specific status code
func (s *AuditLogService) GetByStatusCode(ctx context.Context, statusCode int, limit *int) ([]models.AuditLog, error) {
	return s.auditLogRepo.GetByStatusCode(ctx, statusCode, limit)
}

// GetByDateRange gets audit logs within a date range
func (s *AuditLogService) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit *int) ([]models.AuditLog, error) {
	return s.auditLogRepo.GetByDateRange(ctx, startDate, endDate, limit)
}

