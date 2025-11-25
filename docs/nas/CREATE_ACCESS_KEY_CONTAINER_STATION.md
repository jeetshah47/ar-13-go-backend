# Create Access Keys Using Container Station

Since you can see the MinIO container environment variables in Container Station, here's how to create access keys using the terminal.

## What You Have

From the environment variables shown:
- **MINIO_ROOT_USER**: `minioadmin`
- **MINIO_ROOT_PASSWORD**: `minioadmin`
- **TZ**: `Asia/Kolkata`

These are your **admin credentials** - you'll use them to create access keys.

## Method 1: Using Container Station Terminal (Recommended)

### Step 1: Open Terminal

1. In Container Station, click on your **minio** container
2. Click the **"Attach Terminal"** tab
3. If you see the interactive/TTY error, use **"Execute"** button instead:
   - Click **"Execute"** button
   - Enter: `/bin/sh`
   - Click **Execute**

### Step 2: Install MinIO Client

Once you have a terminal prompt, run:

```bash
# Download MinIO client
wget https://dl.min.io/client/mc/release/linux-amd64/mc

# If wget doesn't work, use curl:
curl https://dl.min.io/client/mc/release/linux-amd64/mc -o mc

# Make it executable
chmod +x mc

# Move to system path
mv mc /usr/local/bin/mc
```

### Step 3: Configure MinIO Connection

```bash
# Use localhost since we're inside the container
# Use the credentials from your environment variables
mc alias set myminio https://localhost:9000 minioadmin minioadmin
```

**Note:** If SSL is not configured, use `http://` instead:
```bash
mc alias set myminio http://localhost:9000 minioadmin minioadmin
```

### Step 4: Create Access Key

```bash
mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
```

### Step 5: Copy the Output

You'll see something like:
```
Access Key: ABCDEFGHIJKLMNOPQRST
Secret Key: abcdefghijklmnopqrstuvwxyz1234567890ABCDEF
```

**⚠️ IMPORTANT:** Copy both keys immediately! You won't see the Secret Key again.

## Method 2: Using MinIO Web Console

### Step 1: Access Console

1. Open browser and go to: `https://YOUR-NAS-IP:9001` (or `http://` if SSL not enabled)
2. Login with:
   - Username: `minioadmin` (from MINIO_ROOT_USER)
   - Password: `minioadmin` (from MINIO_ROOT_PASSWORD)

### Step 2: Navigate to Access Keys

1. Look for **Identity** menu in the left sidebar
2. If you don't see it, try these direct URLs:
   - `https://YOUR-NAS-IP:9001/identity/access-keys`
   - `https://YOUR-NAS-IP:9001/admin/identity/access-keys`

3. Click **"Create Access Key"**

4. Fill in:
   - **Access Key Name:** `ar-13-backend-key`
   - **Policy:** Select `readwrite`

5. Click **Create**

6. **Copy both Access Key and Secret Key immediately!**

## Method 3: Using Execute Button (If Terminal Won't Type)

If the terminal won't let you type, use the **"Execute"** button instead:

1. Click **"Execute"** button in Container Station
2. Enter this all-in-one command:
   ```bash
   wget -q -O /tmp/mc https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x /tmp/mc && /tmp/mc alias set myminio http://localhost:9000 minioadmin minioadmin && /tmp/mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
   ```

3. Click **Execute**

This will:
- Download mc client
- Configure connection
- Create access key
- Show the output with keys

**See `USE_EXECUTE_BUTTON.md` for detailed step-by-step instructions.**

## After Getting Access Keys

Add them to your backend `.env` file:

```env
MINIO_ENDPOINT=your-nas-ip:9000
MINIO_ACCESS_KEY=ABCDEFGHIJKLMNOPQRST
MINIO_SECRET_KEY=abcdefghijklmnopqrstuvwxyz1234567890ABCDEF
MINIO_BUCKET=ar-13-uploads
MINIO_USE_SSL=true
MINIO_INSECURE_SSL=false
```

## Troubleshooting

### "Execute" Command Not Working

Try breaking it into steps:
1. First execute: `wget -q -O /tmp/mc https://dl.min.io/client/mc/release/linux-amd64/mc`
2. Then: `chmod +x /tmp/mc`
3. Then: `/tmp/mc alias set myminio http://localhost:9000 minioadmin minioadmin`
4. Finally: `/tmp/mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite`

### Can't Access Web Console

- Make sure port 9001 is accessible
- Try both `http://` and `https://`
- Check firewall settings

### Time Sync Issues

If you get time sync errors when using mc from your computer, use Container Station terminal instead (Method 1) - it avoids time sync issues.

## Quick Reference

**Root Credentials (from environment):**
- Username: `minioadmin`
- Password: `minioadmin`

**Use these to:**
- Login to web console
- Create access keys
- Admin operations

**Access Keys (what you create):**
- Used by your backend application
- Limited permissions (readwrite policy)
- More secure than using root credentials

