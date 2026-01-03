#!/bin/bash
# Setup script to create .env file from template on NAS server
# Usage: ./setup-env.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"
TEMPLATE_FILE="${SCRIPT_DIR}/env.template"

echo "=========================================="
echo "AR-13 Backend Environment Setup"
echo "=========================================="
echo ""

# Check if .env already exists
if [ -f "$ENV_FILE" ]; then
    echo "⚠️  WARNING: .env file already exists at: $ENV_FILE"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted. Existing .env file preserved."
        exit 0
    fi
    echo "Backing up existing .env to .env.backup"
    cp "$ENV_FILE" "${ENV_FILE}.backup"
fi

# Check if template exists
if [ ! -f "$TEMPLATE_FILE" ]; then
    echo "❌ Error: Template file not found: $TEMPLATE_FILE"
    echo "Please ensure env.template exists in the docker directory."
    exit 1
fi

# Copy template to .env
echo "📋 Creating .env file from template..."
cp "$TEMPLATE_FILE" "$ENV_FILE"

echo "✅ .env file created at: $ENV_FILE"
echo ""
echo "⚠️  IMPORTANT: You must now edit the .env file and fill in your actual values:"
echo "   nano $ENV_FILE"
echo "   or"
echo "   vi $ENV_FILE"
echo ""
echo "Required variables to set:"
echo "  - MONGODB_URI (REQUIRED)"
echo "  - JWT_SECRET (REQUIRED - generate with: openssl rand -base64 32)"
echo "  - REDIS_PASSWORD (if using Redis)"
echo ""
echo "After editing, verify with:"
echo "   docker compose -f docker-compose.nas.yml config"

