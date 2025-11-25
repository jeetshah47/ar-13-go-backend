# FileBrowser Setup Guide

FileBrowser is a web-based file browser that allows you to view, download, and manage files directly from your browser. Unlike MinIO, FileBrowser shows **existing files immediately** without requiring import.

## Why FileBrowser?

- ✅ **Shows existing files** - No need to import or move files
- ✅ **Works with large files** - Streaming downloads for GB-sized files
- ✅ **Simple web UI** - Easy browsing for clients
- ✅ **User authentication** - Secure access control
- ✅ **Upload/Download** - Full file management capabilities

## Quick Start

### For Windows Machine

1. **Start FileBrowser:**
   ```powershell
   cd docs\nas\filebrowser
   docker-compose -f docker-compose.filebrowser.windows.yml up -d
   ```

2. **Access FileBrowser:**
   - Open browser: http://localhost:8081
   - **Get the password from logs:**
     ```powershell
     docker logs filebrowser
     ```
     Look for: `User 'admin' initialized with randomly generated password: <PASSWORD>`
   - Or use default credentials (if set in docker-compose):
     - Username: `admin`
     - Password: `admin` (or check logs for random password)
   - **⚠️ Change password immediately after first login!**

3. **Stop FileBrowser:**
   ```powershell
   docker-compose -f docker-compose.filebrowser.windows.yml down
   ```

### For NAS (Container Station)

1. **The configuration is already set** to use `/share/studio work/` folder
   - If you need to use a different folder, edit `docker-compose.filebrowser.nas.yml`:
   - Change `/share/studio work/` to your desired NAS folder path
   - Common paths:
     - `/share/Public/` - Public folder
     - `/share/Container/` - Container Station folder
     - `/share/Download/` - Download folder
     - Or create custom folder: `/share/YourFolder/`

2. **Deploy via Container Station:**
   - Upload `docker-compose.filebrowser.nas.yml` to Container Station
   - Or use SSH:
     ```bash
     cd /share/Container/filebrowser
     docker-compose -f docker-compose.filebrowser.nas.yml up -d
     ```

3. **Access FileBrowser:**
   - Open browser: http://YOUR-NAS-IP:8081
   - **Get the password from logs:**
     ```bash
     docker logs filebrowser
     ```
     Look for: `User 'admin' initialized with randomly generated password: <PASSWORD>`
   - Or use default credentials (if set in docker-compose):
     - Username: `admin`
     - Password: `admin` (or check logs for random password)
   - **⚠️ Change password immediately after first login!**

## Configuration

### Changing Default Credentials

After first login, go to **Settings** → **User Management** to:
- Change admin password
- Create additional users
- Set permissions per user

### Changing Root Directory

**Windows:**
Edit `docker-compose.filebrowser.windows.yml`:
```yaml
volumes:
  - "D:/YourFolder/:/srv"  # Change D:/YourFolder/ to your path
```

**NAS:**
Edit `docker-compose.filebrowser.nas.yml`:
```yaml
volumes:
  - "/share/YourFolder/:/srv"  # Change to your NAS folder path
```

Then restart:
```bash
docker-compose -f docker-compose.filebrowser.windows.yml down
docker-compose -f docker-compose.filebrowser.windows.yml up -d
```

### Port Configuration

To change the port (default: 8081):

**Windows:**
```yaml
ports:
  - "9000:80"  # Change 9000 to your desired port
```

**NAS:**
```yaml
ports:
  - "9000:80"  # Change 9000 to your desired port
```

## Features

### File Browsing
- Navigate folders and subfolders
- View file details (size, date, permissions)
- Preview images and text files
- Download files (streaming for large files)

### File Management
- Upload files
- Create folders
- Rename files/folders
- Delete files/folders
- Move/copy files

### User Management
- Multiple users
- Permission control (read-only, read-write, admin)
- User-specific home directories

### Security
- User authentication
- HTTPS support (requires reverse proxy)
- IP whitelisting (via reverse proxy)

## Integration with Backend

### Option 1: Direct File Access
FileBrowser provides direct file URLs that can be used in your backend:

```
http://localhost:8081/api/raw/path/to/file.pdf
```

### Option 2: API Integration
FileBrowser has a REST API for programmatic access:

```bash
# List files
curl http://localhost:8081/api/resources/path/to/folder

# Download file
curl http://localhost:8081/api/raw/path/to/file.pdf
```

### Option 3: Embed in Frontend
You can embed FileBrowser in an iframe or redirect users to FileBrowser URL.

## Troubleshooting

### "Permission Denied" Error

**Windows:**
- Ensure Docker Desktop has access to `D:/Resumers/`
- Check folder permissions in Windows
- If config folder has permission issues:
  ```powershell
  Remove-Item -Recurse -Force .\filebrowser-config -ErrorAction SilentlyContinue
  New-Item -ItemType Directory -Path .\filebrowser-config
  docker-compose -f docker-compose.filebrowser.windows.yml restart
  ```

**NAS:**
- **Fix config directory permissions:**
  ```bash
  mkdir -p /share/Container/filebrowser-config
  chmod -R 755 /share/Container/filebrowser-config
  chown -R 1000:1000 /share/Container/filebrowser-config
  docker-compose -f docker-compose.filebrowser.nas.yml restart
  ```
- Set correct permissions on data folder:
  ```bash
  chmod -R 755 "/share/studio work/"
  chown -R 1000:1000 "/share/studio work/"
  ```
- See `FIX_PERMISSIONS.md` for detailed troubleshooting

### Files Not Showing

- Verify the volume mount path is correct
- Check folder permissions
- Ensure files exist in the mounted directory
- Restart container after path changes

### Can't Access Web UI

- Check if port 8081 is available
- Verify container is running: `docker ps`
- Check logs: `docker logs filebrowser`
- Try accessing via IP instead of localhost (for NAS)

### Wrong Password / Can't Login

**Get password from logs:**
```bash
docker logs filebrowser
```
Look for: `User 'admin' initialized with randomly generated password: <PASSWORD>`

**Reset password (if needed):**
```bash
# Stop container
docker-compose -f docker-compose.filebrowser.windows.yml down

# Delete config to reset (WARNING: This deletes all users and settings)
# Windows:
Remove-Item -Recurse -Force .\filebrowser-config

# NAS:
rm -rf /share/Container/filebrowser-config

# Start again - will generate new random password
docker-compose -f docker-compose.filebrowser.windows.yml up -d

# Check logs for new password
docker logs filebrowser
```

**Or use FileBrowser CLI to change password:**
```bash
docker exec -it filebrowser filebrowser users update admin --password YOUR_NEW_PASSWORD
```

### Large File Downloads Failing

- FileBrowser supports streaming, but check:
  - Browser timeout settings
  - Network stability
  - Container memory limits (if set)

## Comparison: FileBrowser vs MinIO

| Feature | FileBrowser | MinIO |
|---------|-------------|-------|
| Shows existing files | ✅ Yes | ❌ No (requires import) |
| File browsing | ✅ Yes | ❌ Bucket-based only |
| Large file support | ✅ Streaming | ✅ Streaming |
| S3 API | ❌ No | ✅ Yes |
| Upload/Download | ✅ Yes | ✅ Yes |
| User management | ✅ Yes | ✅ Yes |
| Use case | File browser | Object storage |

## Security Recommendations

1. **Change default password** immediately
2. **Use HTTPS** via reverse proxy (nginx)
3. **Set up user permissions** - don't give everyone admin access
4. **Restrict access** via firewall/network rules
5. **Regular backups** of FileBrowser config database

## Next Steps

1. Set up FileBrowser using appropriate docker-compose file
2. Change default admin password
3. Create user accounts for clients
4. Configure permissions
5. Integrate with your backend/frontend as needed

## Support

For more information:
- FileBrowser Docs: https://filebrowser.org/
- GitHub: https://github.com/filebrowser/filebrowser

