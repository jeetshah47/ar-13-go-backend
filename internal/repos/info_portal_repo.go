package repos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// InfoPortalRepo handles info portal data operations
type InfoPortalRepo struct {
	*DynamoBaseRepo
}

// NewInfoPortalRepo creates a new info portal repository
func NewInfoPortalRepo() *InfoPortalRepo {
	return &InfoPortalRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("info-portal"),
	}
}

// Folder operations
func (r *InfoPortalRepo) GetAllFolders(ctx context.Context) ([]models.Folder, error) {
	items, err := r.ScanItems(ctx, nil)
	if err != nil {
		return nil, err
	}

	folders := make([]models.Folder, 0)
	for _, item := range items {
		// Filter by type = "folder"
		if itemType, ok := item["type"].(*types.AttributeValueMemberS); ok && itemType.Value == "folder" {
			var folder models.Folder
			if err := UnmarshalItem(item, &folder); err != nil {
				continue
			}
			folders = append(folders, folder)
		}
	}

	return folders, nil
}

func (r *InfoPortalRepo) GetFolderByID(ctx context.Context, folderID string) (*models.Folder, error) {
	// Use prefix if not already present
	id := folderID
	if !strings.HasPrefix(id, "folder-") {
		id = "folder-" + folderID
	}

	item, err := r.DynamoBaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var folder models.Folder
	if err := UnmarshalItem(item, &folder); err != nil {
		return nil, fmt.Errorf("failed to unmarshal folder: %w", err)
	}

	return &folder, nil
}

func (r *InfoPortalRepo) CreateFolder(ctx context.Context, folder *models.Folder) error {
	now := time.Now()
	if folder.ID == "" {
		folder.ID = "folder-" + uuid.New().String()
	} else if !strings.HasPrefix(folder.ID, "folder-") {
		folder.ID = "folder-" + folder.ID
	}
	folder.Created = now

	data := map[string]interface{}{
		"id":      folder.ID,
		"type":    "folder",
		"name":    folder.Name,
		"color":   folder.Color,
		"created": folder.Created.Format(time.RFC3339),
	}

	return r.PutItem(ctx, data)
}

func (r *InfoPortalRepo) UpdateFolder(ctx context.Context, folderID string, updates map[string]interface{}) error {
	id := folderID
	if !strings.HasPrefix(id, "folder-") {
		id = "folder-" + folderID
	}

	// Remove type from updates if present
	delete(updates, "type")
	return r.UpdateItem(ctx, id, updates)
}

func (r *InfoPortalRepo) DeleteFolder(ctx context.Context, folderID string) error {
	id := folderID
	if !strings.HasPrefix(id, "folder-") {
		id = "folder-" + folderID
	}
	return r.DeleteByID(ctx, id)
}

// Page operations
func (r *InfoPortalRepo) GetPageByID(ctx context.Context, pageID string) (*models.Page, error) {
	// Use prefix if not already present
	id := pageID
	if !strings.HasPrefix(id, "page-") {
		id = "page-" + pageID
	}

	item, err := r.DynamoBaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var page models.Page
	if err := UnmarshalItem(item, &page); err != nil {
		return nil, fmt.Errorf("failed to unmarshal page: %w", err)
	}

	return &page, nil
}

func (r *InfoPortalRepo) CreatePage(ctx context.Context, folderID string, page *models.Page) error {
	now := time.Now()
	if page.ID == "" {
		page.ID = "page-" + uuid.New().String()
	} else if !strings.HasPrefix(page.ID, "page-") {
		page.ID = "page-" + page.ID
	}
	page.FolderID = folderID
	page.Created = now

	data := map[string]interface{}{
		"id":       page.ID,
		"type":     "page",
		"title":    page.Title,
		"isActive": page.IsActive,
		"folderId": page.FolderID,
		"created":  page.Created.Format(time.RFC3339),
	}

	return r.PutItem(ctx, data)
}

func (r *InfoPortalRepo) UpdatePage(ctx context.Context, pageID string, updates map[string]interface{}) error {
	id := pageID
	if !strings.HasPrefix(id, "page-") {
		id = "page-" + pageID
	}

	// Remove type from updates if present
	delete(updates, "type")
	return r.UpdateItem(ctx, id, updates)
}

func (r *InfoPortalRepo) DeletePage(ctx context.Context, pageID string) error {
	id := pageID
	if !strings.HasPrefix(id, "page-") {
		id = "page-" + pageID
	}
	return r.DeleteByID(ctx, id)
}

func (r *InfoPortalRepo) UpdatePageSections(ctx context.Context, pageID string, sections []models.Section) error {
	id := pageID
	if !strings.HasPrefix(id, "page-") {
		id = "page-" + pageID
	}

	updates := map[string]interface{}{
		"sections": sections,
	}
	return r.UpdateItem(ctx, id, updates)
}

// Attachment operations
func (r *InfoPortalRepo) CreateAttachment(ctx context.Context, pageID string, attachment *models.Attachment) error {
	now := time.Now()
	if attachment.ID == "" {
		attachment.ID = "attachment-" + uuid.New().String()
	} else if !strings.HasPrefix(attachment.ID, "attachment-") {
		attachment.ID = "attachment-" + attachment.ID
	}
	attachment.PageID = pageID
	attachment.Created = now

	data := map[string]interface{}{
		"id":       attachment.ID,
		"type":     "attachment",
		"name":     attachment.Name,
		"imageUrl": attachment.ImageURL,
		"fileUrl":  attachment.FileURL,
		"fileType": attachment.FileType,
		"fileSize": attachment.FileSize,
		"pageId":   attachment.PageID,
		"created":  attachment.Created.Format(time.RFC3339),
	}

	return r.PutItem(ctx, data)
}

func (r *InfoPortalRepo) DeleteAttachment(ctx context.Context, attachmentID string) error {
	id := attachmentID
	if !strings.HasPrefix(id, "attachment-") {
		id = "attachment-" + attachmentID
	}
	return r.DeleteByID(ctx, id)
}
