package models

// DrawingCategory represents a category of architectural drawings
type DrawingCategory struct {
	Model
	Name        string `json:"name" bson:"name"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Order       int    `json:"order" bson:"order"`
	IsActive    bool   `json:"isActive" bson:"isActive"`
}

// DrawingType represents a specific type of drawing under a category
type DrawingType struct {
	Model
	CategoryID  string `json:"categoryId" bson:"categoryId"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Order       int    `json:"order" bson:"order"`
	IsActive    bool   `json:"isActive" bson:"isActive"`
}

// DrawingCategoryWithTypes represents a category with its associated drawing types
type DrawingCategoryWithTypes struct {
	DrawingCategory
	Types []DrawingType `json:"types" bson:"types"`
}

