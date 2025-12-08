package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/jwt"
	"github.com/ar-13-go-backend/pkg/password"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo             *repos.UserRepo
	notificationRepo     *repos.NotificationRepo
	signupInvitationRepo *repos.SignupInvitationRepo
	config               *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:             repos.NewUserRepo(),
		notificationRepo:     repos.NewNotificationRepo(),
		signupInvitationRepo: repos.NewSignupInvitationRepo(),
		config:               cfg,
	}
}

// LoginResponse represents login response with tokens
type LoginResponse struct {
	AccessToken        string `json:"accessToken"`
	RefreshToken       string `json:"refreshToken"`
	ExpiresIn          int    `json:"expiresIn"` // in seconds
	ForceChangePassword bool  `json:"forceChangePassword,omitempty"` // True if user must change password
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
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		ExpiresIn:           expirationHours * 3600, // convert hours to seconds
		ForceChangePassword: user.ForceChangePassword,
	}, nil
}

// Logout logs out a user
func (s *AuthService) Logout(ctx context.Context, tokenString string, timeTrackingSvc *TimeTrackingService) error {
	// Verify token to get user info
	claims, err := jwt.VerifyToken(tokenString)
	if err != nil {
		return err
	}

	userID := claims.UserID

	// Stop all active time tracking sessions for the user
	if timeTrackingSvc != nil {
		if err := timeTrackingSvc.StopAllUserSessions(ctx, userID); err != nil {
			// Log error but don't fail logout
			log.Printf("Failed to stop time tracking sessions on logout: %v", err)
		}
	}

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

// ValidateSignupToken validates a signup invitation token
func (s *AuthService) ValidateSignupToken(ctx context.Context, token string) (*models.SignupInvitation, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	// Parse and verify the JWT token
	parsedToken, err := jwtv4.ParseWithClaims(token, jwtv4.MapClaims{}, func(token *jwtv4.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv4.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if s.config == nil || s.config.JWTSecret == "" {
			return nil, errors.New("JWT secret not configured")
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwtv4.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Verify token type
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "signup" {
		return nil, errors.New("invalid token type")
	}

	// Get email from token
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("invalid token: email not found")
	}

	// Get invitation from database by token
	invitation, err := s.signupInvitationRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	// Verify email matches
	if invitation.Email != email {
		return nil, errors.New("token email does not match invitation email")
	}

	// Check if already used
	if invitation.HasSignup {
		return nil, errors.New("invitation has already been used")
	}

	// Check if expired
	if time.Now().After(invitation.LinkExpiry) {
		return nil, errors.New("invitation has expired")
	}

	return invitation, nil
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*LoginResponse, error) {
	// Validate signup token
	invitation, err := s.ValidateSignupToken(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid signup token: %w", err)
	}

	// Verify email matches the invitation
	if invitation.Email != req.Email {
		return nil, errors.New("email does not match the invitation")
	}

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

	// Mark invitation as used
	if err := s.signupInvitationRepo.MarkAsSignedUp(ctx, invitation.ID); err != nil {
		// Log error but don't fail registration
		// The user is already created, so we just log the error
		fmt.Printf("Warning: failed to mark invitation as used: %v\n", err)
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
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		ExpiresIn:           expirationHours * 3600,
		ForceChangePassword: user.ForceChangePassword,
	}, nil
}