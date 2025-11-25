#!/bin/bash

# Script to create MinIO Access Keys using MinIO Client (mc)
# Usage: ./create_minio_access_key.sh

set -e

# Configuration - Update these values
MINIO_ENDPOINT="${MINIO_ENDPOINT:-your-nas-ip:32774}"
MINIO_ROOT_USER="${MINIO_ROOT_USER:-ar-web-backend}"
MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD:-bIF62kxCm54D}"
ACCESS_KEY_NAME="${ACCESS_KEY_NAME:-ar-13-backend}"

echo "=========================================="
echo "MinIO Access Key Creation Script"
echo "=========================================="
echo ""

# Check if mc is installed
if ! command -v mc &> /dev/null; then
    echo "ERROR: MinIO Client (mc) is not installed."
    echo ""
    echo "Install it from: https://min.io/docs/minio/linux/reference/minio-mc.html"
    echo "Or on Windows with Chocolatey: choco install minio-client"
    exit 1
fi

echo "Step 1: Configuring MinIO alias..."
mc alias set myminio "http://${MINIO_ENDPOINT}" "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}"

echo ""
echo "Step 2: Creating user for access key..."
USER_NAME="${ACCESS_KEY_NAME}-user"
mc admin user add myminio "${USER_NAME}"

echo ""
echo "Step 3: Creating access key..."
echo "Creating service account for user: ${USER_NAME}"

# Create service account and capture output
OUTPUT=$(mc admin user svcacct add myminio "${USER_NAME}" --name "${ACCESS_KEY_NAME}-key" 2>&1)

echo ""
echo "=========================================="
echo "Access Key Created Successfully!"
echo "=========================================="
echo ""
echo "$OUTPUT"
echo ""
echo "IMPORTANT: Copy the Access Key and Secret Key above!"
echo "Add them to your .env file:"
echo ""
echo "MINIO_ACCESS_KEY=<Access Key from above>"
echo "MINIO_SECRET_KEY=<Secret Key from above>"
echo ""

