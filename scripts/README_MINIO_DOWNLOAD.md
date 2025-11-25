# MinIO Download/Read Script

Script to read and download data from MinIO/NAS server.

## Features

- ✅ List all objects in a bucket
- ✅ Download specific objects
- ✅ Filter by prefix (folder path)
- ✅ Recursive listing/download
- ✅ Reads configuration from environment variables
- ✅ Command-line flags override environment variables

## Usage

### Prerequisites

Set environment variables in your `.env` file or export them:

```env
MINIO_ENDPOINT=your-nas-ip:9000
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
MINIO_BUCKET=ar-13-uploads
MINIO_USE_SSL=true
MINIO_INSECURE_SSL=false
```

### List All Objects in Bucket

**Using environment variables:**
```bash
go run scripts/test_minio_download.go --list
```

**Using command-line flags:**
```bash
go run scripts/test_minio_download.go \
  --endpoint "192.168.0.118:9000" \
  --access-key "your-key" \
  --secret-key "your-secret" \
  --bucket "ar-13-uploads" \
  --list
```

### List Objects with Prefix (Folder)

List objects in a specific folder:

```bash
go run scripts/test_minio_download.go --list --prefix "uploads/2024/"
```

### List Recursively

List all objects including subfolders:

```bash
go run scripts/test_minio_download.go --list --recursive
```

### Download Specific Object

**Using environment variables:**
```bash
go run scripts/test_minio_download.go --object "path/to/file.txt"
```

**Specify output path:**
```bash
go run scripts/test_minio_download.go \
  --object "path/to/file.txt" \
  --output "downloaded_file.txt"
```

**Using command-line flags:**
```bash
go run scripts/test_minio_download.go \
  --endpoint "192.168.0.118:9000" \
  --access-key "your-key" \
  --secret-key "your-secret" \
  --bucket "ar-13-uploads" \
  --object "path/to/file.txt" \
  --output "local_file.txt"
```

## Command-Line Options

| Flag | Description | Required |
|------|-------------|----------|
| `--endpoint` | MinIO endpoint (e.g., nas.example.com:9000) | Yes* |
| `--access-key` | Access Key ID | Yes* |
| `--secret-key` | Secret Access Key | Yes* |
| `--bucket` | Bucket name | No (defaults to `ar-13-uploads`) |
| `--object` | Object name to download | No (if not provided, lists objects) |
| `--output` | Output file path | No (defaults to object name) |
| `--list` | Only list objects, don't download | No |
| `--prefix` | Filter objects by prefix | No |
| `--recursive` | List/download recursively | No |
| `--use-ssl` | Use SSL/TLS connection | No |
| `--insecure-ssl` | Skip SSL certificate verification | No |
| `--region` | Region (default: us-east-1) | No |

*Required if not set in environment variables

## Examples

### Example 1: List all files in bucket

```bash
go run scripts/test_minio_download.go --list
```

Output:
```
=== MinIO/S3 Read/Download Test ===
Endpoint: 192.168.0.118:9000
Use SSL: false
Bucket: ar-13-uploads

Test 1: Testing connection...
  Found 1 bucket(s)
    - ar-13-uploads (created: 2024-11-21T10:00:00Z)
✓ Connection successful!

Test 2: Checking bucket...
✓ Bucket 'ar-13-uploads' exists!

Test 3: Listing objects...
  Objects in bucket:
  --------------------------------------------------------------------------------
  Name                                                 Size      Last Modified
  --------------------------------------------------------------------------------
  uploads/file1.txt                                   1.23 KB   2024-11-21 10:30:00
  uploads/file2.pdf                                   45.67 KB  2024-11-21 11:00:00
  --------------------------------------------------------------------------------
  Total: 2 object(s), 46.90 KB
```

### Example 2: Download a file

```bash
go run scripts/test_minio_download.go --object "uploads/file1.txt"
```

Output:
```
=== MinIO/S3 Read/Download Test ===
...

Test 3: Downloading object...
  Object: uploads/file1.txt
  Size: 1.23 KB
  Content Type: text/plain
  Last Modified: 2024-11-21T10:30:00Z
  ETag: "abc123..."

  Downloading to: file1.txt
  ✓ Downloaded 1.23 KB successfully
```

### Example 3: List files in a folder

```bash
go run scripts/test_minio_download.go --list --prefix "uploads/2024/"
```

### Example 4: Download with custom output path

```bash
go run scripts/test_minio_download.go \
  --object "uploads/document.pdf" \
  --output "downloaded_document.pdf"
```

## PowerShell Usage

**Single line:**
```powershell
go run scripts\test_minio_download.go --list
```

**With line continuation (backticks):**
```powershell
go run scripts\test_minio_download.go `
  --endpoint "192.168.0.118:9000" `
  --access-key "your-key" `
  --secret-key "your-secret" `
  --list
```

## Troubleshooting

### "The difference between the request time and the server's time is too large"

Sync your Windows clock:
```powershell
# Run as Administrator
w32tm /resync /force
```

### "Bucket does not exist"

- Check bucket name is correct
- Verify you have access to the bucket
- Create bucket using upload script or MinIO console

### "Object not found"

- Check object name/path is correct
- Use `--list` to see available objects
- Check if object is in a subfolder (use full path)

### Connection Issues

- Verify endpoint is correct
- Check if SSL is required (`--use-ssl`)
- For self-signed certificates, use `--insecure-ssl`
- Check firewall settings

## Related Scripts

- `test_minio_upload.go` - Upload files to MinIO
- See `docs/nas/` for MinIO setup guides

