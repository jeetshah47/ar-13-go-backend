package services

import (
	"context"
	"encoding/json"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// ActivityLogService handles activity log business logic
type ActivityLogService struct {
	activityLogRepo repos.ActivityLogRepository
	userRepo        repos.UserRepository
	cacheSvc        CacheServiceInterface
}

// NewActivityLogService creates a new activity log service with dependency injection
func NewActivityLogService(
	activityLogRepo repos.ActivityLogRepository,
	userRepo repos.UserRepository,
	cacheSvc CacheServiceInterface,
) *ActivityLogService {
	return &ActivityLogService{
		activityLogRepo: activityLogRepo,
		userRepo:        userRepo,
		cacheSvc:        cacheSvc,
	}
}

// NewActivityLogServiceWithDefaults creates a new activity log service with default dependencies
func NewActivityLogServiceWithDefaults() *ActivityLogService {
	return NewActivityLogService(
		repos.NewActivityLogRepo(),
		repos.NewUserRepo(),
		NewCacheService(),
	)
}

// populateUserDetails populates CreatedByUser for activity logs
func (s *ActivityLogService) populateUserDetails(ctx context.Context, logs []models.ActivityLogBase) ([]models.ActivityLogResponse, error) {
	responses := make([]models.ActivityLogResponse, 0, len(logs))
	userCache := make(map[string]*models.User)

	for _, log := range logs {
		response := models.ActivityLogResponse{
			ActivityLogBase: log,
		}

		// Fetch user if not in cache
		if log.CreatedBy != "" {
			user, exists := userCache[log.CreatedBy]
			if !exists {
				fetchedUser, err := s.userRepo.GetByID(ctx, log.CreatedBy)
				if err != nil {
					// Log error but don't fail the entire request
					// User will be nil if not found
					fetchedUser = nil
				}
				user = fetchedUser
				userCache[log.CreatedBy] = user
			}
			response.CreatedByUser = user
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetByEntity gets activity logs for a specific entity
func (s *ActivityLogService) GetByEntity(ctx context.Context, entityType models.ActivityLogEntityType, entityID string) ([]models.ActivityLogResponse, error) {
	logs, err := s.activityLogRepo.GetByEntity(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}

	return s.populateUserDetails(ctx, logs)
}

// GetByEntityType gets activity logs by entity type
// Uses Redis cache to improve performance
func (s *ActivityLogService) GetByEntityType(ctx context.Context, entityType models.ActivityLogEntityType, limit *int) ([]models.ActivityLogResponse, error) {
	limitValue := 10 // default limit
	if limit != nil {
		limitValue = *limit
	}

	// Try to get from cache first
	cached, err := s.cacheSvc.GetActivityLogs(ctx, string(entityType), limitValue)
	if err == nil && cached != nil {
		// Convert cached interface{} slice to ActivityLogResponse slice
		logs := make([]models.ActivityLogResponse, 0, len(cached))
		for _, item := range cached {
			if logMap, ok := item.(map[string]interface{}); ok {
				var log models.ActivityLogResponse
				if data, err := json.Marshal(logMap); err == nil {
					if err := json.Unmarshal(data, &log); err == nil {
						logs = append(logs, log)
					}
				}
			}
		}
		if len(logs) > 0 {
			return logs, nil
		}
	}

	// Cache miss or error - fetch from DB
	logs, err := s.activityLogRepo.GetByEntityType(ctx, entityType, limit)
	if err != nil {
		return nil, err
	}

	// Populate user details
	responses, err := s.populateUserDetails(ctx, logs)
	if err != nil {
		return nil, err
	}

	// Convert to interface{} slice for caching
	cacheData := make([]interface{}, len(responses))
	for i := range responses {
		cacheData[i] = responses[i]
	}

	// Cache the result (ignore cache errors)
	_ = s.cacheSvc.SetActivityLogs(ctx, string(entityType), limitValue, cacheData)

	return responses, nil
}

// Add adds an activity log
func (s *ActivityLogService) Add(ctx context.Context, log *models.ActivityLogBase) error {
	if err := s.activityLogRepo.Add(ctx, log); err != nil {
		return err
	}
	// Invalidate activity logs cache for this entity type
	_ = s.cacheSvc.InvalidateActivityLogs(ctx, string(log.EntityType))
	return nil
}

// GetByID gets an activity log by ID
func (s *ActivityLogService) GetByID(ctx context.Context, activityLogID string) (*models.ActivityLogBase, error) {
	return s.activityLogRepo.GetByID(ctx, activityLogID)
}