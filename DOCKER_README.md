# Docker Setup Guide

This guide explains how to build and run the AR-13 backend services using Docker.

## Architecture

The application consists of two main services:

1. **Server** - Main API server (Gin framework)
2. **Worker** - Background job service for time tracking and other async tasks

## Build Process

**Important**: This setup uses pre-built binaries. You must build the binaries on your server before building Docker images.

### Step 1: Build Binaries

Build the binaries for Linux (required for Docker containers):

**On Linux/macOS:**
```bash
chmod +x build-binaries.sh
./build-binaries.sh
```

**On Windows:**
```powershell
.\build-binaries.ps1
```

**Manual build:**
```bash
# Create bin directory
mkdir -p bin

# Build server binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o bin/server ./cmd/server/main.go

# Build worker binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o bin/worker ./cmd/worker/main.go

# Make executable
chmod +x bin/server bin/worker
```

This will create:
- `bin/server` - Main API server binary
- `bin/worker` - Background worker binary

### Step 2: Build Docker Images

After binaries are built, create Docker images:

```bash
# Build both images
docker-compose build

# Or build individually
docker build -f docker/Dockerfile.server -t ar-13-server .
docker build -f docker/Dockerfile.worker -t ar-13-worker .
```

## Prerequisites

- Go installed on your build machine
- Docker and Docker Compose installed
- MongoDB instance (can be external or in Docker)
- (Optional) Redis instance for caching

## Quick Start

### 1. Build Binaries

```bash
./build-binaries.sh
```

### 2. Create Environment File

Create a `.env` file with your configuration:

```bash
# Required variables
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=ar13_backend

# Optional variables
PORT=3000
NODE_ENV=production
REDIS_ADDR=redis:6379
# ... (see .env.example for all options)
```

### 3. Build and Start Services

```bash
# Build Docker images (requires binaries in bin/ directory)
docker-compose -f docker/docker-compose.yml build

# Start all services
docker-compose -f docker/docker-compose.yml up -d

# View logs
docker-compose -f docker/docker-compose.yml logs -f

# View logs for specific service
docker-compose -f docker/docker-compose.yml logs -f server
docker-compose -f docker/docker-compose.yml logs -f worker
```

### 4. Stop Services

```bash
docker-compose -f docker/docker-compose.yml down

# Remove volumes (cleans Redis data)
docker-compose -f docker/docker-compose.yml down -v
```

## Building Individual Services

### Build Server Only

```bash
# 1. Build binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o bin/server ./cmd/server/main.go

# 2. Build Docker image
docker build -f docker/Dockerfile.server -t ar-13-server .
```

### Build Worker Only

```bash
# 1. Build binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o bin/worker ./cmd/worker/main.go

# 2. Build Docker image
docker build -f docker/Dockerfile.worker -t ar-13-worker .
```

## Running Individual Containers

### Run Server Container

```bash
docker run -d \
  --name ar-13-server \
  -p 3000:3000 \
  --env-file .env \
  -v $(pwd)/upload:/app/upload \
  ar-13-server
```

### Run Worker Container

```bash
docker run -d \
  --name ar-13-worker \
  --env-file .env \
  ar-13-worker
```

## Memory Limits

The docker/docker-compose.yml includes memory limits:

- **Server**: 512 MB limit, 200 MB reserved
- **Worker**: 256 MB limit, 100 MB reserved
- **Redis**: 200 MB limit, 50 MB reserved

You can adjust these in `docker/docker-compose.yml` based on your needs.

## Health Checks

Both services include health checks:

- **Server**: HTTP health check on `/api/health` endpoint
- **Worker**: Process check to ensure worker is running

View health status:

```bash
docker-compose -f docker/docker-compose.yml ps
```

## Volumes

- `./upload` - Mounted to `/app/upload` in server container for file uploads
- `redis-data` - Persistent storage for Redis data

## Networking

All services are connected via the `ar-13-network` bridge network.

- Server accessible at: `http://localhost:3000` (or your configured PORT)
- Redis accessible at: `redis:6379` (internal network)

## Environment Variables

See `.env.example` for all available environment variables.

### Required Variables

- `JWT_SECRET` - Secret key for JWT token signing
- `MONGODB_URI` - MongoDB connection string
- `MONGODB_DATABASE` - Database name

### Optional Variables

- `REDIS_ADDR` - Redis address (defaults to `redis:6379` if using included Redis)
- `PORT` - Server port (defaults to 3000)
- Email, OAuth, and storage configurations

## Troubleshooting

### Check Container Logs

```bash
# All services
docker-compose -f docker/docker-compose.yml logs

# Specific service
docker-compose -f docker/docker-compose.yml logs server
docker-compose -f docker/docker-compose.yml logs worker

# Follow logs
docker-compose -f docker/docker-compose.yml logs -f server
```

### Check Container Status

```bash
docker-compose -f docker/docker-compose.yml ps
```

### Restart a Service

```bash
docker-compose -f docker/docker-compose.yml restart server
docker-compose -f docker/docker-compose.yml restart worker
```

### Rebuild After Code Changes

**Important**: After code changes, you must rebuild binaries first, then rebuild Docker images:

```bash
# 1. Rebuild binaries
./build-binaries.sh

# 2. Rebuild and restart
docker-compose -f docker/docker-compose.yml up -d --build

# Or rebuild specific service
docker-compose -f docker/docker-compose.yml build server
docker-compose -f docker/docker-compose.yml up -d server
```

### Access Container Shell

```bash
# Server container
docker exec -it ar-13-server sh

# Worker container
docker exec -it ar-13-worker sh
```

### Check Resource Usage

```bash
docker stats
```

### Binary Not Found Error

If you get an error about missing binaries:

```bash
# Make sure binaries exist
ls -la bin/

# If missing, build them
./build-binaries.sh
```

## Production Considerations

1. **Security**:
   - Use strong `JWT_SECRET`
   - Set secure `REDIS_PASSWORD`
   - Use environment variables, not `.env` file in production
   - Consider using Docker secrets for sensitive data

2. **MongoDB**:
   - Use external MongoDB instance or MongoDB Atlas
   - Update `MONGODB_URI` accordingly
   - Ensure proper network connectivity

3. **Redis**:
   - Can use external Redis instance
   - Update `REDIS_ADDR` in environment
   - Comment out Redis service in docker/docker-compose.yml if using external

4. **Storage**:
   - Consider using persistent volumes for uploads
   - Or use MinIO/S3 for object storage

5. **Monitoring**:
   - Set up logging aggregation
   - Monitor container health
   - Set up alerts for container failures

6. **Build Process**:
   - Build binaries on CI/CD pipeline
   - Store binaries in artifact repository if needed
   - Use multi-stage builds in CI for consistency

## Development vs Production

For development, you can:
- Build binaries locally
- Use development mode (`NODE_ENV=development`)
- Enable debug logging

For production:
- Build binaries on build server/CI
- Use production mode (`NODE_ENV=production`)
- Set proper resource limits
- Use orchestration tools (Kubernetes, Docker Swarm)

## File Structure

```
ar-13-go-backend/
├── bin/                    # Pre-built binaries (created by build script)
│   ├── server             # Server binary
│   └── worker             # Worker binary
├── docker/                 # Docker configuration files
│   ├── Dockerfile.server  # Server Dockerfile
│   ├── Dockerfile.worker  # Worker Dockerfile
│   ├── docker-compose.yml # Docker Compose configuration
│   └── .dockerignore      # Files to ignore in Docker build
├── build-binaries.sh      # Build script (Linux/macOS)
└── build-binaries.ps1     # Build script (Windows)
```

## Notes

- Binaries must be built for Linux (`GOOS=linux`) even if building on Windows/macOS
- The `bin/` directory is included in Docker builds (not ignored)
- Binaries are statically linked (CGO_ENABLED=0) for maximum compatibility
- After code changes, always rebuild binaries before rebuilding Docker images
