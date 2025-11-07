package repos

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/internal/models"
	"google.golang.org/api/iterator"
)

// SignupInvitationRepo handles signup invitation data access
type SignupInvitationRepo struct {
	*BaseRepo
}

// NewSignupInvitationRepo creates a new signup invitation repository
func NewSignupInvitationRepo() *SignupInvitationRepo {
	return &SignupInvitationRepo{
		BaseRepo: NewBaseRepo("signupInvitations"),
	}
}

// GetByEmail gets a signup invitation by email
func (r *SignupInvitationRepo) GetByEmail(ctx context.Context, email string) (*models.SignupInvitation, error) {
	iter := r.collection.Where("email", "==", email).Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	data := doc.Data()
	if err := ConvertTimeFieldsInMap(data, []string{"linkExpiry", "created"}); err != nil {
		return nil, err
	}

	var invitation models.SignupInvitation
	if err := doc.DataTo(&invitation); err != nil {
		return nil, err
	}
	invitation.ID = doc.Ref.ID
	return &invitation, nil
}

// GetByToken gets a signup invitation by token
func (r *SignupInvitationRepo) GetByToken(ctx context.Context, token string) (*models.SignupInvitation, error) {
	iter := r.collection.Where("token", "==", token).Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	data := doc.Data()
	if err := ConvertTimeFieldsInMap(data, []string{"linkExpiry", "created"}); err != nil {
		return nil, err
	}

	var invitation models.SignupInvitation
	if err := doc.DataTo(&invitation); err != nil {
		return nil, err
	}
	invitation.ID = doc.Ref.ID
	return &invitation, nil
}

// Add creates a new signup invitation
func (r *SignupInvitationRepo) Add(ctx context.Context, invitation *models.SignupInvitation) error {
	newDocRef := r.collection.NewDoc()
	invitation.ID = newDocRef.ID
	invitation.Created = time.Now()

	data := map[string]interface{}{
		"id":         invitation.ID,
		"email":      invitation.Email,
		"token":      invitation.Token,
		"linkExpiry": invitation.LinkExpiry,
		"hasSignup":  invitation.HasSignup,
		"created":    invitation.Created,
	}

	_, err := newDocRef.Set(ctx, data)
	return err
}

// MarkAsSignedUp marks an invitation as used
func (r *SignupInvitationRepo) MarkAsSignedUp(ctx context.Context, id string) error {
	_, err := r.collection.Doc(id).Update(ctx, []firestore.Update{
		{Path: "hasSignup", Value: true},
	})
	return err
}
