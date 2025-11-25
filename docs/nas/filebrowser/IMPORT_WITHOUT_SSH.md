# Import File Browser Service Without SSH

## Problem
- Container Station Web UI rejects `.tar` files with "Invalid File Format"
- SSH is not available

## Solution Options

### Option 1: Pull from Docker Hub (✅ RECOMMENDED - Easiest Method)

**This is the simplest and most reliable method!** The image is pre-built and available on Docker Hub.

**Quick Steps:**
1. Open Container Station → **Images** tab
2. Click **Pull** button
3. Enter: `jeetshah786/filebrowser-service:latest`
4. Click **Pull** and wait for download
5. Deploy using `docker-compose.filebrowser-service.nas-dockerhub.yml`

**Full Guide:** See [DOCKER_HUB_DEPLOYMENT.md](./DOCKER_HUB_DEPLOYMENT.md) for detailed instructions.

**Advantages:**
- ✅ No SSH required
- ✅ No build process needed
- ✅ No import file format issues
- ✅ Works with Container Station GUI
- ✅ Fast and reliable

### Option 2: Build Image Directly on NAS

Container Station can build Docker images from source code. This avoids the import issue entirely.

#### Step 1: Prepare Files for Upload

You need to upload the entire `services/filebrowser/` folder to your NAS:

**Files to upload:**
- `services/filebrowser/main.go`
- `services/filebrowser/go.mod`
- `services/filebrowser/go.sum`
- `services/filebrowser/Dockerfile`
- `services/filebrowser/.dockerignore`
- `services/filebrowser/handlers/` (entire folder)
- `services/filebrowser/models/` (entire folder)
- `services/filebrowser/middleware/` (entire folder)
- `services/filebrowser/utils/` (entire folder)
- `docker-compose.filebrowser-service.nas.yml`

#### Step 2: Upload to NAS

1. Open **File Station** on your NAS
2. Create a folder: `/share/Public/filebrowser-build/`
3. Upload the entire `services/filebrowser/` folder structure to `/share/Public/filebrowser-build/`
4. Upload `docker-compose.filebrowser-service.nas.yml` to `/share/Public/filebrowser-build/`

#### Step 3: Build Using Container Station

1. Open **Container Station**
2. Go to **Images** tab
3. Click **Build** button (or **Create** → **Build Image**)
4. Configure build:
   - **Build Context:** `/share/Public/filebrowser-build/services/filebrowser/`
   - **Dockerfile:** `Dockerfile` (should auto-detect)
   - **Image Name:** `ar-13-go-backend-filebrowser-service:latest`
   - **Tag:** `latest`
5. Click **Build**
6. Wait for build to complete (may take 5-10 minutes)

#### Step 4: Deploy Container

**Using Container Station GUI:**

1. Go to **Containers** tab
2. Click **Create** → **Create Application**
3. Choose **Compose** tab
4. Upload or paste contents of `docker-compose.filebrowser-service.nas.yml`
5. **Important:** Remove or comment out the `build:` section since image is already built:
   ```yaml
   services:
     filebrowser-service:
       # build:
       #   context: ./services/filebrowser
       #   dockerfile: Dockerfile
       image: ar-13-go-backend-filebrowser-service:latest
       container_name: filebrowser-service
       # ... rest of config
   ```
6. Click **Create**

### Option 2: Use Container Station Terminal (Web-Based)

Some QNAP models have a web-based terminal in Container Station that doesn't require SSH.

1. Open **Container Station**
2. Look for **Terminal** or **Console** option (usually in top menu)
3. If available, open it
4. Run these commands:
   ```bash
   cd /share/Public
   docker load -i filebrowser-service.tar
   docker-compose -f docker-compose.filebrowser-service.nas.yml up -d
   ```

### Option 3: Try Different File Format

Sometimes renaming or compressing helps (though this may not work):

1. Rename `filebrowser-service.tar` to `filebrowser-service.img`
2. Try importing in Container Station
3. If that doesn't work, try `filebrowser-service.docker`

### Option 4: Use Container Station API (Advanced)

If your QNAP supports Container Station API:

1. Check if Container Station has API documentation
2. Use API to load image programmatically
3. This requires API access and may need authentication

### Option 5: Use QNAP App Center Alternative

Some QNAP models allow installing Docker images via App Center or other package managers.

## Recommended: Option 1 (Build on NAS)

**Why this is best:**
- ✅ No SSH required
- ✅ No import issues
- ✅ Works with Container Station GUI
- ✅ Builds directly on NAS
- ✅ Most reliable method

**Steps Summary:**
1. Upload `services/filebrowser/` folder to NAS
2. Build image in Container Station using Dockerfile
3. Deploy using docker-compose (remove build section)

## Modified docker-compose for Pre-built Image

If you build the image on NAS, use this modified docker-compose:

```yaml
services:
  filebrowser-service:
    image: ar-13-go-backend-filebrowser-service:latest  # Use pre-built image
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

## Troubleshooting

**Build fails:**
- Check Container Station logs
- Verify all files uploaded correctly
- Ensure Dockerfile is in correct location

**Container won't start:**
- Check if image exists: Look in Images tab
- Verify docker-compose file syntax
- Check logs in Container Station

**Permission errors:**
- Ensure `/share/studio work/` folder exists
- Check folder permissions in File Station

