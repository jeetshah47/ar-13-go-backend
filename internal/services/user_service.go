package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/email"
	"github.com/golang-jwt/jwt/v4"
)

// UserService handles user business logic
type UserService struct {
	userRepo             repos.UserRepository
	signupInvitationRepo repos.SignupInvitationRepository
	projectService       *ProjectService
	emailClient          EmailClientInterface
	config               *config.Config
}

// NewUserService creates a new user service with dependency injection
func NewUserService(
	userRepo repos.UserRepository,
	signupInvitationRepo repos.SignupInvitationRepository,
	projectService *ProjectService,
	emailClient EmailClientInterface,
	cfg *config.Config,
) *UserService {
	return &UserService{
		userRepo:             userRepo,
		signupInvitationRepo: signupInvitationRepo,
		projectService:       projectService,
		emailClient:          emailClient,
		config:               cfg,
	}
}

// NewUserServiceWithDefaults creates a new user service with default dependencies
func NewUserServiceWithDefaults(cfg *config.Config) *UserService {
	var emailClient EmailClientInterface
	if cfg != nil {
		emailClient = email.NewClient(cfg)
	}
	return NewUserService(
		repos.NewUserRepo(),
		repos.NewSignupInvitationRepo(),
		NewProjectServiceWithDefaults(),
		emailClient,
		cfg,
	)
}

// CreateInvitationRequest represents a signup invitation request
type CreateInvitationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// GetAll gets all users
func (s *UserService) GetAll(ctx context.Context, limit *int) ([]models.User, error) {
	return s.userRepo.GetAll(ctx, limit)
}

// GetByID gets a user by ID
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetByEmail gets a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// Add creates a new user
func (s *UserService) Add(ctx context.Context, user *models.User) error {
	// Check if user already exists
	existing, _ := s.userRepo.GetByEmail(ctx, user.Email)
	if existing != nil {
		return errors.New("user with this email already exists")
	}

	return s.userRepo.Add(ctx, user)
}

// Update updates a user
func (s *UserService) Update(ctx context.Context, user *models.User) error {
	// Fetch existing user to merge updates
	existingUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("user not found")
	}

	// Merge only non-empty fields from the update request
	// This prevents empty strings from overwriting existing values
	if user.Name != "" {
		existingUser.Name = user.Name
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}
	if user.PhoneNumber != "" {
		existingUser.PhoneNumber = user.PhoneNumber
	}
	if user.Role != "" {
		existingUser.Role = user.Role
	}
	if user.Designation != nil {
		existingUser.Designation = user.Designation
	}
	if user.Password != "" {
		existingUser.Password = user.Password
	}

	return s.userRepo.Update(ctx, existingUser)
}

// Delete deletes a user
func (s *UserService) Delete(ctx context.Context, id string) error {
	// Check if user exists
	exists, err := s.userRepo.Persists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user not found")
	}

	return s.userRepo.Delete(ctx, id)
}

// GetProfileResponse represents the user profile response with projects
type GetProfileResponse struct {
	User     *models.User                `json:"user"`
	Projects map[string]interface{}      `json:"projects,omitempty"`
}

// GetProfile gets user profile with related data including projects
// Optimized to avoid N+1 queries by using batch operations
func (s *UserService) GetProfile(ctx context.Context, id string) (*GetProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	response := &GetProfileResponse{
		User:     user,
		Projects: make(map[string]interface{}),
	}

	// Get projects where user is owner or member (single query - no N+1)
	if s.projectService != nil {
		projects, err := s.projectService.GetByUserID(ctx, id)
		if err != nil {
			// Log error but don't fail the request - projects are optional
			log.Printf("Failed to fetch projects for user %s: %v", id, err)
		} else if len(projects) > 0 {
			// Batch fetch all tasks for all projects in a single query to avoid N+1
			projectIDs := make([]string, 0, len(projects))
			for _, project := range projects {
				projectIDs = append(projectIDs, project.ID)
			}

			// Batch fetch tasks for all projects at once (single query - no N+1)
			tasksByProject, err := s.projectService.GetTasksByProjectIDs(ctx, projectIDs)
			if err != nil {
				// Log error but continue - task counts are optional
				log.Printf("Failed to fetch tasks for projects: %v", err)
				tasksByProject = make(map[string][]models.Task)
			}

			// Convert projects to map format expected by frontend
			// Calculate task counts efficiently from pre-fetched tasks (in-memory, no DB calls)
			for _, project := range projects {
				tasks := tasksByProject[project.ID]
				allTasksCount := len(tasks)
				activeTasksCount := 0
				
				// Count active tasks (not completed or rejected) - lightweight in-memory calculation
				// Using direct string comparison for better performance
				for _, task := range tasks {
					status := strings.ToLower(strings.TrimSpace(task.Status))
					// Check for active statuses (not completed/rejected) - efficient comparison
					if status != "completed" && status != "rejected" {
						activeTasksCount++
					}
				}

				// Build project map efficiently
				response.Projects[project.ID] = map[string]interface{}{
					"id":               project.ID,
					"code":             project.Code,
					"name":             project.Title,
					"title":            project.Title,
					"created":          project.Created,
					"allTasksCount":    allTasksCount,
					"activeTasksCount": activeTasksCount,
					"priority":         "", // Projects don't have priority, but frontend expects it
				}
			}
		}
	}

	return response, nil
}

// generateSignupToken generates a secure signup token using JWT
func (s *UserService) generateSignupToken(email string) (string, error) {
	if s.config == nil || s.config.JWTSecret == "" {
		return "", errors.New("JWT secret not configured")
	}

	payload := jwt.MapClaims{
		"email": email,
		"type":  "signup",
		"exp":   time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// CreateInvitation creates a signup invitation and sends an email
func (s *UserService) CreateInvitation(ctx context.Context, req CreateInvitationRequest) error {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return errors.New("email already taken")
	}

	// Check if there's already a pending invitation
	existingInvitation, _ := s.signupInvitationRepo.GetByEmail(ctx, req.Email)
	if existingInvitation != nil && !existingInvitation.HasSignup {
		return errors.New("invitation already sent to this email")
	}

	// Generate token
	token, err := s.generateSignupToken(req.Email)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Set expiry date (7 days from now)
	linkExpiry := time.Now().Add(7 * 24 * time.Hour)

	// Create invitation
	invitation := &models.SignupInvitation{
		Email:      req.Email,
		Token:      token,
		LinkExpiry: linkExpiry,
		HasSignup:  false,
	}

	if err := s.signupInvitationRepo.Add(ctx, invitation); err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	// Build signup link
	frontendURL := "http://localhost:3000"
	if s.config != nil && s.config.FrontendURL != "" {
		frontendURL = s.config.FrontendURL
	}
	signupLink := fmt.Sprintf("%s/auth/register?token=%s", frontendURL, token)

	// Send signup email (non-blocking)
	if s.emailClient != nil {
		go func() {
			// Use email prefix as userName
			emailParts := strings.Split(req.Email, "@")
			userName := emailParts[0]

			signupData := email.SignupEmailData{
				UserEmail:  req.Email,
				UserName:   userName,
				SignupLink: signupLink,
			}

			if err := s.emailClient.SendSignupLinkEmail(signupData); err != nil {
				log.Printf("Failed to send signup email to %s: %v", req.Email, err)
			} else {
				log.Printf("Signup invitation sent to %s", req.Email)
			}
		}()
	}

	return nil
}
