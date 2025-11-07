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
	ID          string    `json:"id" firestore:"id"`
	Name        string    `json:"name" firestore:"name"`
	Email       string    `json:"email" firestore:"email"`
	PhoneNumber string    `json:"phoneNumber" firestore:"phoneNumber"`
	Role        UserRole  `json:"role" firestore:"role"`
	Password    string    `json:"-" firestore:"password"` // Hidden from JSON
	Designation *string   `json:"designation,omitempty" firestore:"designation,omitempty"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
