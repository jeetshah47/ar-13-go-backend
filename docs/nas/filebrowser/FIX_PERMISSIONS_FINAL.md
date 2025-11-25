# Final Fix for FileBrowser Permission Issues

If you're still having permission issues, use this approach with a **named Docker volume** instead of a bind mount.

## Solution: Use Named Volume (Recommended)

I've updated `docker-compose.filebrowser.nas.yml` to use a **named volume** instead of a bind mount. This completely avoids permission issues.

### What Changed:

- **Before:** `/share/Container/filebrowser-config:/filebrowser-config` (bind mount - permission issues)
- **After:** `filebrowser-config:/filebrowser-config` (named volume - no permission issues)

### Steps:

1. **Stop any running container:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml down
   ```

2. **Remove old volume (if exists):**
   ```bash
   docker volume rm filebrowser-config 2>/dev/null || true
   ```

3. **Start FileBrowser:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml up -d
   ```

4. **Check logs:**
   ```bash
   docker logs filebrowser
   ```

   Should see: `Using database: /filebrowser-config/database.db` (no errors)

5. **Access web UI:**
   - http://YOUR-NAS-IP:8081
   - Login with: `admin` / `admin` (or check logs for generated password)

## Alternative: Fix Bind Mount Permissions

If you prefer to keep config in `/share/Container/filebrowser-config`:

### Option 1: Run as Root First (Then Fix)

1. **Keep `user:` commented out** in docker-compose
2. **Start container:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml up -d
   ```

3. **Fix permissions from inside container:**
   ```bash
   docker exec -it filebrowser sh
   chmod -R 777 /filebrowser-config
   ls -la /filebrowser-config
   exit
   ```

4. **Verify it's working:**
   ```bash
   docker logs filebrowser
   ```

5. **Then change ownership on host:**
   ```bash
   chown -R 1000:1000 /share/Container/filebrowser-config
   chmod -R 755 /share/Container/filebrowser-config
   ```

6. **Uncomment user line** and restart:
   ```yaml
   user: "1000:1000"
   ```
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml restart
   ```

### Option 2: Use Different Path

Try using a path that Docker/Container Station can write to:

```yaml
volumes:
  - "/share/Container/.filebrowser:/filebrowser-config"  # Hidden folder
```

Or:

```yaml
volumes:
  - "/share/CACHEDEV1_DATA/.qpkg/ContainerStation/filebrowser:/filebrowser-config"
```

## Verify It's Working

```bash
# Check container status
docker ps | grep filebrowser

# Check logs (should have no permission errors)
docker logs filebrowser

# Check if database was created
docker exec filebrowser ls -la /filebrowser-config/

# Access web UI
# http://YOUR-NAS-IP:8081
```

## Named Volume Location

If using named volume, config is stored in:
- Docker volumes directory (usually `/var/lib/docker/volumes/filebrowser-config/`)
- To backup: `docker run --rm -v filebrowser-config:/data -v $(pwd):/backup alpine tar czf /backup/filebrowser-config-backup.tar.gz -C /data .`

## Troubleshooting

### Still Getting Permission Errors?

1. **Check container user:**
   ```bash
   docker exec filebrowser id
   ```

2. **Check directory permissions inside container:**
   ```bash
   docker exec filebrowser ls -la /filebrowser-config
   ```

3. **Try running without user restriction:**
   - Comment out `user:` line
   - Start container
   - Fix permissions from inside
   - Then add user restriction back

### Container Won't Start?

```bash
# Check detailed logs
docker logs filebrowser 2>&1 | tail -50

# Check if port is in use
netstat -tuln | grep 8081

# Try different port
# Edit docker-compose: "8082:80"
```

## Recommended Approach

**Use the named volume** (already configured in docker-compose). It's the simplest and most reliable solution.

