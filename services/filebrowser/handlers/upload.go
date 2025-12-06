package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ar-13-go-backend/services/filebrowser/models"
	"github.com/ar-13-go-backend/services/filebrowser/utils"
	"github.com/gin-gonic/gin"
)

// UploadHandler handles file uploads
func UploadHandler(dataRoot string) gin.HandlerFunc {
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

		// Build full filesystem path for target directory
		targetDirPath := filepath.Join(dataRoot, path)

		// Security: Ensure the path is within dataRoot
		absDataRoot, err := filepath.Abs(dataRoot)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to resolve data root path",
			})
			return
		}

		absTargetDirPath, err := filepath.Abs(targetDirPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to resolve requested path",
			})
			return
		}

		if !strings.HasPrefix(absTargetDirPath, absDataRoot) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "Access denied: path outside data root",
			})
			return
		}

		// Check if target directory exists
		targetInfo, err := os.Stat(absTargetDirPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   "not_found",
					Message: "Target directory does not exist",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: err.Error(),
			})
			return
		}

		// Check if target is a directory
		if !targetInfo.IsDir() {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Path is not a directory",
			})
			return
		}

		// Parse multipart form
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Failed to parse multipart form: " + err.Error(),
			})
			return
		}

		// Get file from form
		files := form.File["file"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "No file provided. Use 'file' as the form field name",
			})
			return
		}

		// Handle first file (if multiple files, only process the first one)
		fileHeader := files[0]
		filename := fileHeader.Filename

		// Validate filename
		if strings.TrimSpace(filename) == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "bad_request",
				Message: "Filename is required",
			})
			return
		}

		// Build full path for uploaded file
		uploadFilePath := filepath.Join(absTargetDirPath, filename)

		// Security: Ensure the upload file path is still within dataRoot
		absUploadFilePath, err := filepath.Abs(uploadFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to resolve upload file path",
			})
			return
		}

		if !strings.HasPrefix(absUploadFilePath, absDataRoot) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "Access denied: file path outside data root",
			})
			return
		}

		// Check overwrite parameter
		overwrite := c.DefaultQuery("overwrite", "false")
		shouldOverwrite := overwrite == "true"

		// Check if file already exists
		if _, err := os.Stat(absUploadFilePath); err == nil {
			if !shouldOverwrite {
				c.JSON(http.StatusConflict, models.ErrorResponse{
					Error:   "conflict",
					Message: "File already exists. Use overwrite=true to replace it",
				})
				return
			}
			// Remove existing file if overwrite is enabled
			if err := os.Remove(absUploadFilePath); err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "internal_error",
					Message: "Failed to remove existing file: " + err.Error(),
				})
				return
			}
		}

		// Open uploaded file
		src, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to open uploaded file: " + err.Error(),
			})
			return
		}
		defer src.Close()

		// Create destination file
		dst, err := os.Create(absUploadFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to create destination file: " + err.Error(),
			})
			return
		}
		defer dst.Close()

		// Copy file content
		written, err := io.Copy(dst, src)
		if err != nil {
			// Clean up partial file on error
			os.Remove(absUploadFilePath)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to save file: " + err.Error(),
			})
			return
		}

		// Verify file was written successfully (fileInfo not needed, we use written size)

		// Build relative path for response
		relativePath := filepath.Join(path, filename)
		relativePath = strings.ReplaceAll(relativePath, "\\", "/")
		if !strings.HasPrefix(relativePath, "/") {
			relativePath = "/" + relativePath
		}

		// Get MIME type
		mimeType := utils.GetMimeType(absUploadFilePath)

		c.JSON(http.StatusOK, models.UploadResponse{
			Name:       filename,
			Path:       relativePath,
			Size:       written,
			MimeType:   mimeType,
			UploadedAt: time.Now(),
		})
	}
}

