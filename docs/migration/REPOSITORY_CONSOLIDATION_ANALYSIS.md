# Repository Consolidation Analysis

## Analysis Summary

After reviewing all 12 repositories, here are the consolidation opportunities:

### ✅ **No Consolidation Recommended**

All repositories serve distinct purposes and should remain separate:

1. **users** - Core user accounts
2. **projects** - Project information
3. **tasks** - Task management
4. **notifications** - User notifications
5. **calendar_events** - Calendar events
6. **leaveRequests** - Leave/vacation requests
7. **activity_logs** - Activity tracking
8. **info-portal** - Info portal content (folders, pages, attachments)
9. **project_details** - Flexible project metadata (1:1 with projects)
10. **user_account_links** - OAuth account links
11. **signupInvitations** - Signup invitation tokens
12. **role_permissions** - Role-based permissions

### Potential Consolidation (Not Recommended)

#### **project_details → projects** (NOT RECOMMENDED)

**Current Structure:**
- `projects` - Core project data (title, description, owner, members, deadline)
- `project_details` - Flexible metadata stored as `map[string]interface{}`

**Why Keep Separate:**
- ✅ **Flexibility**: Project details can grow independently without affecting core project schema
- ✅ **Performance**: Core project queries don't need to load large metadata objects
- ✅ **Schema Evolution**: Details can change without schema migrations
- ✅ **Clean Separation**: Core vs. extended data separation

**If Merged:**
- Would require loading all details when fetching projects
- Schema changes would affect core project structure
- Less flexible for future extensions

**Recommendation**: Keep separate ✅

### Collection Structure Notes

#### **info-portal** Collection
- Uses a **single collection** with type prefixes (`folder-`, `page-`, `attachment-`)
- Filtered by `type` field for different entity types
- This is already optimized - no consolidation needed

#### **activity_logs** Collection
- Tracks activities across multiple entity types (tasks, projects, users, etc.)
- Uses `entityType` and `entityId` for relationships
- Separate collection is appropriate for audit/logging purposes

## Final Recommendation

**Keep all 12 collections separate** - Each serves a distinct purpose and the current structure is optimal for:
- Query performance
- Schema flexibility
- Data organization
- Future scalability

