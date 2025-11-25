#!/bin/bash
# Script to pull ARM64 version of the image on NAS
# Run this on your NAS

IMAGE_NAME="jeetshah786/ar-13-go-backend-filebrowser-service:latest"

echo "Pulling ARM64 version of $IMAGE_NAME"
echo ""

# Remove existing image if it exists
docker rmi $IMAGE_NAME 2>/dev/null || true

# Pull ARM64 version explicitly
docker pull --platform linux/arm64 $IMAGE_NAME

# Verify architecture
echo ""
echo "Verifying architecture..."
ARCH=$(docker inspect $IMAGE_NAME | grep -i '"Architecture"' | head -1 | cut -d'"' -f4)
echo "Image Architecture: $ARCH"

if [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
    echo "✅ Correct ARM64 architecture detected"
else
    echo "⚠️  Warning: Architecture might not be ARM64"
fi

echo ""
echo "Now you can run the container:"
echo "docker run -d \\"
echo "  --name filebrowser-service \\"
echo "  --restart unless-stopped \\"
echo "  --platform linux/arm64 \\"
echo "  -v \"/share/studio work:/data\" \\"
echo "  -e DATA_ROOT=/data \\"
echo "  -e PORT=8082 \\"
echo "  -e TZ=Asia/Kolkata \\"
echo "  --network filebrowser-network \\"
echo "  --user \"1000:1000\" \\"
echo "  $IMAGE_NAME"

