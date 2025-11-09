# Model Field Audit - DynamoDB Schema Verification

This document audits all model fields to ensure DynamoDB schema documentation is up-to-date.

## Audit Date
**Date**: 2025-01-XX  
**Purpose**: Verify all model fields are properly documented in DynamoDB schema

---

## ✅ Task Model (`tasks` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `subject` (String)
- `code` (String)
- `status` (String)
- `deadline` (String, RFC3339) - **NEW**: Replaces old `duration` field
- `priority` (String)
- `progress` (Number, 0-100, optional) - **NEW**: Completion percentage
- `assignTo` (String, optional)
- `projectId` (String) - Indexed via GSI
- `description` (String, optional)
- `timeSpent` (List)
- `fileAttachments` (List)
- `activityLogs` (List)
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Status: ✅ Documented
- See `DYNAMODB_TABLES.md` - Tasks Table section
- See `docs/TASK_FIELD_MIGRATION.md` for migration details
- See `docs/TASK_DEADLINE_PROGRESS_API.md` for API documentation

---

## ✅ Project Model (`projects` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `title` (String)
- `description` (String)
- `ownerId` (String)
- `membersIds` (List of Strings)
- `deadLine` (String, RFC3339) - **Note**: JSON tag is `deadLine` (camelCase)
- `logoUrl` (String, optional)
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Status: ⚠️ Needs Documentation Update
- Field `deadLine` exists but may not be fully documented
- Verify DynamoDB documentation includes this field

### Action Required:
- Update `DYNAMODB_TABLES.md` to include `deadLine` field in Projects table documentation

---

## ✅ User Model (`users` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `name` (String)
- `email` (String) - Indexed via GSI
- `phoneNumber` (String)
- `role` (String) - "Standard" or "Admin"
- `password` (String) - Hidden from JSON
- `designation` (String, optional)
- `createdAt` (String, RFC3339)
- `updatedAt` (String, RFC3339)

### Status: ✅ Documented
- Standard user fields, no recent changes detected

---

## ✅ LeaveRequest Model (`leaveRequests` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `userId` (String) - Indexed via GSI
- `requestType` (String) - "vacation", "sick_leave", "work_remotely"
- `startDate` (String, RFC3339)
- `endDate` (String, RFC3339, optional)
- `duration` (Number)
- `durationType` (String) - "days" or "hours"
- `status` (String) - Indexed via GSI - "pending", "approved", "rejected", "cancelled"
- `comments` (String, optional)
- `requestedAt` (String, RFC3339)
- `reviewedBy` (String, optional)
- `reviewedAt` (String, RFC3339, optional)
- `reviewComments` (String, optional)
- `workingHours` (Map, optional) - For remote work: `{from: string, to: string}`
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Status: ✅ Documented
- All fields appear standard, no recent changes detected

---

## ✅ CalendarEvent Model (`calendar_events` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `title` (String)
- `category` (String)
- `priority` (String)
- `start` (String, RFC3339)
- `end` (String, RFC3339)
- `time` (String, optional) - HH:MM format
- `description` (String, optional)
- `isRepeating` (Boolean)
- `repeatFrequency` (String, optional) - "daily", "weekly", "monthly"
- `repeatDays` (List of Strings, optional)
- `createdBy` (String)
- `addToGoogleCalendar` (Boolean, optional)
- `googleCalendarEventId` (String, optional)
- `eventType` (String, optional) - "offline" or "online"
- `invitedMemberIds` (List of Strings, optional)
- `duration` (Number, optional) - Duration in minutes for online events
- `googleMeetLink` (String, optional)
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Status: ✅ Documented
- All fields appear standard, no recent changes detected
- Note: Calendar events have a `duration` field (in minutes), which is different from task's old `duration` field

---

## ✅ Notification Model (`notifications` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `title` (String)
- `message` (String)
- `type` (String) - Notification type constant
- `userId` (String) - Indexed via GSI
- `relatedEntityId` (String)
- `relatedEntityType` (String)
- `isRead` (Boolean)
- `createdAt` (String, RFC3339)
- `created` (String, RFC3339) - From Model
- `updated` (String, RFC3339, optional) - From Model

### Status: ✅ Documented
- All fields appear standard, no recent changes detected

---

## ✅ ProjectDetails Model (`project_details` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `projectId` (String) - Indexed via GSI
- `data` (Map) - Flexible data structure
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Status: ✅ Documented
- Flexible schema, no specific field changes needed

---

## ✅ InfoPortal Models (`info-portal` table)

### Folder Fields:
- `id` (String) - Primary key
- `name` (String)
- `color` (String) - Hex color code
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Page Fields:
- `id` (String) - Primary key
- `title` (String)
- `isActive` (Boolean)
- `folderId` (String)
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Section Fields:
- `id` (String) - Primary key
- `title` (String)
- `content` (String)
- `order` (Number)
- `pageId` (String)
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Attachment Fields:
- `id` (String) - Primary key
- `name` (String)
- `imageUrl` (String)
- `fileUrl` (String)
- `fileType` (String)
- `fileSize` (Number)
- `pageId` (String)
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Status: ✅ Documented
- All fields appear standard, no recent changes detected

---

## ✅ SignupInvitation Model (`signupInvitations` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `email` (String) - Indexed via GSI
- `token` (String) - Indexed via GSI
- `linkExpiry` (String, RFC3339)
- `hasSignup` (Boolean)
- `created` (String, RFC3339)

### Status: ✅ Documented
- All fields appear standard, no recent changes detected

---

## ✅ UserAccountLink Model (`userAccountLinks` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `userId` (String) - Indexed via GSI
- `provider` (String) - "google", "microsoft", "github"
- `providerUserId` (String)
- `providerEmail` (String)
- `providerDisplayName` (String, optional)
- `accessToken` (String, optional) - Hidden from JSON
- `refreshToken` (String, optional) - Hidden from JSON
- `expiresAt` (String, RFC3339, optional)
- `isActive` (Boolean)
- `linkedAt` (String, RFC3339)
- `created` (String, RFC3339)
- `updated` (String, RFC3339, optional)

### Status: ✅ Documented
- All fields appear standard, no recent changes detected

---

## ✅ ActivityLog Model (`activity_logs` table)

### Current Fields in Code:
- `id` (String) - Primary key
- `entityType` (String)
- `entityId` (String) - Indexed via GSI
- `action` (String)
- `createdAt` (String, RFC3339)
- `createdBy` (String)
- `fields` (Map, optional)
- `description` (String, optional)
- `metadata` (Map, optional)

### Status: ✅ Documented
- All fields appear standard, no recent changes detected
- Note: New action types added: `progress_updated`, `deadline_updated`

---

## Summary

### ✅ Models with Recent Changes:
1. **Task Model** - Added `deadline` (replaces `duration`) and `progress` fields
   - ✅ Fully documented
   - ✅ Migration script created
   - ✅ API documentation updated

### ⚠️ Models Needing Verification:
1. **Project Model** - Has `deadLine` field (note casing)
   - ⚠️ Verify DynamoDB documentation includes this field

### ✅ Models with No Recent Changes:
- User Model
- LeaveRequest Model
- CalendarEvent Model
- Notification Model
- ProjectDetails Model
- InfoPortal Models
- SignupInvitation Model
- UserAccountLink Model
- ActivityLog Model

---

## Recommendations

1. **Update DYNAMODB_TABLES.md**:
   - Add `deadLine` field to Projects table documentation
   - Verify all field types match code

2. **Verify Project Model**:
   - Check if `deadLine` field is properly handled in repository
   - Ensure API endpoints use correct field name

3. **No Migration Needed**:
   - All other models appear stable
   - No field renames or additions detected

---

## Next Steps

1. ✅ Task model migration - Complete
2. ⚠️ Verify Project model `deadLine` field documentation
3. ✅ All other models - No action needed

