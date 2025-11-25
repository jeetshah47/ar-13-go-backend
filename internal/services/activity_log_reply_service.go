package services

import (
	"context"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// ActivityLogReplyService handles activity log reply business logic
type ActivityLogReplyService struct {
	replyRepo repos.ActivityLogReplyRepository
	userRepo  repos.UserRepository
}

// NewActivityLogReplyService creates a new activity log reply service with dependency injection
func NewActivityLogReplyService(
	replyRepo repos.ActivityLogReplyRepository,
	userRepo repos.UserRepository,
) *ActivityLogReplyService {
	return &ActivityLogReplyService{
		replyRepo: replyRepo,
		userRepo:  userRepo,
	}
}

// NewActivityLogReplyServiceWithDefaults creates a new activity log reply service with default dependencies
func NewActivityLogReplyServiceWithDefaults() *ActivityLogReplyService {
	return NewActivityLogReplyService(
		repos.NewActivityLogReplyRepo(),
		repos.NewUserRepo(),
	)
}

// populateUserDetails populates CreatedByUser for activity log replies
func (s *ActivityLogReplyService) populateUserDetails(ctx context.Context, replies []models.ActivityLogReply) ([]models.ActivityLogReplyResponse, error) {
	responses := make([]models.ActivityLogReplyResponse, 0, len(replies))
	userCache := make(map[string]*models.User)

	for _, reply := range replies {
		response := models.ActivityLogReplyResponse{
			ActivityLogReply: reply,
		}

		// Fetch user if not in cache
		if reply.CreatedBy != "" {
			user, exists := userCache[reply.CreatedBy]
			if !exists {
				fetchedUser, err := s.userRepo.GetByID(ctx, reply.CreatedBy)
				if err != nil {
					// Log error but don't fail the entire request
					// User will be nil if not found
					fetchedUser = nil
				}
				user = fetchedUser
				userCache[reply.CreatedBy] = user
			}
			response.CreatedByUser = user
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// Add adds an activity log reply
func (s *ActivityLogReplyService) Add(ctx context.Context, reply *models.ActivityLogReply) error {
	return s.replyRepo.Add(ctx, reply)
}

// GetByActivityLogID gets all replies for a specific activity log
func (s *ActivityLogReplyService) GetByActivityLogID(ctx context.Context, activityLogID string) ([]models.ActivityLogReplyResponse, error) {
	replies, err := s.replyRepo.GetByActivityLogID(ctx, activityLogID)
	if err != nil {
		return nil, err
	}

	return s.populateUserDetails(ctx, replies)
}

