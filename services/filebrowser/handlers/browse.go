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

// BrowseHandler handles listing files and folders in a directory
func BrowseHandler(dataRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get path parameter (default to root "/")
		path := c.DefaultQuery("path", "/")
		
		// Normalize path (remove leading/trailing slashes except root)
		if path != "/" {
			path = strings.Trim(path, "/")
			path = "/" + path
		}

		// Build full filesystem path
		fullPath := filepath.Join(dataRoot, path)
		
		// Security: Ensure the path is within dataRoot (prevent directory traversal)
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

		// Check if path exists
		info, err := os.Stat(absFullPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   "not_found",
					Message: "Path does not exist",
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
		if !info.IsDir() {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Path is not a directory",
			})
			return
		}

		// Read directory contents
		entries, err := os.ReadDir(absFullPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to read directory: " + err.Error(),
			})
			return
		}

		// Build response
		files := make([]models.FileItem, 0, len(entries))
		for _, entry := range entries {
			entryPath := filepath.Join(absFullPath, entry.Name())
			entryInfo, err := entry.Info()
			if err != nil {
				continue // Skip entries we can't read
			}

			// Build relative path for response
			relativePath := filepath.Join(path, entry.Name())
			// Normalize path separators to forward slashes
			relativePath = strings.ReplaceAll(relativePath, "\\", "/")
			if !strings.HasPrefix(relativePath, "/") {
				relativePath = "/" + relativePath
			}

			fileItem := models.FileItem{
				Name:     entry.Name(),
				Path:     relativePath,
				IsFolder: entry.IsDir(),
				Size:     entryInfo.Size(),
				Modified: entryInfo.ModTime(),
			}

			// Add MIME type for files
			if !entry.IsDir() {
				fileItem.MimeType = utils.GetMimeType(entryPath)
			}

			files = append(files, fileItem)
		}

		// Normalize response path
		responsePath := strings.ReplaceAll(path, "\\", "/")
		if responsePath == "" {
			responsePath = "/"
		}

		c.JSON(http.StatusOK, models.BrowseResponse{
			Path:  responsePath,
			Files: files,
		})
	}
}

