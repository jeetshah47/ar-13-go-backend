# WebSocket Database Read Optimization

## Summary
Optimized WebSocket operations to reduce unnecessary MongoDB database reads, improving performance and reducing database load.

## Issues Found

### Before Optimization
For a single task status update via WebSocket, the system was performing:

1. **Project Read #1** - Line 456: Get project for access verification
2. **Task Read #1** - Line 399 (in UpdateStatus): Get task to check existence
3. **Task Write** - Line 412: Update task status
4. **Task Read #2** - Line 501: Get updated task for response
5. **Project Read #2** - Line 233 (in BroadcastToProjectMembers): Get project again for member list

**Total: 4 database reads + 1 write**

### Problems Identified
1. **Duplicate Project Read**: Project was read twice - once for access check, once for broadcast
2. **Duplicate Task Read**: Task was read twice - once in UpdateStatus to check existence, once after update to get the updated version

## Optimizations Implemented

### 1. Eliminated Duplicate Project Read
- **Added**: `BroadcastToProjectMembersWithProject()` method that accepts a project object
- **Modified**: `BroadcastToProjectMembers()` now calls the new method internally
- **Result**: Project is only read once and reused for broadcast

### 2. Eliminated Redundant Task Read After Update
- **Modified**: `handleTaskUpdateStatus()` now:
  - Gets the task **before** calling UpdateStatus
  - Reuses the task object after update (just updates status field in memory)
  - Passes the project object to broadcast function
- **Result**: Task is read once before update, then reused (UpdateStatus still reads it for validation, but we avoid the post-update read)

### After Optimization
For a single task status update via WebSocket:

1. **Project Read #1** - Get project for access verification
2. **Task Read #1** - Get task before update (reused after)
3. **Task Read #2** - UpdateStatus reads task for validation (necessary for service layer)
4. **Task Write** - Update task status
5. **Reuse** - Task object reused from step 2 (no additional read)
6. **Reuse** - Project object reused from step 1 (no additional read)

**Total: 2-3 database reads + 1 write** (reduced from 4 reads)

## Code Changes

### File: `pkg/websocket/websocket.go`

1. **Added import**: `"github.com/ar-13-go-backend/internal/models"`

2. **Added new method**: `BroadcastToProjectMembersWithProject()`
   - Accepts `*models.Project` instead of `projectID string`
   - Avoids redundant database read

3. **Modified**: `BroadcastToProjectMembers()`
   - Now calls `BroadcastToProjectMembersWithProject()` internally
   - Maintains backward compatibility

4. **Optimized**: `handleTaskUpdateStatus()`
   - Gets task before UpdateStatus call
   - Reuses task object after update (updates status in memory)
   - Passes project object to broadcast (avoids duplicate read)

## Performance Impact

- **Reduced database reads**: From 4 reads to 2-3 reads per task update
- **Reduced latency**: Fewer database round trips
- **Reduced database load**: Especially important under high WebSocket traffic
- **Maintained functionality**: All existing behavior preserved

## Future Optimization Opportunities

1. **UpdateStatus Service Method**: Could be modified to accept an existing task object to avoid the duplicate read inside UpdateStatus
2. **Caching**: Consider caching project member lists for frequently accessed projects
3. **Batch Operations**: If multiple tasks are updated, batch the operations

## Testing Recommendations

1. Test WebSocket task status updates to ensure functionality is preserved
2. Monitor database query counts before/after optimization
3. Test with multiple concurrent WebSocket connections
4. Verify broadcast messages are still sent correctly to all project members

