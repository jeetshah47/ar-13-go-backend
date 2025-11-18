package models

import "time"

// RolePermission represents a role-permission mapping in the database
type RolePermission struct {
	ID         string    `json:"id" bson:"id"`
	Role       UserRole  `json:"role" bson:"role"`
	Permission string    `json:"permission" bson:"permission"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" bson:"updatedAt"`
}

