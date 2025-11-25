# Fix FileBrowser Permission Issues

If you see errors like `Permission denied` when FileBrowser tries to create config files, follow these steps:

## For NAS (Container Station)

### Step 1: Create Config Directory
```bash
mkdir -p /share/Container/filebrowser-config
```

### Step 2: Set Permissions
```bash
# Set directory permissions
chmod -R 755 /share/Container/filebrowser-config

# Set ownership (FileBrowser runs as UID 1000)
chown -R 1000:1000 /share/Container/filebrowser-config
```

### Step 3: Verify Permissions
```bash
ls -la /share/Container/filebrowser-config
```

You should see:
```
drwxr-xr-x 2 1000 1000 4096 Nov 22 10:00 filebrowser-config
```

### Step 4: Restart Container
```bash
docker-compose -f docker-compose.filebrowser.nas.yml down
docker-compose -f docker-compose.filebrowser.nas.yml up -d
```

## For Windows

### Option 1: Let Docker Create the Folder (Recommended)
The `./filebrowser-config` folder will be created automatically. If you get permission errors:

1. **Stop the container:**
   ```powershell
   docker-compose -f docker-compose.filebrowser.windows.yml down
   ```

2. **Delete the config folder** (if it exists):
   ```powershell
   Remove-Item -Recurse -Force .\filebrowser-config -ErrorAction SilentlyContinue
   ```

3. **Create folder with proper permissions:**
   ```powershell
   New-Item -ItemType Directory -Path .\filebrowser-config
   ```

4. **Start container:**
   ```powershell
   docker-compose -f docker-compose.filebrowser.windows.yml up -d
   ```

### Option 2: Use Named Volume (Alternative)
If permission issues persist, use a Docker named volume instead:

Edit `docker-compose.filebrowser.windows.yml`:
```yaml
volumes:
  - "D:/Resumers/:/srv"
  - "filebrowser-config:/config"  # Use named volume instead

# Add volumes section at the end
volumes:
  filebrowser-config:
```

## Troubleshooting

### Check Container Logs
```bash
docker logs filebrowser
```

### Check File Permissions Inside Container
```bash
docker exec -it filebrowser ls -la /filebrowser-config
```

### Fix Permissions from Inside Container
If you can't fix permissions from host:
```bash
docker exec -it filebrowser sh
chmod -R 755 /filebrowser-config
chown -R 1000:1000 /filebrowser-config
exit
```

### Alternative: Run as Root (Not Recommended)
If nothing else works, you can temporarily run as root to test:

Edit docker-compose file and remove or comment out:
```yaml
# user: "1000:1000"  # Comment this out
```

**Warning:** Running as root is not recommended for production.

## Common Issues

### Issue: "Permission denied" on config directory
**Solution:** Ensure the directory exists and has correct permissions (755, owned by 1000:1000)

### Issue: "Permission denied" on data directory
**Solution:** Check permissions on `/share/studio work/` (NAS) or `D:/Resumers/` (Windows)

### Issue: Container keeps restarting
**Solution:** Check logs for permission errors and fix directory permissions

## Verification

After fixing permissions, verify FileBrowser is working:

1. **Check container status:**
   ```bash
   docker ps | grep filebrowser
   ```

2. **Check logs:**
   ```bash
   docker logs filebrowser
   ```

3. **Access web UI:**
   - http://localhost:8081 (Windows)
   - http://YOUR-NAS-IP:8081 (NAS)

If you can access the web UI and login, permissions are fixed correctly!

