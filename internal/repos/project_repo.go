package repos

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

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

// GetAll gets all projects (excludes archived projects by default)
func (r *ProjectRepo) GetAll(ctx context.Context, limit *int) ([]models.Project, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}

	// Filter out archived projects
	filter := bson.M{
		"$or": []bson.M{
			{"isArchived": bson.M{"$ne": true}},
			{"isArchived": bson.M{"$exists": false}},
		},
	}

	items, err := r.FindAll(ctx, filter, limitInt64)
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

// generateProjectCode generates a project code in the format mmyy-title
func generateProjectCode(title string) string {
	now := time.Now()
	month := fmt.Sprintf("%02d", int(now.Month()))
	year := fmt.Sprintf("%02d", now.Year()%100)
	
	// Convert title to URL-friendly format: lowercase, replace spaces with hyphens, remove special chars
	titleSlug := strings.ToLower(title)
	titleSlug = strings.TrimSpace(titleSlug)
	
	// Replace spaces and special characters with hyphens
	var builder strings.Builder
	lastWasHyphen := false
	for _, r := range titleSlug {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
			lastWasHyphen = false
		} else if (r == ' ' || r == '-' || r == '_') && !lastWasHyphen {
			builder.WriteRune('-')
			lastWasHyphen = true
		}
	}
	titleSlug = strings.Trim(builder.String(), "-")
	
	return fmt.Sprintf("%s%s-%s", month, year, titleSlug)
}

// Add creates a new project
func (r *ProjectRepo) Add(ctx context.Context, project *models.Project) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	project.Created = time.Now()
	
	// Generate code if not provided
	if project.Code == "" {
		project.Code = generateProjectCode(project.Title)
	}

	// Default isArchived to false for new projects
	// (This ensures backward compatibility - existing projects without the field will be treated as not archived)
	// The field is already initialized to false in the struct, but we make it explicit here

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
		"project_code": project.Code,
		"updated":     project.Updated,
		"isArchived":  project.IsArchived,
	}

	if project.ProductionDuration != nil {
		updates["productionDuration"] = *project.ProductionDuration
	}
	if project.SiteDuration != nil {
		updates["siteDuration"] = *project.SiteDuration
	}
	if project.LogoURL != nil {
		updates["logoUrl"] = *project.LogoURL
	}
	if project.AgencyContact != nil {
		updates["agencyContact"] = project.AgencyContact
	}

	return r.UpdateOne(ctx, project.ID, updates)
}

// UpdateAgencyContact updates only the agency contact for a project
func (r *ProjectRepo) UpdateAgencyContact(ctx context.Context, projectID string, agencyContact *models.AgencyContact) error {
	now := time.Now()
	updates := bson.M{
		"agencyContact": agencyContact,
		"updated":       &now,
	}
	return r.UpdateOne(ctx, projectID, updates)
}

// Delete deletes a project
func (r *ProjectRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// GetByUserID gets all projects where the user is either the owner or a member
// Optimized with single query and efficient filtering
// Excludes archived projects
func (r *ProjectRepo) GetByUserID(ctx context.Context, userID string) ([]models.Project, error) {
	// Query projects where user is owner OR in members list, and not archived
	// MongoDB will use indexes on ownerId and membersIds if available
	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"ownerId": userID},
					{"membersIds": userID},
				},
			},
			{
				"$or": []bson.M{
					{"isArchived": bson.M{"$ne": true}},
					{"isArchived": bson.M{"$exists": false}},
				},
			},
		},
	}

	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	// Pre-allocate slice with known capacity to avoid reallocations
	projects := make([]models.Project, 0, len(items))
	for _, item := range items {
		var project models.Project
		bsonBytes, err := bson.Marshal(item)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(bsonBytes, &project); err != nil {
			continue
		}
		projects = append(projects, project)
	}

	return projects, nil
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

// Archive archives or unarchives a project
func (r *ProjectRepo) Archive(ctx context.Context, projectID string, isArchived bool) error {
	now := time.Now()
	updates := bson.M{
		"isArchived": isArchived,
		"updated":    &now,
	}
	return r.UpdateOne(ctx, projectID, updates)
}
