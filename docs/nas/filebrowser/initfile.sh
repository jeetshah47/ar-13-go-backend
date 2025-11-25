#!/bin/sh

DAEMON="/share/Public/filebrowser-build/filebrowser"
DAEMON_DIR="/share/Public/filebrowser-build"
PIDFILE="/share/Public/filebrowser-build/filebrowser.pid"
LOGFILE="/share/Public/filebrowser-build/filebrowser.log"

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