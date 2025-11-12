package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// GoogleAccountService handles Google account linking business logic
type GoogleAccountService struct {
	userAccountLinkRepo *repos.UserAccountLinkRepo
	userRepo            *repos.UserRepo
	cfg                 *config.Config
}

// NewGoogleAccountService creates a new Google account service
func NewGoogleAccountService() *GoogleAccountService {
	return &GoogleAccountService{
		userAccountLinkRepo: repos.NewUserAccountLinkRepo(),
		userRepo:            repos.NewUserRepo(),
		cfg:                 config.AppConfig,
	}
}

// BoolString is a custom type that can unmarshal both boolean and string values
type BoolString bool

// UnmarshalJSON implements json.Unmarshaler to handle both string and boolean values
func (b *BoolString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as boolean first
	var boolVal bool
	if err := json.Unmarshal(data, &boolVal); err == nil {
		*b = BoolString(boolVal)
		return nil
	}

	// Try to unmarshal as string
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		*b = BoolString(strVal == "true" || strVal == "True" || strVal == "TRUE" || strVal == "1")
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into BoolString", string(data))
}

// Bool returns the boolean value
func (b BoolString) Bool() bool {
	return bool(b)
}

// GoogleTokenInfo represents Google token information
type GoogleTokenInfo struct {
	Sub           string     `json:"sub"`
	Email         string     `json:"email"`
	EmailVerified BoolString `json:"email_verified"`
	Name          string     `json:"name,omitempty"`
	Picture       string     `json:"picture,omitempty"`
}

// GoogleOAuthTokenResponse represents OAuth token response
type GoogleOAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// GoogleCalendarEvent represents a Google Calendar event
type GoogleCalendarEvent struct {
	ID          string `json:"id,omitempty"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	Start       struct {
		DateTime string `json:"dateTime,omitempty"`
		Date     string `json:"date,omitempty"`
		TimeZone string `json:"timeZone,omitempty"`
	} `json:"start"`
	End struct {
		DateTime string `json:"dateTime,omitempty"`
		Date     string `json:"date,omitempty"`
		TimeZone string `json:"timeZone,omitempty"`
	} `json:"end"`
	Location  string `json:"location,omitempty"`
	Attendees []struct {
		Email       string `json:"email"`
		DisplayName string `json:"displayName,omitempty"`
	} `json:"attendees,omitempty"`
	ConferenceData *struct {
		CreateRequest *struct {
			RequestID             string `json:"requestId"`
			ConferenceSolutionKey *struct {
				Type string `json:"type"`
			} `json:"conferenceSolutionKey"`
		} `json:"createRequest,omitempty"`
		EntryPoints []struct {
			EntryPointType string `json:"entryPointType"`
			URI            string `json:"uri"`
			Label          string `json:"label,omitempty"`
		} `json:"entryPoints,omitempty"`
		ConferenceID       string `json:"conferenceId,omitempty"`
		ConferenceSolution struct {
			Key struct {
				Type string `json:"type"`
			} `json:"key"`
			Name    string `json:"name,omitempty"`
			IconURI string `json:"iconUri,omitempty"`
		} `json:"conferenceSolution,omitempty"`
	} `json:"conferenceData,omitempty"`
	HangoutLink string `json:"hangoutLink,omitempty"` // Google Meet link (deprecated, use conferenceData.entryPoints)
}

// GoogleCalendarEventsResponse represents Google Calendar events response
type GoogleCalendarEventsResponse struct {
	Items         []GoogleCalendarEvent `json:"items"`
	NextPageToken string                `json:"nextPageToken,omitempty"`
}

// VerifyGoogleIDToken verifies Google ID token
func (s *GoogleAccountService) VerifyGoogleIDToken(ctx context.Context, idToken string) (*GoogleTokenInfo, error) {
	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken))
	if err != nil {
		return nil, fmt.Errorf("failed to verify Google token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invalid Google token (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenInfo GoogleTokenInfo
	if err := decodeJSON(resp.Body, &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode token info: %w", err)
	}

	if tokenInfo.Sub == "" || tokenInfo.Email == "" {
		return nil, errors.New("invalid Google token: missing sub or email")
	}

	return &tokenInfo, nil
}

// LinkGoogleAccount links a Google account to a user
func (s *GoogleAccountService) LinkGoogleAccount(ctx context.Context, userID, googleIDToken string) (*models.UserAccountLink, error) {
	// Validate userID
	if userID == "" || userID == "user123" || strings.Contains(userID, "test") {
		return nil, errors.New("invalid userId format")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Verify Google token
	googleInfo, err := s.VerifyGoogleIDToken(ctx, googleIDToken)
	if err != nil {
		return nil, err
	}

	// Check if Google account already linked to another user
	existingLink, err := s.userAccountLinkRepo.GetByProvider(ctx, models.AccountProviderGoogle, googleInfo.Sub)
	if err != nil {
		return nil, err
	}
	if existingLink != nil && existingLink.UserID != userID && existingLink.IsActive {
		return nil, errors.New("this Google account is already linked to another user")
	}

	// Check if user already has Google account linked
	existingUserLink, err := s.userAccountLinkRepo.GetByUserIDAndProvider(ctx, userID, models.AccountProviderGoogle)
	if err != nil {
		return nil, err
	}
	if existingUserLink != nil {
		return nil, errors.New("user already has a Google account linked")
	}

	// Create account link
	link := &models.UserAccountLink{
		UserID:         userID,
		Provider:       models.AccountProviderGoogle,
		ProviderUserID: googleInfo.Sub,
		ProviderEmail:  googleInfo.Email,
		IsActive:       true,
	}
	if googleInfo.Name != "" {
		link.ProviderDisplayName = &googleInfo.Name
	}

	if err := s.userAccountLinkRepo.Add(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to link Google account: %w", err)
	}

	return link, nil
}

// UnlinkGoogleAccount unlinks a Google account from a user
func (s *GoogleAccountService) UnlinkGoogleAccount(ctx context.Context, userID string) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Get existing link
	existingLink, err := s.userAccountLinkRepo.GetByUserIDAndProvider(ctx, userID, models.AccountProviderGoogle)
	if err != nil {
		return err
	}
	if existingLink == nil {
		return errors.New("no Google account linked to this user")
	}

	// Deactivate link
	return s.userAccountLinkRepo.Deactivate(ctx, existingLink.ID)
}

// GetGoogleAccountStatus gets Google account link status
func (s *GoogleAccountService) GetGoogleAccountStatus(ctx context.Context, userID string) (*models.UserAccountLink, error) {
	return s.userAccountLinkRepo.GetByUserIDAndProvider(ctx, userID, models.AccountProviderGoogle)
}

// GetAllLinkedAccounts gets all linked accounts for a user
func (s *GoogleAccountService) GetAllLinkedAccounts(ctx context.Context, userID string) ([]models.UserAccountLink, error) {
	return s.userAccountLinkRepo.GetByUserID(ctx, userID)
}

// GenerateStateToken generates OAuth state token
func (s *GoogleAccountService) GenerateStateToken(userID string) (string, error) {
	if userID == "" || userID == "user123" || strings.Contains(userID, "test") {
		return "", errors.New("invalid userId")
	}

	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	randomHex := hex.EncodeToString(randomBytes)
	timestamp := time.Now().Unix()

	return fmt.Sprintf("%s:%d:%s", userID, timestamp, randomHex), nil
}

// VerifyStateToken verifies and extracts userId from state token
func (s *GoogleAccountService) VerifyStateToken(state string) (string, error) {
	parts := strings.Split(state, ":")
	if len(parts) != 3 {
		return "", errors.New("invalid state token format")
	}

	userID := parts[0]
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", errors.New("invalid state token timestamp")
	}

	// State token expires after 1 hour
	if time.Now().Unix()-timestamp > 3600 {
		return "", errors.New("state token expired")
	}

	return userID, nil
}

// validateGoogleClientID validates the Google OAuth Client ID format
func (s *GoogleAccountService) validateGoogleClientID() error {
	clientID := s.cfg.GoogleClientID
	if clientID == "" {
		return errors.New("GOOGLE_CLIENT_ID is not configured. Please set it in your environment variables")
	}

	// Check if client ID looks like a client secret (common mistake)
	if strings.HasPrefix(clientID, "GOCSPX-") {
		return errors.New("GOOGLE_CLIENT_ID appears to be a client secret (starts with GOCSPX-). Please use the actual Client ID from Google Cloud Console (usually ends with .apps.googleusercontent.com)")
	}

	// Google OAuth Client IDs typically end with .apps.googleusercontent.com
	// or are numeric IDs, but the .apps.googleusercontent.com format is most common
	if !strings.Contains(clientID, ".apps.googleusercontent.com") && !strings.HasPrefix(clientID, "GOCSPX-") {
		// Allow numeric IDs as well, but warn if it doesn't match expected patterns
		if len(clientID) < 10 {
			return errors.New("GOOGLE_CLIENT_ID format appears invalid. Expected format: <number>-<string>.apps.googleusercontent.com")
		}
	}

	return nil
}

// InitiateGoogleOAuth initiates OAuth flow
func (s *GoogleAccountService) InitiateGoogleOAuth(userID, callbackURL string) (string, error) {
	// Validate Google Client ID configuration
	if err := s.validateGoogleClientID(); err != nil {
		return "", fmt.Errorf("invalid Google OAuth configuration: %w", err)
	}

	state, err := s.GenerateStateToken(userID)
	if err != nil {
		return "", err
	}

	scope := strings.Join([]string{
		"openid",
		"email",
		"profile",
		"https://www.googleapis.com/auth/calendar.readonly",
		"https://www.googleapis.com/auth/calendar",
	}, " ")

	params := url.Values{}
	params.Set("client_id", s.cfg.GoogleClientID)
	params.Set("redirect_uri", callbackURL)
	params.Set("response_type", "code")
	params.Set("scope", scope)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	params.Set("state", state)

	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?%s", params.Encode()), nil
}

// HandleGoogleOAuthCallback handles OAuth callback
func (s *GoogleAccountService) HandleGoogleOAuthCallback(ctx context.Context, code, state, callbackURL string) (*models.UserAccountLink, error) {
	// Verify state token
	userID, err := s.VerifyStateToken(state)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired state token: %w", err)
	}

	// Validate userID
	if userID == "user123" || strings.Contains(userID, "test") || userID == "" {
		return nil, errors.New("invalid userId format")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Exchange code for tokens
	tokenResp, err := s.exchangeCodeForTokens(code, callbackURL)
	if err != nil {
		return nil, err
	}

	// Verify ID token
	_, err = s.VerifyGoogleIDToken(ctx, tokenResp.IDToken)
	if err != nil {
		return nil, err
	}

	// Link account
	link, err := s.LinkGoogleAccount(ctx, userID, tokenResp.IDToken)
	if err != nil {
		// If already linked, get existing link
		existingLink, getErr := s.userAccountLinkRepo.GetByUserIDAndProvider(ctx, userID, models.AccountProviderGoogle)
		if getErr != nil {
			return nil, err
		}
		if existingLink == nil {
			return nil, err
		}
		link = existingLink
	}

	// Update with tokens
	if tokenResp.AccessToken != "" || tokenResp.RefreshToken != "" {
		link.AccessToken = &tokenResp.AccessToken
		if tokenResp.RefreshToken != "" {
			link.RefreshToken = &tokenResp.RefreshToken
		}
		if tokenResp.ExpiresIn > 0 {
			expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
			link.ExpiresAt = &expiresAt
		}
		if err := s.userAccountLinkRepo.Update(ctx, link); err != nil {
			return nil, fmt.Errorf("failed to update tokens: %w", err)
		}
	}

	return link, nil
}

// exchangeCodeForTokens exchanges authorization code for tokens
func (s *GoogleAccountService) exchangeCodeForTokens(code, callbackURL string) (*GoogleOAuthTokenResponse, error) {
	// Validate Google Client ID configuration
	if err := s.validateGoogleClientID(); err != nil {
		return nil, fmt.Errorf("invalid Google OAuth configuration: %w", err)
	}

	if s.cfg.GoogleClientSecret == "" {
		return nil, errors.New("GOOGLE_CLIENT_SECRET is not configured. Please set it in your environment variables")
	}

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.cfg.GoogleClientID)
	data.Set("client_secret", s.cfg.GoogleClientSecret)
	data.Set("redirect_uri", callbackURL)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to exchange code for tokens")
	}

	var tokenResp GoogleOAuthTokenResponse
	if err := decodeJSON(resp.Body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// refreshAccessToken refreshes an expired access token using the refresh token
func (s *GoogleAccountService) refreshAccessToken(ctx context.Context, refreshToken string) (*GoogleOAuthTokenResponse, error) {
	if s.cfg.GoogleClientID == "" || s.cfg.GoogleClientSecret == "" {
		return nil, errors.New("Google OAuth credentials not configured")
	}

	data := url.Values{}
	data.Set("client_id", s.cfg.GoogleClientID)
	data.Set("client_secret", s.cfg.GoogleClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to refresh access token (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp GoogleOAuthTokenResponse
	if err := decodeJSON(resp.Body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// getValidAccessToken gets a valid access token, refreshing if necessary
func (s *GoogleAccountService) getValidAccessToken(ctx context.Context, userID string) (string, error) {
	link, err := s.userAccountLinkRepo.GetByUserIDAndProvider(ctx, userID, models.AccountProviderGoogle)
	if err != nil {
		return "", fmt.Errorf("failed to get Google account link: %w", err)
	}
	if link == nil {
		return "", errors.New("Google account not linked")
	}
	if !link.IsActive {
		return "", errors.New("Google account link is not active")
	}

	// Check if access token exists and is valid
	if link.AccessToken != nil && link.ExpiresAt != nil {
		// Token is still valid if expiresAt is in the future (with 5 minute buffer)
		if time.Now().Add(5 * time.Minute).Before(*link.ExpiresAt) {
			return *link.AccessToken, nil
		}
	}

	// Token expired or missing, refresh it
	if link.RefreshToken == nil {
		return "", errors.New("refresh token not available, please re-link Google account")
	}

	tokenResp, err := s.refreshAccessToken(ctx, *link.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update the link with new token
	link.AccessToken = &tokenResp.AccessToken
	if tokenResp.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		link.ExpiresAt = &expiresAt
	}
	if err := s.userAccountLinkRepo.Update(ctx, link); err != nil {
		return "", fmt.Errorf("failed to update access token: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// CreateGoogleCalendarEvent creates a Google Calendar event
func (s *GoogleAccountService) CreateGoogleCalendarEvent(ctx context.Context, userID string, event *GoogleCalendarEvent) (*GoogleCalendarEvent, error) {
	accessToken, err := s.getValidAccessToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build request payload as map to ensure correct JSON structure
	requestPayload := map[string]interface{}{
		"summary": event.Summary,
	}

	if event.Description != "" {
		requestPayload["description"] = event.Description
	}

	requestPayload["start"] = map[string]interface{}{
		"dateTime": event.Start.DateTime,
		"timeZone": event.Start.TimeZone,
	}

	requestPayload["end"] = map[string]interface{}{
		"dateTime": event.End.DateTime,
		"timeZone": event.End.TimeZone,
	}

	if event.Location != "" {
		requestPayload["location"] = event.Location
	}

	if len(event.Attendees) > 0 {
		attendees := make([]map[string]interface{}, len(event.Attendees))
		for i, attendee := range event.Attendees {
			attendees[i] = map[string]interface{}{
				"email": attendee.Email,
			}
			if attendee.DisplayName != "" {
				attendees[i]["displayName"] = attendee.DisplayName
			}
		}
		requestPayload["attendees"] = attendees
	}

	// Add conference data if present
	if event.ConferenceData != nil && event.ConferenceData.CreateRequest != nil {
		conferenceData := map[string]interface{}{
			"createRequest": map[string]interface{}{
				"requestId": event.ConferenceData.CreateRequest.RequestID,
				"conferenceSolutionKey": map[string]interface{}{
					"type": event.ConferenceData.CreateRequest.ConferenceSolutionKey.Type,
				},
			},
		}
		requestPayload["conferenceData"] = conferenceData
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	// Log conference data for debugging
	if event.ConferenceData != nil && event.ConferenceData.CreateRequest != nil {
		fmt.Printf("Creating Google Calendar event with conference data. RequestID: %s\n", event.ConferenceData.CreateRequest.RequestID)
		fmt.Printf("Request JSON: %s\n", string(jsonData))
	}

	// Build URL with conferenceDataVersion parameter if conference data is present
	// Also add sendUpdates to notify attendees
	apiURL := "https://www.googleapis.com/calendar/v3/calendars/primary/events"
	params := make(url.Values)
	if event.ConferenceData != nil && event.ConferenceData.CreateRequest != nil {
		params.Set("conferenceDataVersion", "1")
	}
	if len(event.Attendees) > 0 {
		params.Set("sendUpdates", "all") // Send invites to all attendees
	}
	if len(params) > 0 {
		apiURL += "?" + params.Encode()
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		errorMsg := fmt.Sprintf("failed to create calendar event (status %d): %s", resp.StatusCode, bodyStr)
		fmt.Printf("ERROR: %s\n", errorMsg)
		fmt.Printf("Request URL: %s\n", apiURL)
		fmt.Printf("Request Body: %s\n", string(jsonData))
		return nil, fmt.Errorf("failed to create calendar event (status %d): %s", resp.StatusCode, bodyStr)
	}

	var createdEvent GoogleCalendarEvent
	if err := decodeJSON(resp.Body, &createdEvent); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log conference data from response for debugging
	if createdEvent.ConferenceData != nil {
		if len(createdEvent.ConferenceData.EntryPoints) > 0 {
			fmt.Printf("Google Meet link created: %s\n", createdEvent.ConferenceData.EntryPoints[0].URI)
		}
		if createdEvent.HangoutLink != "" {
			fmt.Printf("Google Meet link (hangoutLink): %s\n", createdEvent.HangoutLink)
		}
	}

	return &createdEvent, nil
}

// UpdateGoogleCalendarEvent updates a Google Calendar event
func (s *GoogleAccountService) UpdateGoogleCalendarEvent(ctx context.Context, userID, eventID string, event *GoogleCalendarEvent) (*GoogleCalendarEvent, error) {
	accessToken, err := s.getValidAccessToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build request payload similar to CreateGoogleCalendarEvent to ensure attendees are included
	requestPayload := map[string]interface{}{
		"summary": event.Summary,
	}

	if event.Description != "" {
		requestPayload["description"] = event.Description
	}

	requestPayload["start"] = map[string]interface{}{
		"dateTime": event.Start.DateTime,
		"timeZone": event.Start.TimeZone,
	}

	requestPayload["end"] = map[string]interface{}{
		"dateTime": event.End.DateTime,
		"timeZone": event.End.TimeZone,
	}

	if event.Location != "" {
		requestPayload["location"] = event.Location
	}

	// Always include attendees if they exist (required for sendUpdates to work)
	if len(event.Attendees) > 0 {
		attendees := make([]map[string]interface{}, len(event.Attendees))
		for i, attendee := range event.Attendees {
			attendees[i] = map[string]interface{}{
				"email": attendee.Email,
			}
			if attendee.DisplayName != "" {
				attendees[i]["displayName"] = attendee.DisplayName
			}
		}
		requestPayload["attendees"] = attendees
	}

	// Add conference data if present
	if event.ConferenceData != nil && event.ConferenceData.CreateRequest != nil {
		conferenceData := map[string]interface{}{
			"createRequest": map[string]interface{}{
				"requestId": event.ConferenceData.CreateRequest.RequestID,
				"conferenceSolutionKey": map[string]interface{}{
					"type": event.ConferenceData.CreateRequest.ConferenceSolutionKey.Type,
				},
			},
		}
		requestPayload["conferenceData"] = conferenceData
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	// Build URL with sendUpdates parameter to notify attendees
	// Always send updates if there are attendees, so attendees are notified of any changes
	apiURL := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/primary/events/%s", eventID)
	params := make(url.Values)
	if len(event.Attendees) > 0 {
		params.Set("sendUpdates", "all") // Send updates to all attendees
	}
	if event.ConferenceData != nil && event.ConferenceData.CreateRequest != nil {
		params.Set("conferenceDataVersion", "1")
	}
	if len(params) > 0 {
		apiURL += "?" + params.Encode()
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "PUT", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update calendar event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update calendar event (status %d): %s", resp.StatusCode, string(body))
	}

	var updatedEvent GoogleCalendarEvent
	if err := decodeJSON(resp.Body, &updatedEvent); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedEvent, nil
}

// DeleteGoogleCalendarEvent deletes a Google Calendar event
func (s *GoogleAccountService) DeleteGoogleCalendarEvent(ctx context.Context, userID, eventID string) error {
	accessToken, err := s.getValidAccessToken(ctx, userID)
	if err != nil {
		return err
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/primary/events/%s", eventID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete calendar event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete calendar event (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetGoogleCalendarEvents fetches Google Calendar events for a time range
func (s *GoogleAccountService) GetGoogleCalendarEvents(ctx context.Context, userID string, timeMin, timeMax time.Time) ([]GoogleCalendarEvent, error) {
	accessToken, err := s.getValidAccessToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build query parameters
	params := url.Values{}
	params.Set("timeMin", timeMin.Format(time.RFC3339))
	params.Set("timeMax", timeMax.Format(time.RFC3339))
	params.Set("singleEvents", "true")
	params.Set("orderBy", "startTime")

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/primary/events?%s", params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendar events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch calendar events (status %d): %s", resp.StatusCode, string(body))
	}

	var eventsResp GoogleCalendarEventsResponse
	if err := decodeJSON(resp.Body, &eventsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return eventsResp.Items, nil
}

// getUserTimezone gets the user's timezone
// For now, using Indian timezone (Asia/Kolkata)
func (s *GoogleAccountService) getUserTimezone(ctx context.Context, userID string) (string, error) {
	// Using Indian Standard Time (IST) - Asia/Kolkata
	return "Asia/Kolkata", nil
}

// ConvertCalendarEventToGoogleEvent converts an internal CalendarEvent to GoogleCalendarEvent
// It uses Indian timezone (Asia/Kolkata) for timezone conversion
func (s *GoogleAccountService) ConvertCalendarEventToGoogleEvent(ctx context.Context, userID string, event *models.CalendarEvent) (*GoogleCalendarEvent, error) {
	googleEvent := &GoogleCalendarEvent{
		Summary: event.Title,
	}

	if event.Description != nil {
		googleEvent.Description = *event.Description
	}

	// Get user's timezone from Google Calendar
	timezone, err := s.getUserTimezone(ctx, userID)
	if err != nil {
		// If we can't get timezone, try to use the timezone from the time itself
		if event.Start.Location() != nil && event.Start.Location().String() != "UTC" {
			timezone = event.Start.Location().String()
		} else {
			timezone = "UTC"
		}
	}

	// Load the timezone location
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// If timezone is invalid, default to UTC
		loc = time.UTC
		timezone = "UTC"
	}

	// Convert event times to the user's timezone
	startInTZ := event.Start.In(loc)
	endInTZ := event.End.In(loc)
	if event.End.IsZero() {
		// Use duration if provided, otherwise default to 1 hour
		duration := 1 * time.Hour
		if event.Duration != nil && *event.Duration > 0 {
			duration = time.Duration(*event.Duration) * time.Minute
		}
		endInTZ = event.Start.Add(duration).In(loc)
	}

	// Format DateTime without timezone offset (Google Calendar expects this when TimeZone is specified)
	// Use format: "2006-01-02T15:04:05"
	googleEvent.Start.DateTime = startInTZ.Format("2006-01-02T15:04:05")
	googleEvent.Start.TimeZone = timezone

	googleEvent.End.DateTime = endInTZ.Format("2006-01-02T15:04:05")
	googleEvent.End.TimeZone = timezone

	// Add Google Meet conference if event type is online
	if event.EventType != nil && *event.EventType == models.EventTypeOnline {
		// Generate a unique request ID for the conference (must be unique per request)
		// Using UUID-like format: eventID + timestamp + random hex
		randomBytes := make([]byte, 4)
		rand.Read(randomBytes)
		randomHex := hex.EncodeToString(randomBytes)
		requestID := fmt.Sprintf("%s-%d-%s", event.ID, time.Now().UnixNano(), randomHex)

		// Create conference solution key
		conferenceSolutionKeyType := "hangoutsMeet"
		conferenceSolutionKey := &struct {
			Type string `json:"type"`
		}{
			Type: conferenceSolutionKeyType,
		}

		// Create request structure
		createRequest := &struct {
			RequestID             string `json:"requestId"`
			ConferenceSolutionKey *struct {
				Type string `json:"type"`
			} `json:"conferenceSolutionKey"`
		}{
			RequestID:             requestID,
			ConferenceSolutionKey: conferenceSolutionKey,
		}

		// Create conference data using the same type as GoogleCalendarEvent
		googleEvent.ConferenceData = &struct {
			CreateRequest *struct {
				RequestID             string `json:"requestId"`
				ConferenceSolutionKey *struct {
					Type string `json:"type"`
				} `json:"conferenceSolutionKey"`
			} `json:"createRequest,omitempty"`
			EntryPoints []struct {
				EntryPointType string `json:"entryPointType"`
				URI            string `json:"uri"`
				Label          string `json:"label,omitempty"`
			} `json:"entryPoints,omitempty"`
			ConferenceID       string `json:"conferenceId,omitempty"`
			ConferenceSolution struct {
				Key struct {
					Type string `json:"type"`
				} `json:"key"`
				Name    string `json:"name,omitempty"`
				IconURI string `json:"iconUri,omitempty"`
			} `json:"conferenceSolution,omitempty"`
		}{
			CreateRequest: createRequest,
		}
	}

	// Add attendees from invited member IDs and invites emails
	attendees := make([]struct {
		Email       string `json:"email"`
		DisplayName string `json:"displayName,omitempty"`
	}, 0)

	// Track emails to avoid duplicates
	emailSet := make(map[string]bool)

	// Get user emails from member IDs
	if len(event.InvitedMemberIds) > 0 {
		for _, memberID := range event.InvitedMemberIds {
			user, err := s.userRepo.GetByID(ctx, memberID)
			if err == nil && user != nil && !emailSet[user.Email] {
				emailSet[user.Email] = true
				attendees = append(attendees, struct {
					Email       string `json:"email"`
					DisplayName string `json:"displayName,omitempty"`
				}{
					Email:       user.Email,
					DisplayName: user.Name,
				})
			}
		}
	}

	// Add emails directly from invites field
	if len(event.Invites) > 0 {
		for _, email := range event.Invites {
			// Skip empty emails and duplicates
			if email != "" && !emailSet[email] {
				emailSet[email] = true
				attendees = append(attendees, struct {
					Email       string `json:"email"`
					DisplayName string `json:"displayName,omitempty"`
				}{
					Email: email,
					// No display name for direct email invites
				})
			}
		}
	}

	// Set attendees if we have any
	if len(attendees) > 0 {
		googleEvent.Attendees = attendees
	}

	return googleEvent, nil
}

// decodeJSON decodes JSON from reader
func decodeJSON(r io.Reader, v interface{}) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}
