package constants

// API path constants
const (
	Base = "/api"
)

// User paths
const (
	UsersBase             = "/users"
	UsersGet              = "/all"
	UsersAdd              = "/add"
	UsersUpdate           = "/update"
	UsersDelete           = "/delete/:id"
	UsersGetProfile       = "/profile/:id"
	UsersCreateInvitation = "/invite"
)

// Auth paths
const (
	AuthBase                = "/auth"
	AuthRegister            = "/register"
	AuthLogin               = "/login"
	AuthLogout              = "/logout"
	AuthValidateSignupToken = "/validate-signup"
)

// Project paths
const (
	ProjectBase   = "/project"
	ProjectGet    = "/all"
	ProjectGetOne = "/:id"
	ProjectAdd    = "/add"
	ProjectUpdate = "/update"
	ProjectDelete = "/delete/:id"
)

// Task paths
const (
	TaskBase                 = "/tasks"
	TaskGet                  = "/all/:projectId"
	TaskGetDetails           = "/all/details/:projectId"
	TaskGetOneDetail         = "/detail/:projectId/:taskId"
	TaskAdd                  = "/add"
	TaskAddMultiple          = "/add-multiple"
	TaskUpdate               = "/update"
	TaskUpdateDeadline       = "/update-deadline/:projectId/:taskId"
	TaskUpdateProgress       = "/update-progress/:projectId/:taskId"
	TaskUpdateDescription    = "/update-description/:projectId/:taskId"
	TaskUpdateStatus         = "/update-status/:projectId/:taskId"
	TaskAddTimeSpent         = "/add-time-spent/:projectId/:taskId"
	TaskUpdateTimeSpent      = "/update-time-spent/:projectId/:taskId/:timeSpentIndex"
	TaskRemoveTimeSpent      = "/remove-time-spent/:projectId/:taskId/:timeSpentIndex"
	TaskGetTimeSpent         = "/time-spent/:projectId/:taskId"
	TaskAddFileAttachment    = "/add-file-attachment/:projectId/:taskId"
	TaskRemoveFileAttachment = "/remove-file-attachment/:projectId/:taskId/:fileAttachmentIndex"
	TaskGetFileAttachments   = "/file-attachments/:projectId/:taskId"
	TaskGetActivityLogs      = "/activity-logs/:projectId/:taskId"
	TaskDelete               = "/delete/:projectId/:taskId"
	TaskAssign               = "/assign/:taskId/:userId"
	TaskClaim                = "/claim/:projectId/:taskId"
	TaskGetAssignableUsers   = "/assignable/:projectId"
	TaskGetStatuses          = "/statuses"
)

// Dashboard paths
const (
	DashboardBase = "/dashboard"
	DashboardGet  = "/stats"
)

// Calendar paths
const (
	CalendarBase    = "/calendar"
	CalendarGet     = "/:year/:month"
	CalendarAdd     = "/add"
	CalendarUpdate  = "/update/:id"
	CalendarDelete  = "/delete/:id"
	CalendarGetById = "/:id"
)

// ProjectDetails paths
const (
	ProjectDetailsBase   = "/project-details"
	ProjectDetailsGet    = "/:projectId"
	ProjectDetailsAdd    = "/:projectId/add"
	ProjectDetailsUpdate = "/:projectId/update"
	ProjectDetailsDelete = "/:projectId/delete/:projectDetailsId"
)

// Notification paths
const (
	NotificationsBase              = "/notifications"
	NotificationsGetAll            = "/all/:userId"
	NotificationsGetUnread         = "/unread/:userId"
	NotificationsGetCount          = "/count/:userId"
	NotificationsMarkAsRead        = "/read/:id"
	NotificationsMarkAllAsRead     = "/read-all/:userId"
	NotificationsDelete            = "/:id"
	NotificationsDeleteAllForUser  = "/user/:userId"
	NotificationsGetConnectionInfo = "/connection-info"
)

// ActivityLog paths
const (
	ActivityLogsBase            = "/activity-logs"
	ActivityLogsGetByEntity     = "/:entityType/:entityId"
	ActivityLogsGetByEntityType = "/:entityType"
	ActivityLogsGetEntityTypes  = "/entity-types"
)

// Vacation paths
const (
	VacationBase                 = "/vacation"
	VacationGetMyRequests        = "/my-requests"
	VacationGetAllRequests       = "/all-requests"
	VacationGetPendingRequests   = "/pending-requests"
	VacationGetOneRequest        = "/request/:requestId"
	VacationCreateRequest        = "/create-request"
	VacationUpdateRequestStatus  = "/update-status/:requestId"
	VacationUpdateRequest        = "/update-request"
	VacationDeleteRequest        = "/delete-request/:requestId"
	VacationGetVacationSummaries = "/summaries"
	VacationGetRequestsByStatus  = "/by-status/:status"
	VacationGetRequestsByType    = "/by-type/:type"
)

// Employee paths
const (
	EmployeeBase     = "/employees"
	EmployeeGetList  = "/list"
	EmployeeGetStats = "/stats/:userId"
)

// InfoPortal paths
const (
	InfoPortalBase                = "/info-portal"
	InfoPortalFoldersGetAll       = "/folders"
	InfoPortalFoldersGetOne       = "/folders/:folderId"
	InfoPortalFoldersCreate       = "/folders"
	InfoPortalFoldersUpdate       = "/folders/:folderId"
	InfoPortalFoldersDelete       = "/folders/:folderId"
	InfoPortalPagesGetOne         = "/pages/:pageId"
	InfoPortalPagesCreate         = "/folders/:folderId/pages"
	InfoPortalPagesUpdate         = "/pages/:pageId"
	InfoPortalPagesDelete         = "/pages/:pageId"
	InfoPortalPagesUpdateSections = "/pages/:pageId/sections"
	InfoPortalAttachmentsUpload   = "/pages/:pageId/attachments"
	InfoPortalAttachmentsDelete   = "/attachments/:attachmentId"
	InfoPortalStatistics          = "/statistics"
)

// GoogleAccount paths
const (
	GoogleAccountBase         = "/google-account"
	GoogleAccountLink         = "/link"
	GoogleAccountUnlink       = "/unlink"
	GoogleAccountGetStatus    = "/status"
	GoogleAccountGetAllLinks  = "/links"
	GoogleAccountInitiateAuth = "/auth/initiate"
	GoogleAccountCallback     = "/auth/callback"
)

// Backup paths
const (
	BackupBase = "/backup"
	BackupAll  = "/all"
)
