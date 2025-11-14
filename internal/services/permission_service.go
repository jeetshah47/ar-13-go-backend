package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// PermissionService handles permission business logic with caching
type PermissionService struct {
	rolePermissionRepo *repos.RolePermissionRepo
	cache              map[models.UserRole][]string
	cacheMutex         sync.RWMutex
}

// NewPermissionService creates a new permission service
func NewPermissionService() *PermissionService {
	return &PermissionService{
		rolePermissionRepo: repos.NewRolePermissionRepo(),
		cache:              make(map[models.UserRole][]string),
	}
}

// GetPermissionsByRole returns the list of permissions for a given role
// Uses database with caching, falls back to hardcoded values if database is empty
func (s *PermissionService) GetPermissionsByRole(ctx context.Context, role models.UserRole) ([]string, error) {
	// Check cache first
	s.cacheMutex.RLock()
	if cached, ok := s.cache[role]; ok {
		s.cacheMutex.RUnlock()
		return cached, nil
	}
	s.cacheMutex.RUnlock()

	// Get from database
	permissions, err := s.rolePermissionRepo.GetByRole(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions from database: %w", err)
	}

	// If database has permissions, use them
	if len(permissions) > 0 {
		permissionStrings := make([]string, len(permissions))
		for i, p := range permissions {
			permissionStrings[i] = p.Permission
		}

		// Update cache
		s.cacheMutex.Lock()
		s.cache[role] = permissionStrings
		s.cacheMutex.Unlock()

		return permissionStrings, nil
	}

	// Fallback to hardcoded values (for backward compatibility and initial setup)
	// This will be used until the database is seeded
	return s.getHardcodedPermissions(role), nil
}

// HasPermission checks if a role has a specific permission
func (s *PermissionService) HasPermission(ctx context.Context, role models.UserRole, permission string) (bool, error) {
	permissions, err := s.GetPermissionsByRole(ctx, role)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// InvalidateCache clears the permission cache (useful after updates)
func (s *PermissionService) InvalidateCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.cache = make(map[models.UserRole][]string)
}

// InvalidateRoleCache clears the cache for a specific role
func (s *PermissionService) InvalidateRoleCache(role models.UserRole) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	delete(s.cache, role)
}

// getHardcodedPermissions returns hardcoded permissions as fallback
// This matches the original GetAdminPermissions and GetStandardPermissions functions
func (s *PermissionService) getHardcodedPermissions(role models.UserRole) []string {
	switch role {
	case models.UserRoleAdmin:
		return []string{
			"projects:read", "projects:write", "projects:delete",
			"tasks:read", "tasks:write", "tasks:delete", "tasks:assign",
			"users:read", "users:write", "users:delete", "users:profile", "users:invite",
			"calendar:read", "calendar:write", "calendar:delete",
			"vacation:read", "vacation:write", "vacation:delete", "vacation:approve",
			"notifications:read", "notifications:write", "notifications:delete",
			"activityLogs:read",
			"dashboard:read",
			"employees:read",
			"infoPortal:read", "infoPortal:write", "infoPortal:delete",
			"googleAccount:read", "googleAccount:write", "googleAccount:link", "googleAccount:unlink",
			"websocket:connect",
			"auth:read",
			"projectDetails:read", "projectDetails:write", "projectDetails:delete",
			"backup:read", "backup:write",
		}
	case models.UserRoleStandard:
		return []string{
			"projects:read",
			"tasks:read", "tasks:write", "tasks:delete",
			"calendar:read", "calendar:write", "calendar:delete",
			"vacation:read", "vacation:write", "vacation:delete",
			"notifications:read", "notifications:write", "notifications:delete",
			"activityLogs:read",
			"dashboard:read",
			"employees:read",
			"infoPortal:read", "infoPortal:write", "infoPortal:delete",
			"googleAccount:read", "googleAccount:write", "googleAccount:link", "googleAccount:unlink",
			"websocket:connect",
			"auth:read",
			"users:profile",
			"projectDetails:read",
		}
	default:
		return []string{}
	}
}
