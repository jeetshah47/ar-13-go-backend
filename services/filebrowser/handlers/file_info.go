package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ar-13-go-backend/services/filebrowser/models"
	"github.com/ar-13-go-backend/services/filebrowser/utils"
	"github.com/gin-gonic/gin"
)

// FileInfoHandler handles getting file metadata
func FileInfoHandler(dataRoot string) gin.HandlerFunc {
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
					Message: "File or directory does not exist",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: err.Error(),
			})
			return
		}

		// Normalize response path
		responsePath := strings.ReplaceAll(path, "\\", "/")
		if responsePath == "" {
			responsePath = "/"
		}

		// Build response
		fileInfo := models.FileInfoResponse{
			Name:     filepath.Base(absFullPath),
			Path:     responsePath,
			IsFolder: info.IsDir(),
			Size:     info.Size(),
			Modified: info.ModTime(),
		}

		// Add MIME type for files
		if !info.IsDir() {
			fileInfo.MimeType = utils.GetMimeType(absFullPath)
		}

		c.JSON(http.StatusOK, fileInfo)
	}
}

