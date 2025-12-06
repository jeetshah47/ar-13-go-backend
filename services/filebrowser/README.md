# File Browser Service

A lightweight Go-based file browser microservice for browsing files in the "studio work" folder on NAS. This service provides REST API endpoints for listing directories, getting file metadata, and downloading files.

## Features

- **Directory Browsing**: List files and folders in a directory
- **File Metadata**: Get file information (size, type, modified date)
- **File Download**: Stream large files efficiently
- **Security**: Path traversal protection
- **CORS Support**: Cross-origin requests enabled for frontend integration
- **Health Check**: Built-in health check endpoint
- **Docker Ready**: Multi-stage Docker build for minimal image size (~15-20MB)

## API Endpoints

### Health Check
```
GET /health
```

Returns service health status.

**Response:**
```json
{
  "status": "healthy",
  "service": "file-browser",
  "timestamp": "2024-11-22T10:30:00Z"
}
```

### Browse Directory
```
GET /api/browse?path=/folder
```

List files and folders at the specified path.

**Query Parameters:**
- `path` (optional): Directory path to browse (default: `/`)

**Response:**
```json
{
  "path": "/folder",
  "files": [
    {
      "name": "file.pdf",
      "path": "/folder/file.pdf",
      "isFolder": false,
      "size": 1024000,
      "modified": "2024-11-22T10:30:00Z",
      "mimeType": "application/pdf"
    },
    {
      "name": "subfolder",
      "path": "/folder/subfolder",
      "isFolder": true,
      "size": 0,
      "modified": "2024-11-22T10:30:00Z"
    }
  ]
}
```

### Get File Info
```
GET /api/file-info?path=/folder/file.pdf
```

Get metadata for a specific file or folder.

**Query Parameters:**
- `path` (required): Path to the file or folder

**Response:**
```json
{
  "name": "file.pdf",
  "path": "/folder/file.pdf",
  "isFolder": false,
  "size": 1024000,
  "modified": "2024-11-22T10:30:00Z",
  "mimeType": "application/pdf"
}
```

### Download File
```
GET /api/download?path=/folder/file.pdf
```

Download a file with streaming support for large files.

**Query Parameters:**
- `path` (required): Path to the file to download

**Response:**
- Binary file stream with appropriate headers

## Configuration

The service can be configured using environment variables:

- `PORT`: Server port (default: `8082`)
- `DATA_ROOT`: Root directory to serve files from (default: `/data`)
- `JWT_SECRET`: JWT secret key for token authentication (required, must match JWT_SECRET in main backend)

## Building

### Direct Binary Build (No Docker)

**Windows (PowerShell):**
```powershell
cd services/filebrowser

# Build for Windows (default)
.\build.ps1

# Build for Linux AMD64
.\build.ps1 -OS linux -Arch amd64

# Build for Linux ARM64 (for NAS)
.\build.ps1 -OS linux -Arch arm64

# Build for macOS
.\build.ps1 -OS darwin -Arch amd64
```

**Linux/macOS (Bash):**
```bash
cd services/filebrowser

# Build for Linux (default)
./build.sh

# Build for Linux ARM64 (for NAS)
./build.sh linux arm64

# Build for Windows
./build.sh windows amd64

# Build for macOS
./build.sh darwin amd64
```

**Manual Build:**
```bash
cd services/filebrowser
go mod download

# Windows
go build -o filebrowser-service.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o filebrowser-service-linux-amd64 .

# Linux ARM64 (for NAS)
GOOS=linux GOARCH=arm64 go build -o filebrowser-service-linux-arm64 .
```

### Docker Build
```bash
cd services/filebrowser
docker build -t filebrowser-service:latest .
```

## Running

### Local Run

**With JWT Secret:**
```bash
export DATA_ROOT=/path/to/data
export PORT=8082
export JWT_SECRET=your-jwt-secret-here
./filebrowser-service
```

**Note:** JWT_SECRET is required and must match the JWT_SECRET in the main backend.

**Windows:**
```powershell
$env:DATA_ROOT = "D:/Resumers/"
$env:PORT = "8082"
$env:JWT_SECRET = "your-jwt-secret-here"
.\filebrowser-service.exe
```

### Docker Run
```bash
docker run -d \
  --name filebrowser \
  -p 8082:8082 \
  -v /path/to/data:/data \
  -e DATA_ROOT=/data \
  -e JWT_SECRET=your-jwt-secret-here \
  filebrowser-service:latest
```

### Docker Compose

See `docker-compose.filebrowser-service.nas.yml` for NAS deployment or `docker-compose.filebrowser-service.windows.yml` for Windows deployment.

## Security

- **JWT Token Authentication**: API endpoints require `Authorization: Bearer <token>` header with a valid JWT token
- **Path Traversal Protection**: All paths are validated to ensure they stay within the `DATA_ROOT` directory
- **Non-root User**: Container runs as non-root user (UID 1000, GID 1000)
- **CORS**: Configurable CORS headers for frontend integration

### Authentication

All `/api/*` endpoints require JWT token authentication via the `Authorization` header:

```bash
curl -H "Authorization: Bearer <your-jwt-token>" http://localhost:8082/api/browse?path=/
```

The JWT token must be signed with the same `JWT_SECRET` used by the main backend. The token is typically obtained by authenticating with the main backend's `/api/auth/login` endpoint.

## Integration

This service is designed to be integrated with the main backend application. The backend can:

1. Proxy requests to this service
2. JWT authentication is already implemented - ensure JWT_SECRET matches the main backend
3. Cache responses for better performance
4. Add rate limiting

## Troubleshooting

### Permission Issues
If you encounter permission issues, ensure:
- The mounted volume has correct permissions
- The container user (UID 1000) has read access to the files

### Port Conflicts
If port 8082 is already in use, change it using the `PORT` environment variable.

### Path Not Found
Ensure the `DATA_ROOT` directory exists and is properly mounted in the container.

## License

Part of the AR-13 project.

