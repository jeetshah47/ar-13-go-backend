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
	var req struct {
		Name        string `json:"name" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=6"`
		PhoneNumber string `json:"phoneNumber"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registerReq := models.RegisterRequest{
		Name:        req.Name,
		Email:       req.Email,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
	}

	response, err := h.authService.Register(c.Request.Context(), registerReq)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(constants.StatusOK, response)
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
