package models

import "time"

// SignupInvitation represents a signup invitation
type SignupInvitation struct {
	ID         string    `json:"id" firestore:"id" bson:"id"`
	Email      string    `json:"email" firestore:"email" bson:"email"`
	Token      string    `json:"token" firestore:"token" bson:"token"`
	LinkExpiry time.Time `json:"linkExpiry" firestore:"linkExpiry" bson:"linkExpiry"`
	HasSignup  bool      `json:"hasSignup" firestore:"hasSignup" bson:"hasSignup"`
	Created    time.Time `json:"created" firestore:"created" bson:"created"`
}
