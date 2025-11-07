package repos

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/internal/models"
	"google.golang.org/api/iterator"
)

// UserAccountLinkRepo handles user account link data operations
type UserAccountLinkRepo struct {
	*BaseRepo
}

// NewUserAccountLinkRepo creates a new user account link repository
func NewUserAccountLinkRepo() *UserAccountLinkRepo {
	return &UserAccountLinkRepo{
		BaseRepo: NewBaseRepo("userAccountLinks"),
	}
}

// GetByUserID gets all account links for a user
func (r *UserAccountLinkRepo) GetByUserID(ctx context.Context, userID string) ([]models.UserAccountLink, error) {
	iter := r.collection.Where("userId", "==", userID).Documents(ctx)
	var links []models.UserAccountLink

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		// Convert time fields from strings/timestamps to time.Time
		if err := ConvertTimeFieldsInMap(data, []string{"expiresAt", "linkedAt", "created", "updated"}); err != nil {
			return nil, err
		}

		var link models.UserAccountLink
		if err := doc.DataTo(&link); err != nil {
			return nil, err
		}
		link.ID = doc.Ref.ID
		links = append(links, link)
	}

	return links, nil
}

// GetByProvider gets account link by provider and provider user ID
func (r *UserAccountLinkRepo) GetByProvider(ctx context.Context, provider models.AccountProvider, providerUserID string) (*models.UserAccountLink, error) {
	iter := r.collection.
		Where("provider", "==", string(provider)).
		Where("providerUserId", "==", providerUserID).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	data := doc.Data()
	// Convert time fields from strings/timestamps to time.Time
	if err := ConvertTimeFieldsInMap(data, []string{"expiresAt", "linkedAt", "created", "updated"}); err != nil {
		return nil, err
	}

	var link models.UserAccountLink
	if err := doc.DataTo(&link); err != nil {
		return nil, err
	}
	link.ID = doc.Ref.ID
	return &link, nil
}

// Add creates a new account link
func (r *UserAccountLinkRepo) Add(ctx context.Context, link *models.UserAccountLink) error {
	newDocRef := r.collection.NewDoc()
	link.ID = newDocRef.ID
	link.Created = time.Now()
	link.LinkedAt = time.Now()

	data := map[string]interface{}{
		"id":             link.ID,
		"userId":         link.UserID,
		"provider":       string(link.Provider),
		"providerUserId": link.ProviderUserID,
		"providerEmail":  link.ProviderEmail,
		"isActive":       link.IsActive,
		"linkedAt":       link.LinkedAt,
		"created":        link.Created,
	}
	if link.ProviderDisplayName != nil {
		data["providerDisplayName"] = *link.ProviderDisplayName
	}
	if link.AccessToken != nil {
		data["accessToken"] = *link.AccessToken
	}
	if link.RefreshToken != nil {
		data["refreshToken"] = *link.RefreshToken
	}
	if link.ExpiresAt != nil {
		data["expiresAt"] = *link.ExpiresAt
	}

	_, err := newDocRef.Set(ctx, data)
	return err
}

// Update updates an account link
func (r *UserAccountLinkRepo) Update(ctx context.Context, link *models.UserAccountLink) error {
	now := time.Now()
	link.Updated = &now

	data := map[string]interface{}{
		"id":             link.ID,
		"userId":         link.UserID,
		"provider":       string(link.Provider),
		"providerUserId": link.ProviderUserID,
		"providerEmail":  link.ProviderEmail,
		"isActive":       link.IsActive,
		"updated":        now,
	}
	if link.ProviderDisplayName != nil {
		data["providerDisplayName"] = *link.ProviderDisplayName
	}
	if link.AccessToken != nil {
		data["accessToken"] = *link.AccessToken
	}
	if link.RefreshToken != nil {
		data["refreshToken"] = *link.RefreshToken
	}
	if link.ExpiresAt != nil {
		data["expiresAt"] = *link.ExpiresAt
	}

	_, err := r.collection.Doc(link.ID).Set(ctx, data)
	return err
}

// GetByUserIDAndProvider gets account link by user ID and provider
func (r *UserAccountLinkRepo) GetByUserIDAndProvider(ctx context.Context, userID string, provider models.AccountProvider) (*models.UserAccountLink, error) {
	iter := r.collection.
		Where("userId", "==", userID).
		Where("provider", "==", string(provider)).
		Where("isActive", "==", true).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	data := doc.Data()
	// Convert time fields from strings/timestamps to time.Time
	if err := ConvertTimeFieldsInMap(data, []string{"expiresAt", "linkedAt", "created", "updated"}); err != nil {
		return nil, err
	}

	var link models.UserAccountLink
	if err := doc.DataTo(&link); err != nil {
		return nil, err
	}
	link.ID = doc.Ref.ID
	return &link, nil
}

// Deactivate deactivates an account link
func (r *UserAccountLinkRepo) Deactivate(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.collection.Doc(id).Update(ctx, []firestore.Update{
		{Path: "isActive", Value: false},
		{Path: "updated", Value: now},
	})
	return err
}

// Delete deletes an account link
func (r *UserAccountLinkRepo) Delete(ctx context.Context, id string) error {
	_, err := r.collection.Doc(id).Delete(ctx)
	return err
}
