# MinIO Setup Guide for NAS (with SSL)

This guide will help you set up MinIO on your NAS with SSL/TLS support.

## Prerequisites

- NAS with Docker/Container Station installed
- SSL certificates (or ability to generate self-signed certificates)
- SSH access to your NAS (recommended)

## Step 1: Prepare Directories

Create the necessary directories on your NAS:

```bash
# Via SSH or File Station
mkdir -p /share/Container/minio/data
mkdir -p /share/Container/minio/config
mkdir -p /share/Container/minio/certs
```

**Note:** If using File Station, navigate to `/share/Container/` and create:
- `minio/data` folder
- `minio/config` folder
- `minio/certs` folder

## Step 2: SSL Certificate Setup

### Option A: Using Existing SSL Certificates

1. Copy your SSL certificates to the NAS:
   - `public.crt` → `/share/Container/minio/certs/public.crt`
   - `private.key` → `/share/Container/minio/certs/private.key`

2. Ensure proper permissions:
   ```bash
   chmod 600 /share/Container/minio/certs/private.key
   chmod 644 /share/Container/minio/certs/public.crt
   ```

### Option B: Generate Self-Signed Certificates

1. SSH into your NAS or use a machine with OpenSSL
2. Generate certificates:
   ```bash
   cd /share/Container/minio/certs
   
   # Generate private key
   openssl genrsa -out private.key 2048
   
   # Generate certificate signing request
   openssl req -new -key private.key -out cert.csr -subj "/CN=your-nas-domain.com"
   
   # Generate self-signed certificate (valid for 365 days)
   openssl x509 -req -days 365 -in cert.csr -signkey private.key -out public.crt
   
   # Set permissions
   chmod 600 private.key
   chmod 644 public.crt
   
   # Clean up
   rm cert.csr
   ```

**Note:** Replace `your-nas-domain.com` with your actual NAS domain or IP.

## Step 3: Upload docker-compose File

1. Copy `docker/docker-compose.minio.nas.yml` to your NAS
2. Place it in: `/share/Container/minio/docker-compose.minio.nas.yml`

## Step 4: Configure MinIO Credentials

**IMPORTANT:** Before starting, change the default credentials:

1. Edit `docker-compose.minio.nas.yml`
2. Change these values:
   ```yaml
   MINIO_ROOT_USER: your-secure-username
   MINIO_ROOT_PASSWORD: your-secure-password-min-8-chars
   ```
   - Use a strong password (at least 12 characters)
   - These are your **admin credentials** for the MinIO Console

## Step 5: Deploy MinIO

### Option A: Using Container Station Web Interface

1. Open **Container Station** on your NAS
2. Go to **Container** → **Create** → **Create Application**
3. Click **Create from YAML**
4. Copy and paste the contents of `docker-compose.minio.nas.yml`
5. Click **Create**
6. Wait for the container to start

### Option B: Using SSH (Recommended)

1. SSH into your NAS
2. Navigate to where you placed the docker-compose file:
   ```bash
   cd /share/Container/minio
   ```
3. Start MinIO:
   ```bash
   docker-compose -f docker/docker-compose.minio.nas.yml up -d
   ```
4. Check if it's running:
   ```bash
   docker ps | grep minio
   ```

## Step 6: Access MinIO Console

1. Open your web browser
2. Navigate to:
   - **HTTPS:** `https://your-nas-ip:9001` or `https://your-nas-domain:9001`
   - If using self-signed certificate, accept the security warning
3. Login with:
   - **Username:** The value you set for `MINIO_ROOT_USER`
   - **Password:** The value you set for `MINIO_ROOT_PASSWORD`

**Important:** MinIO has two interfaces:
- **Object Browser** (default): For browsing/managing files - NO Identity tab
- **Admin Console**: For administration - HAS Identity/Settings tabs

## Step 7: Create Access Keys

### Method 1: Access Admin Console (Recommended)

If you don't see the **Identity** menu, you're in the Object Browser. Switch to Admin Console:

1. **Look for the gear icon (⚙️) or Settings icon** in the top-right corner
2. **Or try these direct URLs:**
   - `https://your-nas-ip:9001/identity/access-keys`
   - `https://your-nas-ip:9001/admin/identity/access-keys`
   - `https://your-nas-ip:9001/settings/identity/access-keys`

3. **Or click on your username** in the top-right corner and look for "Admin" or "Settings" option

4. Once in Admin Console, you should see:
   - **Identity** menu in the left sidebar
   - **Access Keys** under Identity
   - **Settings** menu
   - **Monitoring** menu

5. Navigate to **Identity** → **Access Keys**

6. Click **Create Access Key**

7. Fill in:
   - **Access Key Name:** e.g., `ar-13-backend-key`
   - **Policy:** Select `readwrite` or create custom policy

8. Click **Create**

9. **IMPORTANT:** Copy and save the **Access Key** and **Secret Key** immediately (you won't see the secret key again)

### Method 2: Using Container Station Terminal (Easiest for NAS)

If you have Container Station terminal access, this is the easiest method:

1. **Open Container Station** on your NAS
2. **Find your MinIO container** and click on it
3. **Click "Terminal" or "Console"** to open a terminal inside the container
4. **Install MinIO Client (mc) inside the container:**
   ```bash
   # Download MinIO client
   wget https://dl.min.io/client/mc/release/linux-amd64/mc
   chmod +x mc
   mv mc /usr/local/bin/mc
   ```

5. **Configure the MinIO client:**
   ```bash
   # Use localhost since we're inside the container
   mc alias set myminio https://localhost:9000 minioadmin minioadmin
   ```
   
   **Note:** Replace `minioadmin` with your actual `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD` from docker-compose.

6. **Create a service account (access key):**
   ```bash
   # Create service account
   mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
   ```

7. **The output will show:**
   ```
   Access Key: xxxxxx
   Secret Key: xxxxxx
   ```
   **Copy these immediately!**

**Alternative:** If `wget` is not available, you can use `curl`:
```bash
curl https://dl.min.io/client/mc/release/linux-amd64/mc -o mc
chmod +x mc
mv mc /usr/local/bin/mc
```

### Method 3: Using MinIO Client (mc) - From Your Computer

If you prefer to use MinIO Client from your local machine:

1. **Install MinIO Client:**
   
   **On Windows (PowerShell as Administrator):**
   ```powershell
   # Using Chocolatey
   choco install minio-client
   
   # Or download manually
   # Download from: https://dl.min.io/client/mc/release/windows-amd64/mc.exe
   # Place mc.exe in a folder (e.g., C:\Tools\minio\mc.exe)
   ```

   **On Linux/Mac:**
   ```bash
   wget https://dl.min.io/client/mc/release/linux-amd64/mc
   chmod +x mc
   sudo mv mc /usr/local/bin/
   ```

2. **Configure the MinIO client:**
   ```bash
   # Replace with your actual credentials and endpoint
   mc alias set myminio https://your-nas-ip:9000 minioadmin minioadmin
   ```

3. **Create a service account (access key):**
   ```bash
   # Create service account
   mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
   ```

4. **The output will show:**
   ```
   Access Key: xxxxxx
   Secret Key: xxxxxx
   ```
   **Copy these immediately!**

### Method 4: Using MinIO API (Direct HTTP Request)

You can create access keys using the MinIO Admin API:

1. **Get an admin session token:**
   ```bash
   curl -X POST https://your-nas-ip:9001/api/v1/login \
     -H "Content-Type: application/json" \
     -d '{"accessKey":"minioadmin","secretKey":"minioadmin"}' \
     -k
   ```
   (Replace `minioadmin` with your actual `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD`)

2. **Use the token to create service account:**
   ```bash
   curl -X POST https://your-nas-ip:9001/api/v1/service-accounts \
     -H "Authorization: Bearer YOUR_TOKEN_HERE" \
     -H "Content-Type: application/json" \
     -d '{"accessKey":"ar-13-backend","policy":"readwrite"}' \
     -k
   ```

### Troubleshooting: Can't Find Identity Menu

**If you still can't see Identity menu:**

1. **Make sure you're logged in with root credentials** (MINIO_ROOT_USER and MINIO_ROOT_PASSWORD)

2. **Check MinIO version:**
   ```bash
   docker exec minio minio --version
   ```
   Older versions might have different UI. Update to latest version if needed.

3. **Try accessing via direct URL:**
   - `https://your-nas-ip:9001/identity`
   - `https://your-nas-ip:9001/admin`

4. **Clear browser cache** and try again

5. **Check browser console** (F12) for any JavaScript errors

6. **Use Method 2 (MinIO Client)** as a reliable alternative

## Step 8: Create Bucket

1. In MinIO Console, go to **Buckets**
2. Click **Create Bucket**
3. Name it: `ar-13-uploads`
4. Set access policy (usually **Private** for application use)
5. Click **Create**

## Step 9: Configure Your Backend

Add these to your `.env` file:

```env
# MinIO/NAS Storage Configuration (with SSL)
MINIO_ENDPOINT=your-nas-domain:9000
MINIO_ACCESS_KEY=your-access-key-from-step-7
MINIO_SECRET_KEY=your-secret-key-from-step-7
MINIO_BUCKET=ar-13-uploads
MINIO_USE_SSL=true
MINIO_INSECURE_SSL=false
```

**Note:** 
- Use port `9000` for the API endpoint (internal port, or your mapped port)
- Use `MINIO_USE_SSL=true` for SSL connection
- Set `MINIO_INSECURE_SSL=true` only if using self-signed certificates and you want to skip verification

## Step 10: Test the Connection

Test your MinIO setup using the provided test script:

```bash
go run scripts/test_minio_upload.go \
  --endpoint "your-nas-domain:9000" \
  --access-key "your-access-key" \
  --secret-key "your-secret-key" \
  --bucket "ar-13-uploads" \
  --file "test.txt"
```

## Troubleshooting

### Container Won't Start

1. Check logs:
   ```bash
   docker logs minio
   ```

2. Verify certificates exist and have correct permissions:
   ```bash
   ls -la /share/Container/minio/certs/
   ```

3. Check certificate format:
   ```bash
   openssl x509 -in /share/Container/minio/certs/public.crt -text -noout
   ```

### SSL Certificate Errors

1. **Self-signed certificate warning:** This is normal. Accept the warning in your browser or set `MINIO_INSECURE_SSL=true` in your backend config.

2. **Certificate not found:** Ensure certificates are in `/share/Container/minio/certs/` with correct names (`public.crt` and `private.key`).

3. **Permission denied:** Set correct permissions:
   ```bash
   chmod 600 /share/Container/minio/certs/private.key
   chmod 644 /share/Container/minio/certs/public.crt
   ```

### Can't Access Console

1. Verify the container is running:
   ```bash
   docker ps | grep minio
   ```

2. Check firewall settings on NAS

3. Verify port forwarding is configured correctly

4. Try accessing via local IP: `https://192.168.x.x:9001`

### Connection Refused from Backend

1. Verify SSL is enabled in your `.env` file (`MINIO_USE_SSL=true`)

2. If using self-signed certificate, set `MINIO_INSECURE_SSL=true`

3. Check if the endpoint URL is correct (include port if not standard)

## Port Reference

- **9000** (internal): MinIO API (S3 operations) - SSL enabled
- **9001** (internal): MinIO Console (Web UI) - SSL enabled

If you need to map different external ports, update the `ports` section in docker-compose file.

## Security Recommendations

1. **Change Default Credentials:** Always change `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD`
2. **Use Strong Passwords:** At least 12 characters with mixed case, numbers, and symbols
3. **Use Valid SSL Certificates:** For production, use certificates from a trusted CA (Let's Encrypt, etc.)
4. **Limit Access:** Only create access keys with necessary permissions
5. **Firewall Rules:** Restrict access to MinIO ports from trusted networks only
6. **Regular Backups:** Backup the `/share/Container/minio/data` directory regularly
7. **Keep Certificates Secure:** Protect your private key file

## Reset MinIO (Start Fresh)

If you need to reset everything:

```bash
# Stop and remove container
docker-compose -f docker/docker-compose.minio.nas.yml down

# Remove data (WARNING: This deletes all files!)
rm -rf /share/Container/minio/data/*

# Start again
docker-compose -f docker/docker-compose.minio.nas.yml up -d
```

## Additional Resources

- [MinIO Documentation](https://min.io/docs/)
- [MinIO SSL/TLS Setup](https://min.io/docs/minio/container/operations/network-encryption.html)
- [MinIO Docker Hub](https://hub.docker.com/r/minio/minio)

