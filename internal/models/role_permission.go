package models

import "time"

// RolePermission represents a role-permission mapping in the database
type RolePermission struct {
	ID         string    `json:"id" dynamodbav:"id" bson:"id"`
	Role       UserRole  `json:"role" dynamodbav:"role" bson:"role"`
	Permission string    `json:"permission" dynamodbav:"permission" bson:"permission"`
	CreatedAt  time.Time `json:"createdAt" dynamodbav:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" dynamodbav:"updatedAt" bson:"updatedAt"`
}

