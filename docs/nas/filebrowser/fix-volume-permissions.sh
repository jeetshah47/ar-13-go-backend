#!/bin/bash
# Fix FileBrowser volume permissions
# Run this script if you get permission denied errors

echo "Fixing FileBrowser volume permissions..."

# Stop container
echo "Stopping container..."
docker-compose -f docker-compose.filebrowser.nas.yml down

# Remove old volume
echo "Removing old volume..."
docker volume rm filebrowser-config 2>/dev/null || echo "Volume doesn't exist"

# Create new volume with correct permissions
echo "Creating new volume..."
docker volume create filebrowser-config

# Start container as root to create files
echo "Starting container as root to initialize files..."
docker-compose -f docker-compose.filebrowser.nas.yml up -d

# Wait for container to initialize
echo "Waiting for container to initialize..."
sleep 10

# Fix permissions inside container
echo "Fixing permissions inside container..."
docker exec filebrowser chown -R 1000:1000 /filebrowser-config
docker exec filebrowser chmod -R 755 /filebrowser-config

# Verify permissions
echo "Verifying permissions..."
docker exec filebrowser ls -la /filebrowser-config

echo ""
echo "✅ Permissions fixed!"
echo ""
echo "Now you can:"
echo "1. Uncomment 'user: \"1000:1000\"' in docker-compose file (optional, for security)"
echo "2. Restart container: docker-compose -f docker-compose.filebrowser.nas.yml restart"
echo "3. Access FileBrowser at: http://YOUR-NAS-IP:8081"

