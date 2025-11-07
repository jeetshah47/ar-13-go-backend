package handlers

import (
	"path/filepath"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// BackupHandler handles backup routes
type BackupHandler struct {
	backupService *services.BackupService
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler() *BackupHandler {
	return &BackupHandler{
		backupService: services.NewBackupService(),
	}
}

// BackupAllCollections creates a backup of all Firestore collections
func (h *BackupHandler) BackupAllCollections(c *gin.Context) {
	// Default backup directory
	backupDir := filepath.Join("upload", "backups")

	// Allow custom backup directory via query parameter
	if customDir := c.Query("dir"); customDir != "" {
		backupDir = customDir
	}

	backupFiles, err := h.backupService.BackupAllCollections(c.Request.Context(), backupDir)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{
			"error":   "Failed to create backup",
			"details": err.Error(),
		})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"message":          "Backup completed successfully",
		"backupFiles":      backupFiles,
		"totalCollections": len(backupFiles),
	})
}
