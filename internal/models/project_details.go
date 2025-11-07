package models

// ProjectDetails represents project details
type ProjectDetails struct {
	Model
	ProjectID string                 `json:"projectId" firestore:"projectId"`
	Data      map[string]interface{} `json:"data" firestore:"data"` // Flexible data structure
}
