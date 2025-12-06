# Build File Browser Service Directly on NAS

## Overview

Since Container Station's import feature rejects `.tar` files and SSH is not available, we'll build the Docker image directly on the NAS using Container Station's build feature.

## Step-by-Step Guide

### Step 1: Prepare Files for Upload

You need to upload the source code to your NAS. Here's what to upload:

**Create a folder structure on NAS:**
```
/share/Public/filebrowser-build/
├── services/
│   └── filebrowser/
│       ├── main.go
│       ├── go.mod
│       ├── go.sum
│       ├── Dockerfile
│       ├── .dockerignore
│       ├── handlers/
│       │   ├── browse.go
│       │   ├── file_info.go
│       │   └── download.go
│       ├── models/
│       │   └── file.go
│       ├── middleware/
│       │   └── cors.go
│       └── utils/
│           └── mime.go
└── docker-compose.filebrowser-service.nas-built.yml
```

### Step 2: Upload Files to NAS

1. Open **File Station** on your QNAP NAS
2. Navigate to `/share/Public/`
3. Create folder `filebrowser-build`
4. Upload the entire `ar-13-nas-filebrowser/` folder structure
5. Upload `docker-compose.filebrowser-service.nas-built.yml`

**Quick Upload Method:**
- Use File Station's upload feature
- Select the entire `ar-13-nas-filebrowser/` folder from your Windows machine
- Upload to `/share/Public/filebrowser-build/`

### Step 3: Build Image in Container Station

1. Open **Container Station** on your NAS
2. Go to **Images** tab
3. Click **Build** button (or look for **Create** → **Build Image**)
4. Fill in the build configuration:
   
   **Build Settings:**
   - **Build Context:** `/share/Public/filebrowser-build/ar-13-nas-filebrowser/`
   - **Dockerfile:** `Dockerfile` (should auto-detect)
   - **Image Name:** `ar-13-go-backend-filebrowser-service`
   - **Tag:** `latest`
   
   **Advanced Settings (if available):**
   - Platform: `linux/amd64`
   - Build arguments: Leave empty (not needed)

5. Click **Build** or **Start Build**
6. Wait for build to complete (5-10 minutes)
   - You'll see build progress in the logs
   - Watch for any errors

7. Once complete, verify the image appears in the **Images** list

### Step 4: Deploy Container

1. In Container Station, go to **Containers** tab
2. Click **Create** → **Create Application**
3. Choose **Compose** tab
4. **Option A:** Upload `docker-compose.filebrowser-service.nas-built.yml`
   **Option B:** Paste the contents manually:

```yaml
services:
  filebrowser-service:
    image: ar-13-go-backend-filebrowser-service:latest
    container_name: filebrowser-service
    restart: unless-stopped
    user: "1000:1000"
    ports:
      - "8082:8082"
    volumes:
      - "/share/studio work/:/data"
    environment:
      - DATA_ROOT=/data
      - PORT=8082
      - TZ=Asia/Kolkata
    networks:
      - filebrowser-network
    stdin_open: true
    tty: true
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8082/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

networks:
  filebrowser-network:
    driver: bridge
```

5. Click **Create**
6. Container should start automatically

### Step 5: Verify Installation

1. **Check container status:**
   - Go to **Containers** tab
   - Look for `filebrowser-service`
   - Status should be "Running"

2. **Check logs:**
   - Click on the container
   - View logs to see startup messages
   - Should see: "File Browser Service is running on port 8082"

3. **Test health endpoint:**
   - Open browser: `http://<NAS-IP>:8082/health`
   - Should return JSON with status "healthy"

4. **Test browse endpoint:**
   - Open: `http://<NAS-IP>:8082/api/browse?path=/`
   - Should return list of files/folders

## Troubleshooting

### Build Fails

**Error: "Cannot find Dockerfile"**
- Verify Dockerfile is in `/share/Public/filebrowser-build/ar-13-nas-filebrowser/`
- Check file permissions in File Station

**Error: "go.mod requires go >= 1.24.0"**
- The Dockerfile should handle this automatically
- If issue persists, check Container Station's Go version support

**Error: "Permission denied"**
- Check folder permissions in File Station
- Ensure Container Station has read access to build folder

### Container Won't Start

**Error: "Image not found"**
- Verify image was built successfully
- Check Images tab for `ar-13-go-backend-filebrowser-service:latest`

**Error: "Port already in use"**
- Change port in docker-compose from `8082:8082` to `8083:8082`
- Or stop the service using port 8082

**Error: "Permission denied" accessing files**
- Check `/share/studio work/` folder permissions
- Ensure folder is readable by UID 1000

### Container Starts But API Doesn't Work

1. Check container logs in Container Station
2. Verify port 8082 is accessible
3. Test from NAS itself: `curl http://localhost:8082/health`
4. Check firewall settings on NAS

## Alternative: Manual Container Creation

If docker-compose doesn't work, create container manually:

1. Go to **Containers** → **Create**
2. Choose **Create Container**
3. Fill in:
   - **Image:** `ar-13-go-backend-filebrowser-service:latest`
   - **Name:** `filebrowser-service`
   - **Port:** `8082:8082`
   - **Volume:** `/share/studio work/` → `/data`
   - **Environment Variables:**
     - `DATA_ROOT=/data`
     - `PORT=8082`
     - `TZ=Asia/Kolkata`
   - **User:** `1000:1000`
   - **Restart Policy:** `Unless stopped`
4. Click **Create**

## Next Steps

After successful deployment:
1. Test all API endpoints
2. Integrate with Go backend (Phase 2)
3. Add authentication if needed

