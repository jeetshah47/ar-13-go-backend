package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/dynamodb"
	"github.com/ar-13-go-backend/pkg/password"
	"github.com/google/uuid"
)

func main() {
	// Parse command-line flags
	name := flag.String("name", "", "Admin user's full name (required)")
	email := flag.String("email", "", "Admin user's email address (required)")
	pass := flag.String("password", "", "Admin user's password (required, min 6 characters)")
	phone := flag.String("phone", "", "Admin user's phone number (optional)")
	designation := flag.String("designation", "", "Admin user's designation (optional)")
	flag.Parse()

	// Validate required fields
	if *name == "" || *email == "" || *pass == "" {
		fmt.Println("Error: Missing required fields")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		fmt.Println("\nExample:")
		fmt.Println("  go run scripts/add_admin_user.go -name \"Admin User\" -email admin@example.com -password \"SecurePass123\" -phone \"+1234567890\" -designation \"System Administrator\"")
		os.Exit(1)
	}

	// Validate password length
	if len(*pass) < 6 {
		log.Fatal("Error: Password must be at least 6 characters long")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize DynamoDB
	region := cfg.AWSRegion
	if region == "" {
		region = "us-east-1"
	}

	_, err = dynamodb.InitializeDynamoDB(region)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB: %v", err)
	}
	fmt.Println("✅ DynamoDB initialized successfully")

	// Create user repository
	userRepo := repos.NewUserRepo()
	ctx := context.Background()

	// Check if user already exists
	existingUser, err := userRepo.GetByEmail(ctx, *email)
	if err != nil {
		log.Fatalf("Failed to check if user exists: %v", err)
	}
	if existingUser != nil {
		log.Fatalf("Error: User with email %s already exists (ID: %s)", *email, existingUser.ID)
	}

	// Hash password
	hashedPassword, err := password.HashPassword(*pass)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	user := &models.User{
		ID:          uuid.New().String(),
		Name:        *name,
		Email:       *email,
		Password:    hashedPassword,
		PhoneNumber: *phone,
		Role:        models.UserRoleAdmin,
		CreatedAt:   time.Now(),
	}

	// Set designation if provided
	if *designation != "" {
		user.Designation = designation
	}

	// Add user to database
	if err := userRepo.Add(ctx, user); err != nil {
		log.Fatalf("Failed to add admin user to database: %v", err)
	}

	// Success message
	fmt.Println("\n✅ Admin user created successfully!")
	fmt.Println("\nUser Details:")
	fmt.Printf("  ID:          %s\n", user.ID)
	fmt.Printf("  Name:        %s\n", user.Name)
	fmt.Printf("  Email:       %s\n", user.Email)
	fmt.Printf("  Phone:       %s\n", user.PhoneNumber)
	if user.Designation != nil {
		fmt.Printf("  Designation: %s\n", *user.Designation)
	}
	fmt.Printf("  Role:        %s\n", user.Role)
	fmt.Printf("  Created At:  %s\n", user.CreatedAt.Format(time.RFC3339))
	fmt.Println("\nYou can now login with this admin account.")
}
