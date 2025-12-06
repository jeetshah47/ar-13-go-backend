# Configure MinIO to Use Different NAS Folders

By default, MinIO uses `/share/Container/minio/` but you can configure it to use any folder on your NAS.

## Understanding NAS Folder Structure

On QNAP NAS (and similar), common folder locations:

- `/share/Container/` - Container Station default location
- `/share/Public/` - Public shared folder
- `/share/Download/` - Download folder
- `/share/Multimedia/` - Media files
- `/share/homes/` - User home directories
- `/share/` - Root of all shared folders

You can also create custom shared folders via:
- **File Station** → Right-click → **Create Shared Folder**
- **Control Panel** → **Shared Folders**

## Option 1: Use a Different Shared Folder

### Step 1: Create or Choose a Folder

**Via File Station:**
1. Open **File Station** on your NAS
2. Navigate to `/share/`
3. Create a new folder (e.g., `minio-storage`) or use an existing one
4. Note the full path (e.g., `/share/minio-storage/`)

**Via SSH:**
```bash
mkdir -p /share/minio-storage/data
mkdir -p /share/minio-storage/config
mkdir -p /share/minio-storage/certs
```

### Step 2: Update docker-compose File

Edit `docker/docker-compose.minio.nas.yml` and change the volume paths:

```yaml
version: '3.8'

services:
  minio:
    image: minio/minio:latest
    container_name: minio
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
      TZ: Asia/Kolkata
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      # Change these paths to your desired location
      - /share/minio-storage/data:/data
      - /share/minio-storage/config:/root/.minio
      - /share/minio-storage/certs:/root/.minio/certs
    networks:
      - minio-network
    stdin_open: true
    tty: true
    healthcheck:
      test: ["CMD", "curl", "-f", "https://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
      start_period: 40s

networks:
  minio-network:
    driver: bridge
```

### Step 3: Set Permissions

Ensure the folder has correct permissions:

```bash
# Via SSH
chmod -R 755 /share/minio-storage
chown -R 1000:1000 /share/minio-storage  # MinIO runs as UID 1000
```

**Or via File Station:**
1. Right-click folder → **Properties**
2. Go to **Permissions** tab
3. Set appropriate permissions

### Step 4: Restart Container

```bash
# Stop container
docker-compose -f docker/docker-compose.minio.nas.yml down

# Start with new paths
docker-compose -f docker/docker-compose.minio.nas.yml up -d
```

## Option 2: Use Multiple Storage Locations

You can mount multiple folders for MinIO to use:

```yaml
volumes:
  # Primary storage
  - /share/minio-storage/data:/data
  # Additional storage locations (if MinIO supports multiple drives)
  - /share/storage1:/data/drive1
  - /share/storage2:/data/drive2
```

**Note:** MinIO in single-server mode uses `/data` as the storage root. For multiple drives, you'd need to configure MinIO differently or use a distributed setup.

## Option 3: Use External Drive or USB

If you have an external drive mounted:

1. **Find the mount point:**
   ```bash
   # Via SSH
   df -h
   # Look for your external drive, e.g., /share/external_01/
   ```

2. **Update docker-compose:**
   ```yaml
   volumes:
     - /share/external_01/minio-data:/data
     - /share/external_01/minio-config:/root/.minio
   ```

## Option 4: Access Network Shares

If you want to use a network share (SMB/NFS):

1. **Mount the network share first** (via Control Panel → Network Access)
2. **Use the mount point** in docker-compose

**Example for SMB mount:**
```yaml
volumes:
  - /mnt/smb-share/minio:/data
```

## Common Folder Examples

### Example 1: Use Public Folder
```yaml
volumes:
  - /share/Public/minio-data:/data
  - /share/Public/minio-config:/root/.minio
```

### Example 2: Use Custom Shared Folder
```yaml
volumes:
  - /share/Storage/minio:/data
  - /share/Storage/minio-config:/root/.minio
```

### Example 3: Use Home Directory
```yaml
volumes:
  - /share/homes/admin/minio:/data
  - /share/homes/admin/minio-config:/root/.minio
```

## Finding Available Folders

**Via File Station:**
1. Open File Station
2. Navigate to `/share/`
3. See all available shared folders

**Via SSH:**
```bash
ls -la /share/
```

**Via Container Station:**
- When configuring volumes, Container Station shows available paths

## Permission Issues

If you get permission errors:

1. **Check folder ownership:**
   ```bash
   ls -la /share/your-folder
   ```

2. **Set correct permissions:**
   ```bash
   chmod -R 755 /share/your-folder
   chown -R 1000:1000 /share/your-folder
   ```

3. **Or use a folder that Container Station can access:**
   - Folders in `/share/Container/` usually work without permission issues
   - Other folders might need explicit permissions

## Best Practices

1. **Dedicated Folder:** Create a dedicated folder for MinIO (e.g., `/share/minio-storage/`)
2. **Organized Structure:**
   ```
   /share/minio-storage/
   ├── data/          # MinIO data files
   ├── config/        # MinIO configuration
   └── certs/         # SSL certificates
   ```
3. **Backup Location:** Consider using a folder that's included in your backup strategy
4. **Performance:** Use faster storage (SSD) if available for better performance

## Troubleshooting

### "Permission Denied" Error

**Solution:**
```bash
# Set permissions
chmod -R 755 /share/your-folder
chown -R 1000:1000 /share/your-folder

# Or add Container Station user to folder permissions
```

### "No Such File or Directory"

**Solution:**
- Ensure the folder exists before starting the container
- Check the path is correct (case-sensitive)
- Verify the folder is accessible from Container Station

### Container Can't Access Folder

**Solution:**
- Use folders under `/share/` (shared folders)
- Check folder permissions in File Station
- Ensure Container Station has access to the folder

## Migration from Old Location

If you want to move existing data:

1. **Stop MinIO:**
   ```bash
   docker-compose -f docker/docker-compose.minio.nas.yml down
   ```

2. **Copy data:**
   ```bash
   cp -r /share/Container/minio/data/* /share/new-location/data/
   cp -r /share/Container/minio/config/* /share/new-location/config/
   ```

3. **Update docker-compose** with new paths

4. **Start MinIO:**
   ```bash
   docker-compose -f docker/docker-compose.minio.nas.yml up -d
   ```

5. **Verify data is accessible**

## Summary

- Change volume paths in `docker/docker-compose.minio.nas.yml`
- Use any folder under `/share/` (shared folders)
- Set correct permissions (755 for folders, 1000:1000 ownership)
- Restart container after changes
- Test access after migration

