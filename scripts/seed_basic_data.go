// seed_basic_data.go - Seeds basic data for the application
//
// This script seeds:
//   - Role permissions for Admin and Standard roles
//   - An admin user account
//
// Usage:
//
//	go run scripts/seed_basic_data.go
//	go run scripts/seed_basic_data.go -name "John Doe" -email "admin@company.com" -password "SecurePass123"
//	go run scripts/seed_basic_data.go -skip-user  # Only seed permissions
//	go run scripts/seed_basic_data.go -skip-permissions  # Only create admin user
//
// Flags:
//
//	-name string        Admin user's full name (default: "Admin User")
//	-email string       Admin user's email address (default: "admin@example.com")
//	-password string    Admin user's password, min 6 characters (default: "admin123")
//	-phone string       Admin user's phone number (optional)
//	-designation string Admin user's designation (default: "System Administrator")
//	-skip-user          Skip creating admin user (only seed permissions)
//	-skip-permissions   Skip seeding permissions (only create admin user)
//
// Environment Variables Required:
//
//	MONGODB_URI         MongoDB connection string (default: "mongodb://localhost:27017")
//	MONGODB_DATABASE    MongoDB database name (default: "ar13_backend")
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/mongodb"
	passwordpkg "github.com/ar-13-go-backend/pkg/password"
	"github.com/google/uuid"
)

func main() {
	// Parse command-line flags for admin user
	name := flag.String("name", "Admin User", "Admin user's full name")
	email := flag.String("email", "admin@example.com", "Admin user's email address")
	pass := flag.String("password", "admin123", "Admin user's password (min 6 characters)")
	phone := flag.String("phone", "", "Admin user's phone number (optional)")
	designation := flag.String("designation", "System Administrator", "Admin user's designation (optional)")
	skipUser := flag.Bool("skip-user", false, "Skip creating admin user (only seed permissions)")
	skipPermissions := flag.Bool("skip-permissions", false, "Skip seeding permissions (only create admin user)")
	flag.Parse()

	// Validate password length
	if len(*pass) < 6 {
		log.Fatal("Error: Password must be at least 6 characters long")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB
	_, err = mongodb.InitializeMongoDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	fmt.Println("✅ MongoDB initialized successfully")

	ctx := context.Background()

	// Seed role permissions
	if !*skipPermissions {
		seedRolePermissions(ctx)
	}

	// Create admin user
	if !*skipUser {
		createAdminUser(ctx, *name, *email, *pass, *phone, *designation)
	}

	fmt.Println("\n✅ Basic data seeding completed successfully!")
}

func seedRolePermissions(ctx context.Context) {
	fmt.Println("\n📋 Seeding role permissions...")

	repo := repos.NewRolePermissionRepo()

	// Check if permissions already exist
	existingAdmin, _ := repo.GetByRole(ctx, models.UserRoleAdmin)
	existingStandard, _ := repo.GetByRole(ctx, models.UserRoleStandard)

	if len(existingAdmin) > 0 || len(existingStandard) > 0 {
		fmt.Println("⚠️  Role permissions already exist in database.")
		fmt.Printf("   Admin permissions: %d\n", len(existingAdmin))
		fmt.Printf("   Standard permissions: %d\n", len(existingStandard))
		fmt.Println("   Skipping permission seeding. Use -skip-permissions=false to force update.")
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
	now := time.Now()

	// Add admin permissions
	for _, perm := range adminPermissions {
		allPermissions = append(allPermissions, models.RolePermission{
			ID:         uuid.New().String(),
			Role:       models.UserRoleAdmin,
			Permission: perm,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	// Add standard permissions
	for _, perm := range standardPermissions {
		allPermissions = append(allPermissions, models.RolePermission{
			ID:         uuid.New().String(),
			Role:       models.UserRoleStandard,
			Permission: perm,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	// Batch insert
	if err := repo.BatchAdd(ctx, allPermissions); err != nil {
		log.Fatalf("❌ Failed to seed role permissions: %v", err)
	}

	fmt.Printf("✅ Successfully seeded %d role permissions:\n", len(allPermissions))
	fmt.Printf("   - Admin: %d permissions\n", len(adminPermissions))
	fmt.Printf("   - Standard: %d permissions\n", len(standardPermissions))
}

func createAdminUser(ctx context.Context, name, email, password, phone, designation string) {
	fmt.Println("\n👤 Creating admin user...")

	userRepo := repos.NewUserRepo()

	// Check if user already exists
	existingUser, err := userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Fatalf("❌ Failed to check if user exists: %v", err)
	}
	if existingUser != nil {
		fmt.Printf("⚠️  User with email %s already exists (ID: %s)\n", email, existingUser.ID)
		fmt.Println("   Skipping user creation.")
		return
	}

	// Hash password
	hashedPassword, err := passwordpkg.HashPassword(password)
	if err != nil {
		log.Fatalf("❌ Failed to hash password: %v", err)
	}

	// Create admin user
	now := time.Now()
	user := &models.User{
		ID:          uuid.New().String(),
		Name:        name,
		Email:       email,
		Password:    hashedPassword,
		PhoneNumber: phone,
		Role:        models.UserRoleAdmin,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Set designation if provided
	if designation != "" {
		user.Designation = &designation
	}

	// Add user to database
	if err := userRepo.Add(ctx, user); err != nil {
		log.Fatalf("❌ Failed to add admin user to database: %v", err)
	}

	// Success message
	fmt.Println("✅ Admin user created successfully!")
	fmt.Println("\nUser Details:")
	fmt.Printf("   ID:          %s\n", user.ID)
	fmt.Printf("   Name:        %s\n", user.Name)
	fmt.Printf("   Email:       %s\n", user.Email)
	if user.PhoneNumber != "" {
		fmt.Printf("   Phone:       %s\n", user.PhoneNumber)
	}
	if user.Designation != nil {
		fmt.Printf("   Designation: %s\n", *user.Designation)
	}
	fmt.Printf("   Role:        %s\n", user.Role)
	fmt.Printf("   Created At:  %s\n", user.CreatedAt.Format(time.RFC3339))
	fmt.Println("\n💡 You can now login with this admin account.")
}
