package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/firebase"
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

// FirebaseAuthResponse represents Firebase auth response
type FirebaseAuthResponse struct {
	IDToken      string `json:"idToken"`
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"`
	Registered   bool   `json:"registered"`
}

// Login logs in a user
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (string, error) {
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", s.config.FirebaseWebAPIKey)

	payload := map[string]interface{}{
		"email":             req.Email,
		"password":          req.Password,
		"returnSecureToken": true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("invalid credentials")
	}

	var authResp FirebaseAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", err
	}

	// Create login notification (non-blocking)
	go func() {
		// Get user ID from token
		authClient := firebase.GetFirebaseClient()
		token, err := authClient.VerifyIDToken(context.Background(), authResp.IDToken)
		if err == nil && token != nil {
			userID := token.UID
			_ = s.notificationRepo.Add(context.Background(), &models.Notification{
				Title:             "Login Successful",
				Message:           fmt.Sprintf("You logged in at %s", time.Now().Format("2006-01-02 15:04:05")),
				Type:              models.NotificationTypeUserLogin,
				UserID:            userID,
				RelatedEntityID:   userID,
				RelatedEntityType: models.RelatedEntityTypeUser,
				IsRead:            false,
			})
		}
	}()

	return authResp.IDToken, nil
}

// Logout logs out a user
func (s *AuthService) Logout(ctx context.Context, idToken string) error {
	// Verify token to get user info
	authClient := firebase.GetFirebaseClient()
	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return err
	}

	userID := token.UID

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
