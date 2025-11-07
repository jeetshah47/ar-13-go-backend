package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// ProjectRepo handles project data operations
// TODO: Migrate to DynamoDB
type ProjectRepo struct {
	*DynamoBaseRepo
}

// NewProjectRepo creates a new project repository
func NewProjectRepo() *ProjectRepo {
	return &ProjectRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("projects"),
	}
}

// convertDeadlineIfNeeded converts deadLine from string to time.Time if needed
func (r *ProjectRepo) convertDeadlineIfNeeded(data map[string]interface{}) error {
	return ConvertTimeFieldsInMap(data, []string{"deadLine", "created", "updated"})
}

// dataToProject converts a data map to a Project struct
func (r *ProjectRepo) dataToProject(data map[string]interface{}, id string) (*models.Project, error) {
	project := &models.Project{}
	project.ID = id

	if title, ok := data["title"].(string); ok {
		project.Title = title
	}
	if description, ok := data["description"].(string); ok {
		project.Description = description
	}
	if ownerID, ok := data["ownerId"].(string); ok {
		project.OwnerID = ownerID
	}
	if membersIDs, ok := data["membersIds"].([]interface{}); ok {
		project.MembersIDs = make([]string, 0, len(membersIDs))
		for _, id := range membersIDs {
			if strID, ok := id.(string); ok {
				project.MembersIDs = append(project.MembersIDs, strID)
			}
		}
	}
	// Handle deadLine - use ConvertToTime helper
	if deadLineVal, ok := data["deadLine"]; ok && deadLineVal != nil {
		if deadLine, err := ConvertToTime(deadLineVal); err == nil {
			project.Deadline = deadLine
		}
	}
	if logoURL, ok := data["logoUrl"].(string); ok {
		project.LogoURL = &logoURL
	}
	// Handle created - use ConvertToTime helper
	if createdVal, ok := data["created"]; ok && createdVal != nil {
		if created, err := ConvertToTime(createdVal); err == nil {
			project.Created = created
		}
	}
	// Handle updated - use ConvertToTime helper
	if updatedVal, ok := data["updated"]; ok && updatedVal != nil {
		if updated, err := ConvertToTime(updatedVal); err == nil {
			project.Updated = &updated
		}
	}

	return project, nil
}

// GetByID gets a project by ID
func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*models.Project, error) {
	item, err := r.DynamoBaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var project models.Project
	if err := UnmarshalItem(item, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// GetAll gets all projects
func (r *ProjectRepo) GetAll(ctx context.Context, limit *int) ([]models.Project, error) {
	var limitInt32 *int32
	if limit != nil {
		l := int32(*limit)
		limitInt32 = &l
	}

	items, err := r.ScanItems(ctx, limitInt32)
	if err != nil {
		return nil, err
	}

	projects := make([]models.Project, 0, len(items))
	for _, item := range items {
		var project models.Project
		if err := UnmarshalItem(item, &project); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// Add creates a new project
func (r *ProjectRepo) Add(ctx context.Context, project *models.Project) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	project.Created = time.Now()

	data := map[string]interface{}{
		"id":          project.ID,
		"title":       project.Title,
		"description": project.Description,
		"ownerId":     project.OwnerID,
		"membersIds":  project.MembersIDs,
		"deadLine":    project.Deadline.Format(time.RFC3339),
		"created":     project.Created.Format(time.RFC3339),
	}
	if project.LogoURL != nil {
		data["logoUrl"] = *project.LogoURL
	}

	return r.PutItem(ctx, data)
}

// Update updates a project
func (r *ProjectRepo) Update(ctx context.Context, project *models.Project) error {
	now := time.Now()
	project.Updated = &now

	updates := map[string]interface{}{
		"title":       project.Title,
		"description": project.Description,
		"ownerId":     project.OwnerID,
		"membersIds":  project.MembersIDs,
		"deadLine":    project.Deadline.Format(time.RFC3339),
	}
	if project.LogoURL != nil {
		updates["logoUrl"] = *project.LogoURL
	}

	return r.UpdateItem(ctx, project.ID, updates)
}

// Delete deletes a project
func (r *ProjectRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// Persists checks if a project exists by title
func (r *ProjectRepo) Persists(ctx context.Context, title string) (bool, error) {
	// TODO: Implement DynamoDB scan with filter by title
	items, err := r.ScanItems(ctx, nil)
	if err != nil {
		return false, err
	}

	for _, item := range items {
		if t, ok := item["title"].(*types.AttributeValueMemberS); ok && t.Value == title {
			return true, nil
		}
	}
	return false, nil
}
