package models

import "time"

// UserRole represents user roles
type UserRole string

const (
	UserRoleStandard UserRole = "Standard"
	UserRoleAdmin    UserRole = "Admin"
)

// User represents a user in the system
type User struct {
	ID          string    `json:"id" firestore:"id" bson:"id"`
	Name        string    `json:"name" firestore:"name" bson:"name"`
	Email       string    `json:"email" firestore:"email" bson:"email"`
	PhoneNumber string    `json:"phoneNumber" firestore:"phoneNumber" bson:"phoneNumber"`
	Role        UserRole  `json:"role" firestore:"role" bson:"role"`
	Password    string    `json:"-" firestore:"password" bson:"password"` // Hidden from JSON
	Designation *string   `json:"designation,omitempty" firestore:"designation,omitempty" bson:"designation,omitempty"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt" bson:"updatedAt"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	PhoneNumber string `json:"phoneNumber"`
	Token       string `json:"token" binding:"required"` // Signup invitation token
}
