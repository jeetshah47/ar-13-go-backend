package handlers

import (
	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/middleware"
	"github.com/ar-13-go-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user routes
type UserHandler struct {
	userService       *services.UserService
	permissionService *services.PermissionService
	vacationService   *services.VacationService
}

// NewUserHandler creates a new user handler with dependency injection
func NewUserHandler(
	userService *services.UserService,
	permissionService *services.PermissionService,
	vacationService *services.VacationService,
) *UserHandler {
	return &UserHandler{
		userService:       userService,
		permissionService: permissionService,
		vacationService:   vacationService,
	}
}

// NewUserHandlerWithDefaults creates a new user handler with default dependencies
func NewUserHandlerWithDefaults(cfg *config.Config) *UserHandler {
	return NewUserHandler(
		services.NewUserServiceWithDefaults(cfg),
		services.NewPermissionServiceWithDefaults(),
		services.NewVacationServiceWithDefaults(),
	)
}

// GetUserService returns the user service (for setting notification service)
func (h *UserHandler) GetUserService() *services.UserService {
	return h.userService
}

// GetAll gets all users
// All authenticated users can view the user list (read access)
// Only admins can create, update, or delete users
func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userService.GetAll(c.Request.Context(), nil)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"users": users})
}

// CreateInvitation creates a signup invitation
func (h *UserHandler) CreateInvitation(c *gin.Context) {
	var req services.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.CreateInvitation(c.Request.Context(), req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusCreated, gin.H{"message": "Signup invitation sent successfully"})
}

// Update updates a user
func (h *UserHandler) Update(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Update(c.Request.Context(), &user); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "User updated successfully"})
}

// Delete deletes a user
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(constants.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetProfile gets user profile with permissions
func (h *UserHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	profileResponse, err := h.userService.GetProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	user := profileResponse.User

	// Get permissions for the user's role
	permissions, err := h.permissionService.GetPermissionsByRole(c.Request.Context(), user.Role)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
		return
	}

	// Get leave requests for the user
	leaveRequests, err := h.vacationService.GetByUserID(c.Request.Context(), id)
	if err != nil {
		// Log error but don't fail the request - leave requests are optional
		leaveRequests = []models.LeaveRequest{}
	}

	// Build user object with projects included
	userWithProjects := gin.H{
		"id":                 user.ID,
		"name":               user.Name,
		"email":              user.Email,
		"phoneNumber":        user.PhoneNumber,
		"role":               user.Role,
		"designation":        user.Designation,
		"forceChangePassword": user.ForceChangePassword,
		"createdAt":          user.CreatedAt,
		"updatedAt":          user.UpdatedAt,
		"projects":           profileResponse.Projects,
	}

	c.JSON(constants.StatusOK, gin.H{
		"user":         userWithProjects,
		"role":          user.Role,
		"permissions":   permissions,
		"leaveRequests": leaveRequests,
	})
}

// GetUserPermissions gets permissions for a specific user by ID
func (h *UserHandler) GetUserPermissions(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if user == nil {
		c.JSON(constants.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get permissions for the user's role
	permissions, err := h.permissionService.GetPermissionsByRole(c.Request.Context(), user.Role)
	if err != nil {
		c.JSON(constants.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
		return
	}

	c.JSON(constants.StatusOK, gin.H{
		"userId":      user.ID,
		"role":        user.Role,
		"permissions": permissions,
	})
}

// CreateUser creates a new user by admin with temporary password
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req services.CreateUserByAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.userService.CreateUserByAdmin(c.Request.Context(), req)
	if err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return user without password, but include temp password for admin
	c.JSON(constants.StatusCreated, gin.H{
		"user": gin.H{
			"id":                 response.User.ID,
			"name":               response.User.Name,
			"email":              response.User.Email,
			"phoneNumber":        response.User.PhoneNumber,
			"role":               response.User.Role,
			"designation":        response.User.Designation,
			"forceChangePassword": response.User.ForceChangePassword,
			"createdAt":          response.User.CreatedAt,
			"updatedAt":          response.User.UpdatedAt,
		},
		"tempPassword": response.TempPassword,
		"message":      "User created successfully",
	})
}

// ChangePassword changes a user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by AuthenticateUser middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(constants.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Ensure user can only change their own password (unless admin)
	role := middleware.GetUserRole(c)
	if userID != req.UserID && role != string(models.UserRoleAdmin) {
		c.JSON(constants.StatusForbidden, gin.H{"error": "You can only change your own password"})
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), req); err != nil {
		c.JSON(constants.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constants.StatusOK, gin.H{"message": "Password changed successfully"})
}