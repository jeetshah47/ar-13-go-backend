# FileBrowser Service - Systemd Service Setup

This guide shows how to set up the filebrowser service as a systemd daemon on your NAS.

## Prerequisites

- Filebrowser binary is already built and uploaded to `/opt/filebrowser-service/filebrowser`
- Binary is executable: `chmod +x /opt/filebrowser-service/filebrowser`
- NAS has systemd (most Linux-based NAS systems do)

## Step 1: Create Systemd Service File

SSH into your NAS and run:

```bash
sudo nano /etc/systemd/system/filebrowser.service
```

Or use this one-liner to create it directly:

```bash
sudo tee /etc/systemd/system/filebrowser.service > /dev/null << 'EOF'
[Unit]
Description=File Browser Service
Documentation=https://github.com/ar-13-go-backend
After=network.target

[Service]
Type=simple
User=Administrator
Group=administrators
WorkingDirectory=/opt/filebrowser-service
Environment="DATA_ROOT=/share/studio work"
Environment="PORT=8082"
Environment="TZ=Asia/Kolkata"
ExecStart=/opt/filebrowser-service/filebrowser
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
```

## Step 2: Verify Service File

Check that the file was created correctly:

```bash
cat /etc/systemd/system/filebrowser.service
```

## Step 3: Reload Systemd

Reload systemd to recognize the new service:

```bash
sudo systemctl daemon-reload
```

## Step 4: Enable Service (Start on Boot)

Enable the service to start automatically on boot:

```bash
sudo systemctl enable filebrowser.service
```

## Step 5: Start the Service

Start the service:

```bash
sudo systemctl start filebrowser.service
```

## Step 6: Check Service Status

Check if the service is running:

```bash
sudo systemctl status filebrowser.service
```

You should see output like:
```
● filebrowser.service - File Browser Service
   Loaded: loaded (/etc/systemd/system/filebrowser.service; enabled; vendor preset: disabled)
   Active: active (running) since ...
```

## Step 7: Verify Service is Working

Test the health endpoint:

```bash
curl http://localhost:8082/health
```

Or test from another machine:

```bash
curl http://your-nas-ip:8082/health
```

## Useful Commands

### View Service Logs

```bash
# View all logs
sudo journalctl -u filebrowser.service

# Follow logs in real-time
sudo journalctl -u filebrowser.service -f

# View last 50 lines
sudo journalctl -u filebrowser.service -n 50

# View logs since today
sudo journalctl -u filebrowser.service --since today
```

### Service Management

```bash
# Start service
sudo systemctl start filebrowser.service

# Stop service
sudo systemctl stop filebrowser.service

# Restart service
sudo systemctl restart filebrowser.service

# Reload service (if supported)
sudo systemctl reload filebrowser.service

# Check status
sudo systemctl status filebrowser.service

# Disable auto-start on boot
sudo systemctl disable filebrowser.service

# Enable auto-start on boot
sudo systemctl enable filebrowser.service
```

### Check if Service is Running

```bash
# Check process
ps aux | grep filebrowser

# Check if port is listening
netstat -tlnp | grep 8082
# or
ss -tlnp | grep 8082
```

## Customization

### Change User/Group

If you need to run as a different user, first check available users:

```bash
id
# or
cat /etc/passwd | grep -E "(admin|Administrator|qnap)"
```

Then edit the service file:

```bash
sudo nano /etc/systemd/system/filebrowser.service
```

Change these lines:
```ini
User=your-username
Group=your-group
```

After editing, reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart filebrowser.service
```

### Change Data Root Directory

Edit the service file and change:

```ini
Environment="DATA_ROOT=/path/to/your/data"
```

Then reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart filebrowser.service
```

### Change Port

Edit the service file and change:

```ini
Environment="PORT=8083"
```

Then reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart filebrowser.service
```

**Note:** If you change the port, also update your firewall rules if needed.

## Troubleshooting

### Service Won't Start

1. Check service status:
   ```bash
   sudo systemctl status filebrowser.service
   ```

2. Check logs:
   ```bash
   sudo journalctl -u filebrowser.service -n 50
   ```

3. Verify binary exists and is executable:
   ```bash
   ls -la /opt/filebrowser-service/filebrowser
   chmod +x /opt/filebrowser-service/filebrowser
   ```

4. Test running manually:
   ```bash
   cd /opt/filebrowser-service
   export DATA_ROOT="/share/studio work"
   export PORT=8082
   ./filebrowser
   ```

### Permission Denied Errors

1. Check file permissions:
   ```bash
   ls -la /opt/filebrowser-service/
   ```

2. Ensure user has access to DATA_ROOT:
   ```bash
   ls -la "/share/studio work"
   ```

3. Fix permissions if needed:
   ```bash
   sudo chown -R Administrator:administrators /opt/filebrowser-service
   sudo chmod +x /opt/filebrowser-service/filebrowser
   ```

### Port Already in Use

If port 8082 is already in use:

1. Find what's using it:
   ```bash
   sudo lsof -i :8082
   # or
   sudo netstat -tlnp | grep 8082
   ```

2. Either stop the conflicting service or change the port in the service file.

### Service Keeps Restarting

If the service keeps restarting (check with `systemctl status`):

1. Check logs for errors:
   ```bash
   sudo journalctl -u filebrowser.service -n 100
   ```

2. Verify DATA_ROOT directory exists:
   ```bash
   ls -la "/share/studio work"
   ```

3. Check if the binary is the correct architecture:
   ```bash
   file /opt/filebrowser-service/filebrowser
   # Should show: ELF 64-bit LSB executable, ARM aarch64
   ```

## Complete Setup Script

Here's a complete script to set up everything:

```bash
#!/bin/bash

# Create directory if it doesn't exist
sudo mkdir -p /opt/filebrowser-service

# Set permissions (adjust user/group as needed)
sudo chown Administrator:administrators /opt/filebrowser-service
sudo chmod 755 /opt/filebrowser-service

# Create systemd service file
sudo tee /etc/systemd/system/filebrowser.service > /dev/null << 'EOF'
[Unit]
Description=File Browser Service
Documentation=https://github.com/ar-13-go-backend
After=network.target

[Service]
Type=simple
User=Administrator
Group=administrators
WorkingDirectory=/opt/filebrowser-service
Environment="DATA_ROOT=/share/studio work"
Environment="PORT=8082"
Environment="TZ=Asia/Kolkata"
ExecStart=/opt/filebrowser-service/filebrowser
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable filebrowser.service

# Start service
sudo systemctl start filebrowser.service

# Check status
sudo systemctl status filebrowser.service

echo "Setup complete! Service should be running."
echo "Check logs with: sudo journalctl -u filebrowser.service -f"
```

Save this as `setup-filebrowser-service.sh`, make it executable, and run:

```bash
chmod +x setup-filebrowser-service.sh
./setup-filebrowser-service.sh
```

## Verification Checklist

- [ ] Service file created at `/etc/systemd/system/filebrowser.service`
- [ ] Service enabled: `systemctl is-enabled filebrowser.service` shows "enabled"
- [ ] Service running: `systemctl status filebrowser.service` shows "active (running)"
- [ ] Health check works: `curl http://localhost:8082/health` returns JSON
- [ ] Service starts on boot (reboot and verify)
- [ ] Logs are accessible: `journalctl -u filebrowser.service` shows logs

## Next Steps

After setting up the systemd service:

1. Update your backend `.env` file:
   ```
   FILEBROWSER_SERVICE_URL=http://your-nas-ip:8082
   ```

2. Update your frontend `.env` file:
   ```
   VITE_FILEBROWSER_BASE_URL=http://your-nas-ip:8082
   ```

3. Test the integration from your frontend application.

