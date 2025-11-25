package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type storageService struct {
	client *minio.Client
	config *config.Config
}

// NewStorageService creates a new storage service instance
func NewStorageService(cfg *config.Config) StorageServiceInterface {
	return &storageService{
		config: cfg,
	}
}

// Initialize initializes the MinIO client and ensures the bucket exists
func (s *storageService) Initialize() error {
	if s.config.MinIOEndpoint == "" {
		return fmt.Errorf("MINIO_ENDPOINT is not configured")
	}

	// Strip protocol from endpoint if present (MinIO client expects only hostname:port)
	endpoint := strings.TrimSpace(s.config.MinIOEndpoint)
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	// Remove quotes if present
	endpoint = strings.Trim(endpoint, "\"'")

	// Handle empty credentials (for SeaweedFS or anonymous access)
	accessKey := s.config.MinIOAccessKey
	secretKey := s.config.MinIOSecretKey
	// Empty credentials are allowed for S3-compatible services that support anonymous access

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: s.config.MinIOUseSSL,
		Region: "us-east-1",
	}

	if s.config.MinIOUseSSL && s.config.MinIOInsecureSSL {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		opts.Transport = tr
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %w", err)
	}

	s.client = client

	// Ensure bucket exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, s.config.MinIOBucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, s.config.MinIOBucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// IsInitialized checks if the storage service is initialized
func (s *storageService) IsInitialized() bool {
	return s.client != nil
}

// ListObjects lists files and folders at the given path prefix
func (s *storageService) ListObjects(ctx context.Context, prefix string) ([]StorageObject, error) {
	if s.client == nil {
		return nil, fmt.Errorf("storage service not initialized")
	}

	// Normalize prefix (ensure it ends with / for folders, remove leading /)
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var objects []StorageObject
	objectChan := s.client.ListObjects(ctx, s.config.MinIOBucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
	})

	// Track folders we've seen
	folders := make(map[string]bool)

	for obj := range objectChan {
		if obj.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", obj.Err)
		}

		// Extract relative path
		relativePath := strings.TrimPrefix(obj.Key, prefix)

		// Skip if it's the prefix itself
		if relativePath == "" {
			continue
		}

		// Check if it's a folder (ends with /) or a file
		if strings.HasSuffix(obj.Key, "/") {
			// It's a folder
			folderName := strings.TrimSuffix(relativePath, "/")
			if !folders[folderName] {
				objects = append(objects, StorageObject{
					Name:         folderName,
					Path:         obj.Key,
					IsFolder:     true,
					LastModified: obj.LastModified,
				})
				folders[folderName] = true
			}
		} else {
			// It's a file
			objects = append(objects, StorageObject{
				Name:         filepath.Base(obj.Key),
				Path:         obj.Key,
				IsFolder:     false,
				Size:         obj.Size,
				LastModified: obj.LastModified,
				ContentType:  obj.ContentType,
			})
		}
	}

	return objects, nil
}

// UploadFile uploads a file to MinIO storage
func (s *storageService) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	if s.client == nil {
		return fmt.Errorf("storage service not initialized")
	}

	// Normalize object name (remove leading /)
	objectName = strings.TrimPrefix(objectName, "/")

	_, err := s.client.PutObject(ctx, s.config.MinIOBucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// GetPresignedURL generates a presigned URL for accessing a file
func (s *storageService) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("storage service not initialized")
	}

	// Normalize object name
	objectName = strings.TrimPrefix(objectName, "/")

	url, err := s.client.PresignedGetObject(ctx, s.config.MinIOBucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// DeleteObject deletes a file from MinIO storage
func (s *storageService) DeleteObject(ctx context.Context, objectName string) error {
	if s.client == nil {
		return fmt.Errorf("storage service not initialized")
	}

	objectName = strings.TrimPrefix(objectName, "/")

	err := s.client.RemoveObject(ctx, s.config.MinIOBucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// ObjectExists checks if a file exists in MinIO storage
func (s *storageService) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	if s.client == nil {
		return false, fmt.Errorf("storage service not initialized")
	}

	objectName = strings.TrimPrefix(objectName, "/")

	_, err := s.client.StatObject(ctx, s.config.MinIOBucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

