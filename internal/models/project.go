package models

// AgencyContact represents agency contact information
type AgencyContact struct {
	ContactName       string `json:"contact_name" firestore:"contact_name" bson:"contact_name"`
	ContactAgencyType string `json:"contact_agency_type" firestore:"contact_agency_type" bson:"contact_agency_type"`
	PhoneNumber       string `json:"phone_number" firestore:"phone_number" bson:"phone_number"`
	FirmName          string `json:"firm_name" firestore:"firm_name" bson:"firm_name"`
}

// Project represents a project
type Project struct {
	Model
	Title              string         `json:"title" firestore:"title" bson:"title"`
	Description        string         `json:"description" firestore:"description" bson:"description"`
	OwnerID            string         `json:"ownerId" firestore:"ownerId" bson:"ownerId"`
	MembersIDs         []string       `json:"membersIds" firestore:"membersIds" bson:"membersIds"`
	ProductionDuration *int           `json:"productionDuration,omitempty" firestore:"productionDuration,omitempty" bson:"productionDuration,omitempty"`
	SiteDuration       *int           `json:"siteDuration,omitempty" firestore:"siteDuration,omitempty" bson:"siteDuration,omitempty"`
	LogoURL       *string        `json:"logoUrl,omitempty" firestore:"logoUrl,omitempty" bson:"logoUrl,omitempty"`
	Code          string         `json:"code" firestore:"code" bson:"project_code"`
	AgencyContact *AgencyContact `json:"agencyContact,omitempty" firestore:"agencyContact,omitempty" bson:"agencyContact,omitempty"`
	IsArchived    bool           `json:"isArchived" firestore:"isArchived" bson:"isArchived"`
}
