# Fix Docker Image Pull Error

If you get "unable to get image" or "500 Internal Server Error" when starting MinIO, try these solutions:

## Solution 1: Restart Docker Desktop

1. **Right-click Docker Desktop icon** in system tray
2. Click **Restart**
3. Wait for Docker to fully start
4. Try again:
   ```powershell
   docker-compose -f docker-compose.minio.yml up -d
   ```

## Solution 2: Pull Image Manually

```powershell
docker pull minio/minio:latest
```

Then try starting again:
```powershell
docker-compose -f docker-compose.minio.yml up -d
```

## Solution 3: Use Specific Version Instead of Latest

Edit `docker-compose.minio.yml` and change:

```yaml
image: minio/minio:RELEASE.2024-11-20T23-30-35Z
```

Or use a stable version:
```yaml
image: minio/minio:RELEASE.2024-10-20T18-20-00Z
```

Then:
```powershell
docker-compose -f docker-compose.minio.yml up -d
```

## Solution 4: Check Docker Desktop Status

1. Open **Docker Desktop**
2. Check if it shows "Docker Desktop is running"
3. If not, click **Start** and wait
4. Check **Settings** → **General** → Make sure "Use WSL 2 based engine" is enabled (if using WSL)

## Solution 5: Clear Docker Cache

```powershell
# Stop all containers
docker stop $(docker ps -aq)

# Remove unused images
docker image prune -a

# Try pulling again
docker pull minio/minio:latest
```

## Solution 6: Check Network/Proxy

If behind a corporate proxy:

1. **Docker Desktop** → **Settings** → **Resources** → **Proxies**
2. Configure proxy settings
3. Restart Docker Desktop

## Solution 7: Use Alternative Image Source

If Docker Hub is blocked, try:

```powershell
# Pull from alternative registry
docker pull quay.io/minio/minio:latest
```

Then update docker-compose.yml:
```yaml
image: quay.io/minio/minio:latest
```

## Solution 8: Check Docker API Version

The error mentions API version. Try:

```powershell
# Check Docker version
docker version

# Check Docker info
docker info
```

If there are errors, restart Docker Desktop.

## Quick Fix Summary

**Most common fix:**
1. Restart Docker Desktop
2. Wait 30 seconds for it to fully start
3. Run: `docker pull minio/minio:latest`
4. Run: `docker-compose -f docker-compose.minio.yml up -d`

## Verify Docker is Working

Test with a simple container:

```powershell
docker run hello-world
```

If this works, Docker is fine. If not, Docker Desktop needs to be fixed.

## Still Having Issues?

1. **Check Docker Desktop logs:**
   - Docker Desktop → Troubleshoot → View logs

2. **Reinstall Docker Desktop** (last resort)

3. **Check Windows updates** - Sometimes Docker needs Windows updates

4. **Check antivirus** - May be blocking Docker

