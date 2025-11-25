# Connecting to MinIO from Windows Computer

If you want to use MinIO Client (mc) from your Windows computer instead of Container Station terminal, follow these steps.

## Prerequisites

1. **Find your NAS IP address:**
   - Check your router's admin panel
   - Or check NAS network settings
   - Common format: `192.168.1.x` or `192.168.0.x`
   - **NOT** the container IP (172.29.4.2) - that's internal only!

2. **Verify MinIO is accessible:**
   - Try opening in browser: `https://YOUR-NAS-IP:9001`
   - If it works, you can proceed

## Step 1: Install MinIO Client on Windows

**Option A: Using Chocolatey (Recommended)**
```powershell
# Run PowerShell as Administrator
choco install minio-client
```

**Option B: Manual Download**
1. Download from: https://dl.min.io/client/mc/release/windows-amd64/mc.exe
2. Place `mc.exe` in a folder (e.g., `C:\Tools\minio\mc.exe`)
3. Add to PATH or use full path when running commands

**Option C: Using Scoop**
```powershell
scoop install minio-client
```

## Step 2: Configure MinIO Connection

### If SSL is Enabled (Recommended)

```powershell
# Replace YOUR-NAS-IP with your actual NAS IP address
# Replace minioadmin with your actual MINIO_ROOT_USER and MINIO_ROOT_PASSWORD
mc alias set myminio https://YOUR-NAS-IP:9000 minioadmin minioadmin
```

**Example:**
```powershell
mc alias set myminio https://192.168.1.100:9000 minioadmin minioadmin
```

### If Using Self-Signed Certificate

If you get certificate errors, add `--insecure` flag:

```powershell
mc alias set myminio https://YOUR-NAS-IP:9000 minioadmin minioadmin --insecure
```

### If SSL is NOT Enabled (Not Recommended)

```powershell
mc alias set myminio http://YOUR-NAS-IP:9000 minioadmin minioadmin
```

## Step 3: Create Access Key

```powershell
mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
```

You'll see output like:
```
Access Key: ABCDEFGHIJKLMNOP
Secret Key: abcdefghijklmnopqrstuvwxyz1234567890
```

**⚠️ IMPORTANT:** Copy both keys immediately!

## Common Errors and Solutions

### Error: "dial tcp: i/o timeout"

**Problem:** Can't connect to MinIO server

**Solutions:**
1. **Check IP address:**
   - Use your NAS's external IP (e.g., 192.168.1.100)
   - NOT the container IP (172.29.4.2)
   - NOT localhost (127.0.0.1)

2. **Check port:**
   - Make sure port 9000 is accessible
   - Check firewall settings on NAS
   - Verify port forwarding if accessing from outside network

3. **Test connection:**
   ```powershell
   # Test if port is open
   Test-NetConnection -ComputerName YOUR-NAS-IP -Port 9000
   ```

4. **Check MinIO is running:**
   - Verify in Container Station that container is running
   - Check logs: `docker logs minio`

### Error: "certificate verify failed"

**Problem:** SSL certificate issue

**Solutions:**
1. **Use `--insecure` flag** (for self-signed certificates):
   ```powershell
   mc alias set myminio https://YOUR-NAS-IP:9000 minioadmin minioadmin --insecure
   ```

2. **Or use HTTP** (if SSL not required):
   ```powershell
   mc alias set myminio http://YOUR-NAS-IP:9000 minioadmin minioadmin
   ```

### Error: "Unable to initialize new alias"

**Problem:** Connection or credential issue

**Solutions:**
1. **Verify credentials:**
   - Check `MINIO_ROOT_USER` in docker-compose file
   - Check `MINIO_ROOT_PASSWORD` in docker-compose file
   - Make sure you're using root credentials, not access keys

2. **Check protocol:**
   - If SSL is enabled, use `https://`
   - If SSL is disabled, use `http://`

3. **Verify endpoint:**
   - Format: `https://IP:PORT` or `http://IP:PORT`
   - Port 9000 for API, port 9001 for Console

### Error: "Access Denied"

**Problem:** Wrong credentials or insufficient permissions

**Solutions:**
1. **Use root credentials:**
   - Must use `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD`
   - Regular access keys don't have admin privileges

2. **Check credentials in docker-compose:**
   ```yaml
   MINIO_ROOT_USER: your-username
   MINIO_ROOT_PASSWORD: your-password
   ```

### Error: "The difference between the request time and the server's time is too large"

**Problem:** System clocks are not synchronized between your Windows computer and NAS

**Solutions:**

1. **Sync Windows Time (Recommended):**
   ```powershell
   # Run PowerShell as Administrator
   w32tm /resync
   ```
   
   Or manually:
   - Right-click on clock in taskbar
   - Select "Adjust date/time"
   - Click "Sync now" under "Synchronize your clock"

2. **Check Windows Time Service:**
   ```powershell
   # Check if time service is running
   Get-Service w32time
   
   # If not running, start it
   Start-Service w32time
   
   # Force sync
   w32tm /resync /force
   ```

3. **Sync NAS Time:**
   - Log into your NAS web interface
   - Go to System Settings → Date & Time
   - Enable NTP (Network Time Protocol) sync
   - Or manually set the correct time

4. **Verify Times Match:**
   ```powershell
   # Check Windows time
   Get-Date
   
   # Check NAS time (via SSH or web interface)
   # They should be within 15 minutes of each other
   ```

5. **Quick Fix - Set Time Manually:**
   - If automatic sync doesn't work, manually set both clocks to the same time
   - Make sure timezone is correct on both systems

**Note:** This error occurs because MinIO uses time-based authentication (similar to AWS S3). The clocks must be synchronized for security reasons.

## Finding Your NAS IP Address

### Method 1: Check Router Admin Panel
1. Log into your router (usually `192.168.1.1` or `192.168.0.1`)
2. Look for "Connected Devices" or "DHCP Clients"
3. Find your NAS device

### Method 2: Check NAS Network Settings
1. Log into your NAS web interface
2. Go to Network Settings
3. Check IP address

### Method 3: Use Ping (if you know NAS hostname)
```powershell
ping your-nas-hostname
```

## Testing the Connection

After setting up the alias, test it:

```powershell
# List buckets
mc ls myminio

# Check connection
mc admin info myminio
```

## Quick Reference

**Correct format:**
```powershell
mc alias set ALIAS_NAME PROTOCOL://NAS-IP:PORT USERNAME PASSWORD
```

**Examples:**
```powershell
# With SSL
mc alias set myminio https://192.168.1.100:9000 minioadmin minioadmin

# With SSL (self-signed cert)
mc alias set myminio https://192.168.1.100:9000 minioadmin minioadmin --insecure

# Without SSL (not recommended)
mc alias set myminio http://192.168.1.100:9000 minioadmin minioadmin
```

## Recommendation

**For easiest setup, use Container Station terminal** (as shown in QUICK_ACCESS_KEY_SETUP.md) since:
- No network/firewall issues
- No certificate problems
- Already inside the NAS network
- Works immediately

Use Windows mc client only if you need to manage MinIO from your computer regularly.

