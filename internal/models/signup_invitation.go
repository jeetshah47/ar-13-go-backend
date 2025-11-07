package models

import "time"

// SignupInvitation represents a signup invitation
type SignupInvitation struct {
	ID         string    `json:"id" firestore:"id"`
	Email      string    `json:"email" firestore:"email"`
	Token      string    `json:"token" firestore:"token"`
	LinkExpiry time.Time `json:"linkExpiry" firestore:"linkExpiry"`
	HasSignup  bool      `json:"hasSignup" firestore:"hasSignup"`
	Created    time.Time `json:"created" firestore:"created"`
}
