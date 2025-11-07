package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
)

// ProjectDetailsRepo handles project details data operations
type ProjectDetailsRepo struct {
	*DynamoBaseRepo
}

// NewProjectDetailsRepo creates a new project details repository
func NewProjectDetailsRepo() *ProjectDetailsRepo {
	return &ProjectDetailsRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("project_details"),
	}
}

// Get gets project details for a project
func (r *ProjectDetailsRepo) Get(ctx context.Context, projectID string) (*models.ProjectDetails, error) {
	items, err := r.QueryByIndex(ctx, "projectId-index", "projectId", projectID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	var details models.ProjectDetails
	if err := UnmarshalItem(items[0], &details); err != nil {
		return nil, err
	}
	return &details, nil
}

// Add creates project details
func (r *ProjectDetailsRepo) Add(ctx context.Context, projectID string, details *models.ProjectDetails) error {
	if details.ID == "" {
		details.ID = fmt.Sprintf("%s-details", projectID)
	}
	details.ProjectID = projectID
	details.Created = time.Now()

	data := map[string]interface{}{
		"id":        details.ID,
		"projectId": details.ProjectID,
		"data":      details.Data,
		"created":   details.Created.Format(time.RFC3339),
	}

	return r.PutItem(ctx, data)
}

// Update updates project details
func (r *ProjectDetailsRepo) Update(ctx context.Context, projectID string, details *models.ProjectDetails) error {
	updates := map[string]interface{}{
		"projectId": details.ProjectID,
		"data":      details.Data,
	}

	return r.UpdateItem(ctx, details.ID, updates)
}

// Delete deletes project details
func (r *ProjectDetailsRepo) Delete(ctx context.Context, projectID, detailsID string) error {
	return r.DeleteByID(ctx, detailsID)
}
