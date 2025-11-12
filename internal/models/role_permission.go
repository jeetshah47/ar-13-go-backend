package models

import "time"

// RolePermission represents a role-permission mapping in the database
type RolePermission struct {
	ID         string    `json:"id" dynamodbav:"id"`
	Role       UserRole  `json:"role" dynamodbav:"role"`
	Permission string    `json:"permission" dynamodbav:"permission"`
	CreatedAt  time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}

