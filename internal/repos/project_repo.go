package repos

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/internal/models"
	"google.golang.org/api/iterator"
)

// ProjectRepo handles project data operations
type ProjectRepo struct {
	*BaseRepo
}

// NewProjectRepo creates a new project repository
func NewProjectRepo() *ProjectRepo {
	return &ProjectRepo{
		BaseRepo: NewBaseRepo("projects"),
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
	doc, err := r.collection.Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, nil
	}

	data := doc.Data()
	if err := r.convertDeadlineIfNeeded(data); err != nil {
		return nil, err
	}

	project, err := r.dataToProject(data, doc.Ref.ID)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// GetAll gets all projects
func (r *ProjectRepo) GetAll(ctx context.Context, limit *int) ([]models.Project, error) {
	query := r.collection.OrderBy("created", firestore.Desc)
	if limit != nil {
		query = query.Limit(*limit)
	}

	iter := query.Documents(ctx)
	var projects []models.Project

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		if err := r.convertDeadlineIfNeeded(data); err != nil {
			return nil, err
		}

		project, err := r.dataToProject(data, doc.Ref.ID)
		if err != nil {
			return nil, err
		}
		projects = append(projects, *project)
	}

	return projects, nil
}

// Add creates a new project
func (r *ProjectRepo) Add(ctx context.Context, project *models.Project) error {
	newDocRef := r.collection.NewDoc()
	project.ID = newDocRef.ID
	project.Created = time.Now()

	data := map[string]interface{}{
		"id":          project.ID,
		"title":       project.Title,
		"description": project.Description,
		"ownerId":     project.OwnerID,
		"membersIds":  project.MembersIDs,
		"deadLine":    project.Deadline,
		"created":     project.Created,
	}
	if project.LogoURL != nil {
		data["logoUrl"] = *project.LogoURL
	}

	_, err := newDocRef.Set(ctx, data)
	if err != nil {
		return err
	}

	// Create project details subcollection
	projectDetailsRef := newDocRef.Collection("project_details").NewDoc()
	_, err = projectDetailsRef.Set(ctx, map[string]interface{}{
		"attachments": []interface{}{},
		"comments":    []interface{}{},
		"details":     "",
	})
	return err
}

// Update updates a project
func (r *ProjectRepo) Update(ctx context.Context, project *models.Project) error {
	now := time.Now()
	project.Updated = &now

	data := map[string]interface{}{
		"id":          project.ID,
		"title":       project.Title,
		"description": project.Description,
		"ownerId":     project.OwnerID,
		"membersIds":  project.MembersIDs,
		"deadLine":    project.Deadline,
		"updated":     now,
	}
	if project.LogoURL != nil {
		data["logoUrl"] = *project.LogoURL
	}

	_, err := r.collection.Doc(project.ID).Set(ctx, data)
	return err
}

// Delete deletes a project
func (r *ProjectRepo) Delete(ctx context.Context, id string) error {
	_, err := r.collection.Doc(id).Delete(ctx)
	return err
}

// Persists checks if a project exists by title
func (r *ProjectRepo) Persists(ctx context.Context, title string) (bool, error) {
	iter := r.collection.Where("title", "==", title).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return doc.Exists(), nil
}
