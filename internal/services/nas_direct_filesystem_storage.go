package services

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/config"
)

// nasDirectFilesystemStorage implements StorageServiceInterface using direct filesystem access
// This is used when the backend is deployed directly on the NAS server
type nasDirectFilesystemStorage struct {
	cfg         *config.Config
	basePath    string
	initialized bool
}

// NewNASDirectFilesystemStorage creates a new direct filesystem storage service
func NewNASDirectFilesystemStorage(cfg *config.Config) StorageServiceInterface {
	return &nasDirectFilesystemStorage{
		cfg:         cfg,
		basePath:    cfg.NASBasePath,
		initialized: false,
	}
}

// Initialize verifies the base path exists and is accessible
func (s *nasDirectFilesystemStorage) Initialize() error {
	if s.basePath == "" {
		return fmt.Errorf("NAS_BASE_PATH is not configured")
	}

	// Clean and normalize base path
	s.basePath = filepath.Clean(s.basePath)

	// Verify base path exists and is a directory
	info, err := os.Stat(s.basePath)
	if err != nil {
		return fmt.Errorf("base path does not exist or is not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("base path is not a directory: %s", s.basePath)
	}

	s.initialized = true
	return nil
}

// IsInitialized checks if the storage service is initialized
func (s *nasDirectFilesystemStorage) IsInitialized() bool {
	return s.initialized
}

// toLocalPath converts a NAS path (e.g., "/folder/file.pdf") to a local filesystem path
func (s *nasDirectFilesystemStorage) toLocalPath(nasPath string) string {
	// Normalize path: remove leading slash, clean path
	nasPath = strings.TrimPrefix(nasPath, "/")
	nasPath = filepath.Clean(nasPath)

	// Join with base path
	localPath := filepath.Join(s.basePath, nasPath)
	return localPath
}

// toNASPath converts a local path back to NAS path format
func (s *nasDirectFilesystemStorage) toNASPath(localPath string) string {
	// Remove base path prefix
	nasPath := strings.TrimPrefix(localPath, s.basePath)
	// Normalize to forward slashes and ensure leading slash
	nasPath = filepath.ToSlash(nasPath)
	if !strings.HasPrefix(nasPath, "/") {
		nasPath = "/" + nasPath
	}
	return nasPath
}

// ListObjects lists files and folders at the given path prefix
func (s *nasDirectFilesystemStorage) ListObjects(ctx context.Context, prefix string) ([]StorageObject, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("storage service not initialized")
	}

	localPath := s.toLocalPath(prefix)

	// If prefix is empty or root, list base path
	if prefix == "" || prefix == "/" {
		localPath = s.basePath
	}

	// Read directory
	entries, err := os.ReadDir(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []StorageObject{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var objects []StorageObject
	for _, entry := range entries {
		entryPath := filepath.Join(localPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't stat
		}

		nasPath := s.toNASPath(entryPath)
		obj := StorageObject{
			Name:         entry.Name(),
			Path:         nasPath,
			IsFolder:     entry.IsDir(),
			Size:         info.Size(),
			LastModified: info.ModTime(),
		}

		// Set content type for files
		if !entry.IsDir() {
			ext := filepath.Ext(entry.Name())
			obj.ContentType = mime.TypeByExtension(ext)
			if obj.ContentType == "" {
				obj.ContentType = "application/octet-stream"
			}
		}

		objects = append(objects, obj)
	}

	return objects, nil
}

// UploadFile uploads a file to NAS storage
func (s *nasDirectFilesystemStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	if !s.IsInitialized() {
		return fmt.Errorf("storage service not initialized")
	}

	localPath := s.toLocalPath(objectName)

	// Create parent directory if it doesn't exist
	parentDir := filepath.Dir(localPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Create file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	written, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(localPath) // Clean up on error
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Verify size if provided
	if size > 0 && written != size {
		os.Remove(localPath) // Clean up on error
		return fmt.Errorf("file size mismatch: expected %d, wrote %d", size, written)
	}

	return nil
}

// GetPresignedURL returns the NAS path (for direct filesystem, we return the path)
func (s *nasDirectFilesystemStorage) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if !s.IsInitialized() {
		return "", fmt.Errorf("storage service not initialized")
	}

	// Return the NAS path which the handler can use to read the file directly
	return objectName, nil
}

// DeleteObject deletes a file from NAS storage
func (s *nasDirectFilesystemStorage) DeleteObject(ctx context.Context, objectName string) error {
	if !s.IsInitialized() {
		return fmt.Errorf("storage service not initialized")
	}

	localPath := s.toLocalPath(objectName)

	// Check if it's a directory or file
	info, err := os.Stat(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		// Remove directory recursively
		if err := os.RemoveAll(localPath); err != nil {
			return fmt.Errorf("failed to remove directory: %w", err)
		}
	} else {
		// Remove file
		if err := os.Remove(localPath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
	}

	return nil
}

// ObjectExists checks if a file exists in NAS storage
func (s *nasDirectFilesystemStorage) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	if !s.IsInitialized() {
		return false, fmt.Errorf("storage service not initialized")
	}

	localPath := s.toLocalPath(objectName)

	_, err := os.Stat(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat path: %w", err)
	}

	return true, nil
}

// RenameObject renames a file or folder in NAS storage
func (s *nasDirectFilesystemStorage) RenameObject(ctx context.Context, oldPath string, newName string) error {
	if !s.IsInitialized() {
		return fmt.Errorf("storage service not initialized")
	}

	oldLocalPath := s.toLocalPath(oldPath)

	// Build new path in same directory
	dir := filepath.Dir(oldLocalPath)
	newLocalPath := filepath.Join(dir, newName)

	// Rename
	if err := os.Rename(oldLocalPath, newLocalPath); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}

// CreateFolder creates a new folder in NAS storage
func (s *nasDirectFilesystemStorage) CreateFolder(ctx context.Context, parentPath string, folderName string) error {
	if !s.IsInitialized() {
		return fmt.Errorf("storage service not initialized")
	}

	// Validate folder name
	if strings.Contains(folderName, "/") || strings.Contains(folderName, "\\") {
		return fmt.Errorf("folder name cannot contain path separators")
	}

	// Build folder path
	var localPath string
	if parentPath == "" || parentPath == "/" {
		localPath = filepath.Join(s.basePath, folderName)
	} else {
		parentLocalPath := s.toLocalPath(parentPath)
		localPath = filepath.Join(parentLocalPath, folderName)
	}

	// Create directory
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	return nil
}

// MoveObject moves a file or folder to a new location in NAS storage
func (s *nasDirectFilesystemStorage) MoveObject(ctx context.Context, sourcePath string, destinationPath string) error {
	if !s.IsInitialized() {
		return fmt.Errorf("storage service not initialized")
	}

	sourceLocalPath := s.toLocalPath(sourcePath)
	destLocalPath := s.toLocalPath(destinationPath)

	// Create parent directory if it doesn't exist
	parentDir := filepath.Dir(destLocalPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination parent directory: %w", err)
	}

	// Move (rename)
	if err := os.Rename(sourceLocalPath, destLocalPath); err != nil {
		return fmt.Errorf("failed to move: %w", err)
	}

	return nil
}
