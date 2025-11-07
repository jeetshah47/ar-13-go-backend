package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// UserAccountLinkRepo handles user account link data operations
type UserAccountLinkRepo struct {
	*DynamoBaseRepo
}

// NewUserAccountLinkRepo creates a new user account link repository
func NewUserAccountLinkRepo() *UserAccountLinkRepo {
	return &UserAccountLinkRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("userAccountLinks"),
	}
}

// GetByUserID gets all account links for a user
func (r *UserAccountLinkRepo) GetByUserID(ctx context.Context, userID string) ([]models.UserAccountLink, error) {
	items, err := r.QueryByIndex(ctx, "userId-index", "userId", userID)
	if err != nil {
		return nil, err
	}

	links := make([]models.UserAccountLink, 0, len(items))
	for _, item := range items {
		var link models.UserAccountLink
		if err := UnmarshalItem(item, &link); err != nil {
			return nil, fmt.Errorf("failed to unmarshal account link: %w", err)
		}
		links = append(links, link)
	}

	return links, nil
}

// GetByProvider gets account link by provider and provider user ID
func (r *UserAccountLinkRepo) GetByProvider(ctx context.Context, provider models.AccountProvider, providerUserID string) (*models.UserAccountLink, error) {
	items, err := r.ScanItems(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		p, ok1 := item["provider"].(*types.AttributeValueMemberS)
		pu, ok2 := item["providerUserId"].(*types.AttributeValueMemberS)
		if ok1 && ok2 && p.Value == string(provider) && pu.Value == providerUserID {
			var link models.UserAccountLink
			if err := UnmarshalItem(item, &link); err != nil {
				return nil, fmt.Errorf("failed to unmarshal account link: %w", err)
			}
			return &link, nil
		}
	}

	return nil, nil
}

// Add creates a new account link
func (r *UserAccountLinkRepo) Add(ctx context.Context, link *models.UserAccountLink) error {
	now := time.Now()
	if link.ID == "" {
		link.ID = uuid.New().String()
	}
	link.Created = now
	link.LinkedAt = now

	data := map[string]interface{}{
		"id":             link.ID,
		"userId":         link.UserID,
		"provider":       string(link.Provider),
		"providerUserId": link.ProviderUserID,
		"providerEmail":  link.ProviderEmail,
		"isActive":       link.IsActive,
		"linkedAt":       link.LinkedAt.Format(time.RFC3339),
		"created":        link.Created.Format(time.RFC3339),
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
		data["expiresAt"] = link.ExpiresAt.Format(time.RFC3339)
	}

	return r.PutItem(ctx, data)
}

// Update updates an account link
func (r *UserAccountLinkRepo) Update(ctx context.Context, link *models.UserAccountLink) error {
	now := time.Now()
	link.Updated = &now

	updates := map[string]interface{}{
		"userId":         link.UserID,
		"provider":       string(link.Provider),
		"providerUserId": link.ProviderUserID,
		"providerEmail":  link.ProviderEmail,
		"isActive":       link.IsActive,
		"updated":        link.Updated.Format(time.RFC3339),
	}

	if link.ProviderDisplayName != nil {
		updates["providerDisplayName"] = *link.ProviderDisplayName
	}
	if link.AccessToken != nil {
		updates["accessToken"] = *link.AccessToken
	}
	if link.RefreshToken != nil {
		updates["refreshToken"] = *link.RefreshToken
	}
	if link.ExpiresAt != nil {
		updates["expiresAt"] = link.ExpiresAt.Format(time.RFC3339)
	}

	return r.UpdateItem(ctx, link.ID, updates)
}

// GetByUserIDAndProvider gets account link by user ID and provider
func (r *UserAccountLinkRepo) GetByUserIDAndProvider(ctx context.Context, userID string, provider models.AccountProvider) (*models.UserAccountLink, error) {
	items, err := r.QueryByIndex(ctx, "userId-index", "userId", userID)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if p, ok := item["provider"].(*types.AttributeValueMemberS); ok && p.Value == string(provider) {
			if isActive, ok := item["isActive"].(*types.AttributeValueMemberBOOL); ok && isActive.Value {
				var link models.UserAccountLink
				if err := UnmarshalItem(item, &link); err != nil {
					return nil, fmt.Errorf("failed to unmarshal account link: %w", err)
				}
				return &link, nil
			}
		}
	}

	return nil, nil
}

// Deactivate deactivates an account link
func (r *UserAccountLinkRepo) Deactivate(ctx context.Context, id string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"isActive": false,
		"updated":  now.Format(time.RFC3339),
	}
	return r.UpdateItem(ctx, id, updates)
}

// Delete deletes an account link
func (r *UserAccountLinkRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
