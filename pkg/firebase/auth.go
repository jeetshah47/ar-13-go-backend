package firebase

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

// AuthClient wraps Firebase Auth client for easier usage
type AuthClient struct {
	client *auth.Client
}

// UserToCreate represents a user to be created
type UserToCreate struct {
	email       string
	password    string
	displayName string
}

// NewUserToCreate creates a new UserToCreate
func NewUserToCreate() *UserToCreate {
	return &UserToCreate{}
}

// Email sets the email
func (u *UserToCreate) Email(email string) *UserToCreate {
	u.email = email
	return u
}

// Password sets the password
func (u *UserToCreate) Password(password string) *UserToCreate {
	u.password = password
	return u
}

// DisplayName sets the display name
func (u *UserToCreate) DisplayName(name string) *UserToCreate {
	u.displayName = name
	return u
}

// CreateUser creates a Firebase Auth user
func (a *AuthClient) CreateUser(ctx context.Context, user *UserToCreate) (*auth.UserRecord, error) {
	params := (&auth.UserToCreate{}).
		Email(user.email).
		Password(user.password).
		DisplayName(user.displayName)
	return a.client.CreateUser(ctx, params)
}

// DeleteUser deletes a Firebase Auth user
func (a *AuthClient) DeleteUser(ctx context.Context, uid string) error {
	return a.client.DeleteUser(ctx, uid)
}

// VerifyIDToken verifies an ID token
func (a *AuthClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return a.client.VerifyIDToken(ctx, idToken)
}
