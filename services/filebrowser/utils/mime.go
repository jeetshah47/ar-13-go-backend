package utils

import (
	"mime"
	"path/filepath"
)

// GetMimeType returns the MIME type for a file based on its extension
func GetMimeType(filePath string) string {
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

