#!/bin/bash
# Build script to create binaries for Docker containers
# This script builds the server and worker binaries for Linux x86_64

set -e

# Ignore any arguments passed to the script
# (This script doesn't accept arguments)
if [ $# -gt 0 ]; then
    echo "Warning: This script doesn't accept arguments. Ignoring: $@"
fi

echo "Building AR-13 backend binaries for Linux x86_64..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build configuration for x86_64 (amd64)
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64  # amd64 = x86_64 architecture

# Build flags for optimized binary (properly quoted)
BUILD_FLAGS=(-a -installsuffix cgo -ldflags "-w -s")

# Build server binary
echo "Building server binary (x86_64)..."
go build "${BUILD_FLAGS[@]}" -o bin/server ./cmd/server/main.go

# Build worker binary
echo "Building worker binary (x86_64)..."
go build "${BUILD_FLAGS[@]}" -o bin/worker ./cmd/worker/main.go

# Make binaries executable
chmod +x bin/server bin/worker

echo ""
echo "✅ Binaries built successfully!"
echo "Location: bin/"
echo ""
echo "Server binary: $(du -h bin/server | cut -f1)"
echo "Worker binary: $(du -h bin/worker | cut -f1)"
echo ""
echo "You can now build Docker images:"
echo "  docker build -f Dockerfile.server -t ar-13-server ."
echo "  docker build -f Dockerfile.worker -t ar-13-worker ."
echo "  Or use: docker-compose build"

