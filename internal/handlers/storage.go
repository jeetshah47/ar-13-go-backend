package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
		errorMsg := "Storage service is not initialized."
		if h.cfg != nil && h.cfg.NASBasePath == "" {
			errorMsg += " Please configure NAS_BASE_PATH for direct filesystem access."
		} else if h.cfg != nil && h.cfg.NASBasePath != "" {
			errorMsg += fmt.Sprintf(" NAS_BASE_PATH is set to '%s' but storage initialization failed. Please check the path exists and is accessible.", h.cfg.NASBasePath)
		} else {
			errorMsg += " Please configure storage settings (NAS_BASE_PATH or MinIO)."
		}
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": errorMsg,
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

	// Check if this is NAS direct filesystem storage (returns a path, not a URL)
	// NAS direct filesystem storage returns a path like "/folder/file.pdf" instead of a URL
	isNASDirectFS := !strings.HasPrefix(downloadURL, "http://") && !strings.HasPrefix(downloadURL, "https://")
	if isNASDirectFS && h.cfg != nil && h.cfg.NASBasePath != "" {
		// Read file directly from filesystem
		localPath := filepath.Join(h.cfg.NASBasePath, strings.TrimPrefix(path, "/"))
		localPath = filepath.Clean(localPath)

		// Open file
		file, fileErr := os.Open(localPath)
		if fileErr != nil {
			if os.IsNotExist(fileErr) {
				c.JSON(constants.StatusNotFound, gin.H{"error": "File not found"})
				return
			}
			c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to open file: " + fileErr.Error()})
			return
		}
		defer file.Close()

		// Get file info
		fileInfo, statErr := file.Stat()
		if statErr != nil {
			c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to stat file: " + statErr.Error()})
			return
		}

		// Get filename from path
		filename := filepath.Base(path)
		if filename == "" || filename == "." || filename == "/" {
			filename = "download"
		}

		// Detect content type
		contentType := "application/octet-stream"
		ext := filepath.Ext(filename)
		if ext != "" {
			// Try to detect from extension
			switch strings.ToLower(ext) {
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
			}
		}

		// Set headers
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
		c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

		// Stream the file
		_, copyErr := io.Copy(c.Writer, file)
		if copyErr != nil {
			// Error writing response, but we can't send JSON error at this point
			return
		}
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

// RenameFile renames a file or folder in NAS storage
// PUT /api/storage/rename?path=/folder/oldname
// Body: { "newName": "newname" }
func (h *StorageHandler) RenameFile(c *gin.Context) {
	if !h.storageService.IsInitialized() {
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": "Storage service is not initialized. Please configure storage settings.",
		})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	path := c.Query("path")
	if path == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "path parameter is required"})
		return
	}

	var req struct {
		NewName string `json:"newName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "newName is required"})
		return
	}

	err := h.storageService.RenameObject(c.Request.Context(), path, req.NewName)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message": "File renamed successfully",
		"path":    path,
		"newName": req.NewName,
	})
}

// DeleteFile deletes a file or folder from NAS storage
// DELETE /api/storage/delete?path=/folder/filename
func (h *StorageHandler) DeleteFile(c *gin.Context) {
	if !h.storageService.IsInitialized() {
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": "Storage service is not initialized. Please configure storage settings.",
		})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	path := c.Query("path")
	if path == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "path parameter is required"})
		return
	}

	err := h.storageService.DeleteObject(c.Request.Context(), path)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message": "File deleted successfully",
		"path":    path,
	})
}

// CreateFolder creates a new folder in NAS storage
// POST /api/storage/create-folder?path=/parent/folder
// Body: { "folderName": "NewFolder" }
func (h *StorageHandler) CreateFolder(c *gin.Context) {
	if !h.storageService.IsInitialized() {
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": "Storage service is not initialized. Please configure storage settings.",
		})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	parentPath := c.DefaultQuery("path", "/")

	var req struct {
		FolderName string `json:"folderName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "folderName is required"})
		return
	}

	err := h.storageService.CreateFolder(c.Request.Context(), parentPath, req.FolderName)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message":    "Folder created successfully",
		"parentPath": parentPath,
		"folderName": req.FolderName,
	})
}

// MoveFile moves a file or folder to a new location in NAS storage
// PUT /api/storage/move?sourcePath=/folder/oldname
// Body: { "destinationPath": "/newfolder/newname" }
func (h *StorageHandler) MoveFile(c *gin.Context) {
	if !h.storageService.IsInitialized() {
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": "Storage service is not initialized. Please configure storage settings.",
		})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": constants.MsgUserNotAuthenticated})
		return
	}

	sourcePath := c.Query("sourcePath")
	if sourcePath == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "sourcePath parameter is required"})
		return
	}

	var req struct {
		DestinationPath string `json:"destinationPath" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "destinationPath is required"})
		return
	}

	err := h.storageService.MoveObject(c.Request.Context(), sourcePath, req.DestinationPath)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message":         "File moved successfully",
		"sourcePath":      sourcePath,
		"destinationPath": req.DestinationPath,
	})
}
