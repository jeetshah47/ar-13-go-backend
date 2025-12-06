package services

import (
	"context"
	"errors"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// VacationService handles vacation/leave request business logic
type VacationService struct {
	vacationRepo repos.VacationRepository
	userRepo     repos.UserRepository
}

// NewVacationService creates a new vacation service with dependency injection
func NewVacationService(
	vacationRepo repos.VacationRepository,
	userRepo repos.UserRepository,
) *VacationService {
	return &VacationService{
		vacationRepo: vacationRepo,
		userRepo:     userRepo,
	}
}

// NewVacationServiceWithDefaults creates a new vacation service with default dependencies
func NewVacationServiceWithDefaults() *VacationService {
	return NewVacationService(
		repos.NewVacationRepo(),
		repos.NewUserRepo(),
	)
}

// GetByUserID gets leave requests for a user
func (s *VacationService) GetByUserID(ctx context.Context, userID string) ([]models.LeaveRequest, error) {
	return s.vacationRepo.GetByUserID(ctx, userID)
}

// GetAll gets all leave requests
func (s *VacationService) GetAll(ctx context.Context) ([]models.LeaveRequest, error) {
	return s.vacationRepo.GetAll(ctx)
}

// GetPending gets pending leave requests
func (s *VacationService) GetPending(ctx context.Context) ([]models.LeaveRequest, error) {
	return s.vacationRepo.GetPending(ctx)
}

// GetByID gets a leave request by ID
func (s *VacationService) GetByID(ctx context.Context, id string) (*models.LeaveRequest, error) {
	return s.vacationRepo.GetByID(ctx, id)
}

// GetByStatus gets leave requests by status
func (s *VacationService) GetByStatus(ctx context.Context, status models.LeaveRequestStatus) ([]models.LeaveRequest, error) {
	return s.vacationRepo.GetByStatus(ctx, status)
}

// GetByType gets leave requests by type
func (s *VacationService) GetByType(ctx context.Context, requestType models.LeaveRequestType) ([]models.LeaveRequest, error) {
	return s.vacationRepo.GetByType(ctx, requestType)
}

// Create creates a new leave request
func (s *VacationService) Create(ctx context.Context, userID string, requestType models.LeaveRequestType, startDate time.Time, endDate *time.Time, duration float64, durationType models.DurationType, comments *string, workingHours *models.WorkingHours) (*models.LeaveRequest, error) {
	request := &models.LeaveRequest{
		UserID:       userID,
		RequestType:  requestType,
		StartDate:    startDate,
		EndDate:      endDate,
		Duration:     duration,
		DurationType: durationType,
		Status:       models.LeaveRequestStatusPending,
		Comments:     comments,
		WorkingHours: workingHours,
	}

	if err := s.vacationRepo.Add(ctx, request); err != nil {
		return nil, err
	}

	return request, nil
}

// Update updates a leave request
func (s *VacationService) Update(ctx context.Context, request *models.LeaveRequest) error {
	existing, err := s.vacationRepo.GetByID(ctx, request.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("leave request not found")
	}

	return s.vacationRepo.Update(ctx, request)
}

// UpdateStatus updates leave request status
func (s *VacationService) UpdateStatus(ctx context.Context, id string, status models.LeaveRequestStatus, reviewedBy string, reviewComments *string) error {
	existing, err := s.vacationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("leave request not found")
	}

	return s.vacationRepo.UpdateStatus(ctx, id, status, reviewedBy, reviewComments)
}

// Delete deletes a leave request
func (s *VacationService) Delete(ctx context.Context, id string) error {
	existing, err := s.vacationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("leave request not found")
	}

	return s.vacationRepo.Delete(ctx, id)
}

// GetVacationSummaries gets vacation summaries for all users
func (s *VacationService) GetVacationSummaries(ctx context.Context) ([]models.VacationSummary, error) {
	// Get all leave requests
	requests, err := s.vacationRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Get all users
	users, err := s.userRepo.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Calculate summaries
	userMap := make(map[string]*models.VacationSummary)
	for _, user := range users {
		userMap[user.ID] = &models.VacationSummary{
			UserID:           user.ID,
			UserName:         user.Name,
			UserEmail:        user.Email,
			VacationDays:     0,
			SickLeaveDays:    0,
			WorkRemotelyDays: 0,
			PendingRequests:  0,
		}
		if user.Designation != nil {
			userMap[user.ID].UserAvatar = user.Designation // Using designation as placeholder
		}
	}

	// Aggregate request data
	for _, req := range requests {
		summary, exists := userMap[req.UserID]
		if !exists {
			continue
		}

		if req.Status == models.LeaveRequestStatusPending {
			summary.PendingRequests++
		}

		if req.Status == models.LeaveRequestStatusApproved {
			switch req.RequestType {
			case models.LeaveRequestTypeVacation:
				summary.VacationDays += req.Duration
			case models.LeaveRequestTypeSickLeave:
				summary.SickLeaveDays += req.Duration
			case models.LeaveRequestTypeWorkRemotely:
				summary.WorkRemotelyDays += req.Duration
			}
		}
	}

	// Convert map to slice
	summaries := make([]models.VacationSummary, 0, len(userMap))
	for _, summary := range userMap {
		summaries = append(summaries, *summary)
	}

	return summaries, nil
}
