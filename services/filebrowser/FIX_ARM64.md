# Fix ARM64 Exec Format Error

If you're getting "exec format error" on your NAS, follow these steps:

## Step 1: Rebuild and Push Multi-Arch Image

On your development machine (Windows):

```powershell
cd services\filebrowser
.\build-multi-arch.ps1
```

This will rebuild with the fixed Dockerfile that properly handles ARM64.

## Step 2: On Your NAS - Pull ARM64 Version Explicitly

```bash
# Remove old image
docker rmi jeetshah786/ar-13-go-backend-filebrowser-service:latest

# Pull ARM64 version explicitly
docker pull --platform linux/arm64 jeetshah786/ar-13-go-backend-filebrowser-service:latest

# Verify architecture
docker inspect jeetshah786/ar-13-go-backend-filebrowser-service:latest | grep -i arch
```

## Step 3: Run with Platform Specification

```bash
docker run -d \
  --name filebrowser-service \
  --restart unless-stopped \
  --platform linux/arm64 \
  -v "/share/studio work:/data" \
  -e DATA_ROOT=/data \
  -e PORT=8082 \
  -e TZ=Asia/Kolkata \
  --network filebrowser-network \
  --user "1000:1000" \
  jeetshah786/ar-13-go-backend-filebrowser-service:latest
```

## Alternative: Build Directly on NAS

If pulling doesn't work, build directly on your NAS:

```bash
# Upload the services/filebrowser directory to your NAS
# Then build:

cd /path/to/services/filebrowser
docker build -t jeetshah786/ar-13-go-backend-filebrowser-service:latest .
```

This will automatically build for ARM64 since that's your NAS architecture.

