package services

import (
	"context"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// InfoPortalService handles info portal business logic
type InfoPortalService struct {
	infoPortalRepo repos.InfoPortalRepository
}

// NewInfoPortalService creates a new info portal service with dependency injection
func NewInfoPortalService(infoPortalRepo repos.InfoPortalRepository) *InfoPortalService {
	return &InfoPortalService{
		infoPortalRepo: infoPortalRepo,
	}
}

// NewInfoPortalServiceWithDefaults creates a new info portal service with default dependencies
func NewInfoPortalServiceWithDefaults() *InfoPortalService {
	return NewInfoPortalService(repos.NewInfoPortalRepo())
}

// Folder operations
func (s *InfoPortalService) GetAllFolders(ctx context.Context) ([]models.Folder, error) {
	return s.infoPortalRepo.GetAllFolders(ctx)
}

func (s *InfoPortalService) GetFolderByID(ctx context.Context, folderID string) (*models.Folder, error) {
	return s.infoPortalRepo.GetFolderByID(ctx, folderID)
}

func (s *InfoPortalService) CreateFolder(ctx context.Context, name, color string) (*models.Folder, error) {
	folder := &models.Folder{
		Name:  name,
		Color: color,
	}
	if err := s.infoPortalRepo.CreateFolder(ctx, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *InfoPortalService) UpdateFolder(ctx context.Context, folderID string, name, color *string) (*models.Folder, error) {
	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if color != nil {
		updates["color"] = *color
	}

	if err := s.infoPortalRepo.UpdateFolder(ctx, folderID, updates); err != nil {
		return nil, err
	}

	return s.infoPortalRepo.GetFolderByID(ctx, folderID)
}

func (s *InfoPortalService) DeleteFolder(ctx context.Context, folderID string) error {
	return s.infoPortalRepo.DeleteFolder(ctx, folderID)
}

// Page operations
func (s *InfoPortalService) GetPageByID(ctx context.Context, pageID string) (*models.Page, error) {
	return s.infoPortalRepo.GetPageByID(ctx, pageID)
}

func (s *InfoPortalService) CreatePage(ctx context.Context, folderID string, title string) (*models.Page, error) {
	page := &models.Page{
		Title:    title,
		IsActive: false,
		FolderID: folderID,
	}
	if err := s.infoPortalRepo.CreatePage(ctx, folderID, page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *InfoPortalService) UpdatePage(ctx context.Context, pageID string, title *string, isActive *bool) (*models.Page, error) {
	updates := make(map[string]interface{})
	if title != nil {
		updates["title"] = *title
	}
	if isActive != nil {
		updates["isActive"] = *isActive
	}

	if err := s.infoPortalRepo.UpdatePage(ctx, pageID, updates); err != nil {
		return nil, err
	}

	return s.infoPortalRepo.GetPageByID(ctx, pageID)
}

func (s *InfoPortalService) DeletePage(ctx context.Context, pageID string) error {
	return s.infoPortalRepo.DeletePage(ctx, pageID)
}

func (s *InfoPortalService) UpdatePageSections(ctx context.Context, pageID string, sections []models.Section) ([]models.Section, error) {
	if err := s.infoPortalRepo.UpdatePageSections(ctx, pageID, sections); err != nil {
		return nil, err
	}
	return sections, nil
}

// Attachment operations
func (s *InfoPortalService) CreateAttachment(ctx context.Context, pageID string, name, imageURL, fileURL, fileType string, fileSize int64) (*models.Attachment, error) {
	attachment := &models.Attachment{
		Name:     name,
		ImageURL: imageURL,
		FileURL:  fileURL,
		FileType: fileType,
		FileSize: fileSize,
		PageID:   pageID,
	}
	if err := s.infoPortalRepo.CreateAttachment(ctx, pageID, attachment); err != nil {
		return nil, err
	}
	return attachment, nil
}

func (s *InfoPortalService) DeleteAttachment(ctx context.Context, attachmentID string) error {
	return s.infoPortalRepo.DeleteAttachment(ctx, attachmentID)
}

// Statistics
func (s *InfoPortalService) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	folders, err := s.infoPortalRepo.GetAllFolders(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"totalFolders":     len(folders),
		"totalPages":       0, // TODO: Implement page count
		"totalAttachments": 0, // TODO: Implement attachment count
	}, nil
}
