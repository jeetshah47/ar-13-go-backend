# Quick Import Guide - Container Station

## ⚠️ If Web UI Shows "Invalid File Format" Error

Container Station's web UI sometimes rejects `.tar` files. **Use SSH method instead** (it's actually easier and more reliable).

## Quick Steps (SSH Method - Recommended)

### 1. Transfer Files to NAS

**Easiest way:** Use File Station
- Open File Station on your NAS
- Upload `filebrowser-service.tar` to `/share/Public/`
- Upload `docker-compose.filebrowser-service.nas.yml` to `/share/Public/`

### 2. Enable SSH (If Not Already)

- Control Panel → Network Services → Telnet/SSH → Enable SSH

### 3. SSH into NAS

**Windows PowerShell:**
```powershell
ssh admin@<YOUR-NAS-IP>
```

**Or use PuTTY:**
- Host: `<YOUR-NAS-IP>`
- Port: `22` (or your SSH port)
- Login with admin credentials

### 4. Load Image and Start Container

```bash
# Go to Public folder
cd /share/Public

# Load the Docker image
docker load -i filebrowser-service.tar

# Verify it loaded
docker images | grep filebrowser

# Start the container
docker-compose -f docker-compose.filebrowser-service.nas.yml up -d
```

### 5. Verify It's Running

```bash
# Check container status
docker ps | grep filebrowser

# Check logs
docker logs filebrowser-service
```

### 6. Test the Service

Open in browser: `http://<YOUR-NAS-IP>:8082/health`

Should return:
```json
{
  "status": "healthy",
  "service": "file-browser",
  "timestamp": "..."
}
```

## Alternative: Direct Docker Run (If docker-compose not available)

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

## Troubleshooting

**Image not found after load:**
```bash
# Check if image exists
docker images

# If not found, try loading again
docker load -i /share/Public/filebrowser-service.tar
```

**Permission errors:**
```bash
# Check folder permissions
ls -la "/share/studio work/"

# Fix permissions if needed
sudo chmod -R 755 "/share/studio work/"
```

**Port already in use:**
- Edit docker-compose file and change port from `8082:8082` to `8083:8082`
- Or stop the service using port 8082 first

## That's It!

Your file browser service should now be running on port 8082.

