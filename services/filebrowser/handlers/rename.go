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

// RenameHandler handles renaming files and folders
func RenameHandler(dataRoot string) gin.HandlerFunc {
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
		var req models.RenameRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid request body: " + err.Error(),
			})
			return
		}

		// Validate new name
		if strings.TrimSpace(req.NewName) == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "New name is required",
			})
			return
		}

		// Validate new name doesn't contain path traversal characters
		if strings.Contains(req.NewName, "/") || strings.Contains(req.NewName, "\\") ||
			strings.Contains(req.NewName, "..") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "New name cannot contain path separators or '..'",
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

		// Check if source path exists
		if _, err := os.Stat(absFullPath); err != nil {
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

		// Build new path
		parentDir := filepath.Dir(absFullPath)
		newPath := filepath.Join(parentDir, req.NewName)

		// Security: Ensure the new path is still within dataRoot
		absNewPath, err := filepath.Abs(newPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to resolve new path",
			})
			return
		}

		if !strings.HasPrefix(absNewPath, absDataRoot) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "Access denied: new path outside data root",
			})
			return
		}

		// Check if target name already exists
		if _, err := os.Stat(absNewPath); err == nil {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "conflict",
				Message: "A file or folder with that name already exists",
			})
			return
		}

		// Rename file or folder
		if err := os.Rename(absFullPath, absNewPath); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to rename: " + err.Error(),
			})
			return
		}

		// Get updated file info
		newInfo, err := os.Stat(absNewPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to get updated file info: " + err.Error(),
			})
			return
		}

		// Build relative path for response
		relativePath := filepath.Join(filepath.Dir(path), req.NewName)
		relativePath = strings.ReplaceAll(relativePath, "\\", "/")
		if !strings.HasPrefix(relativePath, "/") {
			relativePath = "/" + relativePath
		}

		// Build response
		response := models.RenameResponse{
			Name:     req.NewName,
			Path:     relativePath,
			IsFolder: newInfo.IsDir(),
			Size:     newInfo.Size(),
			Modified: newInfo.ModTime(),
		}

		// Add MIME type for files
		if !newInfo.IsDir() {
			response.MimeType = utils.GetMimeType(absNewPath)
		}

		c.JSON(http.StatusOK, response)
	}
}

