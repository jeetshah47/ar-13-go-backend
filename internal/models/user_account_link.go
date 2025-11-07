package models

import "time"

// AccountProvider represents the account provider type
type AccountProvider string

const (
	AccountProviderGoogle    AccountProvider = "google"
	AccountProviderMicrosoft AccountProvider = "microsoft"
	AccountProviderGithub    AccountProvider = "github"
)

// UserAccountLink represents a link between a user and an external account
type UserAccountLink struct {
	Model
	UserID              string          `json:"userId" firestore:"userId"`
	Provider            AccountProvider `json:"provider" firestore:"provider"`
	ProviderUserID      string          `json:"providerUserId" firestore:"providerUserId"`
	ProviderEmail       string          `json:"providerEmail" firestore:"providerEmail"`
	ProviderDisplayName *string         `json:"providerDisplayName,omitempty" firestore:"providerDisplayName,omitempty"`
	AccessToken         *string         `json:"-" firestore:"accessToken,omitempty"`  // Hidden from JSON
	RefreshToken        *string         `json:"-" firestore:"refreshToken,omitempty"` // Hidden from JSON
	ExpiresAt           *time.Time      `json:"expiresAt,omitempty" firestore:"expiresAt,omitempty"`
	IsActive            bool            `json:"isActive" firestore:"isActive"`
	LinkedAt            time.Time       `json:"linkedAt" firestore:"linkedAt"`
}
