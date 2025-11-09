package services

import (
	"context"
	"errors"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// AuthorizationService handles authorization logic
type AuthorizationService struct {
	projectRepo *repos.ProjectRepo
	taskRepo    *repos.TaskRepo
	userRepo    *repos.UserRepo
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService() *AuthorizationService {
	return &AuthorizationService{
		projectRepo: repos.NewProjectRepo(),
		taskRepo:    repos.NewTaskRepo(),
		userRepo:    repos.NewUserRepo(),
	}
}

// IsAdmin checks if a user is an admin
func (s *AuthorizationService) IsAdmin(ctx context.Context, userID string) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return user.Role == models.UserRoleAdmin, nil
}

// IsProjectOwnerOrMember checks if a user is the owner or a member of a project
func (s *AuthorizationService) IsProjectOwnerOrMember(ctx context.Context, projectID, userID string) (bool, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return false, err
	}
	if project == nil {
		return false, errors.New(constants.MsgProjectNotFound)
	}

	// Check if user is the owner
	if project.OwnerID == userID {
		return true, nil
	}

	// Check if user is a member
	for _, memberID := range project.MembersIDs {
		if memberID == userID {
			return true, nil
		}
	}

	return false, nil
}

// IsTaskAssignedToUser checks if a task is assigned to a user
func (s *AuthorizationService) IsTaskAssignedToUser(ctx context.Context, projectID, taskID, userID string) (bool, error) {
	task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return false, err
	}
	if task == nil {
		return false, errors.New(constants.MsgTaskNotFound)
	}

	if task.AssignTo != nil && *task.AssignTo == userID {
		return true, nil
	}

	return false, nil
}

// CanModifyProject checks if a user can modify a project (must be owner or member, or admin)
func (s *AuthorizationService) CanModifyProject(ctx context.Context, projectID, userID string) error {
	// Admins have full access
	isAdmin, err := s.IsAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if isAdmin {
		return nil
	}

	canModify, err := s.IsProjectOwnerOrMember(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !canModify {
		return errors.New(constants.MsgCannotModifyProject)
	}
	return nil
}

// CanModifyTask checks if a user can modify a task
// User can modify if they are assigned to the task OR they are project owner/member OR they are admin
func (s *AuthorizationService) CanModifyTask(ctx context.Context, projectID, taskID, userID string) error {
	// Admins have full access
	isAdmin, err := s.IsAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if isAdmin {
		return nil
	}

	// Check if user is assigned to the task
	isAssigned, err := s.IsTaskAssignedToUser(ctx, projectID, taskID, userID)
	if err != nil {
		return err
	}
	if isAssigned {
		return nil
	}

	// Check if user is project owner or member
	canModify, err := s.IsProjectOwnerOrMember(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if canModify {
		return nil
	}

	return errors.New(constants.MsgCannotModifyTask)
}

// CanClaimTask checks if a user can claim a task (must be project member or admin)
func (s *AuthorizationService) CanClaimTask(ctx context.Context, projectID, userID string) error {
	// Admins have full access
	isAdmin, err := s.IsAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if isAdmin {
		return nil
	}

	canModify, err := s.IsProjectOwnerOrMember(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !canModify {
		return errors.New(constants.MsgCannotClaimTask)
	}
	return nil
}

// CanModifyTimeLog checks if a user can modify time logs
// User can modify if they are assigned to the task OR they are project owner/member
func (s *AuthorizationService) CanModifyTimeLog(ctx context.Context, projectID, taskID, userID string) error {
	return s.CanModifyTask(ctx, projectID, taskID, userID)
}

// CanAssignTask checks if a user can assign tasks (must be project owner or member, or admin)
func (s *AuthorizationService) CanAssignTask(ctx context.Context, projectID, userID string) error {
	// Admins have full access
	isAdmin, err := s.IsAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if isAdmin {
		return nil
	}

	canModify, err := s.IsProjectOwnerOrMember(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !canModify {
		return errors.New(constants.MsgCannotAssignTask)
	}
	return nil
}
