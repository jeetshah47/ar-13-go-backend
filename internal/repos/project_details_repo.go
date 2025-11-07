package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/firebase"
)

// ProjectDetailsRepo handles project details data operations
type ProjectDetailsRepo struct {
	*BaseRepo
}

// NewProjectDetailsRepo creates a new project details repository
func NewProjectDetailsRepo() *ProjectDetailsRepo {
	return &ProjectDetailsRepo{
		BaseRepo: NewBaseRepo("projects"),
	}
}

// Get gets project details for a project
func (r *ProjectDetailsRepo) Get(ctx context.Context, projectID string) (*models.ProjectDetails, error) {
	doc, err := firebase.GetCollection("projects").
		Doc(projectID).
		Collection("project_details").
		Doc("details").
		Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, nil
	}

	var details models.ProjectDetails
	if err := doc.DataTo(&details); err != nil {
		return nil, err
	}
	details.ID = doc.Ref.ID
	details.ProjectID = projectID
	return &details, nil
}

// Add creates project details
func (r *ProjectDetailsRepo) Add(ctx context.Context, projectID string, details *models.ProjectDetails) error {
	detailsRef := firebase.GetCollection("projects").
		Doc(projectID).
		Collection("project_details").
		Doc("details")

	details.ID = detailsRef.ID
	details.ProjectID = projectID
	details.Created = time.Now()

	data := map[string]interface{}{
		"id":        details.ID,
		"projectId": details.ProjectID,
		"data":      details.Data,
		"created":   details.Created,
	}

	_, err := detailsRef.Set(ctx, data)
	return err
}

// Update updates project details
func (r *ProjectDetailsRepo) Update(ctx context.Context, projectID string, details *models.ProjectDetails) error {
	now := time.Now()
	details.Updated = &now

	data := map[string]interface{}{
		"id":        details.ID,
		"projectId": details.ProjectID,
		"data":      details.Data,
		"updated":   now,
	}

	_, err := firebase.GetCollection("projects").
		Doc(projectID).
		Collection("project_details").
		Doc("details").
		Set(ctx, data)
	return err
}

// Delete deletes project details
func (r *ProjectDetailsRepo) Delete(ctx context.Context, projectID, detailsID string) error {
	_, err := firebase.GetCollection("projects").
		Doc(projectID).
		Collection("project_details").
		Doc(detailsID).
		Delete(ctx)
	return err
}
