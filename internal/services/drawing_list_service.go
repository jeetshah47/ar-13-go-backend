package services

import (
	"context"
	"errors"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// DrawingListService handles drawing list business logic
type DrawingListService struct {
	categoryRepo *repos.DrawingListRepo
	typeRepo     *repos.DrawingTypeRepo
}

// NewDrawingListService creates a new drawing list service
func NewDrawingListService() *DrawingListService {
	return &DrawingListService{
		categoryRepo: repos.NewDrawingListRepo(),
		typeRepo:     repos.NewDrawingTypeRepo(),
	}
}

// ==================== Category Methods ====================

// GetCategoryByID gets a category by ID
func (s *DrawingListService) GetCategoryByID(ctx context.Context, id string) (*models.DrawingCategory, error) {
	return s.categoryRepo.GetCategoryByID(ctx, id)
}

// GetAllCategories gets all categories
func (s *DrawingListService) GetAllCategories(ctx context.Context) ([]models.DrawingCategory, error) {
	return s.categoryRepo.GetAllCategories(ctx)
}

// AddCategory creates a new category
func (s *DrawingListService) AddCategory(ctx context.Context, category *models.DrawingCategory) error {
	if category.Name == "" {
		return errors.New("category name is required")
	}
	return s.categoryRepo.AddCategory(ctx, category)
}

// UpdateCategory updates a category
func (s *DrawingListService) UpdateCategory(ctx context.Context, category *models.DrawingCategory) error {
	if category.ID == "" {
		return errors.New("category ID is required")
	}
	if category.Name == "" {
		return errors.New("category name is required")
	}
	return s.categoryRepo.UpdateCategory(ctx, category)
}

// DeleteCategory deletes a category
func (s *DrawingListService) DeleteCategory(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("category ID is required")
	}
	return s.categoryRepo.DeleteCategory(ctx, id)
}

// ==================== Type Methods ====================

// GetTypeByID gets a drawing type by ID
func (s *DrawingListService) GetTypeByID(ctx context.Context, id string) (*models.DrawingType, error) {
	return s.typeRepo.GetTypeByID(ctx, id)
}

// GetTypesByCategoryID gets all types for a category
func (s *DrawingListService) GetTypesByCategoryID(ctx context.Context, categoryID string) ([]models.DrawingType, error) {
	return s.typeRepo.GetTypesByCategoryID(ctx, categoryID)
}

// GetAllTypes gets all drawing types
func (s *DrawingListService) GetAllTypes(ctx context.Context) ([]models.DrawingType, error) {
	return s.typeRepo.GetAllTypes(ctx)
}

// AddType creates a new drawing type
func (s *DrawingListService) AddType(ctx context.Context, drawingType *models.DrawingType) error {
	if drawingType.Name == "" {
		return errors.New("drawing type name is required")
	}
	if drawingType.CategoryID == "" {
		return errors.New("category ID is required")
	}
	return s.typeRepo.AddType(ctx, drawingType)
}

// UpdateType updates a drawing type
func (s *DrawingListService) UpdateType(ctx context.Context, drawingType *models.DrawingType) error {
	if drawingType.ID == "" {
		return errors.New("drawing type ID is required")
	}
	if drawingType.Name == "" {
		return errors.New("drawing type name is required")
	}
	if drawingType.CategoryID == "" {
		return errors.New("category ID is required")
	}
	return s.typeRepo.UpdateType(ctx, drawingType)
}

// DeleteType deletes a drawing type
func (s *DrawingListService) DeleteType(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("drawing type ID is required")
	}
	return s.typeRepo.DeleteType(ctx, id)
}

// ==================== Combined Methods ====================

// GetCategoriesWithTypes gets all categories with their associated types
func (s *DrawingListService) GetCategoriesWithTypes(ctx context.Context) ([]models.DrawingCategoryWithTypes, error) {
	return s.categoryRepo.GetCategoriesWithTypes(ctx)
}

