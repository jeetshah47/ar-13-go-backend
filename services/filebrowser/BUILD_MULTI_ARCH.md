# Building Multi-Arch Docker Image

This guide explains how to build and push a multi-architecture Docker image that supports both AMD64 and ARM64 platforms.

## Prerequisites

1. **Docker Desktop** (Windows/Mac) or **Docker Engine** (Linux) with Buildx support
2. **Docker Hub account** with push permissions to `jeetshah786/ar-13-go-backend-filebrowser-service`
3. **Logged in to Docker Hub**:
   ```bash
   docker login
   ```

## Quick Start

### Windows (PowerShell)

```powershell
cd services\filebrowser
.\build-multi-arch.ps1
```

### Linux/macOS (Bash)

```bash
cd services/filebrowser
chmod +x build-multi-arch.sh
./build-multi-arch.sh
```

## Manual Build

### Step 1: Set up Buildx Builder

```bash
# Create a new builder instance
docker buildx create --name multiarch-builder --use

# Bootstrap the builder
docker buildx inspect --bootstrap
```

### Step 2: Build and Push

```bash
cd services/filebrowser

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag jeetshah786/ar-13-go-backend-filebrowser-service:latest \
  --push \
  .
```

### Step 3: Verify

```bash
# Check the image manifest
docker buildx imagetools inspect jeetshah786/ar-13-go-backend-filebrowser-service:latest
```

You should see both `linux/amd64` and `linux/arm64` in the output.

## Building Specific Version

```bash
# Windows
.\build-multi-arch.ps1 -Version v1.0.0

# Linux/macOS
./build-multi-arch.sh v1.0.0
```

## Troubleshooting

### Buildx Not Available

**Windows/Mac:**
- Update Docker Desktop to the latest version
- Buildx is included by default in newer versions

**Linux:**
```bash
# Install buildx plugin
mkdir -p ~/.docker/cli-plugins
curl -L https://github.com/docker/buildx/releases/latest/download/buildx-v0.12.0.linux-amd64 -o ~/.docker/cli-plugins/docker-buildx
chmod +x ~/.docker/cli-plugins/docker-buildx
```

### Authentication Issues

```bash
# Login to Docker Hub
docker login

# Verify
docker info | grep Username
```

### Build Fails for ARM64

The Dockerfile uses `ARG TARGETARCH` which should work automatically. If it doesn't:

1. Check the Dockerfile has:
   ```dockerfile
   ARG TARGETOS=linux
   ARG TARGETARCH=amd64
   RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build ...
   ```

2. Ensure Go supports ARM64:
   ```bash
   go version
   # Should show go1.23 or later
   ```

## Testing the Image

### Test AMD64 (on AMD64 machine)

```bash
docker run --rm --platform linux/amd64 \
  jeetshah786/ar-13-go-backend-filebrowser-service:latest \
  --version
```

### Test ARM64 (on ARM64 machine or emulator)

```bash
docker run --rm --platform linux/arm64 \
  jeetshah786/ar-13-go-backend-filebrowser-service:latest \
  --version
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Push Multi-Arch

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      
      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: ./services/filebrowser
          platforms: linux/amd64,linux/arm64
          push: true
          tags: jeetshah786/ar-13-go-backend-filebrowser-service:latest
```

## Notes

- **Build Time**: Multi-arch builds take longer (2-3x) because they build for multiple platforms
- **Image Size**: The manifest list adds minimal overhead
- **Compatibility**: The image will automatically pull the correct architecture for the platform
- **NAS Support**: Once pushed, your ARM64 NAS can pull the image directly without building locally

