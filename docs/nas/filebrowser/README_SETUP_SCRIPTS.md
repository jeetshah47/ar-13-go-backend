# FileBrowser Setup Scripts

Automated setup scripts for FileBrowser deployment.

## Available Scripts

### For NAS (Linux/QNAP)
- **`setup-filebrowser.sh`** - Bash script for NAS deployment

### For Windows
- **`setup-filebrowser.ps1`** - PowerShell script for Windows deployment

## Usage

### NAS Setup (using ttyd or SSH)

1. **Navigate to the filebrowser directory:**
   ```bash
   cd /path/to/docs/nas/filebrowser
   ```

2. **Make script executable:**
   ```bash
   chmod +x setup-filebrowser.sh
   ```

3. **Run the script:**
   ```bash
   ./setup-filebrowser.sh
   ```

   Or with sudo if needed:
   ```bash
   sudo ./setup-filebrowser.sh
   ```

### Windows Setup

1. **Open PowerShell** in the filebrowser directory:
   ```powershell
   cd docs\nas\filebrowser
   ```

2. **Run the script:**
   ```powershell
   .\setup-filebrowser.ps1
   ```

   If you get execution policy error:
   ```powershell
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   .\setup-filebrowser.ps1
   ```

## What the Scripts Do

### Automated Steps:

1. ✅ **Stop existing container** (if running)
2. ✅ **Setup configuration directory** with proper permissions
3. ✅ **Clean up old volumes** (if using named volume)
4. ✅ **Start FileBrowser container**
5. ✅ **Wait for container to be ready**
6. ✅ **Check logs** for errors
7. ✅ **Verify setup** (container running, database created, port listening)
8. ✅ **Display admin credentials** (username and password)

### Features:

- **Colored output** for better readability
- **Error handling** - stops on errors
- **Verification** - checks if everything is working
- **Password retrieval** - automatically gets admin password from logs
- **Cleanup** - removes old volumes/containers before setup

## Script Configuration

### NAS Script (`setup-filebrowser.sh`)

Edit these variables at the top of the script:

```bash
COMPOSE_FILE="docker-compose.filebrowser.nas.yml"
CONTAINER_NAME="filebrowser"
CONFIG_DIR="/share/Container/filebrowser-config"
USE_NAMED_VOLUME=true  # Set to false to use bind mount
```

### Windows Script (`setup-filebrowser.ps1`)

Edit these variables at the top of the script:

```powershell
$ComposeFile = "docker-compose.filebrowser.windows.yml"
$ContainerName = "filebrowser"
$ConfigDir = ".\filebrowser-config"
```

## Manual Setup (Alternative)

If you prefer manual setup, see:
- `QUICK_START.md` - Quick start guide
- `FIX_PERMISSIONS_FINAL.md` - Permission troubleshooting
- `SETUP_VIA_FILE_STATION.md` - Using File Station GUI

## Troubleshooting

### Script Fails to Start Container

```bash
# Check Docker is running
docker ps

# Check compose file exists
ls -la docker-compose.filebrowser.nas.yml

# Run with verbose output
bash -x setup-filebrowser.sh
```

### Permission Denied (NAS)

```bash
# Run with sudo
sudo ./setup-filebrowser.sh

# Or fix permissions first
chmod +x setup-filebrowser.sh
```

### PowerShell Execution Policy (Windows)

```powershell
# Allow script execution
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Or run with bypass
powershell -ExecutionPolicy Bypass -File .\setup-filebrowser.ps1
```

## After Running Script

### Access FileBrowser:

- **NAS:** http://YOUR-NAS-IP:8081
- **Windows:** http://localhost:8081

### Default Credentials:

- **Username:** `admin`
- **Password:** Check script output (or `admin` if using default)

### Useful Commands:

```bash
# View logs
docker logs filebrowser

# Stop FileBrowser
docker-compose -f docker-compose.filebrowser.nas.yml down

# Restart FileBrowser
docker-compose -f docker-compose.filebrowser.nas.yml restart

# Check container status
docker ps | grep filebrowser
```

## Script Output Example

```
==========================================
FileBrowser Setup Script for NAS
==========================================

[INFO] Starting FileBrowser setup...

[INFO] Stopping existing FileBrowser container...
[INFO] Container stopped.
[INFO] Using named volume - skipping bind mount setup.
[INFO] Starting FileBrowser container...
[INFO] Container started successfully!
[INFO] Waiting for container to be ready...
[INFO] Container is running.
[INFO] Checking container logs...

[INFO] Verifying setup...
✓ Container is running
✓ Database file created successfully
✓ Port 8081 is listening

==========================================
FileBrowser Admin Credentials:
==========================================
Username: admin
Password: xyz123abc456

Access FileBrowser at: http://YOUR-NAS-IP:8081
==========================================

[INFO] Setup complete!
```

## Next Steps

1. ✅ Run the setup script
2. ✅ Access FileBrowser web UI
3. ✅ Login with provided credentials
4. ✅ Change admin password (recommended)
5. ✅ Start browsing your files!

