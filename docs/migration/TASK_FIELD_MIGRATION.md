# Task Field Migration Guide

This document describes the migration of task fields in DynamoDB from `duration` to `deadline` and the addition of the `progress` field.

## Overview

The task model has been updated with the following changes:

1. **Field Renamed**: `duration` → `deadline`
   - Old field: `duration` (deprecated)
   - New field: `deadline` (RFC3339 datetime format)

2. **New Field Added**: `progress`
   - Type: Integer (0-100)
   - Optional: Can be `null` if not set
   - Represents task completion percentage

## DynamoDB Schema

### Tasks Table Structure

The `tasks` table uses the following structure:

**Primary Key:**
- `id` (String) - Partition key

**Global Secondary Index (GSI):**
- `projectId-index` - For querying tasks by project

**Task Fields:**
```json
{
  "id": "string",
  "subject": "string",
  "code": "string",
  "status": "string",
  "deadline": "string (RFC3339 format)",  // NEW: Replaces 'duration'
  "priority": "string",
  "progress": "number (0-100)",           // NEW: Optional field
  "assignTo": "string (optional)",
  "projectId": "string",
  "description": "string (optional)",
  "timeSpent": "array",
  "fileAttachments": "array",
  "activityLogs": "array",
  "created": "string (RFC3339 format)",
  "updated": "string (RFC3339 format, optional)"
}
```

### Field Details

#### `deadline` (Required)
- **Type**: String (RFC3339 datetime format)
- **Format**: `YYYY-MM-DDTHH:MM:SSZ` or `YYYY-MM-DDTHH:MM:SS±HH:MM`
- **Example**: `"2025-01-20T10:00:00Z"`
- **Migration**: Automatically migrated from `duration` field if present

#### `progress` (Optional)
- **Type**: Number (Integer)
- **Range**: 0-100
- **Default**: `null` (if not set)
- **Description**: Task completion percentage
  - `0` = Not started (0%)
  - `1-99` = In progress (1-99%)
  - `100` = Completed (100%)

## Migration Process

### Automatic Migration (On Read)

The application automatically handles backward compatibility:

1. **Reading Tasks**: When a task is read from DynamoDB:
   - If `deadline` exists, it's used
   - If `duration` exists but `deadline` doesn't, `duration` is automatically copied to `deadline`
   - Old `duration` field is preserved until manual migration

2. **Writing Tasks**: New tasks always use `deadline` field

### Manual Migration Script

To migrate existing data in DynamoDB, run the migration script:

```bash
go run scripts/migrate_task_fields.go
```

**What the script does:**
1. Scans all tasks in the `tasks` table
2. For each task:
   - If `duration` exists but `deadline` doesn't: Copies `duration` → `deadline`
   - If both `duration` and `deadline` exist: Removes old `duration` field
   - Leaves tasks that already have `deadline` unchanged
3. Provides a summary of migrated tasks

**Script Output:**
```
Starting migration of task fields...
This will:
  1. Migrate 'duration' field to 'deadline' field
  2. Ensure all tasks have the new field structure

✓ Migrated task task123: duration -> deadline
✓ Cleaned up task task456: removed old duration field
...

Migration Summary:
  Total tasks scanned: 150
  Tasks updated: 45
  Errors: 0

Migration completed!
```

### Manual Migration (AWS CLI)

If you prefer to migrate manually using AWS CLI:

```bash
# 1. Export tasks with duration field
aws dynamodb scan \
  --table-name tasks \
  --filter-expression "attribute_exists(duration)" \
  --projection-expression "id,duration" \
  > tasks_with_duration.json

# 2. For each task, update using UpdateItem
aws dynamodb update-item \
  --table-name tasks \
  --key '{"id": {"S": "task-id-here"}}' \
  --update-expression "SET deadline = duration REMOVE duration" \
  --expression-attribute-names '{"#duration": "duration"}'
```

**Note**: Manual migration is more complex and error-prone. Use the migration script instead.

## Backward Compatibility

### Reading Tasks

The code automatically handles backward compatibility:

```go
// In task_repo.go - convertDurationToDeadline function
// If deadline already exists, use it
// If duration exists but deadline doesn't, migrate it
if durationVal, exists := item["duration"]; exists {
    item["deadline"] = durationVal
}
```

### Writing Tasks

- New tasks: Always use `deadline` field
- Updated tasks: Always use `deadline` field
- Old `duration` field is never written

## Verification

After migration, verify the changes:

### 1. Check Task Structure

```bash
aws dynamodb get-item \
  --table-name tasks \
  --key '{"id": {"S": "your-task-id"}}'
```

Expected result:
- ✅ `deadline` field exists
- ✅ `progress` field may exist (optional)
- ❌ `duration` field should not exist (or will be ignored)

### 2. Query Tasks by Project

```bash
aws dynamodb query \
  --table-name tasks \
  --index-name projectId-index \
  --key-condition-expression "projectId = :pid" \
  --expression-attribute-values '{":pid": {"S": "your-project-id"}}'
```

All returned tasks should have `deadline` field.

### 3. Test API Endpoints

```bash
# Update deadline
curl -X PUT "http://localhost:3000/api/tasks/update-deadline/projectId/taskId" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"deadline": "2025-01-20T10:00:00Z"}'

# Update progress
curl -X PUT "http://localhost:3000/api/tasks/update-progress/projectId/taskId" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"progress": 75}'
```

## Important Notes

1. **No Table Schema Changes Required**
   - DynamoDB is schema-less
   - No ALTER TABLE commands needed
   - Migration is data-only

2. **Zero Downtime**
   - Application continues to work during migration
   - Backward compatibility ensures no breaking changes
   - Migration can be run at any time

3. **Data Safety**
   - Original `duration` values are preserved during migration
   - No data loss occurs
   - Can rollback by reverting code (not recommended)

4. **Performance**
   - Migration script includes throttling delays
   - Processes tasks in batches
   - Safe for production use

5. **Progress Field**
   - New field, no migration needed
   - Can be set incrementally as needed
   - Existing tasks will have `progress: null` until set

## Rollback Plan

If you need to rollback (not recommended):

1. **Code Rollback**: Revert to previous code version that uses `duration`
2. **Data Rollback**: Not needed - old `duration` field is preserved if migration script wasn't run
3. **Note**: If migration script was run, `duration` field was removed. You would need to restore from backup.

## Troubleshooting

### Issue: Migration script fails with "Access Denied"

**Solution**: Ensure AWS credentials are configured:
```bash
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
export AWS_REGION=us-east-1
```

### Issue: Some tasks still have `duration` field

**Solution**: 
- Check if migration script completed successfully
- Re-run migration script (it's safe to run multiple times)
- Verify DynamoDB permissions

### Issue: Tasks missing `deadline` field after migration

**Solution**:
- Check migration script logs for errors
- Verify tasks exist in DynamoDB
- Check if tasks had `duration` field before migration

## Support

For issues or questions:
- Check migration script logs
- Review DynamoDB table structure
- Verify AWS credentials and permissions
- Contact development team

