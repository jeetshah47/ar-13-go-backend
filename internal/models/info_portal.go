package models

// Folder represents an info portal folder
type Folder struct {
	Model
	Name  string `json:"name" firestore:"name"`
	Color string `json:"color" firestore:"color"` // Hex color code
}

// Page represents an info portal page
type Page struct {
	Model
	Title    string `json:"title" firestore:"title"`
	IsActive bool   `json:"isActive" firestore:"isActive"`
	FolderID string `json:"folderId" firestore:"folderId"`
}

// Section represents a section within a page
type Section struct {
	Model
	Title   string `json:"title" firestore:"title"`
	Content string `json:"content" firestore:"content"`
	Order   int    `json:"order" firestore:"order"`
	PageID  string `json:"pageId" firestore:"pageId"`
}

// Attachment represents an attachment on a page
type Attachment struct {
	Model
	Name     string `json:"name" firestore:"name"`
	ImageURL string `json:"imageUrl" firestore:"imageUrl"`
	FileURL  string `json:"fileUrl" firestore:"fileUrl"`
	FileType string `json:"fileType" firestore:"fileType"`
	FileSize int64  `json:"fileSize" firestore:"fileSize"`
	PageID   string `json:"pageId" firestore:"pageId"`
}

// PageWithSections represents a page with its sections
type PageWithSections struct {
	Page
	Sections []Section `json:"sections" firestore:"-"`
}
