#!/bin/bash
# Run FileBrowser Service directly with go run
# Linux/macOS script

export PORT=8082
export DATA_ROOT=/data

echo "Starting FileBrowser Service..."
echo "Port: $PORT"
echo "Data Root: $DATA_ROOT"
echo ""

# Navigate to the filebrowser directory
cd "$(dirname "$0")"

# Run the service
go run main.go


