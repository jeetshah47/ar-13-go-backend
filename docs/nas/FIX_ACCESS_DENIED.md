# Fix: "Access Denied" Error

If you get "Access Denied" when using your access keys, it means the service account doesn't have the necessary permissions.

## Problem

When we created the service account earlier, we couldn't use `--policy readwrite` because it expects a file path, not a policy name. So the service account was created without a policy, meaning it has no permissions.

## Solution: Create and Attach a Policy

### Step 1: Create a Policy File

Create a JSON policy file that grants read/write access. You can do this in Container Station terminal or on your local machine.

**Policy file content** (`readwrite-policy.json`):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:GetBucketLocation",
        "s3:ListBucketMultipartUploads"
      ],
      "Resource": [
        "arn:aws:s3:::ar-13-uploads"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:DeleteObject",
        "s3:AbortMultipartUpload",
        "s3:ListMultipartUploadParts"
      ],
      "Resource": [
        "arn:aws:s3:::ar-13-uploads/*"
      ]
    }
  ]
}
```

### Step 2: Upload Policy to MinIO

**Option A: Using Container Station Terminal**

1. Open Container Station terminal for your MinIO container
2. Create the policy file:
   ```bash
   cat > /tmp/readwrite-policy.json << 'EOF'
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "s3:ListBucket",
           "s3:GetBucketLocation",
           "s3:ListBucketMultipartUploads"
         ],
         "Resource": [
           "arn:aws:s3:::ar-13-uploads"
         ]
       },
       {
         "Effect": "Allow",
         "Action": [
           "s3:PutObject",
           "s3:GetObject",
           "s3:DeleteObject",
           "s3:AbortMultipartUpload",
           "s3:ListMultipartUploadParts"
         ],
         "Resource": [
           "arn:aws:s3:::ar-13-uploads/*"
         ]
       }
     ]
   }
   EOF
   ```

3. Create the policy in MinIO:
   ```bash
   /tmp/mc admin policy create myminio readwrite-policy /tmp/readwrite-policy.json
   ```

4. Attach policy to your service account user:
   ```bash
   /tmp/mc admin policy attach myminio readwrite-policy --user ar-13-backend-user
   ```
   
   **Note:** Use `--user` flag (not `user=`)

**Option B: Using MinIO Console (Web UI)**

1. Go to MinIO Console: `https://your-nas-ip:9001`
2. Login with root credentials
3. Go to **Identity** → **Policies**
4. Click **Create Policy**
5. Name: `readwrite-policy`
6. Paste the JSON policy content
7. Click **Save**
8. Go to **Identity** → **Users**
9. Find `ar-13-backend-user`
10. Click on it and attach the `readwrite-policy`

### Step 3: Verify Permissions

Test the connection again:

```bash
go run scripts\test_minio_upload.go --endpoint "192.168.0.118:9000" --access-key "your-key" --secret-key "your-secret" --file demo.txt
```

## Alternative: Use Root Credentials (Not Recommended for Production)

For testing only, you can temporarily use root credentials:

```env
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
```

**⚠️ Warning:** Don't use root credentials in production! Create proper service accounts with limited permissions.

## Quick Fix: Recreate Service Account with Policy

If you want to start fresh:

1. **Delete the old service account** (via MinIO Console or mc command)
2. **Create the policy** (as shown above)
3. **Create a new service account with the policy attached:**

   ```bash
   # In Container Station terminal
   /tmp/mc admin policy create myminio readwrite-policy /tmp/readwrite-policy.json
   /tmp/mc admin policy attach myminio readwrite-policy --user ar-13-backend-user
   /tmp/mc admin user svcacct add myminio ar-13-backend-user --name ar-13-backend-key
   ```

## Policy Explanation

The policy grants:
- **ListBucket**: List objects in the bucket
- **GetBucketLocation**: Get bucket location
- **PutObject**: Upload files
- **GetObject**: Download files
- **DeleteObject**: Delete files
- **Multipart operations**: For large file uploads

## Troubleshooting

### "Policy not found"

Make sure you created the policy before attaching it:
```bash
/tmp/mc admin policy list myminio
```

### "User not found"

Make sure the user exists:
```bash
/tmp/mc admin user list myminio
```

### Still Getting "Access Denied"

1. **Verify policy is attached:**
   ```bash
   /tmp/mc admin user info myminio ar-13-backend-user
   ```

2. **Check bucket name matches:**
   - Policy uses: `arn:aws:s3:::ar-13-uploads`
   - Make sure your bucket name matches

3. **Try with root credentials** to verify connection works:
   ```bash
   go run scripts\test_minio_upload.go --endpoint "192.168.0.118:9000" --access-key "minioadmin" --secret-key "minioadmin" --file demo.txt
   ```

## Summary

The issue is that service accounts need explicit policies. Create a policy file, add it to MinIO, and attach it to your service account user.

