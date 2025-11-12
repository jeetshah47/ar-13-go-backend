package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
)

// RolePermissionRepo handles role permission data operations
type RolePermissionRepo struct {
	*DynamoBaseRepo
}

// NewRolePermissionRepo creates a new role permission repository
func NewRolePermissionRepo() *RolePermissionRepo {
	return &RolePermissionRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("role_permissions"),
	}
}

// GetByRole gets all permissions for a specific role
func (r *RolePermissionRepo) GetByRole(ctx context.Context, role models.UserRole) ([]models.RolePermission, error) {
	items, err := r.QueryByIndex(ctx, "role-index", "role", string(role))
	if err != nil {
		return nil, err
	}

	permissions := make([]models.RolePermission, 0, len(items))
	for _, item := range items {
		var rp models.RolePermission
		if err := UnmarshalItem(item, &rp); err != nil {
			return nil, err
		}
		permissions = append(permissions, rp)
	}

	return permissions, nil
}

// GetByID gets a role permission by ID
func (r *RolePermissionRepo) GetByID(ctx context.Context, id string) (*models.RolePermission, error) {
	item, err := r.DynamoBaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var rp models.RolePermission
	if err := UnmarshalItem(item, &rp); err != nil {
		return nil, err
	}

	return &rp, nil
}

// GetAll gets all role permissions
func (r *RolePermissionRepo) GetAll(ctx context.Context) ([]models.RolePermission, error) {
	items, err := r.ScanItems(ctx, nil)
	if err != nil {
		return nil, err
	}

	permissions := make([]models.RolePermission, 0, len(items))
	for _, item := range items {
		var rp models.RolePermission
		if err := UnmarshalItem(item, &rp); err != nil {
			return nil, err
		}
		permissions = append(permissions, rp)
	}

	return permissions, nil
}

// Add creates a new role permission
func (r *RolePermissionRepo) Add(ctx context.Context, rp *models.RolePermission) error {
	// Set timestamps
	now := time.Now()
	if rp.CreatedAt.IsZero() {
		rp.CreatedAt = now
	}
	rp.UpdatedAt = now

	// Prepare data for DynamoDB
	rpData := map[string]interface{}{
		"id":         rp.ID,
		"role":       string(rp.Role),
		"permission": rp.Permission,
		"createdAt":  rp.CreatedAt.Format(time.RFC3339),
		"updatedAt":  rp.UpdatedAt.Format(time.RFC3339),
	}

	return r.PutItem(ctx, rpData)
}

// BatchAdd creates multiple role permissions in a single batch operation
func (r *RolePermissionRepo) BatchAdd(ctx context.Context, permissions []models.RolePermission) error {
	now := time.Now()
	items := make([]interface{}, len(permissions))

	for i, rp := range permissions {
		if rp.CreatedAt.IsZero() {
			rp.CreatedAt = now
		}
		rp.UpdatedAt = now

		rpData := map[string]interface{}{
			"id":         rp.ID,
			"role":       string(rp.Role),
			"permission": rp.Permission,
			"createdAt":  rp.CreatedAt.Format(time.RFC3339),
			"updatedAt":  rp.UpdatedAt.Format(time.RFC3339),
		}
		items[i] = rpData
	}

	return r.BatchWriteItems(ctx, items)
}

// Delete deletes a role permission by ID
func (r *RolePermissionRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

// DeleteByRoleAndPermission deletes a specific role-permission mapping
func (r *RolePermissionRepo) DeleteByRoleAndPermission(ctx context.Context, role models.UserRole, permission string) error {
	// First, find the permission by role and permission string
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

