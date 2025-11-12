package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/ar-13-go-backend/pkg/password"
	"github.com/google/uuid"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo         *repos.UserRepo
	notificationRepo *repos.NotificationRepo
	config           *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:         repos.NewUserRepo(),
		notificationRepo: repos.NewNotificationRepo(),
		config:           cfg,
	}
}

// LoginResponse represents login response with tokens
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"` // in seconds
}

// Login logs in a user
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if password is empty (unmarshaling issue)
	if user.Password == "" {
		return nil, fmt.Errorf("user password not found in database")
	}

	// Verify password
	if !password.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	expirationHours := s.config.JWTExpiration
	if expirationHours == 0 {
		expirationHours = 24 // default 24 hours
	}

	accessToken, err := jwt.GenerateToken(user.ID, user.Email, string(user.Role), time.Duration(expirationHours)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate refresh token
	refreshExpirationDays := s.config.RefreshExpiration
	if refreshExpirationDays == 0 {
		refreshExpirationDays = 30 // default 30 days
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create login notification (non-blocking)
	go func() {
		_ = s.notificationRepo.Add(context.Background(), &models.Notification{
			Title:             "Login Successful",
			Message:           fmt.Sprintf("You logged in at %s", time.Now().Format("2006-01-02 15:04:05")),
			Type:              models.NotificationTypeUserLogin,
			UserID:            user.ID,
			RelatedEntityID:   user.ID,
			RelatedEntityType: models.RelatedEntityTypeUser,
			IsRead:            false,
		})
	}()

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expirationHours * 3600, // convert hours to seconds
	}, nil
}

// Logout logs out a user
func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	// Verify token to get user info
	claims, err := jwt.VerifyToken(tokenString)
	if err != nil {
		return err
	}

	userID := claims.UserID

	// Create logout notification (non-blocking)
	go func() {
		_ = s.notificationRepo.Add(context.Background(), &models.Notification{
			Title:             "Logout Successful",
			Message:           fmt.Sprintf("You logged out at %s", time.Now().Format("2006-01-02 15:04:05")),
			Type:              models.NotificationTypeUserLogout,
			UserID:            userID,
			RelatedEntityID:   userID,
			RelatedEntityType: models.RelatedEntityTypeUser,
			IsRead:            false,
		})
	}()

	return nil
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*LoginResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Email:       req.Email,
		Password:    hashedPassword,
		PhoneNumber: req.PhoneNumber,
		Role:        models.UserRoleStandard, // default role
		CreatedAt:   time.Now(),
	}

	if err := s.userRepo.Add(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	expirationHours := s.config.JWTExpiration
	if expirationHours == 0 {
		expirationHours = 24
	}

	accessToken, err := jwt.GenerateToken(user.ID, user.Email, string(user.Role), time.Duration(expirationHours)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expirationHours * 3600,
	}, nil
}
