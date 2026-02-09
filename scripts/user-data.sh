#!/bin/bash
# User Data Script for AR-13 Backend EC2 Instance
# This script runs automatically when the instance launches

set -e  # Exit on error

# Log everything
exec > >(tee /var/log/user-data.log)
exec 2>&1

echo "=========================================="
echo "AR-13 Backend - User Data Script"
echo "Started at: $(date)"
echo "=========================================="

# Update system
echo "Updating system packages..."
yum update -y

# Install Docker
echo "Installing Docker..."
yum install -y docker
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user

# Install Docker Compose
echo "Installing Docker Compose..."
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Verify Docker Compose installation
docker-compose --version

# Create application directories
echo "Creating application directories..."
mkdir -p /home/ec2-user/upload
mkdir -p /home/ec2-user/logs
mkdir -p /home/ec2-user/app
chown -R ec2-user:ec2-user /home/ec2-user

# Install basic tools
echo "Installing basic tools..."
yum install -y curl wget git htop

# Configure swap (critical for 1GB RAM on t3.micro)
echo "Configuring swap file..."
if [ ! -f /swapfile ]; then
    dd if=/dev/zero of=/swapfile bs=1M count=1024
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
    echo "Swap file created: 1GB"
else
    echo "Swap file already exists"
fi

# Optimize for low memory
echo "Optimizing memory settings..."
echo 'vm.swappiness=10' >> /etc/sysctl.conf
sysctl -p

# Configure timezone
echo "Setting timezone to UTC..."
timedatectl set-timezone UTC

# Install CloudWatch agent (optional - for monitoring)
echo "Installing CloudWatch agent..."
yum install -y amazon-cloudwatch-agent

# Create systemd service for auto-start (optional)
echo "Creating systemd service..."
cat > /etc/systemd/system/ar-13-backend.service << 'EOF'
[Unit]
Description=AR-13 Backend Service
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/home/ec2-user
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
User=ec2-user
Group=ec2-user

[Install]
WantedBy=multi-user.target
EOF

# Don't enable by default - user can enable it later
# systemctl enable ar-13-backend

echo ""
echo "=========================================="
echo "User Data Script Completed"
echo "Finished at: $(date)"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. SSH into the instance"
echo "2. Create .env file in /home/ec2-user"
echo "3. Create docker-compose.yml in /home/ec2-user"
echo "4. Deploy your application"
echo ""
echo "To enable auto-start service:"
echo "  sudo systemctl enable ar-13-backend"
echo "  sudo systemctl start ar-13-backend"
echo ""





