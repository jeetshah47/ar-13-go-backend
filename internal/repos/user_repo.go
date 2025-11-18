package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepo handles user data operations with MongoDB
type UserRepo struct {
	*MongoBaseRepo
}

// NewUserRepo creates a new MongoDB user repository
func NewUserRepo() *UserRepo {
	client := mongodb.GetClient()
	if client == nil {
		panic("MongoDB client is not initialized. Please ensure MongoDB is connected before creating repositories.")
	}
	return &UserRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, mongodb.GetDatabaseName(), "users"),
	}
}

// GetByEmail gets a user by email
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	filter := bson.M{"email": email}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var user models.User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByID gets a user by ID
func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	result := r.MongoBaseRepo.GetByID(ctx, id)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var user models.User
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAll gets all users
func (r *UserRepo) GetAll(ctx context.Context, limit *int) ([]models.User, error) {
	var limitInt64 *int64
	if limit != nil {
		l := int64(*limit)
		limitInt64 = &l
	}

	items, err := r.FindAll(ctx, bson.M{}, limitInt64)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0, len(items))
	for _, item := range items {
		var user models.User
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &user); err != nil {
			continue // Skip invalid items
		}
		users = append(users, user)
	}

	return users, nil
}

// Add creates a new user
func (r *UserRepo) Add(ctx context.Context, user *models.User) error {
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	// MongoDB will automatically handle BSON marshaling
	return r.InsertOne(ctx, user)
}

// Update updates a user
func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	updates := bson.M{}

	// Only update fields that are provided (non-empty)
	if user.Name != "" {
		updates["name"] = user.Name
	}
	if user.Email != "" {
		updates["email"] = user.Email
	}
	if user.PhoneNumber != "" {
		updates["phoneNumber"] = user.PhoneNumber
	}
	if user.Role != "" {
		updates["role"] = string(user.Role)
	}

	if user.Designation != nil {
		updates["designation"] = *user.Designation
	}

	if user.Password != "" {
		updates["password"] = user.Password
	}

	return r.UpdateOne(ctx, user.ID, updates)
}

// Delete deletes a user
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// Persists checks if a user exists
func (r *UserRepo) Persists(ctx context.Context, id string) (bool, error) {
	return r.Exists(ctx, id)
}

// BatchGetItems retrieves multiple users by IDs
func (r *UserRepo) BatchGetItems(ctx context.Context, ids []string) (map[string]*models.User, error) {
	items, err := r.MongoBaseRepo.BatchGetItems(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.User)
	for _, item := range items {
		var user models.User
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &user); err != nil {
			continue
		}
		result[user.ID] = &user
	}

	return result, nil
}

