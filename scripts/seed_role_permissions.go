package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/dynamodb"
	"github.com/google/uuid"
)

func main() {
	// Initialize DynamoDB
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	_, err := dynamodb.InitializeDynamoDB(region)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB: %v", err)
	}

	ctx := context.Background()
	repo := repos.NewRolePermissionRepo()

	// Check if permissions already exist
	existingAdmin, _ := repo.GetByRole(ctx, models.UserRoleAdmin)
	existingStandard, _ := repo.GetByRole(ctx, models.UserRoleStandard)

	if len(existingAdmin) > 0 || len(existingStandard) > 0 {
		fmt.Println("Role permissions already exist in database. Skipping seed.")
		fmt.Printf("Admin permissions: %d\n", len(existingAdmin))
		fmt.Printf("Standard permissions: %d\n", len(existingStandard))
		return
	}

	// Admin permissions
	adminPermissions := []string{
		"projects:read", "projects:write", "projects:delete",
		"tasks:read", "tasks:write", "tasks:delete", "tasks:assign",
		"users:read", "users:write", "users:delete", "users:profile", "users:invite",
		"calendar:read", "calendar:write", "calendar:delete",
		"vacation:read", "vacation:write", "vacation:delete", "vacation:approve",
		"notifications:read", "notifications:write", "notifications:delete",
		"activityLogs:read",
		"dashboard:read",
		"employees:read",
		"infoPortal:read", "infoPortal:write", "infoPortal:delete",
		"googleAccount:read", "googleAccount:write", "googleAccount:link", "googleAccount:unlink",
		"websocket:connect",
		"auth:read",
		"projectDetails:read", "projectDetails:write", "projectDetails:delete",
		"backup:read", "backup:write",
	}

	// Standard permissions
	standardPermissions := []string{
		"projects:read",
		"tasks:read", "tasks:write", "tasks:delete",
		"calendar:read", "calendar:write", "calendar:delete",
		"vacation:read", "vacation:write", "vacation:delete",
		"notifications:read", "notifications:write", "notifications:delete",
		"activityLogs:read",
		"dashboard:read",
		"employees:read",
		"infoPortal:read", "infoPortal:write", "infoPortal:delete",
		"googleAccount:read", "googleAccount:write", "googleAccount:link", "googleAccount:unlink",
		"websocket:connect",
		"auth:read",
		"users:profile",
		"projectDetails:read",
	}

	// Create role permissions
	var allPermissions []models.RolePermission

	// Add admin permissions
	for _, perm := range adminPermissions {
		allPermissions = append(allPermissions, models.RolePermission{
			ID:         uuid.New().String(),
			Role:       models.UserRoleAdmin,
			Permission: perm,
		})
	}

	// Add standard permissions
	for _, perm := range standardPermissions {
		allPermissions = append(allPermissions, models.RolePermission{
			ID:         uuid.New().String(),
			Role:       models.UserRoleStandard,
			Permission: perm,
		})
	}

	// Batch insert
	if err := repo.BatchAdd(ctx, allPermissions); err != nil {
		log.Fatalf("Failed to seed role permissions: %v", err)
	}

	fmt.Printf("Successfully seeded %d role permissions:\n", len(allPermissions))
	fmt.Printf("  - Admin: %d permissions\n", len(adminPermissions))
	fmt.Printf("  - Standard: %d permissions\n", len(standardPermissions))
}

