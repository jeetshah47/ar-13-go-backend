package handlers

import (
	"strings"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication routes
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(cfg),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get config from context or use a default (this is a workaround - ideally config should be passed)
	// For now, we'll create a service without email for registration
	userService := services.NewUserService(nil)
	if err := userService.Add(c.Request.Context(), &user); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"token": token})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ValidateSignupToken validates a signup invitation token
func (h *AuthHandler) ValidateSignupToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"valid": false, "reason": "Missing token"})
		return
	}

	// TODO: Implement token validation
	c.JSON(constants.StatusOK, gin.H{"valid": true})
}
