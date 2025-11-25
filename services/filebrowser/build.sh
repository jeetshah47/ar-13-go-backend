#!/bin/bash
# Build FileBrowser Service as a standalone binary
# Bash script for Linux/macOS

set -e

OS="${1:-linux}"
ARCH="${2:-amd64}"
OUTPUT="${3:-}"

# Navigate to the filebrowser directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Determine output filename
if [ -z "$OUTPUT" ]; then
    if [ "$OS" = "windows" ]; then
        OUTPUT="filebrowser-service.exe"
    elif [ "$OS" = "linux" ]; then
        OUTPUT="filebrowser-service-linux-$ARCH"
    elif [ "$OS" = "darwin" ]; then
        OUTPUT="filebrowser-service-darwin-$ARCH"
    else
        OUTPUT="filebrowser-service-$OS-$ARCH"
    fi
fi

echo "Building FileBrowser Service..."
echo "OS: $OS"
echo "Architecture: $ARCH"
echo "Output: $OUTPUT"
echo ""

# Set build environment variables
export GOOS="$OS"
export GOARCH="$ARCH"

# Build the binary
echo "Running: go build -o $OUTPUT ."
go build -o "$OUTPUT" .

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Successfully built binary!"
    echo "Output: $OUTPUT"
    
    # Show file info
    if [ -f "$OUTPUT" ]; then
        SIZE=$(du -h "$OUTPUT" | cut -f1)
        echo "Size: $SIZE"
        echo "Location: $(pwd)/$OUTPUT"
    fi
else
    echo ""
    echo "❌ Build failed!"
    exit 1
fi

