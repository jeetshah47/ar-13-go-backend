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

// CreateFolderHandler handles creating a new folder
func CreateFolderHandler(dataRoot string) gin.HandlerFunc {
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

		// Parse request body
		var req models.CreateFolderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid request body: " + err.Error(),
			})
			return
		}

		// Validate folder name
		if strings.TrimSpace(req.FolderName) == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Folder name is required",
			})
			return
		}

		// Normalize path
		if path != "/" {
			path = strings.Trim(path, "/")
			path = "/" + path
		}

		// Build full filesystem path for parent directory
		parentPath := filepath.Join(dataRoot, path)

		// Security: Ensure the path is within dataRoot
		absDataRoot, err := filepath.Abs(dataRoot)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to resolve data root path",
			})
			return
		}

		absParentPath, err := filepath.Abs(parentPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to resolve requested path",
			})
			return
		}

		if !strings.HasPrefix(absParentPath, absDataRoot) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "Access denied: path outside data root",
			})
			return
		}

		// Check if parent directory exists
		parentInfo, err := os.Stat(absParentPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   "not_found",
					Message: "Parent directory does not exist",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: err.Error(),
			})
			return
		}

		// Check if parent is a directory
		if !parentInfo.IsDir() {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Path is not a directory",
			})
			return
		}

		// Build full path for new folder
		newFolderPath := filepath.Join(absParentPath, req.FolderName)

		// Security: Ensure the new folder path is still within dataRoot
		absNewFolderPath, err := filepath.Abs(newFolderPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to resolve new folder path",
			})
			return
		}

		if !strings.HasPrefix(absNewFolderPath, absDataRoot) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "Access denied: folder path outside data root",
			})
			return
		}

		// Check if folder already exists
		if _, err := os.Stat(absNewFolderPath); err == nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Folder already exists",
			})
			return
		}

		// Create folder (supports nested folder creation)
		if err := os.MkdirAll(absNewFolderPath, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to create folder: " + err.Error(),
			})
			return
		}

		// Build relative path for response
		relativePath := filepath.Join(path, req.FolderName)
		relativePath = strings.ReplaceAll(relativePath, "\\", "/")
		if !strings.HasPrefix(relativePath, "/") {
			relativePath = "/" + relativePath
		}

		c.JSON(http.StatusOK, models.CreateFolderResponse{
			Path:      relativePath,
			CreatedAt: time.Now(),
		})
	}
}

