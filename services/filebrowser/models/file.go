package models

import "time"

// FileItem represents a file or folder in the directory listing
type FileItem struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsFolder bool      `json:"isFolder"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	MimeType string    `json:"mimeType,omitempty"`
}

// BrowseResponse represents the response for the browse endpoint
type BrowseResponse struct {
	Path  string     `json:"path"`
	Files []FileItem `json:"files"`
}

// FileInfoResponse represents the response for the file info endpoint
type FileInfoResponse struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsFolder bool      `json:"isFolder"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	MimeType string    `json:"mimeType,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// AccessResponse represents the response for the access endpoint
type AccessResponse struct {
	URL       string    `json:"url"`       // Time-limited download URL
	Path      string    `json:"path"`      // File path
	ExpiresAt time.Time `json:"expiresAt"` // Expiry time
}

// CreateFolderRequest represents the request for creating a folder
type CreateFolderRequest struct {
	FolderName string `json:"folderName" binding:"required"`
}

// CreateFolderResponse represents the response for creating a folder
type CreateFolderResponse struct {
	Path      string    `json:"path"`      // Created folder path
	CreatedAt time.Time `json:"createdAt"` // Creation timestamp
}

// UploadResponse represents the response for uploading a file
type UploadResponse struct {
	Name       string    `json:"name"`       // File name
	Path       string    `json:"path"`       // File path
	Size       int64     `json:"size"`       // File size in bytes
	MimeType   string    `json:"mimeType"`   // MIME type
	UploadedAt time.Time `json:"uploadedAt"` // Upload timestamp
}

// DeleteResponse represents the response for deleting a file or folder
type DeleteResponse struct {
	Path      string    `json:"path"`      // Deleted file/folder path
	DeletedAt time.Time `json:"deletedAt"` // Deletion timestamp
}

// RenameRequest represents the request for renaming a file or folder
type RenameRequest struct {
	NewName string `json:"newName" binding:"required"`
}

// RenameResponse represents the response for renaming a file or folder
type RenameResponse struct {
	Name     string    `json:"name"`               // New file/folder name
	Path     string    `json:"path"`               // New file/folder path
	IsFolder bool      `json:"isFolder"`           // Whether it's a folder
	Size     int64     `json:"size"`               // File size (0 for folders)
	Modified time.Time `json:"modified"`           // Modification timestamp
	MimeType string    `json:"mimeType,omitempty"` // MIME type (for files only)
}
