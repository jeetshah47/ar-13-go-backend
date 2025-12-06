package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ar-13-go-backend/services/filebrowser/models"
	"github.com/gin-gonic/gin"
)

// DeleteHandler handles deleting files and folders
func DeleteHandler(dataRoot string) gin.HandlerFunc {
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

		// Prevent deletion of the root data directory itself
		if absFullPath == absDataRoot {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "Cannot delete the root data directory",
			})
			return
		}

		// Check if path exists
		info, err := os.Stat(absFullPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   "not_found",
					Message: "File or folder does not exist",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: err.Error(),
			})
			return
		}

		// Delete file or folder
		if info.IsDir() {
			// Delete directory (recursive)
			if err := os.RemoveAll(absFullPath); err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "internal_error",
					Message: "Failed to delete folder: " + err.Error(),
				})
				return
			}
		} else {
			// Delete file
			if err := os.Remove(absFullPath); err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "internal_error",
					Message: "Failed to delete file: " + err.Error(),
				})
				return
			}
		}

		// Normalize response path
		responsePath := strings.ReplaceAll(path, "\\", "/")
		if responsePath == "" {
			responsePath = "/"
		}

		c.JSON(http.StatusOK, models.DeleteResponse{
			Path:      responsePath,
			DeletedAt: time.Now(),
		})
	}
}

