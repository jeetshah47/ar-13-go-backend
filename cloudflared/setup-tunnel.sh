#!/bin/bash
# Cloudflare Tunnel Setup Script
# Uses Docker to run cloudflared - no local installation required

set -e

TUNNEL_NAME="${1:-filebrowser-tunnel}"
HOSTNAME="${2:-}"

echo "Cloudflare Tunnel Setup Script"
echo "==============================="
echo "Using Docker image: cloudflare/cloudflared"
echo ""

# Check if Docker is running
if ! docker ps > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running or not installed!"
    echo "Please start Docker and try again."
    exit 1
fi

echo "✅ Docker is running"

# Get tunnel name if not provided
if [ -z "$TUNNEL_NAME" ]; then
    read -p "Enter tunnel name (default: filebrowser-tunnel): " TUNNEL_NAME
    TUNNEL_NAME="${TUNNEL_NAME:-filebrowser-tunnel}"
fi

# Get hostname
if [ -z "$HOSTNAME" ]; then
    read -p "Enter full hostname (e.g., filebrowser.yourdomain.com): " HOSTNAME
    if [ -z "$HOSTNAME" ]; then
        echo "❌ Error: Hostname is required"
        exit 1
    fi
fi

echo ""
echo "Configuration:"
echo "  Tunnel Name: $TUNNEL_NAME"
echo "  Hostname: $HOSTNAME"
echo ""

read -p "Continue? (y/n): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Create cloudflared directory if it doesn't exist
CLOUDFLARED_DIR="cloudflared"
mkdir -p "$CLOUDFLARED_DIR"
echo "✅ Created cloudflared directory"

# Determine the path for cloudflared data
CLOUDFLARED_DATA_PATH="$HOME/.cloudflared"
mkdir -p "$CLOUDFLARED_DATA_PATH"

# Step 1: Authenticate with Cloudflare
echo ""
echo "Step 1: Authenticating with Cloudflare..."
echo "This will open a browser window for authentication..."

CERT_PATH="$CLOUDFLARED_DATA_PATH/cert.pem"
if [ -f "$CERT_PATH" ]; then
    echo "✅ Already authenticated (cert.pem exists)"
else
    echo "Running cloudflared tunnel login..."
    
    # Run cloudflared login in Docker
    # Set HOME to /etc/cloudflared so cert.pem is saved to the mounted volume
    docker run --rm -it \
        -v "$CLOUDFLARED_DATA_PATH:/etc/cloudflared" \
        -e HOME=/etc/cloudflared \
        cloudflare/cloudflared:latest \
        tunnel login
    
    if [ $? -ne 0 ]; then
        echo "❌ Error: Authentication failed"
        exit 1
    fi
    
    # Wait a moment and verify cert.pem was created
    sleep 3
    if [ ! -f "$CERT_PATH" ]; then
        echo "❌ Error: cert.pem was not created after login"
        echo "The certificate may have been saved inside the container."
        echo "Trying to extract it..."
        
        # Try to copy from container's default location
        CONTAINER_ID=$(docker run -d \
            -v "$CLOUDFLARED_DATA_PATH:/etc/cloudflared" \
            -e HOME=/etc/cloudflared \
            cloudflare/cloudflared:latest \
            sleep 10)
        
        if [ -n "$CONTAINER_ID" ]; then
            docker cp "${CONTAINER_ID}:/home/nonroot/.cloudflared/cert.pem" "$CERT_PATH" 2>/dev/null
            docker rm -f "$CONTAINER_ID" >/dev/null 2>&1
            
            if [ -f "$CERT_PATH" ]; then
                echo "✅ Successfully extracted cert.pem"
            else
                echo "❌ Could not extract cert.pem. Please check $CLOUDFLARED_DATA_PATH"
                exit 1
            fi
        else
            echo "❌ Could not extract cert.pem. Please check $CLOUDFLARED_DATA_PATH"
            exit 1
        fi
    else
        echo "✅ Authentication successful - cert.pem found"
    fi
fi

# Step 2: Create tunnel
echo ""
echo "Step 2: Creating tunnel '$TUNNEL_NAME'..."

# Verify cert.pem exists before proceeding
if [ ! -f "$CERT_PATH" ]; then
    echo "❌ Error: cert.pem not found at $CERT_PATH"
    echo "Please run the authentication step again."
    exit 1
fi

# Set HOME to /etc/cloudflared so it uses the mounted volume
TUNNEL_OUTPUT=$(docker run --rm \
    -v "$CLOUDFLARED_DATA_PATH:/etc/cloudflared" \
    -e HOME=/etc/cloudflared \
    cloudflare/cloudflared:latest \
    tunnel create "$TUNNEL_NAME" 2>&1)

# Extract tunnel ID
TUNNEL_ID=""
if echo "$TUNNEL_OUTPUT" | grep -q "Created tunnel"; then
    TUNNEL_ID=$(echo "$TUNNEL_OUTPUT" | grep -oP 'Created tunnel \K[^ ]+')
    echo "✅ Created tunnel: $TUNNEL_ID"
else
    # Try to get existing tunnel
    TUNNEL_LIST=$(docker run --rm \
        -v "$CLOUDFLARED_DATA_PATH:/etc/cloudflared" \
        -e HOME=/etc/cloudflared \
        cloudflare/cloudflared:latest \
        tunnel list 2>&1)
    
    TUNNEL_ID=$(echo "$TUNNEL_LIST" | grep "$TUNNEL_NAME" | awk '{print $1}' | head -n1)
    if [ -n "$TUNNEL_ID" ]; then
        echo "✅ Using existing tunnel: $TUNNEL_ID"
    else
        echo "❌ Error: Could not create or find tunnel"
        echo "Output: $TUNNEL_OUTPUT"
        exit 1
    fi
fi

# Step 3: Copy credentials.json to project directory
echo ""
echo "Step 3: Setting up credentials..."

CREDENTIALS_SOURCE="$CLOUDFLARED_DATA_PATH/$TUNNEL_ID.json"
CREDENTIALS_DEST="$CLOUDFLARED_DIR/credentials.json"

if [ ! -f "$CREDENTIALS_SOURCE" ]; then
    echo "❌ Error: Credentials file not found at $CREDENTIALS_SOURCE"
    exit 1
fi

cp "$CREDENTIALS_SOURCE" "$CREDENTIALS_DEST"
echo "✅ Copied credentials to $CREDENTIALS_DEST"

# Step 4: Create DNS route
echo ""
echo "Step 4: Creating DNS route..."

DNS_OUTPUT=$(docker run --rm \
    -v "$CLOUDFLARED_DATA_PATH:/etc/cloudflared" \
    -e HOME=/etc/cloudflared \
    cloudflare/cloudflared:latest \
    tunnel route dns "$TUNNEL_NAME" "$HOSTNAME" 2>&1)

if [ $? -eq 0 ]; then
    echo "✅ Created DNS route for $HOSTNAME"
else
    echo "⚠️  Warning: DNS route creation may have failed"
    echo "Output: $DNS_OUTPUT"
fi

# Step 5: Generate config.yml
echo ""
echo "Step 5: Generating config.yml..."

cat > "$CLOUDFLARED_DIR/config.yml" <<EOF
# Cloudflare Tunnel Configuration
# Auto-generated by setup-tunnel.sh

tunnel: $TUNNEL_ID
credentials-file: /etc/cloudflared/credentials.json

# Ingress rules - define how traffic is routed
ingress:
  # Route all traffic to filebrowser service
  - hostname: $HOSTNAME
    service: http://filebrowser-service:8082
  
  # Catch-all rule (must be last)
  - service: http_status:404
EOF

echo "✅ Generated config.yml"

# Step 6: Summary
echo ""
echo "=================================="
echo "✅ Tunnel setup complete!"
echo "=================================="
echo ""
echo "Tunnel Details:"
echo "  ID: $TUNNEL_ID"
echo "  Name: $TUNNEL_NAME"
echo "  Hostname: $HOSTNAME"
echo ""
echo "Files created:"
echo "  - $CLOUDFLARED_DIR/credentials.json"
echo "  - $CLOUDFLARED_DIR/config.yml"
echo ""
echo "Next steps:"
echo "1. Start the services:"
echo "   docker-compose -f docker-compose.filebrowser-cloudflared.yml up -d"
echo ""
echo "2. Check tunnel status:"
echo "   docker logs cloudflared-tunnel"
echo ""
echo "3. Access your service:"
echo "   https://$HOSTNAME"
echo ""

