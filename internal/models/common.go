package models

import "time"

// Model represents the base model interface
type Model struct {
	ID      string     `json:"id" firestore:"id"`
	Created time.Time  `json:"created" firestore:"created"`
	Updated *time.Time `json:"updated,omitempty" firestore:"updated,omitempty"`
}
