# Final Fix for Permission Denied Error

If you're still getting `permission denied` errors even with named volumes, follow these steps:

## Quick Fix (Recommended)

Run this command sequence in your ttyd terminal:

```bash
# 1. Stop container
docker-compose -f docker-compose.filebrowser.nas.yml down

# 2. Remove the problematic volume
docker volume rm filebrowser-config

# 3. Start container (it will create new volume)
docker-compose -f docker-compose.filebrowser.nas.yml up -d

# 4. Wait a few seconds for initialization
sleep 5

# 5. Fix permissions inside the container
docker exec filebrowser chown -R 1000:1000 /filebrowser-config
docker exec filebrowser chmod -R 755 /filebrowser-config

# 6. Check if it's working
docker logs filebrowser
```

## Alternative: Use the Fix Script

```bash
chmod +x fix-volume-permissions.sh
./fix-volume-permissions.sh
```

## Why This Happens

FileBrowser runs as UID 1000 by default, but when Docker creates a named volume, it might be owned by root. By running the container initially without user restriction, it creates files as root, then we fix the ownership.

## After Fixing

1. **Check logs:**
   ```bash
   docker logs filebrowser
   ```
   Should see: `Using database: /filebrowser-config/database.db` (no errors)

2. **Access FileBrowser:**
   - http://YOUR-NAS-IP:8081
   - Username: `admin`
   - Password: Check logs or use `admin` (if using default)

3. **Optional - Add User Restriction (for security):**
   After confirming it works, you can uncomment the user line in docker-compose:
   ```yaml
   user: "1000:1000"
   ```
   Then restart: `docker-compose restart`

## If Still Not Working

Try running completely as root (temporary):

1. **Ensure user line is commented out** in docker-compose
2. **Start container:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml up -d
   ```
3. **It should work now** - FileBrowser will run as root and can write to the volume

**Note:** Running as root is less secure but will work. You can fix permissions later and add user restriction.

