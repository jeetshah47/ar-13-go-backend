# Migrate Data from Firebase Backup to DynamoDB

This guide explains how to migrate your Firebase backup data to DynamoDB using the migration script.

## Prerequisites

1. **AWS Credentials**: You need AWS credentials configured. You can use either:
   - **Option 1 (Recommended)**: Add credentials to `.env` file
   - **Option 2**: Use `aws configure` (credentials stored in `~/.aws/credentials`)

2. **DynamoDB Tables**: All required tables must be created in AWS DynamoDB. See `DYNAMODB_TABLES.md` for table creation commands.

## Setup

### Option 1: Using .env File (Recommended)

1. Create a `.env` file in the project root:

```bash
# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here

# JWT Configuration (required for server, optional for migration)
JWT_SECRET=your_jwt_secret_key_here
JWT_EXPIRATION_HOURS=24
REFRESH_EXPIRATION_DAYS=30
```

2. Replace the placeholder values with your actual AWS credentials.

### Option 2: Using AWS CLI

If you prefer using AWS CLI credentials:

```bash
aws configure
```

This stores credentials in `~/.aws/credentials`.

## Running the Migration

1. **Ensure backup files exist** in `upload/backups/` directory:
   - `users_*.json`
   - `projects_*.json`
   - `notifications_*.json`
   - `calendar_*.json`
   - `activityLogs_*.json`
   - `signupInvitations_*.json`
   - `userAccountLinks_*.json`
   - `infoPortal_*.json`
   - `vacation_*.json` (optional, will be mapped to `leaveRequests`)

2. **Run the migration script**:

```bash
go run scripts/migrate_from_backup.go upload/backups
```

Or if you have a different backup directory:

```bash
go run scripts/migrate_from_backup.go /path/to/backups
```

## What Gets Migrated

The script migrates data in this order:

1. **Users** - User accounts with password hashing
2. **Projects** - Project data
3. **Notifications** - User notifications
4. **Calendar Events** - Calendar entries
5. **Leave Requests** - Vacation/leave requests
6. **Activity Logs** - Activity history
7. **Project Details** - Additional project information
8. **Signup Invitations** - Invitation tokens
9. **User Account Links** - OAuth account links
10. **Info Portal** - Info portal folders, pages, and attachments
11. **Tasks** - Task data (from projects subcollections)

## Migration Details

### Password Handling

- If passwords in backup are already hashed (start with `$2a$`, `$2b$`, or `$2y$`), they are used as-is
- If passwords are plain text, they are automatically hashed using bcrypt

### Data Transformations

- **Collection Names**: Automatically mapped to DynamoDB table names
  - `vacations` → `leaveRequests`
  - `calendar` → `calendar_events`
  - `activityLogs` → `activity_logs`
  - `infoPortal` → `info-portal`
  - etc.

- **Nested Data**: Tasks nested in projects are automatically extracted and migrated

- **Timestamps**: ISO date strings are converted to Go `time.Time` objects

## Verification

After migration, verify the data:

1. **Check AWS Console**: Go to DynamoDB console and verify item counts

2. **Run test script**:
```bash
go run scripts/test_dynamodb_connection.go
```

3. **Start your server** and test the API:
```bash
go run cmd/server/main.go
```

## Troubleshooting

### Error: "The security token included in the request is invalid"

**Solution**: Make sure AWS credentials are correctly set in `.env` file:

```bash
AWS_ACCESS_KEY_ID=your_actual_access_key
AWS_SECRET_ACCESS_KEY=your_actual_secret_key
AWS_REGION=us-east-1
```

### Error: "Table does not exist"

**Solution**: Create the missing table using commands in `DYNAMODB_TABLES.md`

### Error: "Failed to migrate user/project/etc"

**Solution**: 
- Check the error message for specific issues
- Verify the backup JSON file format is correct
- Ensure required fields are present in backup data

### No data migrated

**Solution**:
- Check that backup files exist in the specified directory
- Verify backup file names match expected patterns (e.g., `users_*.json`)
- Check that backup files contain data (not empty arrays)

## Migration Output

The script provides detailed output:

- ✅ Successfully migrated items
- ⚠️ Failed items (with error messages)
- ℹ️ Skipped collections (if backup file not found)

Example output:
```
📦 Migrating users from users_20251107_184435.json...
   ✅ Migrated 45/51 users
✅ Successfully migrated users
```

## Notes

- The migration script is **idempotent** - you can run it multiple times
- If an item already exists (same ID), it will be overwritten
- Failed items are logged but don't stop the migration
- The script processes items sequentially to avoid overwhelming DynamoDB

## Next Steps

After successful migration:

1. **Test your application**: Start the server and test all features
2. **Verify data integrity**: Check that all relationships are intact
3. **Update user passwords**: If passwords were migrated, users may need to reset them
4. **Monitor DynamoDB**: Check CloudWatch metrics for any issues

