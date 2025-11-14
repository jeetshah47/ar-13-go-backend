package models

// ProjectDetails represents project details
type ProjectDetails struct {
	Model
	ProjectID string                 `json:"projectId" firestore:"projectId" bson:"projectId"`
	Data      map[string]interface{} `json:"data" firestore:"data" bson:"data"` // Flexible data structure
}
