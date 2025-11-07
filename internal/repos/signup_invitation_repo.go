package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/google/uuid"
)

// SignupInvitationRepo handles signup invitation data access
type SignupInvitationRepo struct {
	*DynamoBaseRepo
}

// NewSignupInvitationRepo creates a new signup invitation repository
func NewSignupInvitationRepo() *SignupInvitationRepo {
	return &SignupInvitationRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("signupInvitations"),
	}
}

// GetByEmail gets a signup invitation by email
func (r *SignupInvitationRepo) GetByEmail(ctx context.Context, email string) (*models.SignupInvitation, error) {
	items, err := r.QueryByIndex(ctx, "email-index", "email", email)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	var invitation models.SignupInvitation
	if err := UnmarshalItem(items[0], &invitation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invitation: %w", err)
	}

	return &invitation, nil
}

// GetByToken gets a signup invitation by token
func (r *SignupInvitationRepo) GetByToken(ctx context.Context, token string) (*models.SignupInvitation, error) {
	items, err := r.QueryByIndex(ctx, "token-index", "token", token)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	var invitation models.SignupInvitation
	if err := UnmarshalItem(items[0], &invitation); err != nil {
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

	data := map[string]interface{}{
		"id":         invitation.ID,
		"email":      invitation.Email,
		"token":      invitation.Token,
		"linkExpiry": invitation.LinkExpiry.Format(time.RFC3339),
		"hasSignup":  invitation.HasSignup,
		"created":    invitation.Created.Format(time.RFC3339),
	}

	return r.PutItem(ctx, data)
}

// MarkAsSignedUp marks an invitation as used
func (r *SignupInvitationRepo) MarkAsSignedUp(ctx context.Context, id string) error {
	updates := map[string]interface{}{
		"hasSignup": true,
	}
	return r.UpdateItem(ctx, id, updates)
}
