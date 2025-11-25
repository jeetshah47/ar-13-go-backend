#!/bin/bash
# Build and push ARM64 Docker image for filebrowser service
# Linux/macOS script

VERSION="${1:-latest}"
IMAGE_NAME="jeetshah786/ar-13-go-backend-filebrowser-service"
FULL_IMAGE_NAME="$IMAGE_NAME:$VERSION"

echo "Building ARM64 Docker image: $FULL_IMAGE_NAME"
echo "Architecture: linux/arm64"
echo ""

# Navigate to the filebrowser directory
cd "$(dirname "$0")"

# Build and push ARM64 image
echo "Building and pushing ARM64 image..."
echo "This may take a few minutes..."
echo ""

docker buildx build \
    --platform linux/arm64 \
    --tag $IMAGE_NAME:$VERSION \
    --tag $IMAGE_NAME:latest \
    --push \
    .

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Successfully built and pushed ARM64 image!"
    echo "Image: $FULL_IMAGE_NAME"
    echo ""
    echo "On your NAS, pull with:"
    echo "  docker pull $FULL_IMAGE_NAME"
else
    echo ""
    echo "❌ Build failed!"
    exit 1
fi

