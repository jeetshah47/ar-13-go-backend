# Using Container Station "Execute" Button to Create Access Keys

If the terminal won't let you type, use the **"Execute"** button instead. This runs commands directly without needing an interactive terminal.

## Step-by-Step Guide

### Step 1: Open Execute Dialog

1. In Container Station, click on your **minio** container
2. Look for the **"Execute"** button (usually near "Stop", "Edit", "Inspect" buttons)
3. Click **"Execute"**

### Step 2: Install MinIO Client (One-Time Setup)

In the Execute dialog, enter this command:

```bash
wget -q -O /tmp/mc https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x /tmp/mc && /tmp/mc --version
```

Click **Execute**. This will:
- Download MinIO client
- Make it executable
- Verify it works

You should see version information if successful.

### Step 3: Configure MinIO Connection

Enter this command:

```bash
/tmp/mc alias set myminio http://localhost:9000 minioadmin minioadmin
```

**Important Notes:** 
- **Use `http://` for localhost connections** - Even if SSL is configured for external access, use HTTP when connecting from inside the container
- If you get "HTTP response to HTTPS client" error, you're using `https://` - change to `http://`
- Replace `minioadmin` with your actual `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD` if different

Click **Execute**. You should see a success message.

### Step 4: Create Access Key

Enter this command:

```bash
/tmp/mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
```

Click **Execute**.

### Step 5: Copy the Output

You'll see output like:
```
Access Key: ABCDEFGHIJKLMNOPQRST
Secret Key: abcdefghijklmnopqrstuvwxyz1234567890ABCDEF
```

**⚠️ IMPORTANT:** Copy both keys immediately! You won't see the Secret Key again.

## All-in-One Command (Alternative)

If you want to do everything in one go, use this single command:

```bash
wget -q -O /tmp/mc https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x /tmp/mc && /tmp/mc alias set myminio http://localhost:9000 minioadmin minioadmin && /tmp/mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
```

**Note:** Always use `http://` (not `https://`) when connecting to `localhost:9000` from inside the container, even if SSL is configured for external access.

This will:
1. Download mc client
2. Make it executable
3. Configure connection
4. Create access key
5. Show the output

## Troubleshooting

### "wget: command not found"

Use `curl` instead:

```bash
curl -s -o /tmp/mc https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x /tmp/mc && /tmp/mc alias set myminio http://localhost:9000 minioadmin minioadmin && /tmp/mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
```

### "Unable to initialize new alias"

1. **Check if MinIO is running:**
   - Look at container status - should be "Running"
   - Check logs if needed

2. **"HTTP response to HTTPS client" error:**
   - You're using `https://` but MinIO is running on HTTP
   - **Solution:** Use `http://localhost:9000` (not `https://`)
   - Even if SSL is configured for external access, use HTTP for localhost connections

3. **"HTTPS response to HTTP client" error:**
   - MinIO is configured for SSL but you're using HTTP
   - **Solution:** Use `https://localhost:9000` instead

4. **Verify credentials:**
   - Check your environment variables in Container Station
   - Make sure you're using `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD` values

### "Access Denied"

- Make sure you're using root credentials (MINIO_ROOT_USER/MINIO_ROOT_PASSWORD)
- Not a regular access key

### No Output or Empty Response

1. **Check if command executed:**
   - Look for any error messages
   - Try running `/tmp/mc --version` first to verify mc is installed

2. **Try step by step:**
   - Don't use the all-in-one command
   - Run each step separately

## Step-by-Step Breakdown (If All-in-One Doesn't Work)

Run these commands one by one in the Execute dialog:

**Step 1: Download mc**
```bash
wget -q -O /tmp/mc https://dl.min.io/client/mc/release/linux-amd64/mc
```

**Step 2: Make executable**
```bash
chmod +x /tmp/mc
```

**Step 3: Verify installation**
```bash
/tmp/mc --version
```

**Step 4: Configure alias**
```bash
/tmp/mc alias set myminio http://localhost:9000 minioadmin minioadmin
```

**Step 5: Create access key**
```bash
/tmp/mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
```

## After Getting Access Keys

Add to your backend `.env` file:

```env
MINIO_ENDPOINT=your-nas-ip:9000
MINIO_ACCESS_KEY=ABCDEFGHIJKLMNOPQRST
MINIO_SECRET_KEY=abcdefghijklmnopqrstuvwxyz1234567890ABCDEF
MINIO_BUCKET=ar-13-uploads
MINIO_USE_SSL=true
MINIO_INSECURE_SSL=false
```

## Why Use Execute Instead of Terminal?

- **No typing issues** - Just paste and click
- **No TTY requirements** - Works even if terminal attachment is disabled
- **Clear output** - See results immediately
- **No time sync issues** - Runs inside the container

## Quick Reference

**Your credentials (from environment variables):**
- Username: `minioadmin` (MINIO_ROOT_USER)
- Password: `minioadmin` (MINIO_ROOT_PASSWORD)

**Execute command format:**
```bash
/tmp/mc [command] [options]
```

**Common commands:**
- `alias set` - Configure connection
- `admin user svcacct add` - Create access key
- `ls` - List buckets
- `mb` - Create bucket

