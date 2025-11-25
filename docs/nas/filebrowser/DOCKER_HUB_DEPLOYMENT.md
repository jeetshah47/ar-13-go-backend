# Deploy File Browser Service from Docker Hub

## Overview

This is the **easiest method** to deploy the file browser service on your NAS. The image is pre-built and available on Docker Hub, so you just need to pull it - no SSH, no build, no import issues!

## Prerequisites

- QNAP NAS with Container Station installed
- Access to Container Station web interface
- Internet connection on NAS

## Step-by-Step Guide

### Step 1: Pull Image from Docker Hub

1. Open **Container Station** on your QNAP NAS
2. Go to **Images** tab
3. Click **Pull** button
4. Enter image name: `jeetshah786/filebrowser-service:latest`
5. Click **Pull** or **OK**
6. Wait for the image to download (may take 2-5 minutes depending on internet speed)
7. Verify the image appears in the Images list

### Step 2: Deploy Container

**Option A: Using Container Station GUI (Recommended)**

1. In Container Station, go to **Containers** tab
2. Click **Create** → **Create Container**
3. Fill in the configuration:
   - **Image:** Select `jeetshah786/filebrowser-service:latest` from dropdown
   - **Name:** `filebrowser-service`
   - **Port:** 
     - Host: `8082`
     - Container: `8082`
   - **Volume:**
     - Host: `/share/studio work/`
     - Container: `/data`
   - **Environment Variables:**
     - `DATA_ROOT=/data`
     - `PORT=8082`
     - `TZ=Asia/Kolkata`
   - **User:** `1000:1000`
   - **Restart Policy:** `Unless stopped`
4. Click **Create**
5. Container should start automatically

**Option B: Using Docker Compose**

1. Upload `docker-compose.filebrowser-service.nas-dockerhub.yml` to your NAS (e.g., `/share/Public/`)
2. In Container Station, go to **Applications** tab
3. Click **Create** → **Create Application**
4. Choose **Compose** tab
5. Upload or paste contents of `docker-compose.filebrowser-service.nas-dockerhub.yml`
6. Click **Create**
7. Container will be created and started automatically

### Step 3: Verify Installation

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
   - Should return:
     ```json
     {
       "status": "healthy",
       "service": "file-browser",
       "timestamp": "2024-11-22T..."
     }
     ```

4. **Test browse endpoint:**
   - Open: `http://<NAS-IP>:8082/api/browse?path=/`
   - Should return list of files/folders from `/share/studio work/`

## Troubleshooting

### Image Won't Pull

**Error: "Image not found"**
- Verify image name: `jeetshah786/filebrowser-service:latest`
- Check internet connection on NAS
- Try pulling again after a few minutes (Docker Hub may need time to process)

**Error: "Network timeout"**
- Check NAS internet connection
- Verify DNS settings on NAS
- Try again later

### Container Won't Start

**Error: "Port already in use"**
- Change port in container settings from `8082:8082` to `8083:8082`
- Or stop the service using port 8082

**Error: "Permission denied" accessing files**
- Check `/share/studio work/` folder permissions in File Station
- Ensure folder is readable by UID 1000
- Fix permissions: `sudo chmod -R 755 "/share/studio work/"`

**Error: "Volume mount failed"**
- Verify `/share/studio work/` folder exists
- Check folder path is correct (case-sensitive)
- Ensure Container Station has access to the folder

### Container Starts But API Doesn't Work

1. Check container logs in Container Station
2. Verify port 8082 is accessible
3. Check firewall settings on NAS
4. Test from NAS itself: `curl http://localhost:8082/health` (if terminal available)

## Image Information

- **Docker Hub:** https://hub.docker.com/r/jeetshah786/filebrowser-service
- **Image Name:** `jeetshah786/filebrowser-service:latest`
- **Size:** ~25-30 MB (compressed)
- **Platform:** linux/amd64
- **Public:** Yes (anyone can pull)

## API Endpoints

Once deployed, the service provides these endpoints:

- **Health Check:** `GET http://<NAS-IP>:8082/health`
- **Browse Directory:** `GET http://<NAS-IP>:8082/api/browse?path=/folder`
- **File Info:** `GET http://<NAS-IP>:8082/api/file-info?path=/file.pdf`
- **Download File:** `GET http://<NAS-IP>:8082/api/download?path=/file.pdf`

## Updating the Image

When a new version is available:

1. In Container Station, go to **Images** tab
2. Find `jeetshah786/filebrowser-service:latest`
3. Click **Pull** to update to latest version
4. Restart the container to use the new image

## Advantages of This Method

- ✅ No SSH required
- ✅ No build process needed
- ✅ No import file format issues
- ✅ Easy to update (just pull new version)
- ✅ Works with Container Station GUI
- ✅ Fast deployment (just pull and run)

## Next Steps

After successful deployment:

1. Test all API endpoints
2. Integrate with Go backend (Phase 2)
3. Add authentication/authorization if needed
4. Configure reverse proxy if using HTTPS

