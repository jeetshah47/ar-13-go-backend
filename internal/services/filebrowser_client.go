package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

// FileBrowserClient handles communication with FileBrowser API
type FileBrowserClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewFileBrowserClient creates a new FileBrowser client
func NewFileBrowserClient(baseURL, token string) *FileBrowserClient {
	return &FileBrowserClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FileBrowserItem represents a file or folder in FileBrowser
type FileBrowserItem struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Extension string `json:"extension"`
	Modified  string `json:"modified"`
	Mode      int    `json:"mode"`
	IsDir     bool   `json:"isDir"`
	IsSymlink bool   `json:"isSymlink"`
	Type      string `json:"type"`
}

// FileBrowserListResponse represents the response from list files API
type FileBrowserListResponse struct {
	Items    []FileBrowserItem `json:"items"`
	NumDirs  int               `json:"numDirs"`
	NumFiles int               `json:"numFiles"`
}

// ListFiles lists files and folders at the given path
func (c *FileBrowserClient) ListFiles(path string) (*FileBrowserListResponse, error) {
	if path == "" {
		path = "/"
	}
	
	// Ensure path starts with /
	if !filepath.IsAbs(path) && path[0] != '/' {
		path = "/" + path
	}
	
	encodedPath := url.QueryEscape(path)
	apiURL := fmt.Sprintf("%s/api/resources?path=%s", c.baseURL, encodedPath)
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	var result FileBrowserListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &result, nil
}

// GetFileURL returns the direct URL to access a file
func (c *FileBrowserClient) GetFileURL(filePath string) string {
	encodedPath := url.QueryEscape(filePath)
	return fmt.Sprintf("%s/api/raw?path=%s", c.baseURL, encodedPath)
}

// UploadFile uploads a file to FileBrowser
func (c *FileBrowserClient) UploadFile(destPath string, fileData io.Reader, filename string) error {
	dir := filepath.Dir(destPath)
	if dir == "." || dir == "" {
		dir = "/"
	}
	
	// Ensure dir starts with /
	if !filepath.IsAbs(dir) && dir[0] != '/' {
		dir = "/" + dir
	}
	
	encodedDir := url.QueryEscape(dir)
	apiURL := fmt.Sprintf("%s/api/resources?path=%s", c.baseURL, encodedDir)
	
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	
	if _, err := io.Copy(part, fileData); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}
	
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

// DeleteFile deletes a file or folder from FileBrowser
func (c *FileBrowserClient) DeleteFile(filePath string) error {
	encodedPath := url.QueryEscape(filePath)
	apiURL := fmt.Sprintf("%s/api/resources?path=%s", c.baseURL, encodedPath)
	
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

// FileExists checks if a file exists in FileBrowser
func (c *FileBrowserClient) FileExists(filePath string) (bool, error) {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		dir = "/"
	}
	
	result, err := c.ListFiles(dir)
	if err != nil {
		return false, err
	}
	
	// Normalize paths for comparison
	normalizedPath := filePath
	if !filepath.IsAbs(normalizedPath) {
		normalizedPath = "/" + normalizedPath
	}
	
	for _, item := range result.Items {
		if item.Path == normalizedPath {
			return true, nil
		}
	}
	
	return false, nil
}

