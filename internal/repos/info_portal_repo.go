package repos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// InfoPortalRepo handles info portal data operations with MongoDB
type InfoPortalRepo struct {
	*MongoBaseRepo
}

// NewInfoPortalRepo creates a new MongoDB info portal repository
func NewInfoPortalRepo() *InfoPortalRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &InfoPortalRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "info-portal"),
	}
}

// Folder operations
func (r *InfoPortalRepo) GetAllFolders(ctx context.Context) ([]models.Folder, error) {
	filter := bson.M{"type": "folder"}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	folders := make([]models.Folder, 0, len(items))
	for _, item := range items {
		var folder models.Folder
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &folder); err != nil {
			continue
		}
		folders = append(folders, folder)
	}

	return folders, nil
}

func (r *InfoPortalRepo) GetFolderByID(ctx context.Context, folderID string) (*models.Folder, error) {
	id := folderID
	if !strings.HasPrefix(id, "folder-") {
		id = "folder-" + folderID
	}

	result := r.MongoBaseRepo.GetByID(ctx, id)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var folder models.Folder
	if err := result.Decode(&folder); err != nil {
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
	folder.Type = "folder"

	return r.InsertOne(ctx, folder)
}

func (r *InfoPortalRepo) UpdateFolder(ctx context.Context, folderID string, updates map[string]interface{}) error {
	id := folderID
	if !strings.HasPrefix(id, "folder-") {
		id = "folder-" + folderID
	}

	// Remove type from updates if present
	delete(updates, "type")
	return r.UpdateOne(ctx, id, updates)
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
	id := pageID
	if !strings.HasPrefix(id, "page-") {
		id = "page-" + pageID
	}

	result := r.MongoBaseRepo.GetByID(ctx, id)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var page models.Page
	if err := result.Decode(&page); err != nil {
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
	page.Type = "page"

	return r.InsertOne(ctx, page)
}

func (r *InfoPortalRepo) UpdatePage(ctx context.Context, pageID string, updates map[string]interface{}) error {
	id := pageID
	if !strings.HasPrefix(id, "page-") {
		id = "page-" + pageID
	}

	// Remove type from updates if present
	delete(updates, "type")
	return r.UpdateOne(ctx, id, updates)
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

	updates := bson.M{"sections": sections}
	return r.UpdateOne(ctx, id, updates)
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
	attachment.Type = "attachment"

	return r.InsertOne(ctx, attachment)
}

func (r *InfoPortalRepo) DeleteAttachment(ctx context.Context, attachmentID string) error {
	id := attachmentID
	if !strings.HasPrefix(id, "attachment-") {
		id = "attachment-" + attachmentID
	}
	return r.DeleteByID(ctx, id)
}

