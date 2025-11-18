# MinIO/S3 Upload Test Script

This script tests the connection to a QNAP NAS server using MinIO (S3-compatible API) and uploads a file.

## Prerequisites

1. MinIO server running on your QNAP NAS
2. Access credentials (Access Key ID and Secret Access Key)
3. Network access to the NAS MinIO endpoint

## Usage

### Basic Usage (HTTP)

```bash
go run scripts/test_minio_upload.go \
  --endpoint "your-nas-ip:9000" \
  --access-key "your-access-key" \
  --secret-key "your-secret-key" \
  --bucket "test-bucket" \
  --file "path/to/your/file.txt"
```

### With SSL/TLS

```bash
go run scripts/test_minio_upload.go \
  --endpoint "your-nas-ip:9000" \
  --access-key "your-access-key" \
  --secret-key "your-secret-key" \
  --use-ssl \
  --bucket "test-bucket" \
  --file "path/to/your/file.txt"
```

### With Self-Signed Certificate (Insecure SSL)

```bash
go run scripts/test_minio_upload.go \
  --endpoint "your-nas-ip:9000" \
  --access-key "your-access-key" \
  --secret-key "your-secret-key" \
  --use-ssl \
  --insecure-ssl \
  --bucket "test-bucket" \
  --file "path/to/your/file.txt"
```

### Custom Object Name

```bash
go run scripts/test_minio_upload.go \
  --endpoint "your-nas-ip:9000" \
  --access-key "your-access-key" \
  --secret-key "your-secret-key" \
  --bucket "test-bucket" \
  --file "path/to/your/file.txt" \
  --object "custom-name-in-bucket.txt"
```

## Command-Line Options

| Option | Description | Required | Default |
|--------|-------------|----------|---------|
| `--endpoint` | MinIO/S3 endpoint (e.g., nas.example.com:9000) | Yes | - |
| `--access-key` | Access Key ID | Yes | - |
| `--secret-key` | Secret Access Key | Yes | - |
| `--use-ssl` | Use SSL/TLS connection | No | false |
| `--insecure-ssl` | Skip SSL certificate verification (for self-signed certs) | No | false |
| `--bucket` | Bucket name | No | test-bucket |
| `--file` | Path to file to upload | Yes | - |
| `--object` | Object name in bucket (defaults to filename) | No | (filename) |
| `--region` | Region (optional, some S3-compatible services ignore this) | No | us-east-1 |

## What the Script Does

1. **Test Connection**: Connects to the MinIO server and lists existing buckets
2. **Check/Create Bucket**: Verifies the bucket exists, creates it if it doesn't
3. **Upload File**: Uploads the specified file to the bucket
4. **Verify Upload**: Verifies the file was uploaded correctly by checking object metadata

## Example Output

```
=== MinIO/S3 Connection Test ===
Endpoint: nas.example.com:9000
Use SSL: false
Bucket: test-bucket
File: ./test-file.txt
Object Name: test-file.txt

Test 1: Testing connection...
  Found 2 bucket(s)
    - test-bucket (created: 2024-01-15T10:30:00Z)
    - another-bucket (created: 2024-01-14T09:20:00Z)
✓ Connection successful!

Test 2: Checking bucket...
  ✓ Bucket 'test-bucket' already exists
✓ Bucket 'test-bucket' is ready!

Test 3: Uploading file...
  Uploading: ./test-file.txt (1024 bytes)
  Content Type: text/plain
✓ File uploaded successfully!
  Object: test-file.txt
  Size: 1024 bytes
  ETag: "abc123def456"

Test 4: Verifying upload...
  Object found: test-file.txt
  Size: 1024 bytes
  Last Modified: 2024-01-15T10:35:00Z
  ETag: "abc123def456"
  Verifying file integrity...
  ✓ Successfully read 100 bytes from uploaded file
✓ Upload verified! Object 'test-file.txt' exists in bucket 'test-bucket'

=== All tests passed! ===
```

## Building the Executable

You can also build an executable:

```bash
go build -o test_minio_upload.exe scripts/test_minio_upload.go
```

Then run it:

```bash
./test_minio_upload.exe --endpoint "nas.example.com:9000" --access-key "key" --secret-key "secret" --file "test.txt"
```

## Troubleshooting

### Connection Refused
- Verify the endpoint IP and port are correct
- Check firewall settings on the NAS
- Ensure MinIO service is running on the NAS

### SSL Certificate Errors
- Use `--insecure-ssl` flag for self-signed certificates
- Or configure proper SSL certificates on the NAS

### Authentication Failed
- Verify your Access Key ID and Secret Access Key
- Check that the credentials are correct in MinIO configuration

### Bucket Creation Failed
- Ensure your credentials have permission to create buckets
- Check MinIO bucket policies

## QNAP NAS MinIO Setup

1. Install MinIO Server package on QNAP
2. Configure MinIO with your credentials
3. Note the port (default is 9000)
4. Ensure the service is running
5. Configure firewall rules if needed

