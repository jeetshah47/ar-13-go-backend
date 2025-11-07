package repos

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/internal/models"
	"google.golang.org/api/iterator"
)

// InfoPortalRepo handles info portal data operations
type InfoPortalRepo struct {
	*BaseRepo
}

// NewInfoPortalRepo creates a new info portal repository
func NewInfoPortalRepo() *InfoPortalRepo {
	return &InfoPortalRepo{
		BaseRepo: NewBaseRepo("info-portal"),
	}
}

// Helper methods to get subcollections
func (r *InfoPortalRepo) getFolderCollection() *firestore.CollectionRef {
	return r.collection.Doc("data").Collection("folders")
}

func (r *InfoPortalRepo) getPageCollection() *firestore.CollectionRef {
	return r.collection.Doc("data").Collection("pages")
}

func (r *InfoPortalRepo) getAttachmentCollection() *firestore.CollectionRef {
	return r.collection.Doc("data").Collection("attachments")
}

// Folder operations
func (r *InfoPortalRepo) GetAllFolders(ctx context.Context) ([]models.Folder, error) {
	iter := r.getFolderCollection().Documents(ctx)
	var folders []models.Folder

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var folder models.Folder
		if err := doc.DataTo(&folder); err != nil {
			return nil, err
		}
		folder.ID = doc.Ref.ID
		folders = append(folders, folder)
	}

	return folders, nil
}

func (r *InfoPortalRepo) GetFolderByID(ctx context.Context, folderID string) (*models.Folder, error) {
	doc, err := r.getFolderCollection().Doc(folderID).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, nil
	}

	var folder models.Folder
	if err := doc.DataTo(&folder); err != nil {
		return nil, err
	}
	folder.ID = doc.Ref.ID
	return &folder, nil
}

func (r *InfoPortalRepo) CreateFolder(ctx context.Context, folder *models.Folder) error {
	newDocRef := r.getFolderCollection().NewDoc()
	folder.ID = newDocRef.ID
	folder.Created = time.Now()

	data := map[string]interface{}{
		"id":      folder.ID,
		"name":    folder.Name,
		"color":   folder.Color,
		"created": folder.Created,
	}

	_, err := newDocRef.Set(ctx, data)
	return err
}

func (r *InfoPortalRepo) UpdateFolder(ctx context.Context, folderID string, updates map[string]interface{}) error {
	now := time.Now()
	updates["updated"] = now

	_, err := r.getFolderCollection().Doc(folderID).Update(ctx, []firestore.Update{
		{Path: "name", Value: updates["name"]},
		{Path: "color", Value: updates["color"]},
		{Path: "updated", Value: now},
	})
	return err
}

func (r *InfoPortalRepo) DeleteFolder(ctx context.Context, folderID string) error {
	_, err := r.getFolderCollection().Doc(folderID).Delete(ctx)
	return err
}

// Page operations
func (r *InfoPortalRepo) GetPageByID(ctx context.Context, pageID string) (*models.Page, error) {
	doc, err := r.getPageCollection().Doc(pageID).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, nil
	}

	var page models.Page
	if err := doc.DataTo(&page); err != nil {
		return nil, err
	}
	page.ID = doc.Ref.ID
	return &page, nil
}

func (r *InfoPortalRepo) CreatePage(ctx context.Context, folderID string, page *models.Page) error {
	newDocRef := r.getPageCollection().NewDoc()
	page.ID = newDocRef.ID
	page.FolderID = folderID
	page.Created = time.Now()

	data := map[string]interface{}{
		"id":       page.ID,
		"title":    page.Title,
		"isActive": page.IsActive,
		"folderId": page.FolderID,
		"created":  page.Created,
	}

	_, err := newDocRef.Set(ctx, data)
	return err
}

func (r *InfoPortalRepo) UpdatePage(ctx context.Context, pageID string, updates map[string]interface{}) error {
	now := time.Now()
	updates["updated"] = now

	updateList := []firestore.Update{
		{Path: "updated", Value: now},
	}
	if title, ok := updates["title"].(string); ok {
		updateList = append(updateList, firestore.Update{Path: "title", Value: title})
	}
	if isActive, ok := updates["isActive"].(bool); ok {
		updateList = append(updateList, firestore.Update{Path: "isActive", Value: isActive})
	}

	_, err := r.getPageCollection().Doc(pageID).Update(ctx, updateList)
	return err
}

func (r *InfoPortalRepo) DeletePage(ctx context.Context, pageID string) error {
	_, err := r.getPageCollection().Doc(pageID).Delete(ctx)
	return err
}

func (r *InfoPortalRepo) UpdatePageSections(ctx context.Context, pageID string, sections []models.Section) error {
	now := time.Now()
	_, err := r.getPageCollection().Doc(pageID).Update(ctx, []firestore.Update{
		{Path: "sections", Value: sections},
		{Path: "updated", Value: now},
	})
	return err
}

// Attachment operations
func (r *InfoPortalRepo) CreateAttachment(ctx context.Context, pageID string, attachment *models.Attachment) error {
	newDocRef := r.getAttachmentCollection().NewDoc()
	attachment.ID = newDocRef.ID
	attachment.PageID = pageID
	attachment.Created = time.Now()

	data := map[string]interface{}{
		"id":       attachment.ID,
		"name":     attachment.Name,
		"imageUrl": attachment.ImageURL,
		"fileUrl":  attachment.FileURL,
		"fileType": attachment.FileType,
		"fileSize": attachment.FileSize,
		"pageId":   attachment.PageID,
		"created":  attachment.Created,
	}

	_, err := newDocRef.Set(ctx, data)
	return err
}

func (r *InfoPortalRepo) DeleteAttachment(ctx context.Context, attachmentID string) error {
	_, err := r.getAttachmentCollection().Doc(attachmentID).Delete(ctx)
	return err
}
