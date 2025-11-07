package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/firebase"
	"google.golang.org/api/iterator"
)

// UserRepo handles user data operations
type UserRepo struct {
	*BaseRepo
	authClient *firebase.AuthClient
}

// NewUserRepo creates a new user repository
func NewUserRepo() *UserRepo {
	return &UserRepo{
		BaseRepo:   NewBaseRepo("users"),
		authClient: firebase.GetFirebaseClient(),
	}
}

// GetByEmail gets a user by email
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	iter := r.collection.Where("email", "==", email).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	data := doc.Data()
	// Convert time fields from strings/timestamps to time.Time
	if err := ConvertTimeFieldsInMap(data, []string{"createdAt", "updatedAt", "created", "updated"}); err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	user.ID = doc.Ref.ID
	return &user, nil
}

// GetByID gets a user by ID
func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	doc, err := r.collection.Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, nil
	}

	data := doc.Data()
	// Convert time fields from strings/timestamps to time.Time
	if err := ConvertTimeFieldsInMap(data, []string{"createdAt", "updatedAt", "created", "updated"}); err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	user.ID = doc.Ref.ID
	return &user, nil
}

// GetAll gets all users
func (r *UserRepo) GetAll(ctx context.Context, limit *int) ([]models.User, error) {
	query := r.collection.Query
	if limit != nil {
		query = query.Limit(*limit)
	}

	iter := query.Documents(ctx)
	var users []models.User

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
		if err := ConvertTimeFieldsInMap(data, []string{"createdAt", "updatedAt", "created", "updated"}); err != nil {
			return nil, err
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return nil, err
		}
		user.ID = doc.Ref.ID
		users = append(users, user)
	}

	return users, nil
}

// Add creates a new user
func (r *UserRepo) Add(ctx context.Context, user *models.User) error {
	// Create Firebase Auth user
	authUser, err := r.authClient.CreateUser(ctx, firebase.NewUserToCreate().
		Email(user.Email).
		Password(user.Password).
		DisplayName(user.Name))
	if err != nil {
		return err
	}

	// Create Firestore document
	user.ID = authUser.UID
	user.CreatedAt = time.Now()

	data := map[string]interface{}{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"phoneNumber": user.PhoneNumber,
		"role":        string(user.Role),
		"password":    user.Password,
		"createdAt":   user.CreatedAt,
	}
	if user.Designation != nil {
		data["designation"] = *user.Designation
	}

	_, err = r.collection.Doc(user.ID).Set(ctx, data)
	return err
}

// Update updates a user
func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	now := time.Now()
	user.UpdatedAt = now

	data := map[string]interface{}{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"phoneNumber": user.PhoneNumber,
		"role":        string(user.Role),
		"updatedAt":   now,
	}
	if user.Designation != nil {
		data["designation"] = *user.Designation
	}
	if user.Password != "" {
		data["password"] = user.Password
	}

	_, err := r.collection.Doc(user.ID).Set(ctx, data)
	return err
}

// Delete deletes a user
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	// Delete from Firestore
	if _, err := r.collection.Doc(id).Delete(ctx); err != nil {
		return err
	}
	// Delete from Auth
	return r.authClient.DeleteUser(ctx, id)
}

// Persists checks if a user exists
func (r *UserRepo) Persists(ctx context.Context, id string) (bool, error) {
	return r.Exists(ctx, id)
}
