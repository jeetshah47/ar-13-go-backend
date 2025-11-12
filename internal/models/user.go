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
	ID          string    `json:"id" firestore:"id" dynamodbav:"id"`
	Name        string    `json:"name" firestore:"name" dynamodbav:"name"`
	Email       string    `json:"email" firestore:"email" dynamodbav:"email"`
	PhoneNumber string    `json:"phoneNumber" firestore:"phoneNumber" dynamodbav:"phoneNumber"`
	Role        UserRole  `json:"role" firestore:"role" dynamodbav:"role"`
	Password    string    `json:"-" firestore:"password" dynamodbav:"password"` // Hidden from JSON
	Designation *string   `json:"designation,omitempty" firestore:"designation,omitempty" dynamodbav:"designation,omitempty"`
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" firestore:"updatedAt" dynamodbav:"updatedAt"`
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
