# Import File Browser Service to Container Station

## ⚠️ Important: Container Station Web UI Import Issue

**Container Station's web UI may reject `.tar` files with "Invalid File Format" error.**  
**Use Method 2 (SSH/Terminal) instead - it's more reliable and recommended.**

## Files Required

1. **`filebrowser-service.tar`** - Docker image file (saved from Windows)
2. **`docker-compose.filebrowser-service.nas.yml`** - Docker Compose configuration

## Method 1: Import via Container Station Web UI (May Not Work)

⚠️ **Note:** Container Station sometimes rejects `.tar` files. If you get "Invalid File Format" error, use Method 2 instead.

### Step 1: Transfer Files to NAS

1. Copy `filebrowser-service.tar` to your NAS (e.g., to `/share/Public/` or any accessible folder)
2. Copy `docker-compose.filebrowser-service.nas.yml` to your NAS

### Step 2: Import Image in Container Station

1. Open **Container Station** on your QNAP NAS
2. Go to **Images** tab
3. Click **Import** button
4. Browse and select `filebrowser-service.tar`
5. Wait for import to complete (may take a few minutes)
6. Verify the image appears as `ar-13-go-backend-filebrowser-service:latest`

### Step 3: Create Container

**Option A: Using Container Station GUI**

1. In Container Station, go to **Containers** tab
2. Click **Create** → **Create Application**
3. Choose **Compose** tab
4. Upload or paste the contents of `docker-compose.filebrowser-service.nas.yml`
5. Click **Create**
6. Container will start automatically

**Option B: Using SSH/Terminal**

1. SSH into your NAS
2. Navigate to where you saved the docker-compose file
3. Run:
   ```bash
   docker-compose -f docker-compose.filebrowser-service.nas.yml up -d
   ```

## Method 2: Import via SSH/Terminal (✅ RECOMMENDED)

This method is more reliable and works even when Container Station's web UI rejects the file.

### Step 1: Transfer Files to NAS

**Option A: Using File Station (Easiest)**

1. Open **File Station** on your QNAP NAS
2. Navigate to `/share/Public/` (or any folder you prefer)
3. Upload `filebrowser-service.tar` to this folder
4. Upload `docker-compose.filebrowser-service.nas.yml` to this folder

**Option B: Using SCP from Windows**

```powershell
# From PowerShell on Windows
scp filebrowser-service.tar admin@<NAS-IP>:/share/Public/
scp docker-compose.filebrowser-service.nas.yml admin@<NAS-IP>:/share/Public/
```

**Option C: Using Network Drive**

1. Map NAS as network drive in Windows
2. Copy files directly to `/share/Public/` folder

### Step 2: Enable SSH on NAS (If Not Already Enabled)

1. Go to **Control Panel** → **Network Services** → **Telnet/SSH**
2. Enable SSH service
3. Note the SSH port (usually 22)

### Step 3: SSH into NAS and Load Image

**Using PuTTY (Windows):**
1. Open PuTTY
2. Enter NAS IP address and SSH port (usually 22)
3. Login with admin credentials

**Using PowerShell SSH:**
```powershell
ssh admin@<NAS-IP>
```

**Once connected, run these commands:**

```bash
# Navigate to where you uploaded the files
cd /share/Public

# Load the Docker image
docker load -i filebrowser-service.tar

# Verify the image loaded successfully
docker images | grep filebrowser
```

You should see output like:
```
ar-13-go-backend-filebrowser-service   latest   8b38c42c800d   26.7MB
```

### Step 3: Start Container

```bash
docker-compose -f docker-compose.filebrowser-service.nas.yml up -d
```

Or if docker-compose is not available, use docker run:

```bash
docker run -d \
  --name filebrowser-service \
  --restart unless-stopped \
  -p 8082:8082 \
  -v "/share/studio work/:/data" \
  -e DATA_ROOT=/data \
  -e PORT=8082 \
  -e TZ=Asia/Kolkata \
  --user "1000:1000" \
  ar-13-go-backend-filebrowser-service:latest
```

## Verify Installation

1. **Check container status:**
   ```bash
   docker ps | grep filebrowser
   ```

2. **Check logs:**
   ```bash
   docker logs filebrowser-service
   ```

3. **Test health endpoint:**
   Open browser: `http://<NAS-IP>:8082/health`
   
   Should return:
   ```json
   {
     "status": "healthy",
     "service": "file-browser",
     "timestamp": "2024-11-22T..."
   }
   ```

4. **Test browse endpoint:**
   `http://<NAS-IP>:8082/api/browse?path=/`

## Troubleshooting

### Permission Issues

If you get permission errors accessing files:

1. Check folder permissions on `/share/studio work/`:
   ```bash
   ls -la "/share/studio work/"
   ```

2. Ensure the folder is readable by UID 1000:
   ```bash
   sudo chmod -R 755 "/share/studio work/"
   ```

3. If needed, change ownership (be careful with this):
   ```bash
   sudo chown -R 1000:1000 "/share/studio work/"
   ```

### Port Already in Use

If port 8082 is already in use:

1. Edit `docker-compose.filebrowser-service.nas.yml`
2. Change port mapping from `8082:8082` to `8083:8082` (or any available port)
3. Restart the container

### Container Won't Start

1. Check logs:
   ```bash
   docker logs filebrowser-service
   ```

2. Verify the image exists:
   ```bash
   docker images ar-13-go-backend-filebrowser-service
   ```

3. Check if DATA_ROOT directory exists:
   ```bash
   ls -la /share/studio\ work/
   ```

## Next Steps

After successful deployment:

1. Test all API endpoints
2. Integrate with your Go backend (Phase 2)
3. Add authentication/authorization if needed
4. Configure reverse proxy if using HTTPS

## API Endpoints Reference

- **Health:** `GET http://<NAS-IP>:8082/health`
- **Browse:** `GET http://<NAS-IP>:8082/api/browse?path=/folder`
- **File Info:** `GET http://<NAS-IP>:8082/api/file-info?path=/file.pdf`
- **Download:** `GET http://<NAS-IP>:8082/api/download?path=/file.pdf`

