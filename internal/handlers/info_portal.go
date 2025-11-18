package handlers

import (
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// InfoPortalHandler handles info portal routes
type InfoPortalHandler struct {
	infoPortalService *services.InfoPortalService
}

// NewInfoPortalHandler creates a new info portal handler with dependency injection
func NewInfoPortalHandler(infoPortalService *services.InfoPortalService) *InfoPortalHandler {
	return &InfoPortalHandler{
		infoPortalService: infoPortalService,
	}
}

// NewInfoPortalHandlerWithDefaults creates a new info portal handler with default dependencies
func NewInfoPortalHandlerWithDefaults() *InfoPortalHandler {
	return NewInfoPortalHandler(
		services.NewInfoPortalServiceWithDefaults(),
	)
}

// GetAllFolders gets all folders
func (h *InfoPortalHandler) GetAllFolders(c *gin.Context) {
	folders, err := h.infoPortalService.GetAllFolders(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"folders":      folders,
		"totalFolders": len(folders),
	})
}

// GetFolderById gets folder by ID
func (h *InfoPortalHandler) GetFolderById(c *gin.Context) {
	folderID := c.Param("folderId")
	folder, err := h.infoPortalService.GetFolderByID(c.Request.Context(), folderID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if folder == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	c.JSON(constants.StatusOK, folder)
}

// CreateFolder creates a folder
func (h *InfoPortalHandler) CreateFolder(c *gin.Context) {
	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.infoPortalService.CreateFolder(c.Request.Context(), req.Name, req.Color)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{"folder": folder})
}

// UpdateFolder updates a folder
func (h *InfoPortalHandler) UpdateFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	var req struct {
		Name  *string `json:"name,omitempty"`
		Color *string `json:"color,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.infoPortalService.UpdateFolder(c.Request.Context(), folderID, req.Name, req.Color)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"folder": folder})
}

// DeleteFolder deletes a folder
func (h *InfoPortalHandler) DeleteFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	if err := h.infoPortalService.DeleteFolder(c.Request.Context(), folderID); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Folder deleted successfully"})
}

// GetPageById gets page by ID
func (h *InfoPortalHandler) GetPageById(c *gin.Context) {
	pageID := c.Param("pageId")
	page, err := h.infoPortalService.GetPageByID(c.Request.Context(), pageID)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if page == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}

	c.JSON(constants.StatusOK, page)
}

// CreatePage creates a page
func (h *InfoPortalHandler) CreatePage(c *gin.Context) {
	folderID := c.Param("folderId")
	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, err := h.infoPortalService.CreatePage(c.Request.Context(), folderID, req.Title)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{"page": page})
}

// UpdatePage updates a page
func (h *InfoPortalHandler) UpdatePage(c *gin.Context) {
	pageID := c.Param("pageId")
	var req struct {
		Title    *string `json:"title,omitempty"`
		IsActive *bool   `json:"isActive,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, err := h.infoPortalService.UpdatePage(c.Request.Context(), pageID, req.Title, req.IsActive)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"page": page})
}

// DeletePage deletes a page
func (h *InfoPortalHandler) DeletePage(c *gin.Context) {
	pageID := c.Param("pageId")
	if err := h.infoPortalService.DeletePage(c.Request.Context(), pageID); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Page deleted successfully"})
}

// UpdatePageSections updates page sections
func (h *InfoPortalHandler) UpdatePageSections(c *gin.Context) {
	pageID := c.Param("pageId")
	var req struct {
		Sections []models.Section `json:"sections"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sections, err := h.infoPortalService.UpdatePageSections(c.Request.Context(), pageID, req.Sections)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"sections": sections})
}

// UploadAttachment uploads a page attachment
func (h *InfoPortalHandler) UploadAttachment(c *gin.Context) {
	_ = c.Param("pageId") // TODO: Use pageID when implementing file upload
	// TODO: Implement file upload
	c.JSON(constants.StatusCreated, gin.H{"message": "Attachment upload not yet implemented"})
}

// DeleteAttachment deletes an attachment
func (h *InfoPortalHandler) DeleteAttachment(c *gin.Context) {
	attachmentID := c.Param("attachmentId")
	if err := h.infoPortalService.DeleteAttachment(c.Request.Context(), attachmentID); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Attachment deleted successfully"})
}

// GetStatistics gets statistics
func (h *InfoPortalHandler) GetStatistics(c *gin.Context) {
	statistics, err := h.infoPortalService.GetStatistics(c.Request.Context())
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"statistics": statistics})
}
