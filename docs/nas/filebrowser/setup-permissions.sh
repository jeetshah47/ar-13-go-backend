#!/bin/bash
# Setup script for FileBrowser permissions on NAS
# Run this script before starting FileBrowser container

echo "Setting up FileBrowser permissions..."

# Create config directory if it doesn't exist
CONFIG_DIR="/share/Container/filebrowser-config"
echo "Creating directory: $CONFIG_DIR"
mkdir -p "$CONFIG_DIR"

# Set permissions (755 = rwxr-xr-x)
echo "Setting permissions to 755..."
chmod -R 755 "$CONFIG_DIR"

# Set ownership to UID 1000 (FileBrowser default user)
echo "Setting ownership to 1000:1000..."
chown -R 1000:1000 "$CONFIG_DIR"

# Verify permissions
echo ""
echo "Verifying permissions:"
ls -la "$CONFIG_DIR"

echo ""
echo "✅ Permissions setup complete!"
echo ""
echo "Now you can start FileBrowser:"
echo "  docker-compose -f docker-compose.filebrowser.nas.yml up -d"
echo ""
echo "If you still get permission errors, uncomment 'user: \"1000:1000\"' in docker-compose file"

