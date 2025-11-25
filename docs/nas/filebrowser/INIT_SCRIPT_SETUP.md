# FileBrowser Service - Init Script Setup (Non-systemd)

This guide shows how to set up the filebrowser service as a daemon on NAS systems that don't use systemd (like QNAP).

## Check Your Init System

First, check what init system your NAS uses:

```bash
# Check for systemd
ls /etc/systemd/system/

# Check for init.d (QNAP, older Linux)
ls /etc/init.d/

# Check for rc.local
ls /etc/rc.local

# Check process manager
ps aux | head -1
```

## Option 1: Init.d Script (QNAP and most NAS systems)

### Step 1: Create Init Script

```bash
sudo nano /etc/init.d/filebrowser
```

Or create it directly:

```bash
sudo tee /etc/init.d/filebrowser > /dev/null << 'EOF'
#!/bin/sh
### BEGIN INIT INFO
# Provides:          filebrowser
# Required-Start:     $network
# Required-Stop:     
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: File Browser Service
# Description:       File Browser Service for NAS
### END INIT INFO

DAEMON="/share/Public/filebrowser-service/filebrowser"
DAEMON_USER="Administrator"
DAEMON_DIR="/share/Public/filebrowser-service"
PIDFILE="/var/run/filebrowser.pid"
NAME="filebrowser"

# Environment variables
export DATA_ROOT="/share/studio work"
export PORT=8082
export TZ=Asia/Kolkata

. /lib/lsb/init-functions

case "$1" in
  start)
    log_daemon_msg "Starting $NAME"
    if [ -f "$PIDFILE" ] && kill -0 $(cat "$PIDFILE") 2>/dev/null; then
        log_end_msg 1
        exit 1
    fi
    
    cd "$DAEMON_DIR"
    start-stop-daemon --start --quiet --background \
        --make-pidfile --pidfile "$PIDFILE" \
        --chuid "$DAEMON_USER" \
        --exec "$DAEMON" || log_end_msg 1
    log_end_msg 0
    ;;
  stop)
    log_daemon_msg "Stopping $NAME"
    if [ ! -f "$PIDFILE" ]; then
        log_end_msg 1
        exit 1
    fi
    
    start-stop-daemon --stop --quiet --pidfile "$PIDFILE" \
        --retry=TERM/30/KILL/5 || log_end_msg 1
    rm -f "$PIDFILE"
    log_end_msg 0
    ;;
  restart)
    $0 stop
    $0 start
    ;;
  status)
    if [ -f "$PIDFILE" ] && kill -0 $(cat "$PIDFILE") 2>/dev/null; then
        echo "$NAME is running (PID: $(cat $PIDFILE))"
        exit 0
    else
        echo "$NAME is not running"
        exit 1
    fi
    ;;
  *)
    echo "Usage: $0 {start|stop|restart|status}"
    exit 1
    ;;
esac

exit 0
EOF
```

### Step 2: Make Script Executable

```bash
sudo chmod +x /etc/init.d/filebrowser
```

### Step 3: Enable on Boot

```bash
# For QNAP and systems with update-rc.d
sudo update-rc.d filebrowser defaults

# Or manually create symlinks
sudo ln -s /etc/init.d/filebrowser /etc/rc2.d/S99filebrowser
sudo ln -s /etc/init.d/filebrowser /etc/rc3.d/S99filebrowser
sudo ln -s /etc/init.d/filebrowser /etc/rc4.d/S99filebrowser
sudo ln -s /etc/init.d/filebrowser /etc/rc5.d/S99filebrowser
```

### Step 4: Start the Service

```bash
sudo /etc/init.d/filebrowser start
# or
sudo service filebrowser start
```

### Step 5: Check Status

```bash
sudo /etc/init.d/filebrowser status
# or
sudo service filebrowser status
```

## Option 2: Simple Init Script (If start-stop-daemon not available)

If your system doesn't have `start-stop-daemon`, use this simpler version:

```bash
sudo tee /etc/init.d/filebrowser > /dev/null << 'EOF'
#!/bin/sh

DAEMON="/share/Public/filebrowser-service/filebrowser"
DAEMON_DIR="/share/Public/filebrowser-service"
PIDFILE="/var/run/filebrowser.pid"
LOGFILE="/var/log/filebrowser.log"

export DATA_ROOT="/share/studio work"
export PORT=8082
export TZ=Asia/Kolkata

case "$1" in
  start)
    if [ -f "$PIDFILE" ] && kill -0 $(cat "$PIDFILE") 2>/dev/null; then
        echo "Filebrowser is already running (PID: $(cat $PIDFILE))"
        exit 1
    fi
    
    echo "Starting filebrowser..."
    cd "$DAEMON_DIR"
    nohup "$DAEMON" > "$LOGFILE" 2>&1 &
    echo $! > "$PIDFILE"
    echo "Filebrowser started (PID: $(cat $PIDFILE))"
    ;;
  stop)
    if [ ! -f "$PIDFILE" ]; then
        echo "Filebrowser is not running"
        exit 1
    fi
    
    PID=$(cat "$PIDFILE")
    if kill -0 "$PID" 2>/dev/null; then
        echo "Stopping filebrowser (PID: $PID)..."
        kill "$PID"
        rm -f "$PIDFILE"
        echo "Filebrowser stopped"
    else
        echo "Filebrowser process not found"
        rm -f "$PIDFILE"
    fi
    ;;
  restart)
    $0 stop
    sleep 2
    $0 start
    ;;
  status)
    if [ -f "$PIDFILE" ] && kill -0 $(cat "$PIDFILE") 2>/dev/null; then
        echo "Filebrowser is running (PID: $(cat $PIDFILE))"
        exit 0
    else
        echo "Filebrowser is not running"
        exit 1
    fi
    ;;
  *)
    echo "Usage: $0 {start|stop|restart|status}"
    exit 1
    ;;
esac

exit 0
EOF

sudo chmod +x /etc/init.d/filebrowser
```

## Option 3: rc.local (Simple but less robust)

If your system has `/etc/rc.local`:

```bash
sudo nano /etc/rc.local
```

Add before `exit 0`:

```bash
# Start filebrowser service
export DATA_ROOT="/share/studio work"
export PORT=8082
export TZ=Asia/Kolkata
cd /share/Public/filebrowser-service
nohup ./filebrowser > /var/log/filebrowser.log 2>&1 &
echo $! > /var/run/filebrowser.pid
```

Make sure `/etc/rc.local` is executable:

```bash
sudo chmod +x /etc/rc.local
```

## Option 4: Cron @reboot (Universal fallback)

This works on almost all systems:

```bash
sudo crontab -e
```

Add this line:

```
@reboot sleep 30 && cd /share/Public/filebrowser-service && export DATA_ROOT="/share/studio work" && export PORT=8082 && export TZ=Asia/Kolkata && nohup ./filebrowser > /var/log/filebrowser.log 2>&1 &
```

Or create a startup script and call it from cron:

```bash
# Create startup script
sudo tee /share/Public/filebrowser-service/start.sh > /dev/null << 'EOF'
#!/bin/bash
export DATA_ROOT="/share/studio work"
export PORT=8082
export TZ=Asia/Kolkata
cd /share/Public/filebrowser-service
./filebrowser > /var/log/filebrowser.log 2>&1
EOF

sudo chmod +x /share/Public/filebrowser-service/start.sh

# Add to crontab
sudo crontab -e
# Add: @reboot /share/Public/filebrowser-service/start.sh
```

## Option 5: Supervisor (If available)

If your NAS has supervisor installed:

```bash
sudo tee /etc/supervisor/conf.d/filebrowser.conf > /dev/null << 'EOF'
[program:filebrowser]
command=/share/Public/filebrowser-service/filebrowser
directory=/share/Public/filebrowser-service
user=Administrator
autostart=true
autorestart=true
stderr_logfile=/var/log/filebrowser.err.log
stdout_logfile=/var/log/filebrowser.out.log
environment=DATA_ROOT="/share/studio work",PORT="8082",TZ="Asia/Kolkata"
EOF

sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl start filebrowser
```

## Option 6: QNAP-Specific (QTS)

For QNAP QTS systems, you can use the built-in service manager:

### Create QNAP Service Script

```bash
sudo tee /etc/init.d/filebrowser > /dev/null << 'EOF'
#!/bin/sh
# QNAP Init Script for FileBrowser Service

DAEMON="/share/Public/filebrowser-service/filebrowser"
PIDFILE="/var/run/filebrowser.pid"
LOGFILE="/share/Public/filebrowser.log"

export DATA_ROOT="/share/studio work"
export PORT=8082
export TZ=Asia/Kolkata

start() {
    if [ -f "$PIDFILE" ] && kill -0 $(cat "$PIDFILE") 2>/dev/null; then
        echo "Filebrowser is already running"
        return 1
    fi
    
    echo "Starting filebrowser..."
    cd /share/Public/filebrowser-service
    nohup "$DAEMON" >> "$LOGFILE" 2>&1 &
    echo $! > "$PIDFILE"
    echo "Filebrowser started (PID: $(cat $PIDFILE))"
}

stop() {
    if [ ! -f "$PIDFILE" ]; then
        echo "Filebrowser is not running"
        return 1
    fi
    
    PID=$(cat "$PIDFILE")
    if kill -0 "$PID" 2>/dev/null; then
        echo "Stopping filebrowser (PID: $PID)..."
        kill "$PID"
        rm -f "$PIDFILE"
        echo "Filebrowser stopped"
    else
        echo "Filebrowser process not found"
        rm -f "$PIDFILE"
    fi
}

status() {
    if [ -f "$PIDFILE" ] && kill -0 $(cat "$PIDFILE") 2>/dev/null; then
        echo "Filebrowser is running (PID: $(cat $PIDFILE))"
        return 0
    else
        echo "Filebrowser is not running"
        return 1
    fi
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        sleep 2
        start
        ;;
    status)
        status
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac

exit 0
EOF

sudo chmod +x /etc/init.d/filebrowser
```

Then enable it in QTS Control Panel or via command:

```bash
# Enable on boot (QNAP)
/etc/init.d/filebrowser start
```

## Testing Your Setup

After setting up any method, test it:

```bash
# Check if process is running
ps aux | grep filebrowser

# Check if port is listening
netstat -tlnp | grep 8082
# or
ss -tlnp | grep 8082

# Test health endpoint
curl http://localhost:8082/health

# Check logs
tail -f /var/log/filebrowser.log
# or
cat /var/log/filebrowser.log
```

## Manual Start/Stop (For Testing)

Before setting up auto-start, test manually:

```bash
# Start manually
cd /share/Public/filebrowser-service
export DATA_ROOT="/share/studio work"
export PORT=8082
nohup ./filebrowser > /var/log/filebrowser.log 2>&1 &
echo $! > /var/run/filebrowser.pid

# Stop manually
kill $(cat /var/run/filebrowser.pid)
rm /var/run/filebrowser.pid
```

## Troubleshooting

### Check What Init System You Have

```bash
# Check for systemd
which systemctl

# Check for init.d
ls -la /etc/init.d/

# Check for runlevel directories
ls -la /etc/rc*.d/

# Check for rc.local
ls -la /etc/rc.local

# Check OS info
cat /etc/os-release
# or
uname -a
```

### Find Your NAS Type

```bash
# QNAP
cat /etc/version

# Synology
cat /etc/synoinfo.conf | grep productversion

# Generic Linux
cat /etc/issue
```

### Common Issues

1. **Script not executable:**
   ```bash
   sudo chmod +x /etc/init.d/filebrowser
   ```

2. **Permission denied:**
   ```bash
   # Check file ownership
   ls -la /share/Public/filebrowser-service/filebrowser
   # Fix if needed
   sudo chown Administrator:administrators /share/Public/filebrowser-service/filebrowser
   sudo chmod +x /share/Public/filebrowser-service/filebrowser
   ```

3. **Service not starting on boot:**
   - Check if init script has correct runlevel symlinks
   - Check if rc.local is executable
   - Check cron logs: `grep CRON /var/log/syslog`

4. **Can't find start-stop-daemon:**
   - Use Option 2 (simple init script) instead
   - Or install: `sudo apt-get install start-stop-daemon` (if package manager available)

## Recommended Approach

For most NAS systems (especially QNAP), **Option 1 (Init.d Script)** is the most reliable. If that doesn't work, try **Option 2 (Simple Init Script)** or **Option 4 (Cron @reboot)** as fallbacks.

