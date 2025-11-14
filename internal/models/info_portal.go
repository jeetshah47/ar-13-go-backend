package models

// Folder represents an info portal folder
type Folder struct {
	Model
	Name  string `json:"name" firestore:"name" bson:"name"`
	Color string `json:"color" firestore:"color" bson:"color"` // Hex color code
	Type  string `json:"type,omitempty" firestore:"type,omitempty" bson:"type,omitempty"` // "folder" for filtering
}

// Page represents an info portal page
type Page struct {
	Model
	Title    string `json:"title" firestore:"title" bson:"title"`
	IsActive bool   `json:"isActive" firestore:"isActive" bson:"isActive"`
	FolderID string `json:"folderId" firestore:"folderId" bson:"folderId"`
	Type     string `json:"type,omitempty" firestore:"type,omitempty" bson:"type,omitempty"` // "page" for filtering
}

// Section represents a section within a page
type Section struct {
	Model
	Title   string `json:"title" firestore:"title" bson:"title"`
	Content string `json:"content" firestore:"content" bson:"content"`
	Order   int    `json:"order" firestore:"order" bson:"order"`
	PageID  string `json:"pageId" firestore:"pageId" bson:"pageId"`
}

// Attachment represents an attachment on a page
type Attachment struct {
	Model
	Name     string `json:"name" firestore:"name" bson:"name"`
	ImageURL string `json:"imageUrl" firestore:"imageUrl" bson:"imageUrl"`
	FileURL  string `json:"fileUrl" firestore:"fileUrl" bson:"fileUrl"`
	FileType string `json:"fileType" firestore:"fileType" bson:"fileType"`
	FileSize int64  `json:"fileSize" firestore:"fileSize" bson:"fileSize"`
	PageID   string `json:"pageId" firestore:"pageId" bson:"pageId"`
	Type     string `json:"type,omitempty" firestore:"type,omitempty" bson:"type,omitempty"` // "attachment" for filtering
}

// PageWithSections represents a page with its sections
type PageWithSections struct {
	Page
	Sections []Section `json:"sections" firestore:"-"`
}
