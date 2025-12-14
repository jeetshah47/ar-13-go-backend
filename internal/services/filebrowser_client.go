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
	"strings"
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

// normalizePath normalizes a file path to ensure it starts with / and is properly formatted
func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	// Normalize path separators
	path = strings.ReplaceAll(path, "\\", "/")
	// Ensure path starts with /
	if !filepath.IsAbs(path) && path[0] != '/' {
		path = "/" + path
	}
	return path
}

// validateName validates that a file or folder name doesn't contain path separators or dangerous characters
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return fmt.Errorf("name cannot contain path separators or '..'")
	}
	return nil
}

// buildPath builds a full path from a directory and a name
func buildPath(dir string, name string) string {
	dir = normalizePath(dir)
	// Ensure dir ends with / if it's not root
	if dir != "/" && !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	// Remove leading / from name if present (since we're joining)
	name = strings.TrimPrefix(name, "/")
	return normalizePath(dir + name)
}

// getParentDir extracts the parent directory from a file path
func getParentDir(filePath string) string {
	filePath = normalizePath(filePath)
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return "/"
	}
	return normalizePath(dir)
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
	path = normalizePath(path)
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
	filePath = normalizePath(filePath)
	encodedPath := url.QueryEscape(filePath)
	return fmt.Sprintf("%s/api/raw?path=%s", c.baseURL, encodedPath)
}

// UploadFile uploads a file to FileBrowser
func (c *FileBrowserClient) UploadFile(destPath string, fileData io.Reader, filename string) error {
	dir := getParentDir(destPath)
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
	filePath = normalizePath(filePath)
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
	filePath = normalizePath(filePath)
	dir := getParentDir(filePath)

	result, err := c.ListFiles(dir)
	if err != nil {
		return false, err
	}

	for _, item := range result.Items {
		if normalizePath(item.Path) == filePath {
			return true, nil
		}
	}

	return false, nil
}

// RenameFile renames a file or folder in FileBrowser
// FileBrowser doesn't have a direct rename endpoint, so we use move to the same directory with new name
func (c *FileBrowserClient) RenameFile(oldPath string, newName string) error {
	if oldPath == "" {
		return fmt.Errorf("old path is required")
	}
	if err := validateName(newName); err != nil {
		return err
	}

	oldPath = normalizePath(oldPath)
	dir := getParentDir(oldPath)
	newPath := buildPath(dir, newName)

	// Try using move operation to rename (move to same directory with new name)
	// This is the standard way to rename in FileBrowser
	encodedSourcePath := url.QueryEscape(oldPath)
	apiURL := fmt.Sprintf("%s/api/resources?path=%s", c.baseURL, encodedSourcePath)

	destPath := normalizePath(newPath)

	// Create request body with move action
	requestBody := map[string]interface{}{
		"action":      "move",
		"destination": destPath,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// FileBrowser uses POST for resource operations
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rename failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// CreateFolder creates a new folder in FileBrowser
func (c *FileBrowserClient) CreateFolder(parentPath string, folderName string) error {
	if err := validateName(folderName); err != nil {
		return fmt.Errorf("folder name validation failed: %w", err)
	}

	parentPath = normalizePath(parentPath)
	encodedPath := url.QueryEscape(parentPath)
	apiURL := fmt.Sprintf("%s/api/resources?path=%s", c.baseURL, encodedPath)

	// Create request body
	requestBody := map[string]string{
		"action": "mkdir",
		"name":   folderName,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create folder failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// MoveFile moves a file or folder to a new location in FileBrowser
func (c *FileBrowserClient) MoveFile(sourcePath string, destinationPath string) error {
	if sourcePath == "" {
		return fmt.Errorf("source path is required")
	}
	if destinationPath == "" {
		return fmt.Errorf("destination path is required")
	}

	sourcePath = normalizePath(sourcePath)
	destPath := normalizePath(destinationPath)

	encodedSourcePath := url.QueryEscape(sourcePath)
	apiURL := fmt.Sprintf("%s/api/resources?path=%s", c.baseURL, encodedSourcePath)

	// Create request body
	requestBody := map[string]interface{}{
		"action":      "move",
		"destination": destPath,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// FileBrowser uses POST for resource operations, not PATCH
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("move failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
