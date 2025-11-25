#!/bin/bash
# Bash script to start frontend and backend with interactive restart capability
# Press 'r' to restart both processes, 'q' to quit

# Get script directory and set paths relative to it
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_PATH="$SCRIPT_DIR"
FRONTEND_PATH="$(dirname "$SCRIPT_DIR")/ar-13-ui"

FRONTEND_PID=""
BACKEND_PID=""

# Colors for output
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
MAGENTA='\033[0;35m'
RED='\033[0;31m'
NC='\033[0m' # No Color

function start_processes() {
    echo -e "\n${CYAN}========================================"
    echo "Starting Frontend and Backend..."
    echo -e "========================================${NC}\n"
    
    # Start Frontend
    echo -e "${GREEN}[FRONTEND] Starting npm run dev...${NC}"
    cd "$FRONTEND_PATH" || exit 1
    npm run dev > /dev/null 2>&1 &
    FRONTEND_PID=$!
    
    # Start Backend
    echo -e "${YELLOW}[BACKEND] Starting go run cmd/server/main.go...${NC}"
    cd "$BACKEND_PATH" || exit 1
    go run cmd/server/main.go > /dev/null 2>&1 &
    BACKEND_PID=$!
    
    echo -e "\n${CYAN}[INFO] Both processes started!${NC}"
    echo -e "${CYAN}[INFO] Press 'r' to restart both processes${NC}"
    echo -e "${CYAN}[INFO] Press 'q' to quit${NC}\n"
}

function stop_processes() {
    echo -e "\n${YELLOW}[INFO] Stopping processes...${NC}"
    
    # Kill frontend process and its children
    if [ ! -z "$FRONTEND_PID" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
        pkill -P "$FRONTEND_PID" 2>/dev/null
        kill "$FRONTEND_PID" 2>/dev/null
    fi
    
    # Kill backend process and its children
    if [ ! -z "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
        pkill -P "$BACKEND_PID" 2>/dev/null
        kill "$BACKEND_PID" 2>/dev/null
    fi
    
    # Kill any remaining node/vite processes (be careful with this)
    pkill -f "vite" 2>/dev/null
    pkill -f "npm run dev" 2>/dev/null
    
    # Wait a bit for processes to terminate
    sleep 0.5
    
    FRONTEND_PID=""
    BACKEND_PID=""
}

function cleanup() {
    echo -e "\n${YELLOW}[INFO] Cleaning up...${NC}"
    stop_processes
    echo -e "${YELLOW}[INFO] Exiting...${NC}"
    exit 0
}

# Trap signals to cleanup on exit
trap cleanup EXIT INT TERM

# Main loop
start_processes

while true; do
    # Check if processes are still running
    if [ ! -z "$FRONTEND_PID" ] && ! kill -0 "$FRONTEND_PID" 2>/dev/null; then
        echo -e "\n${RED}[WARNING] Frontend process exited!${NC}"
    fi
    
    if [ ! -z "$BACKEND_PID" ] && ! kill -0 "$BACKEND_PID" 2>/dev/null; then
        echo -e "\n${RED}[WARNING] Backend process exited!${NC}"
    fi
    
    # Read a single character (non-blocking)
    # Save terminal settings
    if [ -t 0 ]; then
        old_stty=$(stty -g 2>/dev/null)
        # Set terminal to raw mode for single character input
        stty raw -echo min 0 time 1 2>/dev/null
        # Read single character with timeout
        key=$(dd bs=1 count=1 2>/dev/null)
        # Restore terminal settings
        stty "$old_stty" 2>/dev/null
        
        if [ ! -z "$key" ]; then
            case "$key" in
                r|R)
                    echo -e "\n${MAGENTA}[RESTART] Restarting both processes...${NC}"
                    stop_processes
                    sleep 1
                    start_processes
                    ;;
                q|Q)
                    cleanup
                    ;;
            esac
        fi
    fi
    
    sleep 0.1
done

