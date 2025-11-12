#!/bin/bash

# AR-13 Backend Deployment Script
# This script helps deploy the application to EC2

set -e

echo "🚀 AR-13 Backend Deployment Script"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running on EC2
if [ ! -f /sys/class/dmi/id/product_uuid ] || [ "$(cat /sys/class/dmi/id/product-uuid 2>/dev/null | grep -i ec2)" == "" ]; then
    echo -e "${YELLOW}Warning: This script is designed for EC2 instances${NC}"
fi

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
    echo -e "${RED}Please do not run as root. Use a regular user (ubuntu)${NC}"
    exit 1
fi

# Step 1: Update system
echo -e "\n${GREEN}[1/7] Updating system...${NC}"
sudo apt update && sudo apt upgrade -y

# Step 2: Install Go if not installed
if ! command -v go &> /dev/null; then
    echo -e "\n${GREEN}[2/7] Installing Go...${NC}"
    GO_VERSION="1.21.5"
    wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    echo -e "${GREEN}✓ Go installed: $(go version)${NC}"
else
    echo -e "${GREEN}✓ Go already installed: $(go version)${NC}"
fi

# Step 3: Install Nginx if not installed
if ! command -v nginx &> /dev/null; then
    echo -e "\n${GREEN}[3/7] Installing Nginx...${NC}"
    sudo apt install -y nginx
    echo -e "${GREEN}✓ Nginx installed${NC}"
else
    echo -e "${GREEN}✓ Nginx already installed${NC}"
fi

# Step 4: Install Certbot if not installed
if ! command -v certbot &> /dev/null; then
    echo -e "\n${GREEN}[4/7] Installing Certbot...${NC}"
    sudo apt install -y certbot python3-certbot-nginx
    echo -e "${GREEN}✓ Certbot installed${NC}"
else
    echo -e "${GREEN}✓ Certbot already installed${NC}"
fi

# Step 5: Install Redis if not installed
if ! command -v redis-server &> /dev/null; then
    echo -e "\n${GREEN}[5/7] Installing Redis...${NC}"
    sudo apt install -y redis-server
    sudo systemctl enable redis-server
    sudo systemctl start redis-server
    echo -e "${GREEN}✓ Redis installed and started${NC}"
else
    echo -e "${GREEN}✓ Redis already installed${NC}"
fi

# Step 6: Build application
echo -e "\n${GREEN}[6/7] Building application...${NC}"
cd ~/ar-13-go-backend || { echo -e "${RED}Error: ar-13-go-backend directory not found${NC}"; exit 1; }

echo "Downloading dependencies..."
go mod download

echo "Building application..."
go build -o server cmd/server/main.go

if [ -f server ]; then
    echo -e "${GREEN}✓ Application built successfully${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Step 7: Check if .env exists
if [ ! -f .env ]; then
    echo -e "\n${YELLOW}Warning: .env file not found${NC}"
    echo "Creating .env template..."
    cat > .env << EOF
PORT=3000
NODE_ENV=production
AWS_REGION=us-east-1
JWT_SECRET=CHANGE_THIS_TO_A_STRONG_SECRET_MIN_32_CHARS
JWT_EXPIRATION_HOURS=24
REFRESH_EXPIRATION_DAYS=30
EMAIL_HOST=
EMAIL_PORT=587
EMAIL_USER=
EMAIL_PASSWORD=
EMAIL_FROM=
EMAIL_FROM_NAME=AR-13
FRONTEND_URL=
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
EOF
    echo -e "${YELLOW}Please edit .env file with your configuration${NC}"
    echo "Run: nano ~/ar-13-go-backend/.env"
fi

# Step 8: Check systemd service
echo -e "\n${GREEN}[7/7] Checking systemd service...${NC}"
if [ -f /etc/systemd/system/ar13-backend.service ]; then
    echo -e "${GREEN}✓ Service file exists${NC}"
    echo "Restarting service..."
    sudo systemctl daemon-reload
    sudo systemctl restart ar13-backend
    sudo systemctl status ar13-backend --no-pager
else
    echo -e "${YELLOW}Service file not found. Please create it manually.${NC}"
    echo "See: docs/deployment/EC2_DEPLOYMENT_GUIDE.md"
fi

echo -e "\n${GREEN}====================================${NC}"
echo -e "${GREEN}✓ Deployment completed!${NC}"
echo -e "${GREEN}====================================${NC}"
echo ""
echo "Next steps:"
echo "1. Verify .env file is configured correctly"
echo "2. Check service status: sudo systemctl status ar13-backend"
echo "3. View logs: sudo journalctl -u ar13-backend -f"
echo "4. Configure Nginx (see deployment guide)"
echo "5. Set up SSL: sudo certbot --nginx -d your-domain.com"

