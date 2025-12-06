package handlers

import (
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	storageService services.StorageServiceInterface
	cfg            *config.Config
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(storageService services.StorageServiceInterface, cfg *config.Config) *StorageHandler {
	return &StorageHandler{
		storageService: storageService,
		cfg:            cfg,
	}
}

// ListFiles lists files and folders in NAS storage
// GET /api/storage/files?path=/folder/subfolder
func (h *StorageHandler) ListFiles(c *gin.Context) {
	if !h.storageService.IsInitialized() {
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": "Storage service is not initialized. Please configure MINIO_ENDPOINT and other MinIO settings.",
		})
		return
	}

	path := c.DefaultQuery("path", "")

	objects, err := h.storageService.ListObjects(c.Request.Context(), path)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"files": objects,
		"path":  path,
	})
}

// GetFileURL generates a presigned URL for accessing a file
// GET /api/storage/file-url?path=/folder/file.pdf&expiry=3600
// Returns a URL pointing to the backend's download proxy endpoint
func (h *StorageHandler) GetFileURL(c *gin.Context) {
	if !h.storageService.IsInitialized() {
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": "Storage service is not initialized. Please configure MINIO_ENDPOINT and other MinIO settings.",
		})
		return
	}

	path := c.Query("path")
	if path == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "path parameter is required"})
		return
	}

	expiryStr := c.DefaultQuery("expiry", "3600") // Default 1 hour
	expirySeconds, err := strconv.Atoi(expiryStr)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "invalid expiry parameter"})
		return
	}

	// Build the backend proxy URL instead of returning the filebrowser service URL
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := c.Request.Host
	if host == "" {
		host = c.Request.Header.Get("Host")
	}

	// Construct the backend proxy URL
	proxyURL := scheme + "://" + host + constants.Base + constants.StorageBase + constants.StorageDownload + "?path=" + url.QueryEscape(path)

	c.JSON(constants.StatusOK, gin.H{
		"url":    proxyURL,
		"path":   path,
		"expiry": expirySeconds,
	})
}

// DownloadFile proxies file download requests to the filebrowser service
// GET /api/storage/download?path=/folder/file.pdf
// For filebrowser service: proxies the request
// For MinIO: redirects to the presigned URL
func (h *StorageHandler) DownloadFile(c *gin.Context) {
	if !h.storageService.IsInitialized() {
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": "Storage service is not initialized. Please configure MINIO_ENDPOINT and other MinIO settings.",
		})
		return
	}

	path := c.Query("path")
	if path == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "path parameter is required"})
		return
	}

	// Get the download URL from the storage service
	expiry := time.Duration(3600) * time.Second
	downloadURL, err := h.storageService.GetPresignedURL(c.Request.Context(), path, expiry)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if this is a filebrowser service URL (needs proxying)
	// If the URL contains the filebrowser service base URL, we need to proxy it
	needsProxy := false
	if h.cfg != nil && h.cfg.FileBrowserServiceURL != "" {
		filebrowserBaseURL := strings.TrimSuffix(h.cfg.FileBrowserServiceURL, "/")
		if strings.HasPrefix(downloadURL, filebrowserBaseURL) {
			needsProxy = true
		}
	}

	if needsProxy {
		// Proxy the request to filebrowser service
		req, err := http.NewRequest("GET", downloadURL, nil)
		if err != nil {
			c.JSON(constants.StatusInternalServerError, gin.H{"error": "failed to create download request"})
			return
		}

		// Add JWT token from context to Authorization header
		jwtToken := middleware.GetJWTTokenFromContext(c.Request.Context())
		if jwtToken != "" {
			req.Header.Set("Authorization", "Bearer "+jwtToken)
		}

		// Make the request to filebrowser service
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(constants.StatusInternalServerError, gin.H{"error": "failed to download file from filebrowser service: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			c.JSON(resp.StatusCode, gin.H{"error": "filebrowser service returned error: " + string(bodyBytes)})
			return
		}

		// Get filename from path
		filename := filepath.Base(path)
		if filename == "" || filename == "." || filename == "/" {
			filename = "download"
		}

		// Copy response headers
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)

		// Copy content length if available
		if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
			c.Header("Content-Length", contentLength)
		}

		// Stream the file content
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			// Error writing response, but we can't send JSON error at this point
			return
		}
	} else {
		// For MinIO or other services with presigned URLs, redirect to the URL
		c.Redirect(http.StatusTemporaryRedirect, downloadURL)
	}
}

// UploadFile uploads a file to NAS storage
// POST /api/storage/upload
// Form data: file (multipart file), path (optional, default: root)
func (h *StorageHandler) UploadFile(c *gin.Context) {
	if !h.storageService.IsInitialized() {
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": "Storage service is not initialized. Please configure MINIO_ENDPOINT and other MinIO settings.",
		})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Get optional path
	path := c.DefaultPostForm("path", "")

	// Construct object name
	objectName := file.Filename
	if path != "" {
		objectName = path + "/" + file.Filename
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	// Detect content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from extension
		ext := filepath.Ext(file.Filename)
		switch ext {
		case ".pdf":
			contentType = "application/pdf"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".txt":
			contentType = "text/plain"
		case ".json":
			contentType = "application/json"
		case ".zip":
			contentType = "application/zip"
		default:
			contentType = "application/octet-stream"
		}
	}

	// Upload to MinIO
	err = h.storageService.UploadFile(
		c.Request.Context(),
		objectName,
		src,
		file.Size,
		contentType,
	)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message":    "File uploaded successfully",
		"objectName": objectName,
		"path":       objectName,
		"size":       file.Size,
	})
}
