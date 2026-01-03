package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// FileBrowserServiceClient handles communication with the new filebrowser service (simple Go service)
type FileBrowserServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewFileBrowserServiceClient creates a new filebrowser service client
func NewFileBrowserServiceClient(baseURL string) *FileBrowserServiceClient {
	return &FileBrowserServiceClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
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
// jwtToken: JWT token for authentication (from Authorization header of the original request)
func (c *FileBrowserServiceClient) ListFiles(path string, jwtToken string) (*FileBrowserBrowseResponse, error) {
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

	// Add JWT token in Authorization header
	if jwtToken != "" {
		// Trim any whitespace from token
		cleanToken := strings.TrimSpace(jwtToken)
		req.Header.Set("Authorization", "Bearer "+cleanToken)
	} else {
		// Log warning if token is empty (for debugging)
		fmt.Printf("[FileBrowserClient] Warning: JWT token is empty for request to %s\n", apiURL)
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
// jwtToken: JWT token for authentication (from Authorization header of the original request)
func (c *FileBrowserServiceClient) GetFileInfo(path string, jwtToken string) (*FileBrowserFileInfoResponse, error) {
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

	// Add JWT token in Authorization header
	if jwtToken != "" {
		// Trim any whitespace from token
		cleanToken := strings.TrimSpace(jwtToken)
		req.Header.Set("Authorization", "Bearer "+cleanToken)
	} else {
		// Log warning if token is empty (for debugging)
		fmt.Printf("[FileBrowserClient] Warning: JWT token is empty for request to %s\n", apiURL)
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

// UploadFile uploads a file to the filebrowser service
// jwtToken: JWT token for authentication (from Authorization header of the original request)
// targetPath: Target directory path (e.g., "/folder")
// filename: Name of the file to upload
// fileData: Reader containing the file data
// fileSize: Size of the file in bytes
func (c *FileBrowserServiceClient) UploadFile(targetPath string, filename string, fileData io.Reader, fileSize int64, jwtToken string) error {
	// Normalize target path
	if targetPath == "" {
		targetPath = "/"
	}
	if targetPath != "/" {
		targetPath = strings.Trim(targetPath, "/")
		targetPath = "/" + targetPath
	}

	// Build API URL with path query parameter
	apiURL := fmt.Sprintf("%s/api/upload?path=%s", c.baseURL, url.QueryEscape(targetPath))

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file field
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file data to form
	if _, err := io.Copy(part, fileData); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	// Close writer to finalize multipart form
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add JWT token in Authorization header
	if jwtToken != "" {
		cleanToken := strings.TrimSpace(jwtToken)
		req.Header.Set("Authorization", "Bearer "+cleanToken)
	} else {
		fmt.Printf("[FileBrowserClient] Warning: JWT token is empty for upload request to %s\n", apiURL)
	}

	// Set content type with boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// CheckHealth checks if the filebrowser service is healthy
// Note: Health endpoint doesn't require authentication
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
