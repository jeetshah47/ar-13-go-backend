# DynamoDB Tables Setup Guide

## Required Tables

You need to create the following DynamoDB tables. All tables use **on-demand billing** (pay per request) which is cost-effective for 10-15 users.

## Table Creation Commands

### 1. Users Table
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
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 2. Projects Table
```bash
aws dynamodb create-table \
    --table-name projects \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

**Project Table Fields:**
- `id` (String) - Primary key
- `title` (String) - Project title
- `description` (String) - Project description
- `ownerId` (String) - Owner user ID
- `membersIds` (List) - Array of member user IDs
- `deadLine` (String) - Project deadline in RFC3339 format (note: JSON field name is `deadLine` in camelCase)
- `logoUrl` (String, optional) - Project logo URL
- `created` (String) - Creation timestamp (RFC3339)
- `updated` (String, optional) - Last update timestamp (RFC3339)

### 3. Tasks Table
```bash
aws dynamodb create-table \
    --table-name tasks \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=projectId,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=projectId-index,KeySchema=[{AttributeName=projectId,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

**Task Table Fields:**
- `id` (String) - Primary key
- `subject` (String) - Task subject/title
- `code` (String) - Task code
- `status` (String) - Task status
- `deadline` (String) - Deadline in RFC3339 format (replaces old `duration` field)
- `priority` (String) - Task priority
- `progress` (Number) - Completion percentage (0-100, optional)
- `assignTo` (String, optional) - Assigned user ID
- `projectId` (String) - Project ID (indexed via GSI)
- `description` (String, optional) - Task description
- `timeSpent` (List) - Array of time spent entries
- `fileAttachments` (List) - Array of file attachments
- `activityLogs` (List) - Array of activity log entries
- `created` (String) - Creation timestamp (RFC3339)
- `updated` (String, optional) - Last update timestamp (RFC3339)

**Note**: See `docs/TASK_FIELD_MIGRATION.md` for migration details from `duration` to `deadline`.

### 4. Notifications Table
```bash
aws dynamodb create-table \
    --table-name notifications \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=userId,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=userId-index,KeySchema=[{AttributeName=userId,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 5. Calendar Events Table
```bash
aws dynamodb create-table \
    --table-name calendar_events \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 6. Vacations Table
```bash
aws dynamodb create-table \
    --table-name vacations \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=userId,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=userId-index,KeySchema=[{AttributeName=userId,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 7. Activity Logs Table
```bash
aws dynamodb create-table \
    --table-name activity_logs \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=entityId,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=entityId-index,KeySchema=[{AttributeName=entityId,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 8. Info Portal Table
```bash
aws dynamodb create-table \
    --table-name info_portal \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 9. Project Details Table
```bash
aws dynamodb create-table \
    --table-name project_details \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=projectId,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=projectId-index,KeySchema=[{AttributeName=projectId,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 10. User Account Links Table
```bash
aws dynamodb create-table \
    --table-name user_account_links \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=userId,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=userId-index,KeySchema=[{AttributeName=userId,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 11. Signup Invitations Table
```bash
aws dynamodb create-table \
    --table-name signup_invitations \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=token,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=token-index,KeySchema=[{AttributeName=token,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 12. Role Permissions Table
```bash
aws dynamodb create-table \
    --table-name role_permissions \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=role,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=role-index,KeySchema=[{AttributeName=role,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

**Role Permissions Table Fields:**
- `id` (String) - Primary key
- `role` (String) - User role (Admin, Standard) - indexed via GSI
- `permission` (String) - Permission string (e.g., "projects:read", "tasks:write")
- `createdAt` (String) - Creation timestamp (RFC3339)
- `updatedAt` (String) - Last update timestamp (RFC3339)

**Note**: After creating this table, run the seed script to populate initial permissions:
```bash
go run scripts/seed_role_permissions.go
```

## Using AWS Console

Alternatively, you can create tables using the AWS Console:

1. Go to DynamoDB Console
2. Click "Create table"
3. Enter table name
4. Set partition key (usually `id` as String)
5. Add Global Secondary Indexes if needed (for email, userId, etc.)
6. Choose **"Provisioned"** billing mode
7. Set **Read capacity**: 5 units, **Write capacity**: 5 units
8. For each GSI, also set 5 RCU and 5 WCU
9. Click "Create table"

## Verify Tables

After creating, verify tables exist:
```bash
aws dynamodb list-tables
```

## DynamoDB Free Tier

DynamoDB offers a **forever free tier** (not just 12 months) that includes:

### Free Tier Benefits:
- **25 GB of storage** - Free forever
- **25 provisioned Write Capacity Units (WCU)** - Free forever
- **25 provisioned Read Capacity Units (RCU)** - Free forever
- **2.5 million stream read requests** - Free forever

### For 10-15 Users:

**Option 1: Provisioned Capacity (Recommended for Free Tier)**
- Use provisioned capacity with 5 RCU and 5 WCU
- Well within free tier limits (25 RCU/WCU free)
- **Cost: $0/month** ✅

**Option 2: On-Demand (Pay Per Request)**
- First 12 months: Free tier covers:
  - 25 GB storage
  - 200 million read requests
  - 200 million write requests
- After 12 months: Pay per use (~$1-2/month for 10-15 users)

## Cost Estimate

### With Provisioned Capacity (Free Tier):
- **Storage**: Free (up to 25 GB)
- **Reads**: Free (up to 25 RCU)
- **Writes**: Free (up to 25 WCU)
- **Total**: **$0/month** ✅ (Forever free for your scale)

### With On-Demand (After Free Tier):
- **Storage**: ~$0.25/GB-month (likely < 1GB = ~$0.25/month)
- **Reads**: $0.25 per million (likely < 1M/month = ~$0.25/month)
- **Writes**: $1.25 per million (likely < 500K/month = ~$0.63/month)
- **Total**: ~$1-2/month

## Free Tier Capacity Planning

For 10-15 users with provisioned capacity:

**Recommended Settings:**
- **Base Table**: 5 RCU, 5 WCU (well within 25 RCU/WCU free tier)
- **GSI (each)**: 5 RCU, 5 WCU (each GSI has separate free tier allocation)

**Capacity Calculation:**
- 5 RCU = 5 reads/second = 432,000 reads/day (more than enough)
- 5 WCU = 5 writes/second = 432,000 writes/day (more than enough)

**Free Tier Coverage:**
- You're using 5/25 RCU = 20% of free tier
- You're using 5/25 WCU = 20% of free tier
- **Plenty of headroom for growth!**

## Notes

- All tables use `id` as the primary key (partition key)
- Global Secondary Indexes (GSI) are used for querying by email, userId, projectId, etc.
- **Provisioned capacity is recommended** to stay within free tier
- Tables are created in the region specified by `AWS_REGION` environment variable
- Free tier is **forever free** (not just 12 months) for provisioned capacity
- You can switch to on-demand later if needed (but will incur costs)

