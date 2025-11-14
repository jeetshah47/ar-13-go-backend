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

// SignupInvitationRepo handles signup invitation data operations with MongoDB
type SignupInvitationRepo struct {
	*MongoBaseRepo
}

// NewSignupInvitationRepo creates a new MongoDB signup invitation repository
func NewSignupInvitationRepo() *SignupInvitationRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &SignupInvitationRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "signupInvitations"),
	}
}

// GetByEmail gets a signup invitation by email
func (r *SignupInvitationRepo) GetByEmail(ctx context.Context, email string) (*models.SignupInvitation, error) {
	filter := bson.M{"email": email}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var invitation models.SignupInvitation
	if err := result.Decode(&invitation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invitation: %w", err)
	}

	return &invitation, nil
}

// GetByToken gets a signup invitation by token
func (r *SignupInvitationRepo) GetByToken(ctx context.Context, token string) (*models.SignupInvitation, error) {
	filter := bson.M{"token": token}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var invitation models.SignupInvitation
	if err := result.Decode(&invitation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invitation: %w", err)
	}

	return &invitation, nil
}

// Add creates a new signup invitation
func (r *SignupInvitationRepo) Add(ctx context.Context, invitation *models.SignupInvitation) error {
	now := time.Now()
	if invitation.ID == "" {
		invitation.ID = uuid.New().String()
	}
	invitation.Created = now

	return r.InsertOne(ctx, invitation)
}

// MarkAsSignedUp marks an invitation as used
func (r *SignupInvitationRepo) MarkAsSignedUp(ctx context.Context, id string) error {
	updates := bson.M{"hasSignup": true}
	return r.UpdateOne(ctx, id, updates)
}

