package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// filebrowserServiceStorage implements StorageServiceInterface using the new filebrowser service
type filebrowserServiceStorage struct {
	client *FileBrowserServiceClient
}

// NewFileBrowserServiceStorage creates a new storage service using the filebrowser service
func NewFileBrowserServiceStorage(baseURL, secretKey string) StorageServiceInterface {
	return &filebrowserServiceStorage{
		client: NewFileBrowserServiceClient(baseURL, secretKey),
	}
}

// Initialize initializes the filebrowser service storage
func (s *filebrowserServiceStorage) Initialize() error {
	// Test connection by checking health
	if err := s.client.CheckHealth(); err != nil {
		return fmt.Errorf("failed to connect to filebrowser service: %w", err)
	}
	return nil
}

// IsInitialized checks if the storage service is initialized
func (s *filebrowserServiceStorage) IsInitialized() bool {
	return s.client != nil
}

// ListObjects lists files and folders at the given path prefix
func (s *filebrowserServiceStorage) ListObjects(ctx context.Context, prefix string) ([]StorageObject, error) {
	if s.client == nil {
		return nil, fmt.Errorf("storage service not initialized")
	}

	// Normalize prefix
	if prefix == "" {
		prefix = "/"
	}

	// Ensure prefix starts with /
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	// Remove trailing slash for directory listing (filebrowser service handles this)
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix == "" {
		prefix = "/"
	}

	result, err := s.client.ListFiles(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var objects []StorageObject
	for _, item := range result.Files {
		objects = append(objects, StorageObject{
			Name:         item.Name,
			Path:         item.Path,
			IsFolder:     item.IsFolder,
			Size:         item.Size,
			LastModified: item.Modified,
			ContentType:  item.MimeType,
		})
	}

	return objects, nil
}

// UploadFile uploads a file to filebrowser storage
// Note: The current filebrowser service doesn't support uploads via API
// This would need to be implemented in the filebrowser service or handled differently
func (s *filebrowserServiceStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	// The current filebrowser service doesn't support uploads
	// You would need to either:
	// 1. Add upload endpoint to the filebrowser service
	// 2. Use direct file system access
	// 3. Fall back to MinIO for uploads
	return fmt.Errorf("upload not supported by filebrowser service - use MinIO or add upload endpoint to filebrowser service")
}

// GetPresignedURL generates a URL for accessing a file
// Filebrowser service provides direct download URLs (no presigning needed)
func (s *filebrowserServiceStorage) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("storage service not initialized")
	}

	// Normalize object name
	objectName = filepath.Clean(objectName)
	if !strings.HasPrefix(objectName, "/") {
		objectName = "/" + objectName
	}

	return s.client.GetFileDownloadURL(objectName), nil
}

// DeleteObject deletes a file from filebrowser storage
// Note: The current filebrowser service doesn't support deletes via API
func (s *filebrowserServiceStorage) DeleteObject(ctx context.Context, objectName string) error {
	// The current filebrowser service doesn't support deletes
	return fmt.Errorf("delete not supported by filebrowser service - use MinIO or add delete endpoint to filebrowser service")
}

// ObjectExists checks if a file exists in filebrowser storage
func (s *filebrowserServiceStorage) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	if s.client == nil {
		return false, fmt.Errorf("storage service not initialized")
	}

	// Normalize object name
	objectName = filepath.Clean(objectName)
	if !strings.HasPrefix(objectName, "/") {
		objectName = "/" + objectName
	}

	// Get file info to check if it exists
	_, err := s.client.GetFileInfo(objectName)
	if err != nil {
		// Check if it's a "not found" error (404 status or error message contains not_found/not found)
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "404") || 
		   strings.Contains(errStr, "not found") || 
		   strings.Contains(errStr, "not_found") ||
		   strings.Contains(errStr, "does not exist") {
			return false, nil
		}
		// For authentication errors (401), return the error so it can be handled properly
		if strings.Contains(errStr, "401") || 
		   strings.Contains(errStr, "unauthorized") ||
		   strings.Contains(errStr, "invalid secret key") ||
		   strings.Contains(errStr, "secret key required") {
			return false, fmt.Errorf("authentication failed with filebrowser service: %w", err)
		}
		return false, err
	}

	return true, nil
}
