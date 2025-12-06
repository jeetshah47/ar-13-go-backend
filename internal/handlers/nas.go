package handlers

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// NASHandler handles NAS-related operations
type NASHandler struct {
	cfg            *config.Config
	sessionManager *services.QNAPSessionManager
}

// NewNASHandler creates a new NAS handler
func NewNASHandler(cfg *config.Config) *NASHandler {
	var sessionManager *services.QNAPSessionManager

	// Initialize QNAP client if configured
	if cfg.QNAPNASIP != "" && cfg.QNAPServiceUser != "" {
		qnapClient := services.NewQNAPClient(
			cfg.QNAPNASIP,
			cfg.QNAPAPIPort,
			cfg.QNAPServiceUser,
			cfg.QNAPServicePassword,
		)
		sessionManager = services.NewQNAPSessionManager(qnapClient)
	}

	return &NASHandler{
		cfg:            cfg,
		sessionManager: sessionManager,
	}
}

// MountCredentialsResponse represents the response for mount credentials
type MountCredentialsResponse struct {
	NASIP     string    `json:"nasIP"`
	ShareName string    `json:"shareName"`
	SMBPath   string    `json:"smbPath"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	SID       string    `json:"sid"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// GetMountCredentials returns SMB mount credentials for Electron app
// GET /api/nas/mount-credentials
func (h *NASHandler) GetMountCredentials(c *gin.Context) {
	// Validate user is authenticated (JWT middleware should handle this)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if QNAP is configured
	if h.cfg.QNAPNASIP == "" || h.cfg.QNAPServiceUser == "" {
		c.JSON(constants.StatusServiceUnavailable, gin.H{
			"error": "QNAP NAS is not configured",
		})
		return
	}

	// Get QNAP session ID
	sid, err := h.sessionManager.GetSessionID(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{
			"error": "Failed to authenticate with QNAP NAS: " + err.Error(),
		})
		return
	}

	// Build SMB path
	// Windows format: \\192.168.1.100\studio-work
	// Unix format: smb://192.168.1.100/studio-work
	smbPath := "\\\\" + h.cfg.QNAPNASIP + "\\" + h.cfg.QNAPShareName
	smbPath = strings.ReplaceAll(smbPath, "\\", "/") // Normalize for cross-platform

	// Calculate expiration (25 minutes from now, matching session expiration)
	expiresAt := time.Now().Add(25 * time.Minute)

	response := MountCredentialsResponse{
		NASIP:     h.cfg.QNAPNASIP,
		ShareName: h.cfg.QNAPShareName,
		SMBPath:   smbPath,
		Username:  h.cfg.QNAPServiceUser,
		Password:  h.cfg.QNAPServicePassword, // In production, consider using temporary token
		SID:       sid,
		ExpiresAt: expiresAt,
	}

	c.JSON(constants.StatusOK, response)
}

// GetFilePath resolves a file ID to a file path
// GET /api/nas/file-path/:fileId
func (h *NASHandler) GetFilePath(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileID := c.Param("fileId")
	if fileID == "" {
		c.JSON(constants.StatusBadRequest, gin.H{"error": "fileId is required"})
		return
	}

	// TODO: Implement file ID to path mapping
	// This would require a file metadata service or database lookup
	// For now, return an error indicating this needs to be implemented
	c.JSON(constants.StatusNotImplemented, gin.H{
		"error":   "File ID to path mapping not yet implemented",
		"message": "This endpoint requires file metadata storage to map file IDs to paths",
	})
}

// BuildSMBPath builds the SMB path for a given file path
func (h *NASHandler) BuildSMBPath(filePath string) string {
	// Normalize the file path
	if filePath != "/" {
		filePath = strings.Trim(filePath, "/")
		filePath = "/" + filePath
	}

	// Build full SMB path
	// Windows: \\192.168.1.100\studio-work\folder\file.pdf
	smbPath := "\\\\" + h.cfg.QNAPNASIP + "\\" + h.cfg.QNAPShareName

	// Add file path if not root
	if filePath != "/" {
		// Convert forward slashes to backslashes for Windows
		filePath = strings.ReplaceAll(filePath, "/", "\\")
		smbPath = filepath.Join(smbPath, filePath)
	}

	return smbPath
}
