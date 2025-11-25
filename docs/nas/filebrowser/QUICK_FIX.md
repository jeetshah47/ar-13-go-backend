# Quick Fix for Permission Denied Error

If you're seeing `Error: open /filebrowser-config/database.db: permission denied`, follow these steps:

## Option 1: Fix Permissions (Recommended)

### For NAS:

**Option A: Using File Station (GUI - Recommended)**
1. **Open File Station** on your NAS
2. **Navigate to** `/share/Container/`
3. **Create folder** named `filebrowser-config`
4. **Right-click** → **Properties** → **Permissions**
5. **Set permissions to 755** (rwxr-xr-x)
6. **Set owner** to user with UID 1000 (usually `admin`)
7. **Apply recursively** to subfolders
8. See `SETUP_VIA_FILE_STATION.md` for detailed steps

**Option B: Using Command Line**
1. **Stop the container:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml down
   ```

2. **Run the setup script:**
   ```bash
   cd docs/nas/filebrowser
   chmod +x setup-permissions.sh
   ./setup-permissions.sh
   ```

   **Or manually:**
   ```bash
   mkdir -p /share/Container/filebrowser-config
   chmod -R 755 /share/Container/filebrowser-config
   chown -R 1000:1000 /share/Container/filebrowser-config
   ```

3. **Uncomment the user line** in `docker-compose.filebrowser.nas.yml`:
   ```yaml
   user: "1000:1000"  # Uncomment this line
   ```

4. **Start the container:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml up -d
   ```

### For Windows:

1. **Stop the container:**
   ```powershell
   docker-compose -f docker-compose.filebrowser.windows.yml down
   ```

2. **Remove and recreate config folder:**
   ```powershell
   Remove-Item -Recurse -Force .\filebrowser-config -ErrorAction SilentlyContinue
   New-Item -ItemType Directory -Path .\filebrowser-config
   ```

3. **Start the container:**
   ```powershell
   docker-compose -f docker-compose.filebrowser.windows.yml up -d
   ```

## Option 2: Run as Root (Temporary Fix)

If Option 1 doesn't work, you can temporarily run as root:

1. **Keep `user: "1000:1000"` commented out** (or remove it)
2. **Start the container:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml up -d
   ```

3. **After it starts, fix permissions from inside container:**
   ```bash
   docker exec -it filebrowser sh
   chmod -R 755 /filebrowser-config
   chown -R 1000:1000 /filebrowser-config
   exit
   ```

4. **Then uncomment `user: "1000:1000"` and restart:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml restart
   ```

## Option 3: Use Named Volume (Alternative)

If bind mounts keep causing issues, use a Docker named volume:

Edit `docker-compose.filebrowser.nas.yml`:
```yaml
volumes:
  - "/share/studio work/:/srv"
  - "filebrowser-config:/filebrowser-config"  # Use named volume

# Add at the end:
volumes:
  filebrowser-config:
```

**Note:** Named volumes store data in Docker's volume directory, not in `/share/Container/`

## Verify It's Working

After fixing permissions, check:

```bash
# Check container logs
docker logs filebrowser

# Should see:
# Using database: /filebrowser-config/database.db
# (No permission errors)

# Check if database file was created
docker exec filebrowser ls -la /filebrowser-config/

# Access web UI
# http://YOUR-NAS-IP:8081
```

## Common Issues

### "Permission denied" persists
- Ensure directory exists: `mkdir -p /share/Container/filebrowser-config`
- Check ownership: `ls -la /share/Container/ | grep filebrowser-config`
- Try running without `user:` restriction first, then add it back

### Container keeps restarting
- Check logs: `docker logs filebrowser`
- Verify directory permissions
- Try Option 2 (run as root temporarily)

### Can't change ownership
- You may need sudo: `sudo chown -R 1000:1000 /share/Container/filebrowser-config`
- Or run as admin user on NAS

