package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
)

// UserRepo handles user data operations
type UserRepo struct {
	*DynamoBaseRepo
}

// NewUserRepo creates a new user repository
func NewUserRepo() *UserRepo {
	return &UserRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("users"),
	}
}

// GetByEmail gets a user by email using GSI
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	items, err := r.QueryByIndex(ctx, "email-index", "email", email)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	var user models.User
	if err := UnmarshalItem(items[0], &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByID gets a user by ID
func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	item, err := r.DynamoBaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var user models.User
	if err := UnmarshalItem(item, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAll gets all users
func (r *UserRepo) GetAll(ctx context.Context, limit *int) ([]models.User, error) {
	var limitInt32 *int32
	if limit != nil {
		l := int32(*limit)
		limitInt32 = &l
	}

	items, err := r.ScanItems(ctx, limitInt32)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0, len(items))
	for _, item := range items {
		var user models.User
		if err := UnmarshalItem(item, &user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// Add creates a new user
func (r *UserRepo) Add(ctx context.Context, user *models.User) error {
	// Set timestamps
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	// Prepare user data for DynamoDB
	userData := map[string]interface{}{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"phoneNumber": user.PhoneNumber,
		"role":        string(user.Role),
		"password":    user.Password, // Already hashed
		"createdAt":   user.CreatedAt.Format(time.RFC3339),
		"updatedAt":   user.UpdatedAt.Format(time.RFC3339),
	}

	if user.Designation != nil {
		userData["designation"] = *user.Designation
	}

	return r.PutItem(ctx, userData)
}

// Update updates a user
func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	updates := map[string]interface{}{
		"name":        user.Name,
		"email":       user.Email,
		"phoneNumber": user.PhoneNumber,
		"role":        string(user.Role),
	}

	if user.Designation != nil {
		updates["designation"] = *user.Designation
	}

	if user.Password != "" {
		updates["password"] = user.Password // Already hashed
	}

	return r.UpdateItem(ctx, user.ID, updates)
}

// Delete deletes a user
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// Persists checks if a user exists
func (r *UserRepo) Persists(ctx context.Context, id string) (bool, error) {
	return r.Exists(ctx, id)
}
