# Quick Start Guide - After Tables Created ✅

## ✅ You've Created Tables on AWS!

Now let's verify everything works and start using your application.

## Step 1: Configure AWS Credentials

### For Local Development

**Quick Setup:**
```bash
aws configure
```

Enter:
- AWS Access Key ID
- AWS Secret Access Key  
- Default region (e.g., `us-east-1`)
- Default output format (`json`)

**OR use environment variables:**
```bash
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1
```

### For EC2 Deployment
- Attach IAM role with DynamoDB permissions to your EC2 instance
- No need for access keys when using IAM roles

## Step 2: Test DynamoDB Connection

Run the test script to verify everything is set up correctly:

```bash
go run scripts/test_dynamodb_connection.go
```

**Expected Output:**
```
Testing DynamoDB connection in region: us-east-1

✅ DynamoDB client initialized successfully

Found 11 table(s) in DynamoDB:
  - users
  - notifications
  - projects
  ...

Checking required tables...
  ✅ users - exists
  ✅ notifications - exists
  ...

🎉 All checks passed! DynamoDB connection is working correctly.
```

## Step 3: Configure Environment Variables

Create or update your `.env` file:

```env
PORT=3000
NODE_ENV=development

# AWS DynamoDB
AWS_REGION=us-east-1  # Match your table region

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
JWT_EXPIRATION_HOURS=24
REFRESH_EXPIRATION_DAYS=30

# Email (optional for now)
EMAIL_HOST=smtp.example.com
EMAIL_PORT=587
EMAIL_USER=your-email@example.com
EMAIL_PASSWORD=your-password
EMAIL_FROM=noreply@example.com
EMAIL_FROM_NAME=AR-13

# Frontend
FRONTEND_URL=http://localhost:3000

# Google OAuth (optional)
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
```

**Important:**
- `AWS_REGION` must match the region where you created your tables
- `JWT_SECRET` should be a strong random string (at least 32 characters)

## Step 4: Start the Server

```bash
go run cmd/server/main.go
```

Look for:
```
DynamoDB initialized successfully
Server running on port 3000
```

## Step 5: Test Basic Operations

### Test User Registration

```bash
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "testpassword123",
    "phoneNumber": "+1234567890"
  }'
```

**Expected Response:**
```json
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "eyJhbGc...",
  "expiresIn": 86400,
  "user": {
    "id": "uuid-here",
    "name": "Test User",
    "email": "test@example.com"
  }
}
```

### Test User Login

```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123"
  }'
```

## Step 6: Verify Data in AWS Console

1. Go to [AWS Console → DynamoDB](https://console.aws.amazon.com/dynamodb/)
2. Select your region
3. Click on `users` table
4. Go to "Items" tab
5. You should see your test user!

## Table Name Reference

Make sure your AWS table names match exactly:

| Code Table Name | AWS Table Name |
|----------------|----------------|
| `users` | `users` |
| `notifications` | `notifications` |
| `projects` | `projects` |
| `calendar_events` | `calendar_events` |
| `tasks` | `tasks` |
| `project_details` | `project_details` |
| `activity_logs` | `activity_logs` |
| `leaveRequests` | `leaveRequests` |
| `signupInvitations` | `signupInvitations` |
| `userAccountLinks` | `userAccountLinks` |
| `info-portal` | `info-portal` |

## Troubleshooting

### ❌ "NoCredentialProviders" Error
**Fix:** Run `aws configure` or set environment variables

### ❌ "ResourceNotFoundException" 
**Fix:** 
- Check table name matches exactly (case-sensitive!)
- Verify `AWS_REGION` matches table region
- List tables: `aws dynamodb list-tables`

### ❌ "AccessDeniedException"
**Fix:** Check IAM permissions - need DynamoDB read/write access

### ❌ Table Not Found
**Fix:** Verify table exists in AWS Console and region matches

## Next Steps

1. ✅ Test all API endpoints
2. ✅ Monitor DynamoDB usage in AWS Console
3. ✅ Check CloudWatch metrics
4. ✅ Set up production environment variables
5. ✅ Deploy to EC2 (if ready)

## Quick Commands

```bash
# Test connection
go run scripts/test_dynamodb_connection.go

# Start server
go run cmd/server/main.go

# Build for production
go build -o server ./cmd/server

# List AWS tables
aws dynamodb list-tables --region us-east-1

# Check table details
aws dynamodb describe-table --table-name users --region us-east-1
```

---

**You're all set!** 🎉 Your application is ready to use DynamoDB.

For detailed testing instructions, see `TESTING_GUIDE.md`

