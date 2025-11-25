#!/bin/bash
# Build and push multi-arch Docker image for filebrowser service
# Supports both AMD64 and ARM64 architectures

set -e

IMAGE_NAME="jeetshah786/ar-13-go-backend-filebrowser-service"
VERSION="${1:-latest}"

echo "Building multi-arch Docker image: $IMAGE_NAME:$VERSION"
echo "Architectures: linux/amd64, linux/arm64"
echo ""

# Check if buildx is available
if ! docker buildx version > /dev/null 2>&1; then
    echo "❌ Error: docker buildx is not available"
    echo "Please install Docker Buildx or update Docker Desktop"
    exit 1
fi

# Create and use a new builder instance (if needed)
BUILDER_NAME="multiarch-builder"
if ! docker buildx inspect $BUILDER_NAME > /dev/null 2>&1; then
    echo "Creating new buildx builder: $BUILDER_NAME"
    docker buildx create --name $BUILDER_NAME --use
    docker buildx inspect --bootstrap
else
    echo "Using existing buildx builder: $BUILDER_NAME"
    docker buildx use $BUILDER_NAME
fi

# Navigate to the filebrowser directory
cd "$(dirname "$0")"

# Build and push multi-arch image
echo ""
echo "Building and pushing multi-arch image..."
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag $IMAGE_NAME:$VERSION \
    --tag $IMAGE_NAME:latest \
    --push \
    .

echo ""
echo "✅ Successfully built and pushed multi-arch image!"
echo "Image: $IMAGE_NAME:$VERSION"
echo ""
echo "To verify, run:"
echo "  docker buildx imagetools inspect $IMAGE_NAME:$VERSION"

