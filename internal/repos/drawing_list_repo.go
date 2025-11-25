package repos

import (
	"context"
	"errors"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DrawingListRepo handles drawing list data operations with MongoDB
type DrawingListRepo struct {
	*MongoBaseRepo
}

// NewDrawingListRepo creates a new MongoDB drawing list repository
func NewDrawingListRepo() *DrawingListRepo {
	client := mongodb.GetClient()
	if client == nil {
		panic("MongoDB client is not initialized. Please ensure MongoDB is connected before creating repositories.")
	}
	return &DrawingListRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, mongodb.GetDatabaseName(), "drawing_categories"),
	}
}

// ==================== Drawing Category Methods ====================

// GetCategoryByID gets a drawing category by ID
func (r *DrawingListRepo) GetCategoryByID(ctx context.Context, id string) (*models.DrawingCategory, error) {
	result := r.GetByID(ctx, id)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var category models.DrawingCategory
	if err := result.Decode(&category); err != nil {
		return nil, err
	}

	return &category, nil
}

// GetAllCategories gets all drawing categories ordered by order field
func (r *DrawingListRepo) GetAllCategories(ctx context.Context) ([]models.DrawingCategory, error) {
	filter := bson.M{}
	items, err := r.FindAll(ctx, filter, nil, bson.M{"order": 1})
	if err != nil {
		return nil, err
	}

	categories := make([]models.DrawingCategory, 0, len(items))
	for _, item := range items {
		var category models.DrawingCategory
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &category); err != nil {
			continue
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// AddCategory creates a new drawing category
func (r *DrawingListRepo) AddCategory(ctx context.Context, category *models.DrawingCategory) error {
	now := time.Now()
	// Always ensure ID is set - critical for unique index
	// If ID is empty, generate a new UUID
	if category.ID == "" {
		category.ID = uuid.New().String()
	}
	// Verify ID is set before insertion (safety check)
	if category.ID == "" {
		return errors.New("failed to generate category ID")
	}
	category.Created = now

	return r.InsertOne(ctx, category)
}

// UpdateCategory updates a drawing category
func (r *DrawingListRepo) UpdateCategory(ctx context.Context, category *models.DrawingCategory) error {
	now := time.Now()
	category.Updated = &now

	updates := bson.M{
		"name":        category.Name,
		"description": category.Description,
		"order":       category.Order,
		"isActive":    category.IsActive,
		"updated":     category.Updated,
	}

	return r.UpdateOne(ctx, category.ID, updates)
}

// DeleteCategory deletes a drawing category
func (r *DrawingListRepo) DeleteCategory(ctx context.Context, id string) error {
	// Check if category has associated types
	typeRepo := NewDrawingTypeRepo()
	types, err := typeRepo.GetTypesByCategoryID(ctx, id)
	if err != nil {
		return err
	}
	if len(types) > 0 {
		return errors.New("cannot delete category: it has associated drawing types")
	}

	return r.DeleteByID(ctx, id)
}

// ==================== Drawing Type Methods ====================

// DrawingTypeRepo handles drawing type data operations
type DrawingTypeRepo struct {
	*MongoBaseRepo
}

// NewDrawingTypeRepo creates a new MongoDB drawing type repository
func NewDrawingTypeRepo() *DrawingTypeRepo {
	client := mongodb.GetClient()
	if client == nil {
		panic("MongoDB client is not initialized. Please ensure MongoDB is connected before creating repositories.")
	}
	return &DrawingTypeRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, mongodb.GetDatabaseName(), "drawing_types"),
	}
}

// GetTypeByID gets a drawing type by ID
func (r *DrawingTypeRepo) GetTypeByID(ctx context.Context, id string) (*models.DrawingType, error) {
	result := r.GetByID(ctx, id)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var drawingType models.DrawingType
	if err := result.Decode(&drawingType); err != nil {
		return nil, err
	}

	return &drawingType, nil
}

// GetTypesByCategoryID gets all drawing types for a category
func (r *DrawingTypeRepo) GetTypesByCategoryID(ctx context.Context, categoryID string) ([]models.DrawingType, error) {
	filter := bson.M{"categoryId": categoryID}
	items, err := r.FindAll(ctx, filter, nil, bson.M{"order": 1})
	if err != nil {
		return nil, err
	}

	types := make([]models.DrawingType, 0, len(items))
	for _, item := range items {
		var drawingType models.DrawingType
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &drawingType); err != nil {
			continue
		}
		types = append(types, drawingType)
	}

	return types, nil
}

// GetAllTypes gets all drawing types ordered by category and order
func (r *DrawingTypeRepo) GetAllTypes(ctx context.Context) ([]models.DrawingType, error) {
	filter := bson.M{}
	opts := options.Find().SetSort(bson.D{
		{Key: "categoryId", Value: 1},
		{Key: "order", Value: 1},
	})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []bson.M
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	types := make([]models.DrawingType, 0, len(items))
	for _, item := range items {
		var drawingType models.DrawingType
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &drawingType); err != nil {
			continue
		}
		types = append(types, drawingType)
	}

	return types, nil
}

// AddType creates a new drawing type
func (r *DrawingTypeRepo) AddType(ctx context.Context, drawingType *models.DrawingType) error {
	// Verify category exists
	categoryRepo := NewDrawingListRepo()
	category, err := categoryRepo.GetCategoryByID(ctx, drawingType.CategoryID)
	if err != nil {
		return err
	}
	if category == nil {
		return errors.New("category not found")
	}

	now := time.Now()
	if drawingType.ID == "" {
		drawingType.ID = uuid.New().String()
	}
	drawingType.Created = now

	return r.InsertOne(ctx, drawingType)
}

// UpdateType updates a drawing type
func (r *DrawingTypeRepo) UpdateType(ctx context.Context, drawingType *models.DrawingType) error {
	// If categoryId is being updated, verify new category exists
	if drawingType.CategoryID != "" {
		categoryRepo := NewDrawingListRepo()
		category, err := categoryRepo.GetCategoryByID(ctx, drawingType.CategoryID)
		if err != nil {
			return err
		}
		if category == nil {
			return errors.New("category not found")
		}
	}

	now := time.Now()
	drawingType.Updated = &now

	updates := bson.M{
		"categoryId":  drawingType.CategoryID,
		"name":        drawingType.Name,
		"description": drawingType.Description,
		"order":       drawingType.Order,
		"isActive":    drawingType.IsActive,
		"updated":     drawingType.Updated,
	}

	return r.UpdateOne(ctx, drawingType.ID, updates)
}

// DeleteType deletes a drawing type
func (r *DrawingTypeRepo) DeleteType(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// GetCategoriesWithTypes gets all categories with their associated types
func (r *DrawingListRepo) GetCategoriesWithTypes(ctx context.Context) ([]models.DrawingCategoryWithTypes, error) {
	categories, err := r.GetAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	typeRepo := NewDrawingTypeRepo()
	result := make([]models.DrawingCategoryWithTypes, 0, len(categories))

	for _, category := range categories {
		types, err := typeRepo.GetTypesByCategoryID(ctx, category.ID)
		if err != nil {
			continue
		}

		result = append(result, models.DrawingCategoryWithTypes{
			DrawingCategory: category,
			Types:           types,
		})
	}

	return result, nil
}
