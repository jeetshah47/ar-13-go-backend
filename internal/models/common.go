package models

import "time"

// Model represents the base model interface
// Using bson:",inline" to flatten embedded struct fields to root level in MongoDB
type Model struct {
	ID      string     `json:"id" firestore:"id" bson:"id,omitempty"`
	Created time.Time  `json:"created" firestore:"created" bson:"created,omitempty"`
	Updated *time.Time `json:"updated,omitempty" firestore:"updated,omitempty" bson:"updated,omitempty"`
}
