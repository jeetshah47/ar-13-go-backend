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
- `SECRET_KEY`: Secret key for API authentication (optional, if not set, all requests are allowed)

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

**With Secret Key:**
```bash
export DATA_ROOT=/path/to/data
export PORT=8082
export SECRET_KEY=your-secret-key-here
./filebrowser-service
```

**Without Secret Key (Development):**
```bash
export DATA_ROOT=/path/to/data
export PORT=8082
./filebrowser-service
```

**Windows:**
```powershell
$env:DATA_ROOT = "D:/Resumers/"
$env:PORT = "8082"
$env:SECRET_KEY = "your-secret-key-here"
.\filebrowser-service.exe
```

### Docker Run
```bash
docker run -d \
  --name filebrowser \
  -p 8082:8082 \
  -v /path/to/data:/data \
  -e DATA_ROOT=/data \
  filebrowser-service:latest
```

### Docker Compose

See `docker-compose.filebrowser-service.nas.yml` for NAS deployment or `docker-compose.filebrowser-service.windows.yml` for Windows deployment.

## Security

- **Secret Key Authentication**: API endpoints require `X-API-Key` header with the secret key (if `SECRET_KEY` is set)
- **Path Traversal Protection**: All paths are validated to ensure they stay within the `DATA_ROOT` directory
- **Non-root User**: Container runs as non-root user (UID 1000, GID 1000)
- **CORS**: Configurable CORS headers for frontend integration

### Authentication

When `SECRET_KEY` environment variable is set, all `/api/*` endpoints require the `X-API-Key` header:

```bash
curl -H "X-API-Key: your-secret-key" http://localhost:8082/api/browse?path=/
```

If `SECRET_KEY` is not set, all requests are allowed (useful for development).

## Integration

This service is designed to be integrated with the main backend application. The backend can:

1. Proxy requests to this service
2. Add authentication/authorization middleware
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

