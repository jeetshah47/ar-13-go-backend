#!/bin/bash

# AR-13 Backend Deployment Script for EC2
# This script builds the application and deploys it to EC2 using Docker
#
# Usage:
#   ./scripts/deploy-to-ec2.sh [OPTIONS]
#
# Options:
#   --ec2-ip IP              EC2 instance IP (default: 13.201.254.55)
#   --ec2-user USER          EC2 user (default: ubuntu)
#   --ssh-key PATH           Path to SSH private key (default: ~/.ssh/id_rsa)
#   --skip-build             Skip building binaries (use existing)
#   --skip-docker-build      Skip Docker image build on EC2 (use existing images)
#   --app-dir PATH           Application directory on EC2 (default: ~/ar-13-go-backend)
#   --env-file PATH          Path to .env file to copy (default: .env)
#   --help                   Show this help message

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
EC2_IP="13.201.254.55"
EC2_USER="ubuntu"
SSH_KEY="${HOME}/.ssh/id_rsa"
APP_DIR="~/ar-13-go-backend"
SKIP_BUILD=false
SKIP_DOCKER_BUILD=false
ENV_FILE=".env"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --ec2-ip)
            EC2_IP="$2"
            shift 2
            ;;
        --ec2-user)
            EC2_USER="$2"
            shift 2
            ;;
        --ssh-key)
            SSH_KEY="$2"
            shift 2
            ;;
        --app-dir)
            APP_DIR="$2"
            shift 2
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-docker-build)
            SKIP_DOCKER_BUILD=true
            shift
            ;;
        --env-file)
            ENV_FILE="$2"
            shift 2
            ;;
        --help)
            echo "AR-13 Backend Deployment Script for EC2"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --ec2-ip IP              EC2 instance IP (default: 13.201.254.55)"
            echo "  --ec2-user USER          EC2 user (default: ubuntu)"
            echo "  --ssh-key PATH           Path to SSH private key (default: ~/.ssh/id_rsa)"
            echo "  --skip-build             Skip building binaries (use existing)"
            echo "  --skip-docker-build      Skip Docker image build on EC2"
            echo "  --app-dir PATH           Application directory on EC2 (default: ~/ar-13-go-backend)"
            echo "  --env-file PATH          Path to .env file to copy (default: .env)"
            echo "  --help                   Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Expand tilde in paths
SSH_KEY="${SSH_KEY/#\~/$HOME}"
ENV_FILE="${ENV_FILE/#\~/$HOME}"

# Function to print colored output
print_step() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Function to check prerequisites
check_prerequisites() {
    print_step "Checking Prerequisites"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        exit 1
    fi
    print_success "Go is installed: $(go version)"
    
    # Check if SSH key exists
    if [ ! -f "$SSH_KEY" ]; then
        print_error "SSH key not found: $SSH_KEY"
        print_warning "Please provide the correct path using --ssh-key option"
        exit 1
    fi
    print_success "SSH key found: $SSH_KEY"
    
    # Check if .env file exists
    if [ ! -f "$PROJECT_ROOT/$ENV_FILE" ]; then
        print_warning ".env file not found: $PROJECT_ROOT/$ENV_FILE"
        print_warning "You may need to create it on the EC2 instance"
    else
        print_success ".env file found: $PROJECT_ROOT/$ENV_FILE"
    fi
    
    # Check SSH connection
    print_step "Testing SSH Connection"
    if ssh -i "$SSH_KEY" -o ConnectTimeout=10 -o StrictHostKeyChecking=no "$EC2_USER@$EC2_IP" "echo 'Connection successful'" &>/dev/null; then
        print_success "SSH connection successful"
    else
        print_error "Failed to connect to EC2 instance"
        print_warning "Please check:"
        print_warning "  - EC2 IP address: $EC2_IP"
        print_warning "  - SSH key path: $SSH_KEY"
        print_warning "  - Security group allows SSH (port 22)"
        exit 1
    fi
    
    # Check if Docker is installed on EC2
    if ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "command -v docker &> /dev/null"; then
        print_success "Docker is installed on EC2"
    else
        print_error "Docker is not installed on EC2"
        print_warning "Please install Docker first:"
        print_warning "  ssh -i $SSH_KEY $EC2_USER@$EC2_IP"
        print_warning "  curl -fsSL https://get.docker.com -o get-docker.sh"
        print_warning "  sudo sh get-docker.sh"
        print_warning "  sudo usermod -aG docker $EC2_USER"
        exit 1
    fi
    
    # Check if docker-compose is installed on EC2
    if ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "command -v docker-compose &> /dev/null || docker compose version &> /dev/null 2>&1"; then
        print_success "Docker Compose is installed on EC2"
    else
        print_error "Docker Compose is not installed on EC2"
        print_warning "Please install Docker Compose first"
        exit 1
    fi
}

# Function to build binaries
build_binaries() {
    if [ "$SKIP_BUILD" = true ]; then
        print_warning "Skipping binary build (--skip-build flag used)"
        if [ ! -f "$PROJECT_ROOT/bin/server" ] || [ ! -f "$PROJECT_ROOT/bin/worker" ]; then
            print_error "Binaries not found in bin/ directory"
            print_warning "Please build binaries first or remove --skip-build flag"
            exit 1
        fi
        print_success "Using existing binaries"
        return
    fi
    
    print_step "Building Binaries"
    
    cd "$PROJECT_ROOT"
    
    # Run the build script
    if [ -f "$PROJECT_ROOT/build-binaries.sh" ]; then
        bash "$PROJECT_ROOT/build-binaries.sh"
    else
        print_warning "build-binaries.sh not found, building manually..."
        mkdir -p bin
        
        export CGO_ENABLED=0
        export GOOS=linux
        export GOARCH=amd64
        
        print_success "Building server binary..."
        go build -a -installsuffix cgo -ldflags "-w -s" -o bin/server ./cmd/server/main.go
        
        print_success "Building worker binary..."
        go build -a -installsuffix cgo -ldflags "-w -s" -o bin/worker ./cmd/worker/main.go
        
        chmod +x bin/server bin/worker
    fi
    
    if [ ! -f "$PROJECT_ROOT/bin/server" ] || [ ! -f "$PROJECT_ROOT/bin/worker" ]; then
        print_error "Binary build failed"
        exit 1
    fi
    
    print_success "Binaries built successfully"
    print_success "Server: $(du -h "$PROJECT_ROOT/bin/server" | cut -f1)"
    print_success "Worker: $(du -h "$PROJECT_ROOT/bin/worker" | cut -f1)"
}

# Function to copy files to EC2
copy_files_to_ec2() {
    print_step "Copying Files to EC2"
    
    # Create app directory on EC2 if it doesn't exist
    ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "mkdir -p $APP_DIR" || true
    
    # Files and directories to copy
    print_success "Copying project files..."
    
    # Use rsync if available, otherwise use scp
    if command -v rsync &> /dev/null; then
        rsync -avz --progress \
            -e "ssh -i $SSH_KEY -o StrictHostKeyChecking=no" \
            --exclude='.git' \
            --exclude='node_modules' \
            --exclude='tmp' \
            --exclude='*.log' \
            --exclude='.env' \
            --exclude='bin/server' \
            --exclude='bin/worker' \
            "$PROJECT_ROOT/" "$EC2_USER@$EC2_IP:$APP_DIR/"
        
        # Copy binaries separately
        print_success "Copying binaries..."
        rsync -avz --progress \
            -e "ssh -i $SSH_KEY -o StrictHostKeyChecking=no" \
            "$PROJECT_ROOT/bin/" "$EC2_USER@$EC2_IP:$APP_DIR/bin/"
    else
        print_warning "rsync not found, using scp (slower)..."
        
        # Create necessary directories on EC2
        ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "mkdir -p $APP_DIR/bin $APP_DIR/cmd $APP_DIR/internal $APP_DIR/pkg $APP_DIR/config $APP_DIR/scripts"
        
        # Copy essential files
        scp -i "$SSH_KEY" -r "$PROJECT_ROOT/bin" "$EC2_USER@$EC2_IP:$APP_DIR/"
        scp -i "$SSH_KEY" "$PROJECT_ROOT/Dockerfile.server" "$EC2_USER@$EC2_IP:$APP_DIR/"
        scp -i "$SSH_KEY" "$PROJECT_ROOT/Dockerfile.worker" "$EC2_USER@$EC2_IP:$APP_DIR/"
        scp -i "$SSH_KEY" "$PROJECT_ROOT/docker-compose.yml" "$EC2_USER@$EC2_IP:$APP_DIR/"
        scp -i "$SSH_KEY" "$PROJECT_ROOT/go.mod" "$EC2_USER@$EC2_IP:$APP_DIR/"
        scp -i "$SSH_KEY" "$PROJECT_ROOT/go.sum" "$EC2_USER@$EC2_IP:$APP_DIR/"
        scp -i "$SSH_KEY" -r "$PROJECT_ROOT/cmd" "$EC2_USER@$EC2_IP:$APP_DIR/"
        scp -i "$SSH_KEY" -r "$PROJECT_ROOT/internal" "$EC2_USER@$EC2_IP:$APP_DIR/"
        scp -i "$SSH_KEY" -r "$PROJECT_ROOT/pkg" "$EC2_USER@$EC2_IP:$APP_DIR/"
    fi
    
    # Copy .env file if it exists (with warning)
    if [ -f "$PROJECT_ROOT/$ENV_FILE" ]; then
        print_success "Copying .env file..."
        scp -i "$SSH_KEY" "$PROJECT_ROOT/$ENV_FILE" "$EC2_USER@$EC2_IP:$APP_DIR/.env"
    else
        print_warning ".env file not found. Make sure to create it on EC2: $APP_DIR/.env"
    fi
    
    # Copy deployment script to EC2
    print_success "Copying deployment script..."
    scp -i "$SSH_KEY" "$PROJECT_ROOT/scripts/deploy-on-ec2.sh" "$EC2_USER@$EC2_IP:$APP_DIR/deploy-on-ec2.sh"
    ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "chmod +x $APP_DIR/deploy-on-ec2.sh"
    
    print_success "Files copied successfully"
}

# Function to deploy on EC2
deploy_on_ec2() {
    print_step "Deploying on EC2"
    
    # Execute deployment script on EC2
    ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "cd $APP_DIR && bash deploy-on-ec2.sh --skip-docker-build=$SKIP_DOCKER_BUILD" || {
        print_error "Deployment failed on EC2"
        exit 1
    }
    
    print_success "Deployment completed on EC2"
}

# Function to verify deployment
verify_deployment() {
    print_step "Verifying Deployment"
    
    # Check if containers are running
    print_success "Checking container status..."
    ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "cd $APP_DIR && (command -v docker-compose &> /dev/null && docker-compose ps || docker compose ps)"
    
    # Wait a bit for services to start
    sleep 5
    
    # Check server health
    print_success "Checking server health..."
    if ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "curl -f http://localhost:3000/api/health &>/dev/null"; then
        print_success "Server health check passed"
    else
        print_warning "Server health check failed (may need more time to start)"
        print_warning "Check logs: ssh -i $SSH_KEY $EC2_USER@$EC2_IP 'cd $APP_DIR && docker-compose logs server'"
    fi
    
    # Check if containers are running
    SERVER_RUNNING=$(ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "docker ps --filter 'name=ar-13-server' --format '{{.Names}}' | grep -q ar-13-server && echo 'yes' || echo 'no'")
    WORKER_RUNNING=$(ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "docker ps --filter 'name=ar-13-worker' --format '{{.Names}}' | grep -q ar-13-worker && echo 'yes' || echo 'no'")
    REDIS_RUNNING=$(ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "docker ps --filter 'name=ar-13-redis' --format '{{.Names}}' | grep -q ar-13-redis && echo 'yes' || echo 'no'")
    
    if [ "$SERVER_RUNNING" = "yes" ]; then
        print_success "Server container is running"
    else
        print_error "Server container is not running"
    fi
    
    if [ "$WORKER_RUNNING" = "yes" ]; then
        print_success "Worker container is running"
    else
        print_error "Worker container is not running"
    fi
    
    if [ "$REDIS_RUNNING" = "yes" ]; then
        print_success "Redis container is running"
    else
        print_error "Redis container is not running"
    fi
}

# Main deployment flow
main() {
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║     AR-13 Backend Deployment Script for EC2             ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo -e "${BLUE}Configuration:${NC}"
    echo "  EC2 IP: $EC2_IP"
    echo "  EC2 User: $EC2_USER"
    echo "  SSH Key: $SSH_KEY"
    echo "  App Directory: $APP_DIR"
    echo "  Skip Build: $SKIP_BUILD"
    echo "  Skip Docker Build: $SKIP_DOCKER_BUILD"
    echo ""
    
    check_prerequisites
    build_binaries
    copy_files_to_ec2
    deploy_on_ec2
    verify_deployment
    
    print_step "Deployment Complete!"
    echo -e "${GREEN}✓ Deployment successful!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Check logs: ssh -i $SSH_KEY $EC2_USER@$EC2_IP 'cd $APP_DIR && (command -v docker-compose &> /dev/null && docker-compose logs -f || docker compose logs -f)'"
    echo "  2. Check status: ssh -i $SSH_KEY $EC2_USER@$EC2_IP 'cd $APP_DIR && (command -v docker-compose &> /dev/null && docker-compose ps || docker compose ps)'"
    echo "  3. Restart services: ssh -i $SSH_KEY $EC2_USER@$EC2_IP 'cd $APP_DIR && (command -v docker-compose &> /dev/null && docker-compose restart || docker compose restart)'"
    echo ""
}

# Run main function
main

