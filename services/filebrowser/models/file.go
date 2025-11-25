package models

import "time"

// FileItem represents a file or folder in the directory listing
type FileItem struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsFolder  bool      `json:"isFolder"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
	MimeType  string    `json:"mimeType,omitempty"`
}

// BrowseResponse represents the response for the browse endpoint
type BrowseResponse struct {
	Path  string     `json:"path"`
	Files []FileItem `json:"files"`
}

// FileInfoResponse represents the response for the file info endpoint
type FileInfoResponse struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsFolder  bool      `json:"isFolder"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
	MimeType  string    `json:"mimeType,omitempty"`
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

