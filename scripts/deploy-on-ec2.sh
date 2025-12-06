#!/bin/bash

# EC2-side deployment script
# This script is executed on the EC2 instance to deploy the application
# It should be run from the application directory

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SKIP_DOCKER_BUILD=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-docker-build)
            SKIP_DOCKER_BUILD=true
            shift
            ;;
        *)
            shift
            ;;
    esac
done

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

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$SCRIPT_DIR"

# Change to application directory
cd "$APP_DIR" || {
    print_error "Failed to change to application directory: $APP_DIR"
    exit 1
}

print_step "AR-13 Backend Deployment on EC2"
echo "Application directory: $APP_DIR"
echo "Skip Docker build: $SKIP_DOCKER_BUILD"
echo ""

# Check if .env file exists
if [ ! -f "$APP_DIR/.env" ]; then
    print_warning ".env file not found in $APP_DIR"
    print_warning "Please create .env file with required environment variables"
    print_warning "You can copy from .env.example if available"
fi

# Check if binaries exist
if [ ! -f "$APP_DIR/bin/server" ] || [ ! -f "$APP_DIR/bin/worker" ]; then
    print_error "Binaries not found in bin/ directory"
    print_error "Expected: bin/server and bin/worker"
    exit 1
fi
print_success "Binaries found"

# Check if Docker is running
if ! docker info &>/dev/null; then
    print_error "Docker is not running or user doesn't have permission"
    print_warning "Try: sudo systemctl start docker"
    print_warning "Or add user to docker group: sudo usermod -aG docker $USER"
    exit 1
fi
print_success "Docker is running"

# Check if docker-compose is available
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE_CMD="docker compose"
else
    print_error "Docker Compose is not installed"
    exit 1
fi
print_success "Docker Compose found: $DOCKER_COMPOSE_CMD"

# Stop existing containers
print_step "Stopping Existing Containers"
$DOCKER_COMPOSE_CMD down || {
    print_warning "Some containers may not have been running"
}

# Remove old containers if they exist
docker ps -a --filter "name=ar-13" --format "{{.Names}}" | xargs -r docker rm -f || true
print_success "Old containers stopped and removed"

# Build Docker images
if [ "$SKIP_DOCKER_BUILD" = false ]; then
    print_step "Building Docker Images"
    
    # Build server image
    print_success "Building server image..."
    docker build -f Dockerfile.server -t ar-13-server:latest . || {
        print_error "Failed to build server image"
        exit 1
    }
    
    # Build worker image
    print_success "Building worker image..."
    docker build -f Dockerfile.worker -t ar-13-worker:latest . || {
        print_error "Failed to build worker image"
        exit 1
    }
    
    print_success "Docker images built successfully"
else
    print_warning "Skipping Docker build (using existing images)"
fi

# Create upload directory if it doesn't exist
mkdir -p "$APP_DIR/upload"
chmod 755 "$APP_DIR/upload" || true

# Start services with docker-compose
print_step "Starting Services"
$DOCKER_COMPOSE_CMD up -d || {
    print_error "Failed to start services"
    exit 1
}

print_success "Services started"

# Wait for services to be ready
print_step "Waiting for Services to Start"
sleep 5

# Check container status
print_step "Container Status"
$DOCKER_COMPOSE_CMD ps

# Check if containers are running
SERVER_RUNNING=$(docker ps --filter "name=ar-13-server" --format "{{.Names}}" | grep -q ar-13-server && echo "yes" || echo "no")
WORKER_RUNNING=$(docker ps --filter "name=ar-13-worker" --format "{{.Names}}" | grep -q ar-13-worker && echo "yes" || echo "no")
REDIS_RUNNING=$(docker ps --filter "name=ar-13-redis" --format "{{.Names}}" | grep -q ar-13-redis && echo "yes" || echo "no")

if [ "$SERVER_RUNNING" = "yes" ]; then
    print_success "Server container is running"
else
    print_error "Server container is not running"
    print_warning "Check logs: $DOCKER_COMPOSE_CMD logs server"
fi

if [ "$WORKER_RUNNING" = "yes" ]; then
    print_success "Worker container is running"
else
    print_error "Worker container is not running"
    print_warning "Check logs: $DOCKER_COMPOSE_CMD logs worker"
fi

if [ "$REDIS_RUNNING" = "yes" ]; then
    print_success "Redis container is running"
else
    print_error "Redis container is not running"
    print_warning "Check logs: $DOCKER_COMPOSE_CMD logs redis"
fi

# Health check
print_step "Health Check"
sleep 3

if curl -f http://localhost:3000/api/health &>/dev/null; then
    print_success "Server health check passed"
else
    print_warning "Server health check failed (may need more time)"
    print_warning "Check logs: $DOCKER_COMPOSE_CMD logs server"
fi

# Clean up old images (keep only latest)
print_step "Cleaning Up Old Images"
OLD_IMAGES=$(docker images "ar-13-*" --format "{{.ID}}" | tail -n +4)
if [ -n "$OLD_IMAGES" ]; then
    echo "$OLD_IMAGES" | xargs -r docker rmi -f || true
    print_success "Old images removed"
else
    print_success "No old images to clean up"
fi

# Show disk usage
print_step "Disk Usage"
df -h / | tail -1
docker system df

print_step "Deployment Complete!"
echo -e "${GREEN}✓ Deployment successful!${NC}"
echo ""
echo "Useful commands:"
echo "  View logs: $DOCKER_COMPOSE_CMD logs -f"
echo "  View server logs: $DOCKER_COMPOSE_CMD logs -f server"
echo "  View worker logs: $DOCKER_COMPOSE_CMD logs -f worker"
echo "  Restart services: $DOCKER_COMPOSE_CMD restart"
echo "  Stop services: $DOCKER_COMPOSE_CMD down"
echo "  Check status: $DOCKER_COMPOSE_CMD ps"
echo ""

