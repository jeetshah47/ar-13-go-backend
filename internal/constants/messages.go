package constants

// Authorization messages
const (
	MsgUnauthorizedAccess       = "You are not authorized to access this resource"
	MsgUserNotAuthenticated     = "Authentication required. Please log in to continue"
	MsgProjectOwnerOrMemberOnly = "Only project owners and members can perform this operation"
	MsgTaskAssignedOrMemberOnly = "You must be assigned to this task or be a project member to perform this operation"
	MsgProjectMemberOnly        = "Only project members can perform this operation"
	MsgTaskAssignedOnly         = "You must be assigned to this task to perform this operation"
	MsgCannotModifyProject      = "You do not have permission to modify this project. Only project owners and members can make changes"
	MsgCannotModifyTask         = "You do not have permission to modify this task. You must be assigned to the task or be a project member"
	MsgCannotClaimTask          = "You cannot claim this task. You must be a member of the project to claim tasks"
	MsgCannotAssignTask         = "You do not have permission to assign tasks. Only project owners and members can assign tasks"
	MsgCannotModifyTimeLog      = "You do not have permission to modify time logs. You must be assigned to the task or be a project member"
)

// Project messages
const (
	MsgProjectNotFound       = "Project not found"
	MsgProjectCreated        = "Project created successfully"
	MsgProjectUpdated        = "Project updated successfully"
	MsgProjectDeleted        = "Project deleted successfully"
	MsgProjectRetrieved      = "Project retrieved successfully"
	MsgProjectsRetrieved     = "Projects retrieved successfully"
	MsgProjectStatsRetrieved = "Project statistics retrieved successfully"
	MsgInvalidProjectID      = "Invalid project ID"
	MsgProjectIDRequired     = "Project ID is required"
)

// Task messages
const (
	MsgTaskNotFound           = "Task not found"
	MsgTaskCreated            = "Task created successfully"
	MsgTaskUpdated            = "Task updated successfully"
	MsgTaskDeleted            = "Task deleted successfully"
	MsgTaskRetrieved          = "Task retrieved successfully"
	MsgTasksRetrieved         = "Tasks retrieved successfully"
	MsgTaskAssigned           = "Task assigned successfully"
	MsgTaskClaimed            = "Task claimed successfully"
	MsgTaskTransferred        = "Task transferred successfully"
	MsgTaskStatusUpdated      = "Task status updated successfully"
	MsgTaskDeadlineUpdated    = "Task deadline updated successfully"
	MsgTaskProgressUpdated    = "Task progress updated successfully"
	MsgTaskDescriptionUpdated = "Task description updated successfully"
	MsgInvalidTaskID          = "Invalid task ID"
	MsgTaskIDRequired         = "Task ID is required"
	MsgTaskAlreadyAssigned    = "This task is already assigned to you"
)

// Time log messages
const (
	MsgTimeSpentAdded         = "Time log entry added successfully"
	MsgTimeSpentUpdated       = "Time log entry updated successfully"
	MsgTimeSpentRemoved       = "Time log entry removed successfully"
	MsgTimeSpentRetrieved     = "Time log entries retrieved successfully"
	MsgInvalidTimeSpentIndex  = "Invalid time spent entry index"
	MsgTimeSpentIndexRequired = "Time spent entry index is required"
)

// File attachment messages
const (
	MsgFileAttachmentAdded         = "File attachment added successfully"
	MsgFileAttachmentRemoved       = "File attachment removed successfully"
	MsgFileAttachmentsRetrieved    = "File attachments retrieved successfully"
	MsgInvalidFileAttachmentIndex  = "Invalid file attachment index"
	MsgFileAttachmentIndexRequired = "File attachment index is required"
)

// Activity log messages
const (
	MsgActivityLogsRetrieved = "Activity logs retrieved successfully"
)

// General error messages
const (
	MsgInvalidRequest        = "Invalid request. Please check your input and try again"
	MsgInternalServerError   = "An internal server error occurred. Please try again later"
	MsgBadRequest            = "Bad request. Please check your input"
	MsgNotFound              = "Resource not found"
	MsgInvalidFormat         = "Invalid format. Please check your input"
	MsgMissingRequiredField  = "Required field is missing"
	MsgInvalidDeadlineFormat = "Invalid deadline format. Please use RFC3339 format (e.g., 2025-01-20T10:00:00Z)"
	MsgInvalidProgressValue  = "Invalid progress value. Progress must be between 0 and 100"
)

// Validation messages
const (
	MsgInvalidJSON                = "Invalid JSON format"
	MsgInvalidEmail               = "Invalid email format"
	MsgInvalidPassword            = "Invalid password format"
	MsgInvalidToken               = "Invalid or expired token"
	MsgTokenRequired              = "Authorization token is required"
	MsgInvalidAuthorizationHeader = "Invalid authorization header format. Expected: Bearer <token>"
)
