package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProjectRepo handles project data operations with MongoDB
type ProjectRepo struct {
	*MongoBaseRepo
}

// NewProjectRepo creates a new MongoDB project repository
func NewProjectRepo() *ProjectRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &ProjectRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "projects"),
	}
}

// GetByID gets a project by ID
func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*models.Project, error) {
	// Query by id field - check both root level and nested "model.id"
	// (MongoDB may store embedded structs as nested objects)
	filter := bson.M{
		"$or": []bson.M{
			{"id": id},
			{"model.id": id},
			{"_id": id},
		},
	}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	// Decode into a map first for better compatibility with MongoDB document structure
	var rawDoc bson.M
	if err := result.Decode(&rawDoc); err != nil {
		return nil, err
	}

	// Convert map to BSON bytes, then unmarshal to struct
	// This approach handles type conversions and missing fields better
	bsonBytes, err := bson.Marshal(rawDoc)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := bson.Unmarshal(bsonBytes, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// GetAll gets all projects
func (r *ProjectRepo) GetAll(ctx context.Context, limit *int) ([]models.Project, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}

	items, err := r.FindAll(ctx, bson.M{}, limitInt64)
	if err != nil {
		return nil, err
	}

	projects := make([]models.Project, 0, len(items))
	for _, item := range items {
		var project models.Project
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &project); err != nil {
			continue
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

	return r.InsertOne(ctx, project)
}

// Update updates a project
func (r *ProjectRepo) Update(ctx context.Context, project *models.Project) error {
	now := time.Now()
	project.Updated = &now

	updates := bson.M{
		"title":       project.Title,
		"description": project.Description,
		"ownerId":     project.OwnerID,
		"membersIds":  project.MembersIDs,
		"deadLine":    project.Deadline,
		"updated":     project.Updated,
	}

	if project.StartDate != nil {
		updates["startDate"] = *project.StartDate
	}
	if project.EndDate != nil {
		updates["endDate"] = *project.EndDate
	}
	if project.LogoURL != nil {
		updates["logoUrl"] = *project.LogoURL
	}

	return r.UpdateOne(ctx, project.ID, updates)
}

// Delete deletes a project
func (r *ProjectRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// Persists checks if a project exists by title
func (r *ProjectRepo) Persists(ctx context.Context, title string) (bool, error) {
	filter := bson.M{"title": title}
	count, err := r.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
