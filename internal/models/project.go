package models

import "time"

// AgencyContact represents agency contact information
type AgencyContact struct {
	ContactName       string `json:"contact_name" firestore:"contact_name" bson:"contact_name"`
	ContactAgencyType string `json:"contact_agency_type" firestore:"contact_agency_type" bson:"contact_agency_type"`
}

// Project represents a project
type Project struct {
	Model
	Title         string         `json:"title" firestore:"title" bson:"title"`
	Description   string         `json:"description" firestore:"description" bson:"description"`
	OwnerID       string         `json:"ownerId" firestore:"ownerId" bson:"ownerId"`
	MembersIDs    []string       `json:"membersIds" firestore:"membersIds" bson:"membersIds"`
	StartDate     *time.Time     `json:"startDate,omitempty" firestore:"startDate,omitempty" bson:"startDate,omitempty"`
	EndDate       *time.Time     `json:"endDate,omitempty" firestore:"endDate,omitempty" bson:"endDate,omitempty"`
	Deadline      time.Time      `json:"deadLine" firestore:"deadLine" bson:"deadLine"`
	LogoURL       *string        `json:"logoUrl,omitempty" firestore:"logoUrl,omitempty" bson:"logoUrl,omitempty"`
	Code          string         `json:"code" firestore:"code" bson:"code"`
	AgencyContact *AgencyContact `json:"agencyContact,omitempty" firestore:"agencyContact,omitempty" bson:"agencyContact,omitempty"`
}
