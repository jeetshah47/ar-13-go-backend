package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProjectDetailsRepo handles project details data operations with MongoDB
type ProjectDetailsRepo struct {
	*MongoBaseRepo
}

// NewProjectDetailsRepo creates a new MongoDB project details repository
func NewProjectDetailsRepo() *ProjectDetailsRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &ProjectDetailsRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "project_details"),
	}
}

// Get gets project details for a project
func (r *ProjectDetailsRepo) Get(ctx context.Context, projectID string) (*models.ProjectDetails, error) {
	filter := bson.M{"projectId": projectID}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var details models.ProjectDetails
	if err := result.Decode(&details); err != nil {
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

	return r.InsertOne(ctx, details)
}

// Update updates project details
func (r *ProjectDetailsRepo) Update(ctx context.Context, projectID string, details *models.ProjectDetails) error {
	updates := bson.M{
		"projectId": details.ProjectID,
		"data":      details.Data,
	}

	return r.UpdateOne(ctx, details.ID, updates)
}

// Delete deletes project details
func (r *ProjectDetailsRepo) Delete(ctx context.Context, projectID, detailsID string) error {
	return r.DeleteByID(ctx, detailsID)
}

