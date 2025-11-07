# Firebase to DynamoDB + JWT Migration Guide

## Overview
This guide documents the migration from Firebase (Firestore + Firebase Auth) to DynamoDB + Self-hosted JWT authentication.

## Migration Summary

### Services Replaced
1. **Firestore** → **DynamoDB** (Database)
2. **Firebase Auth** → **JWT** (Authentication)
3. **Firebase Realtime** → **Socket.IO** (Already implemented)

## Completed Changes

### 1. New Packages Created
- ✅ `pkg/dynamodb/dynamodb.go` - DynamoDB client initialization
- ✅ `pkg/jwt/jwt.go` - JWT token generation and verification
- ✅ `pkg/password/password.go` - Password hashing with bcrypt

### 2. Configuration Updates
- ✅ Removed Firebase config (FirebaseWebAPIKey, FirebaseProjectID, FirebaseClientEmail, FirebasePrivateKey)
- ✅ Added DynamoDB config (AWSRegion)
- ✅ Enhanced JWT config (JWTExpiration, RefreshExpiration)

### 3. Authentication Updates
- ✅ Updated `internal/middleware/auth.go` to use JWT instead of Firebase Auth
- ✅ Updated `internal/services/auth_service.go` to use JWT and password hashing
- ✅ Added Register method to AuthService

## Remaining Tasks

### 4. Repository Migration (Critical)
All repositories need to be migrated from Firestore to DynamoDB:

**Files to Update:**
- `internal/repos/base.go` - Create DynamoDB base repository
- `internal/repos/user_repo.go` - Remove Firebase Auth dependency
- `internal/repos/project_repo.go`
- `internal/repos/task_repo.go`
- `internal/repos/notification_repo.go`
- `internal/repos/calendar_repo.go`
- `internal/repos/vacation_repo.go`
- `internal/repos/activity_log_repo.go`
- `internal/repos/info_portal_repo.go`
- `internal/repos/project_details_repo.go`
- `internal/repos/user_account_link_repo.go`
- `internal/repos/signup_invitation_repo.go`

### 5. Handler Updates
- ✅ Update `internal/handlers/auth.go` to handle new LoginResponse format
- Update any handlers that directly use Firebase services

### 6. Dependencies
Update `go.mod`:
```bash
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue
go get golang.org/x/crypto/bcrypt
go get github.com/google/uuid
```

Remove Firebase dependencies:
```bash
go mod tidy  # Will remove unused Firebase packages
```

### 7. Environment Variables
Update `.env` file:
```env
# Remove Firebase
# FIREBASE_WEB_API_KEY=
# FIREBASE_PROJECT_ID=
# FIREBASE_CLIENT_EMAIL=
# FIREBASE_PRIVATE_KEY=

# Add DynamoDB
AWS_REGION=us-east-1

# Update JWT
JWT_SECRET=your-secret-key-here
JWT_EXPIRATION_HOURS=24
REFRESH_EXPIRATION_DAYS=30
```

## DynamoDB Setup

### 1. Create Tables
You need to create DynamoDB tables for each collection. Example for users table:

```bash
aws dynamodb create-table \
    --table-name users \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=email,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=email-index,KeySchema=[{AttributeName=email,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PAY_PER_REQUEST
```

### 2. Required Tables
- users
- projects
- tasks
- notifications
- calendar_events
- vacations
- activity_logs
- info_portal
- project_details
- user_account_links
- signup_invitations

## Migration Steps

### Step 1: Install Dependencies
```bash
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue
go get golang.org/x/crypto/bcrypt
go get github.com/google/uuid
```

### Step 2: Create DynamoDB Tables
Use AWS Console or CLI to create all required tables.

### Step 3: Migrate Base Repository
Create `internal/repos/dynamodb_base.go` with DynamoDB operations.

### Step 4: Migrate Each Repository
Update each repository file to use DynamoDB instead of Firestore.

### Step 5: Update User Repository
Remove Firebase Auth dependency from user creation.

### Step 6: Test Authentication
- Test login
- Test register
- Test token verification
- Test protected routes

### Step 7: Data Migration (if needed)
If you have existing Firestore data, create a migration script to move data to DynamoDB.

### Step 8: Remove Firebase Package
Delete `pkg/firebase/` directory and remove from imports.

## Socket.IO Usage

Socket.IO is already implemented in `internal/handlers/socketio.go`. No changes needed for realtime functionality.

## Testing Checklist

- [ ] User registration
- [ ] User login
- [ ] JWT token verification
- [ ] Protected routes
- [ ] User CRUD operations
- [ ] Project CRUD operations
- [ ] Task CRUD operations
- [ ] Notification system
- [ ] WebSocket/Socket.IO connections
- [ ] All API endpoints

## Rollback Plan

If issues occur:
1. Keep Firebase code in a separate branch
2. Revert to Firebase by switching branches
3. Fix issues and retry migration

## Notes

- Password hashing uses bcrypt (industry standard)
- JWT tokens include user ID, email, and role
- Refresh tokens have 30-day expiration
- Access tokens have 24-hour expiration (configurable)
- DynamoDB uses on-demand billing (pay per request)

