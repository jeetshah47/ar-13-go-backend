# Setup MinIO with "studio work" Folder

This guide shows how to configure MinIO to use your "studio work" folder on the NAS.

## Step 1: Prepare "studio work" Folder

The "studio work" folder will be used directly by MinIO. All files uploaded through MinIO will be visible in this folder.

### Via File Station:

1. Open **File Station** on your NAS
2. Navigate to **"studio work"** folder (should be at `/share/studio work/`)
3. The folder can contain existing files - MinIO will add its bucket structure here
4. MinIO will create hidden folders (`.minio-config` and `.minio-certs`) for its internal use

### Via SSH:

```bash
# Create hidden folders for MinIO config (won't show in normal file browsing)
mkdir -p "/share/studio work/.minio-config"
mkdir -p "/share/studio work/.minio-certs"
```

**Note:** The `.minio-config` and `.minio-certs` folders start with a dot, so they're hidden in most file browsers but won't interfere with your files.

## Step 2: Set Permissions

Set correct permissions so MinIO can read/write to the entire "studio work" folder:

```bash
# Set permissions on the main folder
chmod -R 755 "/share/studio work"
chown -R 1000:1000 "/share/studio work"

# Set permissions on hidden config folders
chmod -R 755 "/share/studio work/.minio-config"
chmod -R 755 "/share/studio work/.minio-certs"
chown -R 1000:1000 "/share/studio work/.minio-config"
chown -R 1000:1000 "/share/studio work/.minio-certs"
```

**Note:** The quotes are important because the folder name has a space!

## Step 3: Update docker-compose File

The `docker/docker-compose.minio.nas.yml` file has been updated to use:
- `/share/studio work/minio-data`
- `/share/studio work/minio-config`
- `/share/studio work/minio-certs`

**Important:** The paths are quoted in the YAML file to handle the space in the folder name.

## Step 4: Migrate Existing Data (If Any)

If you have existing MinIO data in `/share/Container/minio/`:

1. **Stop MinIO:**
   ```bash
   docker-compose -f docker/docker-compose.minio.nas.yml down
   ```

2. **Copy data:**
   ```bash
   # Copy MinIO bucket data directly to "studio work" folder
   cp -r /share/Container/minio/data/* "/share/studio work/"
   # Copy config to hidden folder
   cp -r /share/Container/minio/config/* "/share/studio work/.minio-config/"
   cp -r /share/Container/minio/certs/* "/share/studio work/.minio-certs/"
   ```

3. **Start MinIO with new paths:**
   ```bash
   docker-compose -f docker/docker-compose.minio.nas.yml up -d
   ```

**Note:** After migration, your MinIO buckets (like `ar-13-uploads`) will appear as folders in "studio work" that you can see via File Station and network shares.

## Step 5: Deploy Updated Configuration

### Via Container Station:

1. Open **Container Station**
2. Find your **minio** container
3. Click **Stop**
4. Click **Edit**
5. Update the YAML with the new volume paths (already done in the file)
6. Click **Apply** or **Update**
7. Click **Start**

### Via SSH:

```bash
cd /share/Container/minio  # or wherever your docker-compose file is
   docker-compose -f docker/docker-compose.minio.nas.yml down
   docker-compose -f docker/docker-compose.minio.nas.yml up -d
```

## Step 6: Verify It Works

1. **Check container is running:**
   ```bash
   docker ps | grep minio
   ```

2. **Check logs:**
   ```bash
   docker logs minio
   ```

3. **Access MinIO Console:**
   - Go to `https://your-nas-ip:9001`
   - Login and verify you can see your buckets/data

4. **Test upload/download:**
   - Use the test scripts to verify read/write works

## Troubleshooting

### "Permission Denied" Error

**Solution:**
```bash
# Set permissions on the entire folder
chmod -R 755 "/share/studio work"
chown -R 1000:1000 "/share/studio work"
```

### "No Such File or Directory"

**Solution:**
- Make sure the folders exist
- Check the path is correct (case-sensitive)
- Verify the folder name is exactly "studio work" (with space)

### Container Won't Start

**Solution:**
1. Check logs: `docker logs minio`
2. Verify folder paths in docker-compose file
3. Make sure folders exist and have correct permissions

### Spaces in Folder Names

**Important:** When using folders with spaces:
- Always use quotes in commands: `"/share/studio work/..."`
- YAML file already has quotes around the paths
- Be careful with paths in scripts

## Folder Structure

After setup, your "studio work" folder will have:

```
studio work/
├── ar-13-uploads/       # Your MinIO bucket (visible folder)
│   └── [your files]    # All files uploaded through MinIO
├── [other buckets]/     # Any other MinIO buckets you create
├── [your existing files] # Any files you already had in "studio work"
├── .minio-config/       # MinIO configuration (hidden, starts with dot)
└── .minio-certs/        # SSL certificates (hidden, starts with dot)
```

## Accessing Files

All files uploaded through MinIO will be stored directly in:
```
/share/studio work/ar-13-uploads/
```

You can access these files in multiple ways:

1. **Via MinIO Console:** Web interface at port 9001
   - Login and browse buckets
   - Upload/download through web UI

2. **Via File Station:** Navigate to "studio work/ar-13-uploads/"
   - See all files directly
   - Can copy/move files manually
   - Files uploaded via MinIO appear here immediately

3. **Via Network Share (SMB):** 
   - Map "studio work" as a network drive
   - Access `\\your-nas-ip\studio work\ar-13-uploads\`
   - All files visible and accessible

4. **Via your application:** Using the MinIO API
   - Backend uses MinIO SDK to upload/download
   - Files are stored in "studio work" folder

## Benefits

✅ **All files visible** - No hidden subfolders, everything in "studio work"  
✅ **Direct access** - Can access files via network share without MinIO  
✅ **Backend integration** - Your app uses MinIO API for uploads  
✅ **Flexible** - Can manually add files to folders, MinIO will see them  
✅ **Backup friendly** - Just backup "studio work" folder

## Summary

✅ Updated docker-compose to use "studio work" folder  
✅ Paths are quoted to handle spaces  
✅ Create folders and set permissions  
✅ Restart container  
✅ All MinIO data will be in "studio work" folder

