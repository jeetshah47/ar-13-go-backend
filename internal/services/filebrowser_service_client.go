package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// FileBrowserServiceClient handles communication with the new filebrowser service (simple Go service)
type FileBrowserServiceClient struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client
}

// NewFileBrowserServiceClient creates a new filebrowser service client
func NewFileBrowserServiceClient(baseURL, secretKey string) *FileBrowserServiceClient {
	return &FileBrowserServiceClient{
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		secretKey: secretKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FileBrowserFileItem represents a file or folder from the filebrowser service
type FileBrowserFileItem struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsFolder bool      `json:"isFolder"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	MimeType string    `json:"mimeType,omitempty"`
}

// FileBrowserBrowseResponse represents the response from the browse API
type FileBrowserBrowseResponse struct {
	Path  string                `json:"path"`
	Files []FileBrowserFileItem `json:"files"`
}

// FileBrowserFileInfoResponse represents the response from the file-info API
type FileBrowserFileInfoResponse struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsFolder bool      `json:"isFolder"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	MimeType string    `json:"mimeType,omitempty"`
}

// ListFiles lists files and folders at the given path
func (c *FileBrowserServiceClient) ListFiles(path string) (*FileBrowserBrowseResponse, error) {
	// Normalize path
	if path == "" {
		path = "/"
	}
	if path != "/" {
		path = strings.Trim(path, "/")
		path = "/" + path
	}

	apiURL := fmt.Sprintf("%s/api/browse?path=%s", c.baseURL, url.QueryEscape(path))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add secret key header if configured (trim whitespace)
	if c.secretKey != "" {
		req.Header.Set("X-API-Key", strings.TrimSpace(c.secretKey))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result FileBrowserBrowseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetFileInfo gets metadata for a specific file or folder
func (c *FileBrowserServiceClient) GetFileInfo(path string) (*FileBrowserFileInfoResponse, error) {
	// Normalize path
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	if path != "/" {
		path = strings.Trim(path, "/")
		path = "/" + path
	}

	apiURL := fmt.Sprintf("%s/api/file-info?path=%s", c.baseURL, url.QueryEscape(path))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add secret key header if configured (trim whitespace)
	if c.secretKey != "" {
		req.Header.Set("X-API-Key", strings.TrimSpace(c.secretKey))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result FileBrowserFileInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetFileDownloadURL returns the direct URL to download a file
func (c *FileBrowserServiceClient) GetFileDownloadURL(filePath string) string {
	// Normalize path
	if filePath != "/" {
		filePath = strings.Trim(filePath, "/")
		filePath = "/" + filePath
	}
	return fmt.Sprintf("%s/api/download?path=%s", c.baseURL, url.QueryEscape(filePath))
}

// CheckHealth checks if the filebrowser service is healthy
// Note: Health endpoint doesn't require authentication, so we don't send the secret key
func (c *FileBrowserServiceClient) CheckHealth() error {
	apiURL := fmt.Sprintf("%s/health", c.baseURL)
	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}
