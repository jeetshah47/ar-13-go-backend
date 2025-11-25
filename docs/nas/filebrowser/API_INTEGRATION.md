# FileBrowser API Integration Guide

This guide explains how FileBrowser is integrated into the AR-13 backend project.

## Overview

FileBrowser has been integrated as an alternative storage backend to MinIO. The integration maintains compatibility with the existing `StorageServiceInterface`, allowing seamless switching between MinIO and FileBrowser.

## Architecture

```
┌─────────────────┐
│  StorageHandler │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│ StorageServiceInterface │
└────────┬────────────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌──────────────────────┐
│  MinIO  │ │   FileBrowser       │
│ Service │ │   Storage Service    │
└─────────┘ └──────────────────────┘
                    │
                    ▼
            ┌──────────────────┐
            │ FileBrowserClient│
            │  (REST API)      │
            └──────────────────┘
```

## Configuration

Add the following environment variables to your `.env` file:

```env
# Enable FileBrowser (set to "true" to use FileBrowser instead of MinIO)
FILEBROWSER_ENABLED=true

# FileBrowser URL (default: http://localhost:8081)
FILEBROWSER_URL=http://localhost:8081

# FileBrowser API Token (required if FILEBROWSER_ENABLED=true)
# Generate this token from FileBrowser web UI: Settings → API Tokens
FILEBROWSER_TOKEN=your_api_token_here
```

### Getting API Token

1. Start FileBrowser container
2. Access FileBrowser web UI: http://localhost:8081
3. Login with admin credentials
4. Go to **Settings** → **API Tokens**
5. Click **Generate Token**
6. Copy the token and add it to your `.env` file

## How It Works

### Service Selection

The backend automatically selects the storage service based on configuration:

1. **If `FILEBROWSER_ENABLED=true`:**
   - Uses `FileBrowserStorageService`
   - Connects to FileBrowser API
   - Shows existing files immediately

2. **If `FILEBROWSER_ENABLED=false` (default):**
   - Uses `MinIOStorageService`
   - Connects to MinIO/S3-compatible storage
   - Requires files to be imported into buckets

### API Endpoints

The existing storage endpoints work with both MinIO and FileBrowser:

#### List Files
```
GET /api/storage/files?path=/folder/subfolder
Authorization: Bearer <your_jwt_token>
```

**Response:**
```json
{
  "files": [
    {
      "name": "file.pdf",
      "path": "/folder/file.pdf",
      "isFolder": false,
      "size": 1024000,
      "lastModified": "2024-11-22T10:30:00Z",
      "contentType": "application/pdf"
    },
    {
      "name": "subfolder",
      "path": "/folder/subfolder",
      "isFolder": true,
      "size": 0,
      "lastModified": "2024-11-22T10:30:00Z"
    }
  ],
  "path": "/folder/subfolder"
}
```

#### Get File URL
```
GET /api/storage/file-url?path=/folder/file.pdf&expiry=3600
Authorization: Bearer <your_jwt_token>
```

**Response:**
```json
{
  "url": "http://localhost:8081/api/raw?path=/folder/file.pdf",
  "path": "/folder/file.pdf",
  "expiry": 3600
}
```

**Note:** FileBrowser doesn't use presigned URLs with expiry. The URL is direct and requires authentication via API token.

#### Upload File
```
POST /api/storage/upload
Authorization: Bearer <your_jwt_token>
Content-Type: multipart/form-data

file: [binary file data]
path: /folder (optional, default: root)
```

**Response:**
```json
{
  "message": "File uploaded successfully",
  "objectName": "/folder/filename.pdf",
  "path": "/folder/filename.pdf",
  "size": 1024000
}
```

## Implementation Details

### FileBrowser Client (`filebrowser_client.go`)

Handles direct communication with FileBrowser REST API:

- `ListFiles(path)` - Lists files and folders
- `GetFileURL(path)` - Gets direct file access URL
- `UploadFile(path, reader, filename)` - Uploads file
- `DeleteFile(path)` - Deletes file/folder
- `FileExists(path)` - Checks if file exists

### FileBrowser Storage Service (`filebrowser_storage_service.go`)

Implements `StorageServiceInterface`:

- `Initialize()` - Verifies FileBrowser connectivity
- `ListObjects(ctx, prefix)` - Lists files/folders
- `UploadFile(ctx, objectName, reader, size, contentType)` - Uploads file
- `GetPresignedURL(ctx, objectName, expiry)` - Returns direct URL (expiry ignored)
- `DeleteObject(ctx, objectName)` - Deletes file
- `ObjectExists(ctx, objectName)` - Checks file existence

### Handler Integration (`handler.go`)

The handler automatically selects the storage service:

```go
if cfg.FileBrowserEnabled {
    storageService = services.NewFileBrowserStorageService(cfg.FileBrowserURL, cfg.FileBrowserToken)
} else {
    storageService = services.NewStorageService(cfg) // MinIO
}
```

## Advantages of FileBrowser

1. **Shows Existing Files** - No need to import files first
2. **Simple Integration** - REST API is straightforward
3. **Large File Support** - Streaming downloads work well
4. **File Management** - Built-in upload, download, delete capabilities
5. **User Management** - Built-in authentication and permissions

## Migration from MinIO

To switch from MinIO to FileBrowser:

1. **Start FileBrowser:**
   ```powershell
   cd docs\nas\filebrowser
   docker-compose -f docker-compose.filebrowser.windows.yml up -d
   ```

2. **Get API Token:**
   - Login to FileBrowser web UI
   - Generate API token from Settings

3. **Update `.env`:**
   ```env
   FILEBROWSER_ENABLED=true
   FILEBROWSER_URL=http://localhost:8081
   FILEBROWSER_TOKEN=your_token_here
   ```

4. **Restart Backend:**
   - The backend will automatically use FileBrowser
   - No code changes needed!

## Switching Back to MinIO

To switch back to MinIO:

1. **Update `.env`:**
   ```env
   FILEBROWSER_ENABLED=false
   # Ensure MinIO config is set
   MINIO_ENDPOINT=your-minio-endpoint
   MINIO_ACCESS_KEY=your-access-key
   MINIO_SECRET_KEY=your-secret-key
   ```

2. **Restart Backend**

## Troubleshooting

### "Storage service not initialized" Error

**Cause:** FileBrowser token is missing or invalid

**Solution:**
1. Verify `FILEBROWSER_TOKEN` is set in `.env`
2. Check token is valid in FileBrowser web UI
3. Verify FileBrowser is running: `docker ps | grep filebrowser`

### Files Not Showing

**Cause:** Path format issue or FileBrowser not accessible

**Solution:**
1. Verify FileBrowser is accessible: `curl http://localhost:8081/api/resources?path=/`
2. Check path format (should start with `/`)
3. Verify token has read permissions

### Upload Fails

**Cause:** Token doesn't have write permissions or path doesn't exist

**Solution:**
1. Verify token has write permissions in FileBrowser
2. Ensure parent directory exists
3. Check FileBrowser logs: `docker logs filebrowser`

## Testing

### Test FileBrowser Connection

```bash
# List root directory
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8081/api/resources?path=/

# Upload test file
curl -X POST \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test.txt" \
  http://localhost:8081/api/resources?path=/
```

### Test Backend Integration

```bash
# List files via backend
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:3000/api/storage/files?path=/

# Upload file via backend
curl -X POST \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@test.pdf" \
  http://localhost:3000/api/storage/upload
```

## Next Steps

1. ✅ FileBrowser client service created
2. ✅ FileBrowser storage service implemented
3. ✅ Configuration added
4. ✅ Handler integration complete
5. 🔄 Test the integration
6. 🔄 Generate API token and configure
7. 🔄 Deploy to production

## References

- FileBrowser GitHub: https://github.com/filebrowser/filebrowser
- FileBrowser API Docs: Access Swagger UI at `http://localhost:8081/swagger/` (if enabled)
- Project Storage Service: `internal/services/storage_service.go`
- Project Storage Handler: `internal/handlers/storage.go`

