# Troubleshooting MinIO Client (mc) on Windows

## Error: "The process cannot access the file because it is being used by another process"

This error means `mc.exe` is locked by another process. Here's how to fix it:

### Solution 1: Check for Running Processes

1. **Check if mc.exe is already running:**
   ```powershell
   Get-Process | Where-Object {$_.ProcessName -like "*mc*"}
   ```

2. **If found, kill the process:**
   ```powershell
   Stop-Process -Name "mc" -Force
   ```

3. **Or use Task Manager:**
   - Press `Ctrl + Shift + Esc`
   - Look for `mc.exe` or `mc` in the processes
   - Right-click → End Task

### Solution 2: Check File Lock

1. **Check what's locking the file:**
   ```powershell
   # Download Handle.exe from Sysinternals if needed
   # Or use Process Explorer
   ```

2. **Or simply:**
   - Close all PowerShell/Command Prompt windows
   - Restart your computer (if needed)
   - Try again

### Solution 3: Download Fresh Copy

1. **Delete the existing mc.exe:**
   ```powershell
   Remove-Item "C:\path\to\mc.exe" -Force
   ```

2. **Download fresh copy:**
   - Go to: https://dl.min.io/client/mc/release/windows-amd64/mc.exe
   - Save to a different location (e.g., `C:\Tools\mc.exe`)

3. **Try running from new location:**
   ```powershell
   C:\Tools\mc.exe --version
   ```

### Solution 4: Use Different Location

Instead of installing globally, use mc from a specific folder:

1. **Create a folder:**
   ```powershell
   New-Item -ItemType Directory -Path "C:\Tools\minio" -Force
   ```

2. **Download mc.exe to that folder:**
   - Download: https://dl.min.io/client/mc/release/windows-amd64/mc.exe
   - Save as: `C:\Tools\minio\mc.exe`

3. **Use full path:**
   ```powershell
   C:\Tools\minio\mc.exe --version
   C:\Tools\minio\mc.exe alias set myminio http://your-nas-ip:32774 ar-web-backend bIF62kxCm54D
   ```

### Solution 5: Use Docker Instead

If mc.exe keeps having issues, use Docker:

```powershell
# Run mc in a Docker container
docker run --rm -it minio/mc --version

# Configure alias
docker run --rm -it minio/mc alias set myminio http://your-nas-ip:32774 ar-web-backend bIF62kxCm54D

# Create user
docker run --rm -it minio/mc admin user add myminio ar-13-backend-user

# Create access key
docker run --rm -it minio/mc admin user svcacct add myminio ar-13-backend-user --name ar-13-backend-key
```

### Solution 6: Check Antivirus

Sometimes antivirus software locks executable files:

1. **Temporarily disable antivirus** (for testing)
2. **Or add exception** for mc.exe in your antivirus settings
3. **Or download to a trusted folder** (like `C:\Tools\`)

### Quick Test: Check if File is Actually Locked

```powershell
# Try to rename the file (if it's locked, this will fail)
Rename-Item "C:\path\to\mc.exe" "mc_backup.exe"

# If rename works, rename it back
Rename-Item "C:\path\to\mc_backup.exe" "mc.exe"
```

