package handlers

import (
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// DrawingListHandler handles drawing list routes
type DrawingListHandler struct {
	drawingListService *services.DrawingListService
}

// NewDrawingListHandler creates a new drawing list handler with dependency injection
func NewDrawingListHandler(drawingListService *services.DrawingListService) *DrawingListHandler {
	return &DrawingListHandler{
		drawingListService: drawingListService,
	}
}

// NewDrawingListHandlerWithDefaults creates a new drawing list handler with default dependencies
func NewDrawingListHandlerWithDefaults() *DrawingListHandler {
	return NewDrawingListHandler(
		services.NewDrawingListService(),
	)
}

// ==================== Category Handlers ====================

// GetCategoryByID gets a category by ID
func (h *DrawingListHandler) GetCategoryByID(c *gin.Context) {
	id := c.Param("id")
	category, err := h.drawingListService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if category == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"category": category})
}

// GetAllCategories gets all categories
func (h *DrawingListHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.drawingListService.GetAllCategories(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{
		"categories": categories,
		"total":      len(categories),
	})
}

// AddCategory creates a new category
func (h *DrawingListHandler) AddCategory(c *gin.Context) {
	var category models.DrawingCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.drawingListService.AddCategory(c.Request.Context(), &category); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{
		"message":  "Category created successfully",
		"category": category,
	})
}

// UpdateCategory updates a category
func (h *DrawingListHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.DrawingCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.ID = id
	if err := h.drawingListService.UpdateCategory(c.Request.Context(), &category); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message":  "Category updated successfully",
		"category": category,
	})
}

// DeleteCategory deletes a category
func (h *DrawingListHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.drawingListService.DeleteCategory(c.Request.Context(), id); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// ==================== Type Handlers ====================

// GetTypeByID gets a drawing type by ID
func (h *DrawingListHandler) GetTypeByID(c *gin.Context) {
	id := c.Param("id")
	drawingType, err := h.drawingListService.GetTypeByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if drawingType == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Drawing type not found"})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"type": drawingType})
}

// GetTypesByCategoryID gets all types for a category
func (h *DrawingListHandler) GetTypesByCategoryID(c *gin.Context) {
	categoryID := c.Param("categoryId")
	types, err := h.drawingListService.GetTypesByCategoryID(c.Request.Context(), categoryID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{
		"types": types,
		"total": len(types),
	})
}

// GetAllTypes gets all drawing types
func (h *DrawingListHandler) GetAllTypes(c *gin.Context) {
	types, err := h.drawingListService.GetAllTypes(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{
		"types": types,
		"total":  len(types),
	})
}

// AddType creates a new drawing type
func (h *DrawingListHandler) AddType(c *gin.Context) {
	var drawingType models.DrawingType
	if err := c.ShouldBindJSON(&drawingType); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.drawingListService.AddType(c.Request.Context(), &drawingType); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{
		"message": "Drawing type created successfully",
		"type":    drawingType,
	})
}

// UpdateType updates a drawing type
func (h *DrawingListHandler) UpdateType(c *gin.Context) {
	id := c.Param("id")
	var drawingType models.DrawingType
	if err := c.ShouldBindJSON(&drawingType); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	drawingType.ID = id
	if err := h.drawingListService.UpdateType(c.Request.Context(), &drawingType); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message": "Drawing type updated successfully",
		"type":    drawingType,
	})
}

// DeleteType deletes a drawing type
func (h *DrawingListHandler) DeleteType(c *gin.Context) {
	id := c.Param("id")
	if err := h.drawingListService.DeleteType(c.Request.Context(), id); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Drawing type deleted successfully"})
}

// ==================== Combined Handlers ====================

// GetCategoriesWithTypes gets all categories with their associated types
func (h *DrawingListHandler) GetCategoriesWithTypes(c *gin.Context) {
	categoriesWithTypes, err := h.drawingListService.GetCategoriesWithTypes(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{
		"categories": categoriesWithTypes,
		"total":      len(categoriesWithTypes),
	})
}

