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

// RolePermissionRepo handles role permission data operations with MongoDB
type RolePermissionRepo struct {
	*MongoBaseRepo
}

// NewRolePermissionRepo creates a new MongoDB role permission repository
func NewRolePermissionRepo() *RolePermissionRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &RolePermissionRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "role_permissions"),
	}
}

// GetByRole gets all permissions for a specific role
func (r *RolePermissionRepo) GetByRole(ctx context.Context, role models.UserRole) ([]models.RolePermission, error) {
	filter := bson.M{"role": string(role)}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	permissions := make([]models.RolePermission, 0, len(items))
	for _, item := range items {
		var rp models.RolePermission
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &rp); err != nil {
			continue
		}
		permissions = append(permissions, rp)
	}

	return permissions, nil
}

// GetByID gets a role permission by ID
func (r *RolePermissionRepo) GetByID(ctx context.Context, id string) (*models.RolePermission, error) {
	result := r.MongoBaseRepo.GetByID(ctx, id)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var rp models.RolePermission
	if err := result.Decode(&rp); err != nil {
		return nil, err
	}

	return &rp, nil
}

// GetAll gets all role permissions
func (r *RolePermissionRepo) GetAll(ctx context.Context) ([]models.RolePermission, error) {
	items, err := r.FindAll(ctx, bson.M{}, nil)
	if err != nil {
		return nil, err
	}

	permissions := make([]models.RolePermission, 0, len(items))
	for _, item := range items {
		var rp models.RolePermission
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &rp); err != nil {
			continue
		}
		permissions = append(permissions, rp)
	}

	return permissions, nil
}

// Add creates a new role permission
func (r *RolePermissionRepo) Add(ctx context.Context, rp *models.RolePermission) error {
	now := time.Now()
	if rp.ID == "" {
		rp.ID = uuid.New().String()
	}
	if rp.CreatedAt.IsZero() {
		rp.CreatedAt = now
	}
	rp.UpdatedAt = now

	return r.InsertOne(ctx, rp)
}

// BatchAdd creates multiple role permissions in a single batch operation
func (r *RolePermissionRepo) BatchAdd(ctx context.Context, permissions []models.RolePermission) error {
	now := time.Now()
	items := make([]interface{}, len(permissions))

	for i, rp := range permissions {
		if rp.ID == "" {
			rp.ID = uuid.New().String()
		}
		if rp.CreatedAt.IsZero() {
			rp.CreatedAt = now
		}
		rp.UpdatedAt = now
		items[i] = rp
	}

	return r.BatchInsertMany(ctx, items)
}

// Delete deletes a role permission by ID
func (r *RolePermissionRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// DeleteByRoleAndPermission deletes a specific role-permission mapping
func (r *RolePermissionRepo) DeleteByRoleAndPermission(ctx context.Context, role models.UserRole, permission string) error {
	permissions, err := r.GetByRole(ctx, role)
	if err != nil {
		return err
	}

	for _, rp := range permissions {
		if rp.Permission == permission {
			return r.Delete(ctx, rp.ID)
		}
	}

	return fmt.Errorf("role permission not found: role=%s, permission=%s", role, permission)
}

// HasPermission checks if a role has a specific permission
func (r *RolePermissionRepo) HasPermission(ctx context.Context, role models.UserRole, permission string) (bool, error) {
	permissions, err := r.GetByRole(ctx, role)
	if err != nil {
		return false, err
	}

	for _, rp := range permissions {
		if rp.Permission == permission {
			return true, nil
		}
	}

	return false, nil
}

