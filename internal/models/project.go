package models

import "time"

// Project represents a project
type Project struct {
	Model
	Title       string    `json:"title" firestore:"title"`
	Description string    `json:"description" firestore:"description"`
	OwnerID     string    `json:"ownerId" firestore:"ownerId"`
	MembersIDs  []string  `json:"membersIds" firestore:"membersIds"`
	Deadline    time.Time `json:"deadLine" firestore:"deadLine"`
	LogoURL     *string   `json:"logoUrl,omitempty" firestore:"logoUrl,omitempty"`
}
