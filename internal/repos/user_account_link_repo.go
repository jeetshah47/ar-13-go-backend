package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserAccountLinkRepo handles user account link data operations with MongoDB
type UserAccountLinkRepo struct {
	*MongoBaseRepo
}

// NewUserAccountLinkRepo creates a new MongoDB user account link repository
func NewUserAccountLinkRepo() *UserAccountLinkRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &UserAccountLinkRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "user_account_links"),
	}
}

// GetByUserID gets all account links for a user
func (r *UserAccountLinkRepo) GetByUserID(ctx context.Context, userID string) ([]models.UserAccountLink, error) {
	filter := bson.M{"userId": userID}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	links := make([]models.UserAccountLink, 0, len(items))
	for _, item := range items {
		var link models.UserAccountLink
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &link); err != nil {
			continue
		}
		links = append(links, link)
	}

	return links, nil
}

// GetByProvider gets account link by provider and provider user ID
func (r *UserAccountLinkRepo) GetByProvider(ctx context.Context, provider models.AccountProvider, providerUserID string) (*models.UserAccountLink, error) {
	filter := bson.M{
		"provider":       string(provider),
		"providerUserId": providerUserID,
	}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var link models.UserAccountLink
	if err := result.Decode(&link); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account link: %w", err)
	}

	return &link, nil
}

// Add creates a new account link
func (r *UserAccountLinkRepo) Add(ctx context.Context, link *models.UserAccountLink) error {
	now := time.Now()
	if link.ID == "" {
		link.ID = uuid.New().String()
	}
	link.Created = now
	link.LinkedAt = now

	return r.InsertOne(ctx, link)
}

// Update updates an account link
func (r *UserAccountLinkRepo) Update(ctx context.Context, link *models.UserAccountLink) error {
	now := time.Now()
	link.Updated = &now

	updates := bson.M{
		"userId":         link.UserID,
		"provider":       string(link.Provider),
		"providerUserId": link.ProviderUserID,
		"providerEmail":  link.ProviderEmail,
		"isActive":       link.IsActive,
		"updated":        link.Updated,
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
		updates["expiresAt"] = link.ExpiresAt
	}

	return r.UpdateOne(ctx, link.ID, updates)
}

// GetByUserIDAndProvider gets account link by user ID and provider
func (r *UserAccountLinkRepo) GetByUserIDAndProvider(ctx context.Context, userID string, provider models.AccountProvider) (*models.UserAccountLink, error) {
	filter := bson.M{
		"userId":   userID,
		"provider": string(provider),
		"isActive": true,
	}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var link models.UserAccountLink
	if err := result.Decode(&link); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account link: %w", err)
	}

	return &link, nil
}

// Deactivate deactivates an account link
func (r *UserAccountLinkRepo) Deactivate(ctx context.Context, id string) error {
	updates := bson.M{"isActive": false}
	return r.UpdateOne(ctx, id, updates)
}

// Delete deletes an account link
func (r *UserAccountLinkRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

