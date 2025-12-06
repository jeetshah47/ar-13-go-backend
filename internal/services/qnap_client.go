package services

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// QNAPClient handles communication with QNAP File Station API
type QNAPClient struct {
	nasURL     string
	username   string
	password   string
	sid        string
	expiresAt  time.Time
	httpClient *http.Client
}

// QNAPAuthResponse represents the authentication response from QNAP
type QNAPAuthResponse struct {
	XMLName    xml.Name `xml:"QDocRoot"`
	AuthPassed int      `xml:"authPassed"` // 1 = success, 0 = failure
	SID        string   `xml:"sid"`
	ErrorValue int      `xml:"errorValue"`
	Username   string   `xml:"username"`
}

// NewQNAPClient creates a new QNAP API client
func NewQNAPClient(nasIP string, apiPort int, username, password string) *QNAPClient {
	nasURL := fmt.Sprintf("http://%s:%d", nasIP, apiPort)
	if apiPort == 443 || apiPort == 8443 {
		nasURL = fmt.Sprintf("https://%s:%d", nasIP, apiPort)
	}

	return &QNAPClient{
		nasURL:   nasURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate authenticates with QNAP and gets a session ID
func (c *QNAPClient) Authenticate() error {
	// QNAP requires base64 encoded password
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(c.password))

	// Build authentication URL
	authURL := fmt.Sprintf("%s/cgi-bin/authLogin.cgi", c.nasURL)
	params := url.Values{}
	params.Set("user", c.username)
	params.Set("pwd", encodedPassword)
	params.Set("service", "1") // File Station service

	fullURL := fmt.Sprintf("%s?%s", authURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate with QNAP: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	// QNAP returns XML response
	bodyStr := string(body)

	// First try to extract values directly from XML string (more reliable for CDATA)
	// Try both lowercase and different case variations
	authPassed := extractCDATAInt(bodyStr, "authPassed", 0)
	if authPassed == 0 {
		authPassed = extractCDATAInt(bodyStr, "auth_passed", 0)
	}

	// QNAP uses authSid tag, not sid
	sid := extractCDATAString(bodyStr, "authSid", "")
	if sid == "" {
		sid = extractCDATAString(bodyStr, "sid", "")
	}
	if sid == "" {
		sid = extractCDATAString(bodyStr, "SID", "")
	}
	if sid == "" {
		sid = extractCDATAString(bodyStr, "sessionId", "")
	}
	if sid == "" {
		sid = extractCDATAString(bodyStr, "session_id", "")
	}

	errorCode := extractCDATAInt(bodyStr, "errorValue", -1)
	if errorCode == -1 {
		errorCode = extractCDATAInt(bodyStr, "error_value", -1)
	}

	// If extraction failed, try XML unmarshaling as fallback
	if authPassed == 0 && sid == "" {
		var authResp QNAPAuthResponse
		if err := xml.Unmarshal([]byte(bodyStr), &authResp); err == nil {
			if authResp.AuthPassed > 0 {
				authPassed = authResp.AuthPassed
			}
			if authResp.SID != "" {
				sid = authResp.SID
			}
			if authResp.ErrorValue != 0 {
				errorCode = authResp.ErrorValue
			}
		}
	}

	if authPassed != 1 {
		return fmt.Errorf("QNAP authentication failed (error code: %d). Response: %s", errorCode, bodyStr[:min(500, len(bodyStr))])
	}

	if sid == "" {
		// Try to find SID in the response using a more flexible search
		// Search for any tag that might contain session ID
		sidPatterns := []*regexp.Regexp{
			regexp.MustCompile(`<sid><!\[CDATA\[([^\]]+)\]\]></sid>`),
			regexp.MustCompile(`<SID><!\[CDATA\[([^\]]+)\]\]></SID>`),
			regexp.MustCompile(`<sessionId><!\[CDATA\[([^\]]+)\]\]></sessionId>`),
			regexp.MustCompile(`<session_id><!\[CDATA\[([^\]]+)\]\]></session_id>`),
			regexp.MustCompile(`<authSid><!\[CDATA\[([^\]]+)\]\]></authSid>`),
			regexp.MustCompile(`<auth_sid><!\[CDATA\[([^\]]+)\]\]></auth_sid>`),
			regexp.MustCompile(`<sid>([^<]+)</sid>`),
			regexp.MustCompile(`<SID>([^<]+)</SID>`),
		}

		for _, pattern := range sidPatterns {
			matches := pattern.FindStringSubmatch(bodyStr)
			if len(matches) > 1 && matches[1] != "" {
				sid = matches[1]
				break
			}
		}

		if sid == "" {
			return fmt.Errorf("QNAP authentication succeeded but no SID returned. Response length: %d chars", len(bodyStr))
		}
	}

	c.sid = sid
	// QNAP sessions typically expire in 30 minutes, but we'll set it to 25 minutes for safety
	c.expiresAt = time.Now().Add(25 * time.Minute)

	return nil
}

// GetSessionID returns the current session ID, authenticating if needed
func (c *QNAPClient) GetSessionID() (string, error) {
	// Check if we have a valid session
	if c.sid != "" && time.Now().Before(c.expiresAt) {
		return c.sid, nil
	}

	// Authenticate to get a new session
	if err := c.Authenticate(); err != nil {
		return "", err
	}

	return c.sid, nil
}

// RefreshSession refreshes the QNAP session
func (c *QNAPClient) RefreshSession() error {
	c.sid = ""
	c.expiresAt = time.Time{}
	return c.Authenticate()
}

// IsSessionValid checks if the current session is still valid
func (c *QNAPClient) IsSessionValid() bool {
	return c.sid != "" && time.Now().Before(c.expiresAt)
}

// GetNASURL returns the NAS URL
func (c *QNAPClient) GetNASURL() string {
	return c.nasURL
}

// extractCDATAInt extracts integer value from CDATA section
func extractCDATAInt(xmlStr, tagName string, defaultValue int) int {
	// Pattern: <tagName><![CDATA[value]]></tagName>
	// Need to escape special regex characters in tagName
	escapedTag := regexp.QuoteMeta(tagName)
	pattern := fmt.Sprintf(`<%s><!\[CDATA\[(\d+)\]\]></%s>`, escapedTag, escapedTag)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(xmlStr)
	if len(matches) > 1 {
		var value int
		if _, err := fmt.Sscanf(matches[1], "%d", &value); err == nil {
			return value
		}
	}
	return defaultValue
}

// extractCDATAString extracts string value from CDATA section
func extractCDATAString(xmlStr, tagName string, defaultValue string) string {
	// Pattern: <tagName><![CDATA[value]]></tagName>
	// Need to escape special regex characters in tagName
	escapedTag := regexp.QuoteMeta(tagName)

	// Try multiple patterns with different whitespace handling
	patterns := []string{
		// Standard CDATA
		fmt.Sprintf(`<%s><!\[CDATA\[(.*?)\]\]></%s>`, escapedTag, escapedTag),
		// CDATA with whitespace
		fmt.Sprintf(`<%s>\s*<!\[CDATA\[(.*?)\]\]>\s*</%s>`, escapedTag, escapedTag),
		// CDATA with newlines
		fmt.Sprintf(`<%s>[\s\n]*<!\[CDATA\[(.*?)\]\]>[\s\n]*</%s>`, escapedTag, escapedTag),
		// Without CDATA
		fmt.Sprintf(`<%s>([^<]+)</%s>`, escapedTag, escapedTag),
		// Without CDATA with whitespace
		fmt.Sprintf(`<%s>\s*([^<]+)\s*</%s>`, escapedTag, escapedTag),
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(xmlStr)
		if len(matches) > 1 && matches[1] != "" {
			return matches[1]
		}
	}

	return defaultValue
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
