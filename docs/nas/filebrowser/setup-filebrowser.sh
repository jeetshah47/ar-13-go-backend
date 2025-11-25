#!/bin/bash
# FileBrowser Setup Script for NAS
# This script sets up FileBrowser with proper permissions and configuration

set -e  # Exit on error

echo "=========================================="
echo "FileBrowser Setup Script for NAS"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.filebrowser.nas.yml"
CONTAINER_NAME="filebrowser"
CONFIG_DIR="/share/Container/filebrowser-config"
USE_NAMED_VOLUME=true  # Set to false to use bind mount

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root or with sudo
check_permissions() {
    if [ "$EUID" -ne 0 ]; then
        print_warn "Not running as root. Some operations may require sudo."
        SUDO_CMD="sudo"
    else
        SUDO_CMD=""
    fi
}

# Stop existing container
stop_container() {
    print_info "Stopping existing FileBrowser container..."
    docker-compose -f "$COMPOSE_FILE" down 2>/dev/null || true
    print_info "Container stopped."
}

# Setup bind mount (if not using named volume)
setup_bind_mount() {
    if [ "$USE_NAMED_VOLUME" = false ]; then
        print_info "Setting up bind mount at $CONFIG_DIR..."
        
        # Create directory
        if [ ! -d "$CONFIG_DIR" ]; then
            print_info "Creating directory: $CONFIG_DIR"
            $SUDO_CMD mkdir -p "$CONFIG_DIR"
        else
            print_info "Directory already exists: $CONFIG_DIR"
        fi
        
        # Set permissions
        print_info "Setting permissions to 755..."
        $SUDO_CMD chmod -R 755 "$CONFIG_DIR"
        
        # Set ownership
        print_info "Setting ownership to 1000:1000..."
        $SUDO_CMD chown -R 1000:1000 "$CONFIG_DIR"
        
        # Verify
        print_info "Verifying permissions..."
        ls -la "$(dirname $CONFIG_DIR)" | grep "$(basename $CONFIG_DIR)" || print_warn "Could not verify permissions"
        
        print_info "Bind mount setup complete."
    else
        print_info "Using named volume - skipping bind mount setup."
    fi
}

# Remove old volumes (cleanup)
cleanup_volumes() {
    print_info "Cleaning up old volumes..."
    docker volume rm filebrowser-config 2>/dev/null || print_warn "Volume doesn't exist or already removed"
}

# Start container
start_container() {
    print_info "Starting FileBrowser container..."
    docker-compose -f "$COMPOSE_FILE" up -d
    
    if [ $? -eq 0 ]; then
        print_info "Container started successfully!"
    else
        print_error "Failed to start container!"
        exit 1
    fi
}

# Wait for container to be ready
wait_for_container() {
    print_info "Waiting for container to be ready..."
    sleep 5
    
    if docker ps | grep -q "$CONTAINER_NAME"; then
        print_info "Container is running."
    else
        print_error "Container is not running!"
        print_info "Checking logs..."
        docker logs "$CONTAINER_NAME" 2>&1 | tail -20
        exit 1
    fi
}

# Check container logs
check_logs() {
    print_info "Checking container logs..."
    echo ""
    docker logs "$CONTAINER_NAME" 2>&1 | tail -30
    echo ""
}

# Verify setup
verify_setup() {
    print_info "Verifying setup..."
    
    # Check if container is running
    if docker ps | grep -q "$CONTAINER_NAME"; then
        print_info "✓ Container is running"
    else
        print_error "✗ Container is not running"
        return 1
    fi
    
    # Check if database file exists
    if docker exec "$CONTAINER_NAME" test -f /filebrowser-config/database.db 2>/dev/null; then
        print_info "✓ Database file created successfully"
    else
        print_warn "⚠ Database file not found (may be initializing)"
    fi
    
    # Check if port is accessible
    if netstat -tuln 2>/dev/null | grep -q ":8081" || ss -tuln 2>/dev/null | grep -q ":8081"; then
        print_info "✓ Port 8081 is listening"
    else
        print_warn "⚠ Port 8081 may not be accessible"
    fi
}

# Get admin password from logs
get_admin_password() {
    print_info "Retrieving admin password from logs..."
    PASSWORD=$(docker logs "$CONTAINER_NAME" 2>&1 | grep "randomly generated password" | tail -1 | awk -F': ' '{print $2}' | tr -d '[:space:]')
    
    if [ -n "$PASSWORD" ]; then
        echo ""
        print_info "=========================================="
        print_info "FileBrowser Admin Credentials:"
        print_info "=========================================="
        echo -e "Username: ${GREEN}admin${NC}"
        echo -e "Password: ${GREEN}$PASSWORD${NC}"
        echo ""
        print_info "Access FileBrowser at: http://YOUR-NAS-IP:8081"
        print_info "=========================================="
    else
        print_warn "Could not retrieve password from logs."
        print_info "Default credentials: admin / admin"
        print_info "Check logs with: docker logs $CONTAINER_NAME"
    fi
}

# Main execution
main() {
    echo ""
    print_info "Starting FileBrowser setup..."
    echo ""
    
    # Check permissions
    check_permissions
    
    # Stop existing container
    stop_container
    
    # Setup bind mount (if not using named volume)
    if [ "$USE_NAMED_VOLUME" = false ]; then
        setup_bind_mount
    else
        cleanup_volumes
    fi
    
    # Start container
    start_container
    
    # Wait for container
    wait_for_container
    
    # Check logs
    check_logs
    
    # Verify setup
    verify_setup
    
    # Get admin password
    get_admin_password
    
    echo ""
    print_info "Setup complete!"
    print_info "To view logs: docker logs $CONTAINER_NAME"
    print_info "To stop: docker-compose -f $COMPOSE_FILE down"
    print_info "To restart: docker-compose -f $COMPOSE_FILE restart"
    echo ""
}

# Run main function
main

