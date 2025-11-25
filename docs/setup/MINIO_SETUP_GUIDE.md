# MinIO Setup Guide (Local Development)

Simple guide to set up MinIO for local development using Docker.

**For NAS/Production setup, see:** `docs/nas/MINIO_SETUP_GUIDE.md`  
**For detailed local setup, see:** `docs/setup/MINIO_LOCAL_SETUP.md`

## Quick Start

1. **Start MinIO:**
   ```bash
   docker-compose -f docker-compose.minio.yml up -d
   ```

2. **Access MinIO Console:**
   - Open: `http://localhost:9001`
   - Login with:
     - Username: `minioadmin`
     - Password: `minioadmin`

3. **Create a Bucket:**
   - In the Console, go to **Buckets**
   - Click **Create Bucket**
   - Name it: `ar-13-uploads`
   - Click **Create**

4. **Create Access Keys:**
   - Go to **Identity** → **Access Keys**
   - Click **Create Access Key**
   - Save the **Access Key** and **Secret Key** (you'll need these)

5. **Configure Your Backend:**
   
   Add to your `.env` file:
```env
   MINIO_ENDPOINT=localhost:9000
   MINIO_ACCESS_KEY=your-access-key-here
   MINIO_SECRET_KEY=your-secret-key-here
MINIO_BUCKET=ar-13-uploads
MINIO_USE_SSL=false
MINIO_INSECURE_SSL=false
```

6. **Test the Connection:**
```bash
go run scripts/test_minio_upload.go \
     --endpoint "localhost:9000" \
  --access-key "your-access-key" \
  --secret-key "your-secret-key" \
  --bucket "ar-13-uploads" \
  --file "test.txt"
   ```

## Ports

- **9000**: MinIO API (S3 operations)
- **9001**: MinIO Console (Web UI)

## Change Default Credentials

Edit `docker-compose.minio.yml` and change:
```yaml
environment:
  MINIO_ROOT_USER: your-username
  MINIO_ROOT_PASSWORD: your-secure-password
```

Then restart:
```bash
docker-compose -f docker-compose.minio.yml down
docker-compose -f docker-compose.minio.yml up -d
```

## Troubleshooting

**Container won't start:**
   ```bash
   docker logs minio
   ```

**Reset everything:**
```bash
docker-compose -f docker-compose.minio.yml down
rm -rf ./minio-data
docker-compose -f docker-compose.minio.yml up -d
```

## Production Notes

For production:
- Change default credentials
- Use strong passwords
- Enable SSL/TLS
- Configure proper access policies
- Set up backups
