package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
)

// TaskRepository defines the interface for task data operations
type TaskRepository interface {
	GetByID(ctx context.Context, projectID, taskID string) (*models.Task, error)
	GetByIDWithUserDetails(ctx context.Context, projectID, taskID string) (*models.TaskWithUserDetails, error)
	GetAll(ctx context.Context, projectID string) ([]models.Task, error)
	GetAllWithUserDetails(ctx context.Context, projectID string) ([]models.TaskWithUserDetails, error)
	GetAllByProjectIDs(ctx context.Context, projectIDs []string) (map[string][]models.Task, error)
	Add(ctx context.Context, task *models.Task) error
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, projectID, taskID string) error
	UpdateDeadline(ctx context.Context, projectID, taskID string, deadline time.Time) error
	UpdateProgress(ctx context.Context, projectID, taskID string, progress int) error
	UpdateDescription(ctx context.Context, projectID, taskID string, description string) error
	UpdateStatus(ctx context.Context, projectID, taskID, status string) error
	AddTimeSpent(ctx context.Context, projectID, taskID string, timeSpent models.TimeSpent) error
	UpdateTimeSpent(ctx context.Context, projectID, taskID string, index int, timeSpent models.TimeSpent) error
	RemoveTimeSpent(ctx context.Context, projectID, taskID string, index int) error
	AddFileAttachment(ctx context.Context, projectID, taskID string, attachment models.FileAttachment) error
	RemoveFileAttachment(ctx context.Context, projectID, taskID string, index int) error
	GetTimeSpent(ctx context.Context, projectID, taskID string) ([]models.TimeSpent, error)
	GetFileAttachments(ctx context.Context, projectID, taskID string) ([]models.FileAttachment, error)
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetAll(ctx context.Context, limit *int) ([]models.User, error)
	BatchGetItems(ctx context.Context, ids []string) (map[string]*models.User, error)
	Add(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	Persists(ctx context.Context, id string) (bool, error)
}

// ProjectRepository defines the interface for project data operations
type ProjectRepository interface {
	GetByID(ctx context.Context, id string) (*models.Project, error)
	GetAll(ctx context.Context, limit *int) ([]models.Project, error)
	GetByUserID(ctx context.Context, userID string) ([]models.Project, error)
	Add(ctx context.Context, project *models.Project) error
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id string) error
	UpdateAgencyContact(ctx context.Context, projectID string, agencyContact *models.AgencyContact) error
	Archive(ctx context.Context, projectID string, isArchived bool) error
}

// ActivityLogRepository defines the interface for activity log operations
type ActivityLogRepository interface {
	Add(ctx context.Context, log *models.ActivityLogBase) error
	GetByEntity(ctx context.Context, entityType models.ActivityLogEntityType, entityID string) ([]models.ActivityLogBase, error)
	GetByEntityType(ctx context.Context, entityType models.ActivityLogEntityType, limit *int) ([]models.ActivityLogBase, error)
	GetByID(ctx context.Context, activityLogID string) (*models.ActivityLogBase, error)
}

// ActivityLogReplyRepository defines the interface for activity log reply operations
type ActivityLogReplyRepository interface {
	Add(ctx context.Context, reply *models.ActivityLogReply) error
	GetByActivityLogID(ctx context.Context, activityLogID string) ([]models.ActivityLogReply, error)
}

// TaskStatusRepository defines the interface for task status operations
type TaskStatusRepository interface {
	GetAll(ctx context.Context) ([]map[string]interface{}, error)
}

// SignupInvitationRepository defines the interface for signup invitation operations
type SignupInvitationRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.SignupInvitation, error)
	GetByToken(ctx context.Context, token string) (*models.SignupInvitation, error)
	Add(ctx context.Context, invitation *models.SignupInvitation) error
	MarkAsSignedUp(ctx context.Context, id string) error
}

// RolePermissionRepository defines the interface for role permission operations
type RolePermissionRepository interface {
	GetByRole(ctx context.Context, role models.UserRole) ([]models.RolePermission, error)
	GetByID(ctx context.Context, id string) (*models.RolePermission, error)
	GetAll(ctx context.Context) ([]models.RolePermission, error)
	Add(ctx context.Context, rp *models.RolePermission) error
	BatchAdd(ctx context.Context, permissions []models.RolePermission) error
	Delete(ctx context.Context, id string) error
	DeleteByRoleAndPermission(ctx context.Context, role models.UserRole, permission string) error
	HasPermission(ctx context.Context, role models.UserRole, permission string) (bool, error)
}

// CalendarEventRepository defines the interface for calendar event operations
type CalendarEventRepository interface {
	GetByMonth(ctx context.Context, month, year int) ([]models.CalendarEvent, error)
	GetByID(ctx context.Context, id string) (*models.CalendarEvent, error)
	Add(ctx context.Context, event *models.CalendarEvent) error
	Update(ctx context.Context, event *models.CalendarEvent) error
	Delete(ctx context.Context, id string) error
}

// NotificationRepository defines the interface for notification operations
type NotificationRepository interface {
	GetByID(ctx context.Context, id string) (*models.Notification, error)
	GetAll(ctx context.Context, userID string) ([]models.Notification, error)
	GetUnread(ctx context.Context, userID string) ([]models.Notification, error)
	GetCount(ctx context.Context, userID string) (total, unread int, err error)
	Add(ctx context.Context, notification *models.Notification) error
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}

// VacationRepository defines the interface for vacation/leave request operations
type VacationRepository interface {
	GetByID(ctx context.Context, id string) (*models.LeaveRequest, error)
	GetByUserID(ctx context.Context, userID string) ([]models.LeaveRequest, error)
	GetAll(ctx context.Context) ([]models.LeaveRequest, error)
	GetPending(ctx context.Context) ([]models.LeaveRequest, error)
	GetByStatus(ctx context.Context, status models.LeaveRequestStatus) ([]models.LeaveRequest, error)
	GetByType(ctx context.Context, requestType models.LeaveRequestType) ([]models.LeaveRequest, error)
	Add(ctx context.Context, request *models.LeaveRequest) error
	Update(ctx context.Context, request *models.LeaveRequest) error
	UpdateStatus(ctx context.Context, id string, status models.LeaveRequestStatus, reviewedBy string, reviewComments *string) error
	Delete(ctx context.Context, id string) error
}

// InfoPortalRepository defines the interface for info portal operations
type InfoPortalRepository interface {
	GetAllFolders(ctx context.Context) ([]models.Folder, error)
	GetFolderByID(ctx context.Context, folderID string) (*models.Folder, error)
	CreateFolder(ctx context.Context, folder *models.Folder) error
	UpdateFolder(ctx context.Context, folderID string, updates map[string]interface{}) error
	DeleteFolder(ctx context.Context, folderID string) error
	GetPageByID(ctx context.Context, pageID string) (*models.Page, error)
	CreatePage(ctx context.Context, folderID string, page *models.Page) error
	UpdatePage(ctx context.Context, pageID string, updates map[string]interface{}) error
	DeletePage(ctx context.Context, pageID string) error
	UpdatePageSections(ctx context.Context, pageID string, sections []models.Section) error
	CreateAttachment(ctx context.Context, pageID string, attachment *models.Attachment) error
	DeleteAttachment(ctx context.Context, attachmentID string) error
}

// ProjectDetailsRepository defines the interface for project details operations
type ProjectDetailsRepository interface {
	Get(ctx context.Context, projectID string) (*models.ProjectDetails, error)
	Add(ctx context.Context, projectID string, details *models.ProjectDetails) error
	Update(ctx context.Context, projectID string, details *models.ProjectDetails) error
	Delete(ctx context.Context, projectID, detailsID string) error
}

// UserAccountLinkRepository defines the interface for user account link operations
type UserAccountLinkRepository interface {
	GetByUserID(ctx context.Context, userID string) ([]models.UserAccountLink, error)
	GetByProvider(ctx context.Context, provider models.AccountProvider, providerUserID string) (*models.UserAccountLink, error)
	GetByUserIDAndProvider(ctx context.Context, userID string, provider models.AccountProvider) (*models.UserAccountLink, error)
	Add(ctx context.Context, link *models.UserAccountLink) error
	Update(ctx context.Context, link *models.UserAccountLink) error
	Deactivate(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

// TimeTrackingRepository defines the interface for time tracking session operations
type TimeTrackingRepository interface {
	Add(ctx context.Context, session *models.TimeTrackingSession) error
	GetByID(ctx context.Context, sessionID string) (*models.TimeTrackingSession, error)
	GetActiveByTaskAndUser(ctx context.Context, projectID, taskID, userID string) (*models.TimeTrackingSession, error)
	GetAllActive(ctx context.Context) ([]models.TimeTrackingSession, error)
	GetByTask(ctx context.Context, projectID, taskID string) ([]models.TimeTrackingSession, error)
	Update(ctx context.Context, session *models.TimeTrackingSession) error
	StopSession(ctx context.Context, sessionID string) error
	UpdateActivity(ctx context.Context, sessionID string, lastActive time.Time) error
	UpdateTotalMinutes(ctx context.Context, sessionID string, totalMinutes int) error
}

// Verify that concrete types implement interfaces at compile time
var (
	_ UserRepository             = (*UserRepo)(nil)
	_ SignupInvitationRepository = (*SignupInvitationRepo)(nil)
	_ RolePermissionRepository   = (*RolePermissionRepo)(nil)
	_ CalendarEventRepository    = (*CalendarEventRepo)(nil)
	_ NotificationRepository     = (*NotificationRepo)(nil)
	_ VacationRepository         = (*VacationRepo)(nil)
	_ InfoPortalRepository       = (*InfoPortalRepo)(nil)
	_ ProjectDetailsRepository   = (*ProjectDetailsRepo)(nil)
	_ UserAccountLinkRepository  = (*UserAccountLinkRepo)(nil)
	_ TimeTrackingRepository     = (*TimeTrackingRepo)(nil)
)
