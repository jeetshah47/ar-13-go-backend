package services

import (
	"context"
	"errors"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// ProjectDetailsService handles project details business logic
type ProjectDetailsService struct {
	projectDetailsRepo *repos.ProjectDetailsRepo
}

// NewProjectDetailsService creates a new project details service
func NewProjectDetailsService() *ProjectDetailsService {
	return &ProjectDetailsService{
		projectDetailsRepo: repos.NewProjectDetailsRepo(),
	}
}

// Get gets project details
func (s *ProjectDetailsService) Get(ctx context.Context, projectID string) (*models.ProjectDetails, error) {
	return s.projectDetailsRepo.Get(ctx, projectID)
}

// Add creates project details
func (s *ProjectDetailsService) Add(ctx context.Context, projectID string, details *models.ProjectDetails) error {
	details.ProjectID = projectID
	return s.projectDetailsRepo.Add(ctx, projectID, details)
}

// Update updates project details
func (s *ProjectDetailsService) Update(ctx context.Context, projectID string, details *models.ProjectDetails) error {
	existing, err := s.projectDetailsRepo.Get(ctx, projectID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("project details not found")
	}

	details.ProjectID = projectID
	details.ID = existing.ID
	return s.projectDetailsRepo.Update(ctx, projectID, details)
}

// Delete deletes project details
func (s *ProjectDetailsService) Delete(ctx context.Context, projectID, detailsID string) error {
	return s.projectDetailsRepo.Delete(ctx, projectID, detailsID)
}
