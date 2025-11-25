package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ar-13-go-backend/services/filebrowser/models"
	"github.com/ar-13-go-backend/services/filebrowser/utils"
	"github.com/gin-gonic/gin"
)

// DownloadHandler handles file downloads with streaming support
// Supports expiry parameter for time-limited access
func DownloadHandler(dataRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get path parameter
		path := c.Query("path")
		if path == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Path parameter is required",
			})
			return
		}

		// Check expiry if provided (for time-limited access)
		if expiresStr := c.Query("expires"); expiresStr != "" {
			expiresAt, err := time.Parse(time.RFC3339, expiresStr)
			if err == nil {
				if time.Now().After(expiresAt) {
					c.JSON(http.StatusForbidden, models.ErrorResponse{
						Error:   "forbidden",
						Message: "Access link has expired",
					})
					return
				}
			}
		}

		// Normalize path
		if path != "/" {
			path = strings.Trim(path, "/")
			path = "/" + path
		}

		// Build full filesystem path
		fullPath := filepath.Join(dataRoot, path)
		
		// Security: Ensure the path is within dataRoot
		absDataRoot, err := filepath.Abs(dataRoot)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to resolve data root path",
			})
			return
		}

		absFullPath, err := filepath.Abs(fullPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to resolve requested path",
			})
			return
		}

		if !strings.HasPrefix(absFullPath, absDataRoot) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "Access denied: path outside data root",
			})
			return
		}

		// Get file info
		info, err := os.Stat(absFullPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   "not_found",
					Message: "File does not exist",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: err.Error(),
			})
			return
		}

		// Check if it's a directory
		if info.IsDir() {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Cannot download a directory",
			})
			return
		}

		// Open file
		file, err := os.Open(absFullPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to open file: " + err.Error(),
			})
			return
		}
		defer file.Close()

		// Set headers for file download
		filename := filepath.Base(absFullPath)
		mimeType := utils.GetMimeType(absFullPath)
		
		c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
		c.Header("Content-Type", mimeType)
		c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))

		// Stream file to response
		c.DataFromReader(http.StatusOK, info.Size(), mimeType, file, nil)
	}
}

