# FileBrowser Quick Start

## Windows Machine Setup

### Step 1: Start FileBrowser
```powershell
cd docs\nas\filebrowser
docker-compose -f docker-compose.filebrowser.windows.yml up -d
```

### Step 2: Get Password from Logs
```powershell
docker logs filebrowser
```
Look for: `User 'admin' initialized with randomly generated password: <PASSWORD>`

### Step 3: Access Web UI
- Open browser: **http://localhost:8081**
- Login with:
  - Username: `admin`
  - Password: `<PASSWORD_FROM_LOGS>` (or `admin` if you set FB_DEFAULT_PASS)
- **⚠️ Change password immediately!**

### Step 3: Browse Files
- All files from `D:/Resumers/` will be visible
- Navigate folders, download files, upload new files
- Large files (GBs) will stream/download without issues

### Stop FileBrowser
```powershell
docker-compose -f docker-compose.filebrowser.windows.yml down
```

---

## NAS Setup (Container Station)

### Step 1: Folder Path
The configuration is already set to use `/share/studio work/` folder.
If you need a different folder, edit `docker-compose.filebrowser.nas.yml`:
```yaml
volumes:
  - "/share/studio work/:/srv"  # Default: studio work folder
```

### Step 2: Deploy
**Via Container Station:**
1. Upload `docker-compose.filebrowser.nas.yml`
2. Create stack from compose file
3. Start the stack

**Via SSH:**
```bash
cd /share/Container/filebrowser
docker-compose -f docker-compose.filebrowser.nas.yml up -d
```

### Step 3: Get Password from Logs
```bash
docker logs filebrowser
```
Look for: `User 'admin' initialized with randomly generated password: <PASSWORD>`

### Step 4: Access Web UI
- Open browser: **http://YOUR-NAS-IP:8081**
- Login with:
  - Username: `admin`
  - Password: `<PASSWORD_FROM_LOGS>` (or `admin` if you set FB_DEFAULT_PASS)
- **⚠️ Change password immediately!**

### Step 5: Browse Files
- All files from `/share/studio work/` folder will be visible
- Navigate, download, upload files as needed

---

## First-Time Setup Checklist

- [ ] Started FileBrowser container
- [ ] Accessed web UI (http://localhost:8081 or http://NAS-IP:8081)
- [ ] Logged in with default credentials
- [ ] Changed admin password
- [ ] Verified files are visible
- [ ] Created user accounts (if needed)
- [ ] Configured permissions

---

## Common NAS Folder Paths

- `/share/Public/` - Public shared folder
- `/share/Container/` - Container Station folder
- `/share/Download/` - Download folder
- `/share/Multimedia/` - Media files
- `/share/homes/` - User home directories

Create custom folder via File Station if needed.

---

## Troubleshooting

**Can't access web UI?**
- Check if container is running: `docker ps`
- Verify port 8081 is not in use
- Check logs: `docker logs filebrowser`

**Files not showing?**
- Verify volume mount path is correct
- Check folder permissions
- Restart container after path changes

**Permission errors?**
- Windows: Ensure Docker has access to folder
- NAS: Set permissions: `chmod -R 755 /share/YourFolder/`

