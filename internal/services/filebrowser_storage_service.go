package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"
)

// filebrowserStorageService implements StorageServiceInterface using FileBrowser API
type filebrowserStorageService struct {
	client *FileBrowserClient
}

// NewFileBrowserStorageService creates a new FileBrowser storage service
func NewFileBrowserStorageService(baseURL, token string) StorageServiceInterface {
	return &filebrowserStorageService{
		client: NewFileBrowserClient(baseURL, token),
	}
}

// Initialize initializes the FileBrowser storage service
// FileBrowser doesn't require initialization like MinIO, but we verify connectivity
func (s *filebrowserStorageService) Initialize() error {
	// Test connection by listing root directory
	_, err := s.client.ListFiles("/")
	if err != nil {
		return fmt.Errorf("failed to connect to FileBrowser: %w", err)
	}
	return nil
}

// IsInitialized checks if the storage service is initialized
func (s *filebrowserStorageService) IsInitialized() bool {
	return s.client != nil
}

// ListObjects lists files and folders at the given path prefix
func (s *filebrowserStorageService) ListObjects(ctx context.Context, prefix string) ([]StorageObject, error) {
	if s.client == nil {
		return nil, fmt.Errorf("storage service not initialized")
	}

	// Normalize prefix
	if prefix == "" {
		prefix = "/"
	}

	// Ensure prefix starts with /
	if !filepath.IsAbs(prefix) && prefix[0] != '/' {
		prefix = "/" + prefix
	}

	// Ensure prefix ends with / for directories
	if !filepath.IsAbs(prefix) || prefix[len(prefix)-1] != '/' {
		prefix = prefix + "/"
	}

	result, err := s.client.ListFiles(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var objects []StorageObject
	for _, item := range result.Items {
		// Parse modified time
		modified, err := time.Parse(time.RFC3339, item.Modified)
		if err != nil {
			// If parsing fails, use current time
			modified = time.Now()
		}

		objects = append(objects, StorageObject{
			Name:         item.Name,
			Path:         item.Path,
			IsFolder:     item.IsDir,
			Size:         item.Size,
			LastModified: modified,
			ContentType:  getContentTypeFromExtension(item.Extension),
		})
	}

	return objects, nil
}

// UploadFile uploads a file to FileBrowser storage
func (s *filebrowserStorageService) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	if s.client == nil {
		return fmt.Errorf("storage service not initialized")
	}

	// Normalize object name
	objectName = filepath.Clean(objectName)
	if !filepath.IsAbs(objectName) {
		objectName = "/" + objectName
	}

	filename := filepath.Base(objectName)
	if filename == "" || filename == "." {
		return fmt.Errorf("invalid object name: %s", objectName)
	}

	return s.client.UploadFile(objectName, reader, filename)
}

// GetPresignedURL generates a URL for accessing a file
// FileBrowser doesn't use presigned URLs, but provides direct access URLs
func (s *filebrowserStorageService) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("storage service not initialized")
	}

	// Normalize object name
	objectName = filepath.Clean(objectName)
	if !filepath.IsAbs(objectName) {
		objectName = "/" + objectName
	}

	return s.client.GetFileURL(objectName), nil
}

// DeleteObject deletes a file from FileBrowser storage
func (s *filebrowserStorageService) DeleteObject(ctx context.Context, objectName string) error {
	if s.client == nil {
		return fmt.Errorf("storage service not initialized")
	}

	// Normalize object name
	objectName = filepath.Clean(objectName)
	if !filepath.IsAbs(objectName) {
		objectName = "/" + objectName
	}

	return s.client.DeleteFile(objectName)
}

// ObjectExists checks if a file exists in FileBrowser storage
func (s *filebrowserStorageService) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	if s.client == nil {
		return false, fmt.Errorf("storage service not initialized")
	}

	// Normalize object name
	objectName = filepath.Clean(objectName)
	if !filepath.IsAbs(objectName) {
		objectName = "/" + objectName
	}

	return s.client.FileExists(objectName)
}

// RenameObject renames a file or folder in FileBrowser storage
func (s *filebrowserStorageService) RenameObject(ctx context.Context, oldPath string, newName string) error {
	if s.client == nil {
		return fmt.Errorf("storage service not initialized")
	}

	// Normalize old path
	oldPath = filepath.Clean(oldPath)
	if !filepath.IsAbs(oldPath) {
		oldPath = "/" + oldPath
	}

	return s.client.RenameFile(oldPath, newName)
}

// CreateFolder creates a new folder in FileBrowser storage
func (s *filebrowserStorageService) CreateFolder(ctx context.Context, parentPath string, folderName string) error {
	if s.client == nil {
		return fmt.Errorf("storage service not initialized")
	}

	// Normalize parent path
	if parentPath == "" {
		parentPath = "/"
	}
	parentPath = filepath.Clean(parentPath)
	if !filepath.IsAbs(parentPath) {
		parentPath = "/" + parentPath
	}

	return s.client.CreateFolder(parentPath, folderName)
}

// MoveObject moves a file or folder to a new location in FileBrowser storage
func (s *filebrowserStorageService) MoveObject(ctx context.Context, sourcePath string, destinationPath string) error {
	if s.client == nil {
		return fmt.Errorf("storage service not initialized")
	}

	// Normalize paths
	sourcePath = filepath.Clean(sourcePath)
	if !filepath.IsAbs(sourcePath) {
		sourcePath = "/" + sourcePath
	}

	destinationPath = filepath.Clean(destinationPath)
	if !filepath.IsAbs(destinationPath) {
		destinationPath = "/" + destinationPath
	}

	return s.client.MoveFile(sourcePath, destinationPath)
}

// getContentTypeFromExtension returns content type based on file extension
func getContentTypeFromExtension(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".zip":
		return "application/zip"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".ppt", ".pptx":
		return "application/vnd.ms-powerpoint"
	default:
		return "application/octet-stream"
	}
}
