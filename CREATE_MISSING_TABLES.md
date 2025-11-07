# Create Missing DynamoDB Tables

## Missing Tables

You need to create these 4 tables with the **exact names** your code expects:

1. `leaveRequests` (not `vacations`)
2. `signupInvitations` (not `signup_invitations`)
3. `userAccountLinks` (not `user_account_links`)
4. `info-portal` (not `info_portal`)

## Table Creation Commands

### 1. leaveRequests Table

```bash
aws dynamodb create-table \
    --table-name leaveRequests \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=userId,AttributeType=S \
        AttributeName=status,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=userId-index,KeySchema=[{AttributeName=userId,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
        IndexName=status-index,KeySchema=[{AttributeName=status,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 2. signupInvitations Table

```bash
aws dynamodb create-table \
    --table-name signupInvitations \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=email,AttributeType=S \
        AttributeName=token,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=email-index,KeySchema=[{AttributeName=email,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
        IndexName=token-index,KeySchema=[{AttributeName=token,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### 3. userAccountLinks Table

```bash
aws dynamodb create-table \
    --table-name userAccountLinks \
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

### 4. info-portal Table

```bash
aws dynamodb create-table \
    --table-name info-portal \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --billing-mode PROVISIONED \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

## Quick Copy-Paste (All at Once)

Copy and run these commands one by one:

```bash
# 1. leaveRequests
aws dynamodb create-table --table-name leaveRequests --attribute-definitions AttributeName=id,AttributeType=S AttributeName=userId,AttributeType=S AttributeName=status,AttributeType=S --key-schema AttributeName=id,KeyType=HASH --global-secondary-indexes IndexName=userId-index,KeySchema=[{AttributeName=userId,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} IndexName=status-index,KeySchema=[{AttributeName=status,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} --billing-mode PROVISIONED --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

# 2. signupInvitations
aws dynamodb create-table --table-name signupInvitations --attribute-definitions AttributeName=id,AttributeType=S AttributeName=email,AttributeType=S AttributeName=token,AttributeType=S --key-schema AttributeName=id,KeyType=HASH --global-secondary-indexes IndexName=email-index,KeySchema=[{AttributeName=email,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} IndexName=token-index,KeySchema=[{AttributeName=token,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} --billing-mode PROVISIONED --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

# 3. userAccountLinks
aws dynamodb create-table --table-name userAccountLinks --attribute-definitions AttributeName=id,AttributeType=S AttributeName=userId,AttributeType=S --key-schema AttributeName=id,KeyType=HASH --global-secondary-indexes IndexName=userId-index,KeySchema=[{AttributeName=userId,KeyType=HASH}],Projection={ProjectionType=ALL},ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5} --billing-mode PROVISIONED --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

# 4. info-portal
aws dynamodb create-table --table-name info-portal --attribute-definitions AttributeName=id,AttributeType=S --key-schema AttributeName=id,KeyType=HASH --billing-mode PROVISIONED --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

## Verify Tables Created

After creating, verify:

```bash
aws dynamodb list-tables
```

Or run the test script again:

```bash
go run scripts/test_dynamodb_connection.go
```

## Important Notes

- **Table names are case-sensitive** - must match exactly:
  - `leaveRequests` (camelCase)
  - `signupInvitations` (camelCase)
  - `userAccountLinks` (camelCase)
  - `info-portal` (with hyphen)

- **GSIs Required:**
  - `leaveRequests`: `userId-index`, `status-index`
  - `signupInvitations`: `email-index`, `token-index`
  - `userAccountLinks`: `userId-index`
  - `info-portal`: No GSI needed

- **All tables use PROVISIONED billing mode** with 5 RCU/5 WCU to stay within free tier

