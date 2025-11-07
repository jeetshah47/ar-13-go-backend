# DynamoDB Testing Guide

## ✅ Tables Created on AWS

Great! Now let's verify everything is working correctly.

## Step 1: Verify AWS Credentials

### For Local Development

**Option 1: AWS CLI Configuration**
```bash
aws configure
# Enter your:
# - AWS Access Key ID
# - AWS Secret Access Key
# - Default region (e.g., us-east-1)
# - Default output format (json)
```

**Option 2: Environment Variables**
```bash
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1
```

**Option 3: .env File**
Add to your `.env` file:
```env
AWS_REGION=us-east-1
# Note: AWS credentials are loaded from AWS CLI config or environment variables
```

### For EC2 Deployment
- Use IAM roles instead of access keys
- Attach a policy with DynamoDB permissions to your EC2 instance role

## Step 2: Test DynamoDB Connection

### Quick Test Script

Run the test script to verify connection and table existence:

```bash
go run scripts/test_dynamodb_connection.go
```

This will:
- ✅ Test DynamoDB client initialization
- ✅ List all tables in your account
- ✅ Verify all required tables exist
- ✅ Test basic table access

### Expected Output

```
Testing DynamoDB connection in region: us-east-1

✅ DynamoDB client initialized successfully

Found 12 table(s) in DynamoDB:
  - users
  - notifications
  - projects
  - calendar_events
  - tasks
  - project_details
  - activity_logs
  - leaveRequests
  - signupInvitations
  - userAccountLinks
  - info-portal

Checking required tables...
  ✅ users - exists
  ✅ notifications - exists
  ✅ projects - exists
  ✅ calendar_events - exists
  ✅ tasks - exists
  ✅ project_details - exists
  ✅ activity_logs - exists
  ✅ leaveRequests - exists
  ✅ signupInvitations - exists
  ✅ userAccountLinks - exists
  ✅ info-portal - exists

Testing basic operation on 'users' table...
✅ Successfully accessed 'users' table

🎉 All checks passed! DynamoDB connection is working correctly.
```

## Step 3: Verify Table Names Match

Ensure your table names in AWS match exactly:

| Repository | Table Name |
|------------|------------|
| UserRepo | `users` |
| NotificationRepo | `notifications` |
| ProjectRepo | `projects` |
| CalendarEventRepo | `calendar_events` |
| TaskRepo | `tasks` |
| ProjectDetailsRepo | `project_details` |
| ActivityLogRepo | `activity_logs` |
| VacationRepo | `leaveRequests` |
| SignupInvitationRepo | `signupInvitations` |
| UserAccountLinkRepo | `userAccountLinks` |
| InfoPortalRepo | `info-portal` |

## Step 4: Verify Global Secondary Indexes (GSIs)

Check that all required GSIs are created. You can verify in AWS Console or using:

```bash
aws dynamodb describe-table --table-name users
```

### Required GSIs by Table

**users:**
- `email-index` (email)

**notifications:**
- `userId-index` (userId)

**tasks:**
- `projectId-index` (projectId)

**project_details:**
- `projectId-index` (projectId)

**activity_logs:**
- `entityId-index` (entityId)

**leaveRequests:**
- `userId-index` (userId)
- `status-index` (status)

**signupInvitations:**
- `email-index` (email)
- `token-index` (token)

**userAccountLinks:**
- `userId-index` (userId)

## Step 5: Test Application Startup

Start the server and verify DynamoDB initialization:

```bash
go run cmd/server/main.go
```

Look for:
```
DynamoDB initialized successfully
```

If you see errors, check:
1. AWS credentials are configured correctly
2. Region matches your tables
3. IAM permissions allow DynamoDB access

## Step 6: Test Basic Operations

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

This should:
- Create a user in DynamoDB `users` table
- Hash the password
- Return JWT tokens

### Test User Login

```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123"
  }'
```

This should:
- Query user by email using `email-index` GSI
- Verify password
- Return JWT tokens

## Step 7: Verify Data in AWS Console

1. Go to AWS Console → DynamoDB
2. Select your region
3. Check each table:
   - Verify items are being created
   - Check GSI usage
   - Monitor read/write capacity

## Troubleshooting

### Error: "NoCredentialProviders: no valid providers in chain"

**Solution:** Configure AWS credentials:
```bash
aws configure
# OR
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
```

### Error: "ResourceNotFoundException: Requested resource not found"

**Solution:** 
- Verify table name matches exactly (case-sensitive)
- Check table exists in the correct region
- Verify `AWS_REGION` environment variable

### Error: "AccessDeniedException: User is not authorized"

**Solution:**
- Check IAM permissions for DynamoDB
- Required permissions:
  - `dynamodb:GetItem`
  - `dynamodb:PutItem`
  - `dynamodb:UpdateItem`
  - `dynamodb:DeleteItem`
  - `dynamodb:Query`
  - `dynamodb:Scan`
  - `dynamodb:DescribeTable`

### Table Not Found

**Solution:**
- Verify table name in code matches AWS table name
- Check region matches
- Use AWS CLI to list tables:
  ```bash
  aws dynamodb list-tables --region us-east-1
  ```

## Next Steps

Once connection is verified:

1. ✅ Test user registration and login
2. ✅ Test CRUD operations for each repository
3. ✅ Monitor DynamoDB metrics in AWS Console
4. ✅ Check GSI usage and optimize if needed
5. ✅ Set up CloudWatch alarms for capacity issues

## Performance Tips

- Monitor read/write capacity units
- Use GSI queries instead of Scans when possible
- Consider adding more GSIs for frequently queried attributes
- Set up auto-scaling if needed (beyond free tier)

---

**Status**: Ready to test! Run `go run scripts/test_dynamodb_connection.go` to verify everything is working.

