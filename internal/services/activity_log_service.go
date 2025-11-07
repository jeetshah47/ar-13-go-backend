package services

import (
	"context"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// ActivityLogService handles activity log business logic
type ActivityLogService struct {
	activityLogRepo *repos.ActivityLogRepo
	userRepo        *repos.UserRepo
}

// NewActivityLogService creates a new activity log service
func NewActivityLogService() *ActivityLogService {
	return &ActivityLogService{
		activityLogRepo: repos.NewActivityLogRepo(),
		userRepo:        repos.NewUserRepo(),
	}
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
func (s *ActivityLogService) GetByEntityType(ctx context.Context, entityType models.ActivityLogEntityType, limit *int) ([]models.ActivityLogResponse, error) {
	logs, err := s.activityLogRepo.GetByEntityType(ctx, entityType, limit)
	if err != nil {
		return nil, err
	}

	return s.populateUserDetails(ctx, logs)
}

// Add adds an activity log
func (s *ActivityLogService) Add(ctx context.Context, log *models.ActivityLogBase) error {
	return s.activityLogRepo.Add(ctx, log)
}
