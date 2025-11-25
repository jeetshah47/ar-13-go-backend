# MinIO Local Setup Guide (Docker)

This guide will help you set up MinIO locally using Docker for development.

## Prerequisites

- Docker Desktop installed and running
- Docker Compose (usually included with Docker Desktop)
- Basic knowledge of Docker

## Quick Start

### Step 1: Start MinIO

```bash
docker-compose -f docker-compose.minio.yml up -d
```

### Step 2: Access MinIO Console

1. Open your browser
2. Go to: `http://localhost:9001`
3. Login with:
   - Username: `minioadmin`
   - Password: `minioadmin`

### Step 3: Create a Bucket

1. In MinIO Console, click **Buckets** in the left sidebar
2. Click **Create Bucket**
3. Name it: `ar-13-uploads`
4. Set access policy to **Private**
5. Click **Create**

### Step 4: Create Access Keys

1. In MinIO Console, go to **Identity** → **Access Keys**
2. Click **Create Access Key**
3. Fill in:
   - **Access Key Name:** `ar-13-backend-key`
   - **Policy:** Leave empty (or select readwrite if available)
4. Click **Create**
5. **Copy both Access Key and Secret Key immediately!**

### Step 5: Configure Your Backend

Add to your `.env` file:

```env
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=your-access-key-here
MINIO_SECRET_KEY=your-secret-key-here
MINIO_BUCKET=ar-13-uploads
MINIO_USE_SSL=false
MINIO_INSECURE_SSL=false
```

### Step 6: Test the Connection

```bash
go run scripts/test_minio_upload.go --file demo.txt
```

## Docker Compose File

The `docker-compose.minio.yml` file contains:

```yaml
version: '3.8'

services:
  minio:
    image: minio/minio:latest
    container_name: minio
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"  # MinIO API
      - "9001:9001"  # MinIO Console
    volumes:
      - ./minio-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
      start_period: 40s
```

## Detailed Setup Steps

### Step 1: Verify Docker is Running

```bash
docker ps
```

If you see an error, start Docker Desktop.

### Step 2: Start MinIO Container

**Option A: Using Docker Compose (Recommended)**

```bash
cd D:\Projects\ar-13-go-backend
docker-compose -f docker-compose.minio.yml up -d
```

**Option B: Using Docker Run**

```bash
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  -v ./minio-data:/data \
  minio/minio:latest \
  server /data --console-address ":9001"
```

### Step 3: Verify Container is Running

```bash
docker ps | grep minio
```

You should see the container running.

### Step 4: Check Logs (if needed)

```bash
docker logs minio
```

### Step 5: Access MinIO Console

Open: `http://localhost:9001`

Login with:
- Username: `minioadmin`
- Password: `minioadmin`

## Creating Access Keys

### Method 1: Using MinIO Console (Easiest)

1. Go to `http://localhost:9001`
2. Click **Identity** → **Access Keys**
3. Click **Create Access Key**
4. Name: `ar-13-backend-key`
5. Click **Create**
6. Copy both keys immediately

### Method 2: Using MinIO Client (mc)

**Install MinIO Client:**

**Windows (PowerShell as Administrator):**
```powershell
choco install minio-client
```

**Or download manually:**
- Download: https://dl.min.io/client/mc/release/windows-amd64/mc.exe
- Place in a folder and add to PATH

**Configure and Create Access Key:**

```powershell
# Configure connection
mc alias set localminio http://localhost:9000 minioadmin minioadmin

# Create user
mc admin user add localminio ar-13-backend-user

# Create access key
mc admin user svcacct add localminio ar-13-backend-user --name ar-13-backend-key
```

## Setting Up Policies (Optional)

If you need specific permissions, create a policy:

### Step 1: Create Policy File

Create `readwrite-policy.json`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": [
        "arn:aws:s3:::*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads"
      ],
      "Resource": [
        "arn:aws:s3:::*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:AbortMultipartUpload",
        "s3:ListMultipartUploadParts"
      ],
      "Resource": [
        "arn:aws:s3:::*/*"
      ]
    }
  ]
}
```

### Step 2: Create Policy in MinIO

```powershell
mc admin policy create localminio readwrite-policy readwrite-policy.json
```

### Step 3: Attach Policy to User

```powershell
mc admin policy attach localminio readwrite-policy --user ar-13-backend-user
```

## Configuration

### Environment Variables

Add to your `.env` file:

```env
# MinIO Local Development Configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
MINIO_BUCKET=ar-13-uploads
MINIO_USE_SSL=false
MINIO_INSECURE_SSL=false
```

### Ports

- **9000**: MinIO API (S3 operations) - Used by your backend
- **9001**: MinIO Console (Web UI) - For administration

## Data Storage

MinIO data is stored in:
```
./minio-data/
```

This folder is created automatically when you start the container.

**To reset MinIO:**
```bash
docker-compose -f docker-compose.minio.yml down
rm -rf ./minio-data
docker-compose -f docker-compose.minio.yml up -d
```

## Testing

### Test Upload

```bash
go run scripts/test_minio_upload.go --file demo.txt
```

### Test Download

```bash
go run scripts/test_minio_download.go --list
```

### Using Environment Variables

If your `.env` is configured:

```bash
# Upload
go run scripts/test_minio_upload.go --file demo.txt

# List files
go run scripts/test_minio_download.go --list

# Download file
go run scripts/test_minio_download.go --object "demo.txt"
```

## Troubleshooting

### Container Won't Start

**Check if ports are in use:**
```bash
netstat -ano | findstr :9000
netstat -ano | findstr :9001
```

**Kill process using port (Windows):**
```powershell
# Find process ID
netstat -ano | findstr :9000

# Kill process (replace PID)
taskkill /PID <PID> /F
```

**Or change ports in docker-compose.yml:**
```yaml
ports:
  - "9002:9000"  # Change external port
  - "9003:9001"  # Change external port
```

### Can't Access Console

1. **Check container is running:**
   ```bash
   docker ps | grep minio
   ```

2. **Check logs:**
   ```bash
   docker logs minio
   ```

3. **Try different URL:**
   - `http://127.0.0.1:9001`
   - `http://localhost:9001`

### Access Denied Errors

1. **Verify credentials** in `.env` file
2. **Check bucket exists** in MinIO Console
3. **Verify access keys** are correct
4. **Create policy** if needed (see above)

### Permission Errors

**On Windows:**
- Make sure Docker Desktop has permission to access the drive
- Check folder permissions: Right-click `minio-data` → Properties → Security

**On Linux/Mac:**
```bash
chmod -R 755 ./minio-data
```

## Stopping MinIO

```bash
docker-compose -f docker-compose.minio.yml down
```

**To stop and remove data:**
```bash
docker-compose -f docker-compose.minio.yml down -v
rm -rf ./minio-data
```

## Starting MinIO Again

```bash
docker-compose -f docker-compose.minio.yml up -d
```

## Updating MinIO

```bash
docker-compose -f docker-compose.minio.yml pull
docker-compose -f docker-compose.minio.yml up -d
```

## File Locations

- **Data:** `./minio-data/` (relative to project root)
- **Config:** Stored inside container at `/root/.minio`
- **Logs:** `docker logs minio`

## Comparison: Local vs NAS Setup

| Feature | Local | NAS |
|---------|-------|-----|
| Endpoint | `localhost:9000` | `192.168.0.118:9000` |
| Data Location | `./minio-data/` | `/share/studio work/` |
| Access | Local only | Network accessible |
| Use Case | Development | Production |
| SSL | Not needed | Recommended |

## Best Practices

1. **Don't commit `minio-data/`** - Add to `.gitignore`
2. **Use different credentials** for production
3. **Backup data** before major changes
4. **Use environment variables** for configuration
5. **Test locally** before deploying to NAS

## Next Steps

1. ✅ MinIO running locally
2. ✅ Bucket created
3. ✅ Access keys configured
4. ✅ Backend `.env` updated
5. ✅ Test upload/download
6. ✅ Ready for development!

## Additional Resources

- [MinIO Documentation](https://min.io/docs/)
- [MinIO Docker Hub](https://hub.docker.com/r/minio/minio)
- See `MINIO_SETUP_GUIDE.md` for NAS setup
- See `docs/nas/` for production deployment guides

