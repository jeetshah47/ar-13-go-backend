# IAM Role Setup for EC2 - DynamoDB & S3 Access

## Why Use IAM Role?

Instead of hardcoding AWS credentials in your application, use an IAM role attached to your EC2 instance. This is:
- ✅ More secure (no credentials in code)
- ✅ Automatically rotated
- ✅ Easier to manage
- ✅ AWS best practice

---

## Step 1: Create IAM Role

### 1.1 Go to IAM Console
1. Navigate to **IAM** → **Roles**
2. Click **Create role**

### 1.2 Select Trusted Entity
- **Trusted entity type:** AWS service
- **Use case:** EC2
- Click **Next**

### 1.3 Add Permissions

**Policy 1: DynamoDB Full Access** (or create custom policy)
- Search: `AmazonDynamoDBFullAccess`
- Select it

**Policy 2: S3 Access** (for file uploads)
- Search: `AmazonS3FullAccess` (or create custom policy for specific bucket)
- Select it

**OR Create Custom Policy** (More Secure):

Click **Create policy** → **JSON** tab, paste:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:PutItem",
                "dynamodb:GetItem",
                "dynamodb:UpdateItem",
                "dynamodb:DeleteItem",
                "dynamodb:Query",
                "dynamodb:Scan",
                "dynamodb:BatchGetItem",
                "dynamodb:BatchWriteItem"
            ],
            "Resource": [
                "arn:aws:dynamodb:*:*:table/*",
                "arn:aws:dynamodb:*:*:table/*/index/*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:DeleteObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::ar-13-uploads",
                "arn:aws:s3:::ar-13-uploads/*"
            ]
        }
    ]
}
```

Name: `AR13BackendPolicy`
Click **Create policy**

Then attach this custom policy to your role.

### 1.4 Name Role
- **Role name:** `AR13BackendRole`
- **Description:** `IAM role for AR-13 backend EC2 instance`
- Click **Create role**

---

## Step 2: Attach Role to EC2 Instance

### 2.1 Go to EC2 Console
1. Navigate to **EC2** → **Instances**
2. Select your EC2 instance
3. Click **Actions** → **Security** → **Modify IAM role**

### 2.2 Select Role
- **IAM role:** Select `AR13BackendRole`
- Click **Update IAM role**

---

## Step 3: Verify Role is Working

### 3.1 SSH into EC2
```bash
ssh -i your-key.pem ubuntu@YOUR_EC2_IP
```

### 3.2 Install AWS CLI (if not installed)
```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
sudo apt install -y unzip
unzip awscliv2.zip
sudo ./aws/install
```

### 3.3 Test Access
```bash
# Check current identity (should show role, not user)
aws sts get-caller-identity

# Test DynamoDB access
aws dynamodb list-tables --region us-east-1

# Test S3 access (if bucket exists)
aws s3 ls s3://ar-13-uploads
```

---

## Step 4: Update Application

### 4.1 Remove AWS Credentials from .env
Your `.env` file should **NOT** have:
```env
# REMOVE THESE - Use IAM role instead
# AWS_ACCESS_KEY_ID=...
# AWS_SECRET_ACCESS_KEY=...
```

### 4.2 Application Will Auto-Detect Role
The AWS SDK will automatically use the IAM role credentials. No code changes needed!

---

## Troubleshooting

### Role Not Working?
1. **Check instance has role attached:**
   ```bash
   aws sts get-caller-identity
   ```

2. **Check role permissions:**
   - Go to IAM → Roles → AR13BackendRole
   - Check attached policies

3. **Verify trust relationship:**
   - Role should trust `ec2.amazonaws.com`

### Access Denied Errors?
1. Check DynamoDB table names match your region
2. Verify S3 bucket name and region
3. Check IAM policy resources match your resources

---

## Security Best Practices

1. ✅ **Use IAM roles** instead of access keys
2. ✅ **Principle of least privilege** - Only grant needed permissions
3. ✅ **Use custom policies** instead of full access policies
4. ✅ **Regularly audit** IAM roles and permissions
5. ✅ **Use resource-specific ARNs** in policies

---

## Example: Custom Policy for Specific Tables

If you want to restrict access to only your tables:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:PutItem",
                "dynamodb:GetItem",
                "dynamodb:UpdateItem",
                "dynamodb:DeleteItem",
                "dynamodb:Query",
                "dynamodb:Scan",
                "dynamodb:BatchGetItem",
                "dynamodb:BatchWriteItem"
            ],
            "Resource": [
                "arn:aws:dynamodb:us-east-1:YOUR_ACCOUNT_ID:table/users",
                "arn:aws:dynamodb:us-east-1:YOUR_ACCOUNT_ID:table/users/index/*",
                "arn:aws:dynamodb:us-east-1:YOUR_ACCOUNT_ID:table/projects",
                "arn:aws:dynamodb:us-east-1:YOUR_ACCOUNT_ID:table/tasks",
                "arn:aws:dynamodb:us-east-1:YOUR_ACCOUNT_ID:table/tasks/index/*"
            ]
        }
    ]
}
```

Replace `YOUR_ACCOUNT_ID` with your AWS account ID.

