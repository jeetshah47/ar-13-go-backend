package handlers

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// GoogleAccountHandler handles Google account routes
type GoogleAccountHandler struct {
	googleAccountService *services.GoogleAccountService
	cfg                  *config.Config
}

// NewGoogleAccountHandler creates a new Google account handler with dependency injection
func NewGoogleAccountHandler(googleAccountService *services.GoogleAccountService, cfg *config.Config) *GoogleAccountHandler {
	return &GoogleAccountHandler{
		googleAccountService: googleAccountService,
		cfg:                  cfg,
	}
}

// NewGoogleAccountHandlerWithDefaults creates a new Google account handler with default dependencies
func NewGoogleAccountHandlerWithDefaults() *GoogleAccountHandler {
	return NewGoogleAccountHandler(
		services.NewGoogleAccountServiceWithDefaults(),
		config.AppConfig,
	)
}

// LinkGoogleAccount links a Google account
func (h *GoogleAccountHandler) LinkGoogleAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		GoogleIDToken string `json:"googleIdToken"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "Google ID token is required"})
		return
	}

	link, err := h.googleAccountService.LinkGoogleAccount(c.Request.Context(), userID, req.GoogleIDToken)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message": "Google account linked successfully",
		"link":    link,
	})
}

// UnlinkGoogleAccount unlinks a Google account
func (h *GoogleAccountHandler) UnlinkGoogleAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.googleAccountService.UnlinkGoogleAccount(c.Request.Context(), userID); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Google account unlinked successfully"})
}

// GetGoogleAccountStatus gets Google account status
func (h *GoogleAccountHandler) GetGoogleAccountStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	link, err := h.googleAccountService.GetGoogleAccountStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"linked": link != nil,
		"link":   link,
	})
}

// GetAllLinkedAccounts gets all linked accounts
func (h *GoogleAccountHandler) GetAllLinkedAccounts(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	links, err := h.googleAccountService.GetAllLinkedAccounts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"links": links})
}

// InitiateGoogleOAuth initiates OAuth flow
func (h *GoogleAccountHandler) InitiateGoogleOAuth(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	callbackURL := c.Query("callbackUrl")
	if callbackURL == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		callbackURL = scheme + "://" + c.Request.Host + "/api/google-account/auth/callback"
	}

	authURL, err := h.googleAccountService.InitiateGoogleOAuth(userID, callbackURL)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate OAuth URL: %s", err.Error())})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"authUrl": authURL,
		"message": "OAuth URL generated successfully",
	})
}

// HandleGoogleOAuthCallback handles OAuth callback
func (h *GoogleAccountHandler) HandleGoogleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	oauthError := c.Query("error")

	if oauthError != "" {
		frontendURL := h.cfg.FrontendURL
		errorURL := frontendURL + "/account/linked?success=false&error=" + url.QueryEscape(oauthError)
		c.Redirect(302, errorURL)
		return
	}

	if code == "" || state == "" {
		frontendURL := h.cfg.FrontendURL
		errorURL := frontendURL + "/account/linked?success=false&error=" + url.QueryEscape("Missing authorization code or state")
		c.Redirect(302, errorURL)
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	callbackURL := scheme + "://" + c.Request.Host + "/api/google-account/auth/callback"

	link, err := h.googleAccountService.HandleGoogleOAuthCallback(c.Request.Context(), code, state, callbackURL)
	if err != nil {
		frontendURL := h.cfg.FrontendURL
		errorMsg := err.Error()
		errorURL := frontendURL + "/account/linked?success=false&error=" + url.QueryEscape(errorMsg)
		c.Redirect(302, errorURL)
		return
	}

	frontendURL := h.cfg.FrontendURL
	successURL := frontendURL + "/account/linked?success=true&userId=" + link.UserID
	c.Redirect(302, successURL)
}

// GetGoogleCalendarEvents fetches Google Calendar events for a time range
func (h *GoogleAccountHandler) GetGoogleCalendarEvents(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse query parameters for time range
	timeMinStr := c.Query("timeMin")
	timeMaxStr := c.Query("timeMax")

	var timeMin, timeMax time.Time
	var err error

	if timeMinStr == "" {
		// Default to start of current month
		now := time.Now()
		timeMin = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else {
		timeMin, err = time.Parse(time.RFC3339, timeMinStr)
		if err != nil {
			c.JSON(constants.StatusBadRequest, gin.H{"error": "Invalid timeMin format. Use RFC3339 format (e.g., 2024-01-01T00:00:00Z)"})
			return
		}
	}

	if timeMaxStr == "" {
		// Default to end of current month
		now := time.Now()
		timeMax = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, time.UTC)
	} else {
		timeMax, err = time.Parse(time.RFC3339, timeMaxStr)
		if err != nil {
			c.JSON(constants.StatusBadRequest, gin.H{"error": "Invalid timeMax format. Use RFC3339 format (e.g., 2024-01-31T23:59:59Z)"})
			return
		}
	}

	events, err := h.googleAccountService.GetGoogleCalendarEvents(c.Request.Context(), userID, timeMin, timeMax)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}
