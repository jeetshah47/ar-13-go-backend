# Repository Migration Status

## Migration Status Overview

**Total Repositories**: 12  
**Fully Migrated**: 4 ✅  
**Pending Migration**: 8 ⚠️

---

## ✅ Fully Migrated Repositories

### 1. UserRepo (`user_repo.go`)
- ✅ **Status**: Fully migrated to DynamoDB
- ✅ **Methods Implemented**:
  - `GetByID()` - Uses DynamoDB GetItem
  - `GetByEmail()` - Uses GSI query on `email-index`
  - `GetAll()` - Uses DynamoDB Scan
  - `Add()` - Uses DynamoDB PutItem
  - `Update()` - Uses DynamoDB UpdateItem
  - `Delete()` - Uses DynamoDB DeleteItem
- ✅ **Table**: `users`
- ✅ **GSI**: `email-index` (email)

### 2. NotificationRepo (`notification_repo.go`)
- ✅ **Status**: Fully migrated to DynamoDB
- ✅ **Methods Implemented**:
  - `GetByID()` - Uses DynamoDB GetItem
  - `GetAll()` - Uses GSI query on `userId-index`
  - `GetUnread()` - Uses GSI query with filtering
  - `GetCount()` - Uses GSI query with counting
  - `Add()` - Uses DynamoDB PutItem
  - `MarkAsRead()` - Uses DynamoDB UpdateItem
  - `MarkAllAsRead()` - Batch updates via GSI query
  - `Delete()` - Uses DynamoDB DeleteItem
  - `DeleteAllForUser()` - Batch deletes via GSI query
- ✅ **Table**: `notifications`
- ✅ **GSI**: `userId-index` (userId)

---

### 3. ProjectRepo (`project_repo.go`)
- ✅ **Status**: Fully migrated to DynamoDB
- ✅ **Methods Implemented**:
  - `GetByID()` - Uses DynamoDB GetItem
  - `GetAll()` - Uses DynamoDB Scan
  - `Add()` - Uses DynamoDB PutItem
  - `Update()` - Uses DynamoDB UpdateItem
  - `Delete()` - Uses DynamoDB DeleteItem
  - `Persists()` - Uses DynamoDB Scan with filtering
- ✅ **Table**: `projects`
- ⚠️ **Note**: `GetByOwnerID()` method not found (may need to be added)

### 4. CalendarEventRepo (`calendar_repo.go`)
- ✅ **Status**: Fully migrated to DynamoDB
- ✅ **Methods Implemented**:
  - `GetByID()` - Uses DynamoDB GetItem
  - `GetByMonth()` - Uses DynamoDB Scan with date filtering
  - `Add()` - Uses DynamoDB PutItem
  - `Update()` - Uses DynamoDB UpdateItem
  - `Delete()` - Uses DynamoDB DeleteItem
- ✅ **Table**: `calendar_events`
- ⚠️ **Note**: Uses Scan for date range queries (could be optimized with GSI)

---

## ⚠️ Pending Migration Repositories

All repositories below are currently **stubbed** and return `fmt.Errorf("... migration to DynamoDB pending")` when called.

### 5. TaskRepo (`task_repo.go`)
- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `notifications`
- 📋 **Required GSIs**:
  - `userId-index` (userId)
  - `isRead-index` (isRead)
- 📋 **Methods to Implement**:
  - `GetByID()`
  - `GetByUserID()`
  - `GetUnreadByUserID()`
  - `GetAll()`
  - `Add()`
  - `MarkAsRead()`
  - `MarkAllAsRead()`
  - `Delete()`

### 6. ProjectDetailsRepo (`project_details_repo.go`)
- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `tasks`
- 📋 **Required GSIs**:
  - `projectId-index` (projectId)
- 📋 **Methods to Implement**:
  - `GetByID(projectID, taskID)`
  - `GetAll(projectID)`
  - `Add()`
  - `Update()`
  - `Delete()`
  - `UpdateDuration()`
  - `UpdateDescription()`
  - `UpdateStatus()`
  - `AddTimeSpent()`
  - `UpdateTimeSpent()`
  - `RemoveTimeSpent()`
  - `AddFileAttachment()`
  - `RemoveFileAttachment()`
  - `GetTimeSpent()`
  - `GetFileAttachments()`

### 7. ActivityLogRepo (`activity_log_repo.go`)
- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `projects`
- 📋 **Required GSIs**:
  - `ownerId-index` (ownerId)
- 📋 **Methods to Implement**:
  - `GetByID()`
  - `GetAll()`
  - `GetByOwnerID()`
  - `Add()`
  - `Update()`
  - `Delete()`

### 8. VacationRepo (`vacation_repo.go`)
- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `projectDetails`
- 📋 **Required GSIs**: None (uses projectId as partition key)
- 📋 **Methods to Implement**:
  - `GetByProjectID()`
  - `Add()`
  - `Update()`
  - `Delete()`

### 9. SignupInvitationRepo (`signup_invitation_repo.go`)
- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `activityLogs`
- 📋 **Required GSIs**:
  - `entityType-entityId-index` (entityType, entityId)
  - `createdBy-index` (createdBy)
  - `projectId-index` (projectId)
- 📋 **Methods to Implement**:
  - `GetByID()`
  - `GetByEntity()`
  - `GetByUserID()`
  - `GetByProjectID()`
  - `Add()`
  - `Delete()`

### 10. UserAccountLinkRepo (`user_account_link_repo.go`)
- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `calendarEvents`
- 📋 **Required GSIs**:
  - `createdBy-index` (createdBy)
  - `start-index` (start) - for date range queries
- 📋 **Methods to Implement**:
  - `GetByID()`
  - `GetAll()`
  - `GetByUserID()`
  - `GetByDateRange()`
  - `Add()`
  - `Update()`
  - `Delete()`

### 8. VacationRepo (`vacation_repo.go`)
- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `leaveRequests`
- 📋 **Required GSIs**:
  - `userId-index` (userId)
  - `status-index` (status)
  - `requestType-index` (requestType)
- 📋 **Methods to Implement**:
  - `GetByID()`
  - `GetByUserID()`
  - `GetAll()`
  - `GetPending()`
  - `GetByStatus()`
  - `GetByType()`
  - `Add()`
  - `Update()`
  - `UpdateStatus()`
  - `Delete()`

### 9. SignupInvitationRepo (`signup_invitation_repo.go`)
- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `signupInvitations`
- 📋 **Required GSIs**:
  - `email-index` (email)
  - `token-index` (token)
- 📋 **Methods to Implement**:
  - `GetByEmail()`
  - `GetByToken()`
  - `Add()`
  - `MarkAsSignedUp()`

### 10. UserAccountLinkRepo (`user_account_link_repo.go`)
- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `userAccountLinks`
- 📋 **Required GSIs**:
  - `userId-index` (userId)
  - `provider-providerUserId-index` (provider, providerUserId)
  - `userId-provider-index` (userId, provider)
- 📋 **Methods to Implement**:
  - `GetByUserID()`
  - `GetByProvider()`
  - `GetByUserIDAndProvider()`
  - `Add()`
  - `Update()`
  - `Deactivate()`
  - `Delete()`

- ⚠️ **Status**: Stubbed (migration pending)
- 📋 **Table**: `info-portal`
- 📋 **Note**: This repo has subcollections (folders, pages, attachments) which need special handling in DynamoDB
- 📋 **Methods to Implement**:
  - **Folders**:
    - `GetAllFolders()`
    - `GetFolderByID()`
    - `CreateFolder()`
    - `UpdateFolder()`
    - `DeleteFolder()`
  - **Pages**:
    - `GetPageByID()`
    - `CreatePage()`
    - `UpdatePage()`
    - `DeletePage()`
    - `UpdatePageSections()`
  - **Attachments**:
    - `CreateAttachment()`
    - `DeleteAttachment()`

- ⚠️ **Status**: Deprecated (replaced by DynamoBaseRepo)
- 📋 **Note**: This is the old Firestore base repository. All methods are stubbed to return errors, forcing migration to `DynamoBaseRepo`.

---

## Migration Priority Recommendations

### High Priority (Core Features)
1. **TaskRepo** - Core feature for task management (fully stubbed)

### Medium Priority
4. **ActivityLogRepo** - Important for audit trails
5. **CalendarEventRepo** - Calendar functionality
6. **VacationRepo** - Leave request management

### Lower Priority
7. **SignupInvitationRepo** - Used during user registration
8. **UserAccountLinkRepo** - OAuth account linking
9. **ProjectDetailsRepo** - Project metadata
10. **InfoPortalRepo** - Info portal content (may need special DynamoDB design)

---

## Migration Checklist Template

For each repository migration:

- [ ] Review existing Firestore queries and operations
- [ ] Design DynamoDB table structure (partition key, sort key)
- [ ] Identify required Global Secondary Indexes (GSIs)
- [ ] Create/verify DynamoDB table exists (see `DYNAMODB_TABLES.md`)
- [ ] Implement `GetByID()` using `DynamoBaseRepo.GetByID()`
- [ ] Implement query methods using `DynamoBaseRepo.QueryByIndex()`
- [ ] Implement `Add()` using `DynamoBaseRepo.PutItem()`
- [ ] Implement `Update()` using `DynamoBaseRepo.UpdateItem()`
- [ ] Implement `Delete()` using `DynamoBaseRepo.DeleteByID()`
- [ ] Test all methods with real DynamoDB operations
- [ ] Remove stubs and "migration pending" errors
- [ ] Update any handlers/services that use the repository

---

## Notes

- All stubbed repositories will return errors when called - this is intentional until migration is complete
- The `DynamoBaseRepo` provides generic CRUD operations that can be used by all repositories
- GSI design is critical for efficient queries in DynamoDB
- Some repositories may need composite keys or special data structures for DynamoDB

---

**Last Updated**: After Firebase removal completion  
**Next Step**: Migrate NotificationRepo, ProjectRepo, or TaskRepo (high priority)
