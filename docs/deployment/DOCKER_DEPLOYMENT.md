# Docker Deployment Guide for EC2

This guide explains how to deploy the AR-13 backend (server and worker microservices) to an EC2 instance using Docker.

## Overview

The deployment process uses two scripts:
1. **`scripts/deploy-to-ec2.sh`** - Runs on your local machine, builds binaries, and copies files to EC2
2. **`scripts/deploy-on-ec2.sh`** - Runs on EC2, builds Docker images and starts containers

## Prerequisites

### Local Machine
- Go 1.21+ installed
- SSH access to EC2 instance
- SSH key with access to EC2
- `rsync` (optional, but recommended for faster file transfers)

### EC2 Instance
- Ubuntu 20.04+ or similar Linux distribution
- Docker installed and running
- Docker Compose installed
- User added to docker group (to run docker without sudo)
- Port 3000 open in security group (if accessing directly)

## Quick Start

### 1. First Time Setup on EC2

SSH into your EC2 instance and install Docker:

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Log out and log back in for group changes to take effect
exit
```

### 2. Configure Environment Variables

On your local machine, create a `.env` file in the project root:

```bash
cp .env.example .env
# Edit .env with your configuration
```

**Important variables:**
- `JWT_SECRET` - Must be at least 32 characters
- `MONGODB_URI` - Your MongoDB connection string
- `REDIS_ADDR` - Set to `redis:6379` (uses Docker service name)
- `FRONTEND_URL` - Your frontend URL

### 3. Deploy

Run the deployment script from your local machine:

```bash
./scripts/deploy-to-ec2.sh
```

Or with custom options:

```bash
./scripts/deploy-to-ec2.sh \
  --ec2-ip 13.201.254.55 \
  --ec2-user ubuntu \
  --ssh-key ~/.ssh/my-key.pem
```

## Deployment Script Options

### `deploy-to-ec2.sh` Options

```bash
./scripts/deploy-to-ec2.sh [OPTIONS]

Options:
  --ec2-ip IP              EC2 instance IP (default: 13.201.254.55)
  --ec2-user USER          EC2 user (default: ubuntu)
  --ssh-key PATH           Path to SSH private key (default: ~/.ssh/id_rsa)
  --skip-build             Skip building binaries (use existing)
  --skip-docker-build      Skip Docker image build on EC2 (use existing images)
  --app-dir PATH           Application directory on EC2 (default: ~/ar-13-go-backend)
  --env-file PATH          Path to .env file to copy (default: .env)
  --help                   Show help message
```

## What the Deployment Does

1. **Checks Prerequisites**
   - Verifies Go is installed
   - Tests SSH connection to EC2
   - Checks Docker is installed on EC2

2. **Builds Binaries**
   - Builds `bin/server` and `bin/worker` for Linux amd64
   - Uses optimized build flags

3. **Copies Files to EC2**
   - Copies project files (excluding .git, node_modules, etc.)
   - Copies binaries
   - Copies `.env` file
   - Copies deployment script

4. **Deploys on EC2**
   - Stops existing containers
   - Builds Docker images (if not skipped)
   - Starts containers with docker-compose
   - Performs health checks

5. **Verifies Deployment**
   - Checks container status
   - Tests server health endpoint
   - Shows useful commands

## Manual Deployment Steps

If you prefer to deploy manually:

### On Local Machine

```bash
# Build binaries
./build-binaries.sh

# Copy files to EC2
rsync -avz --progress \
  -e "ssh -i ~/.ssh/your-key.pem" \
  --exclude='.git' \
  --exclude='node_modules' \
  --exclude='tmp' \
  ./ ubuntu@13.201.254.55:~/ar-13-go-backend/
```

### On EC2 Instance

```bash
cd ~/ar-13-go-backend

# Build Docker images
docker build -f Dockerfile.server -t ar-13-server:latest .
docker build -f Dockerfile.worker -t ar-13-worker:latest .

# Start services
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f
```

## Managing Services

### View Logs

```bash
# All services
ssh -i ~/.ssh/your-key.pem ubuntu@13.201.254.55 \
  'cd ~/ar-13-go-backend && docker-compose logs -f'

# Specific service
ssh -i ~/.ssh/your-key.pem ubuntu@13.201.254.55 \
  'cd ~/ar-13-go-backend && docker-compose logs -f server'
```

### Restart Services

```bash
ssh -i ~/.ssh/your-key.pem ubuntu@13.201.254.55 \
  'cd ~/ar-13-go-backend && docker-compose restart'
```

### Stop Services

```bash
ssh -i ~/.ssh/your-key.pem ubuntu@13.201.254.55 \
  'cd ~/ar-13-go-backend && docker-compose down'
```

### Update and Redeploy

```bash
# On local machine
git pull
./scripts/deploy-to-ec2.sh
```

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to EC2 via SSH

**Solutions:**
- Check security group allows SSH (port 22) from your IP
- Verify SSH key permissions: `chmod 400 ~/.ssh/your-key.pem`
- Check EC2 instance is running
- Verify IP address is correct

### Docker Permission Issues

**Problem:** Permission denied when running docker commands

**Solutions:**
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and log back in
exit
```

### Container Not Starting

**Problem:** Container exits immediately

**Solutions:**
```bash
# Check logs
docker-compose logs server
docker-compose logs worker

# Check .env file exists and is valid
cat .env

# Verify binaries exist
ls -la bin/
```

### Health Check Failing

**Problem:** Server health endpoint returns error

**Solutions:**
- Check server logs: `docker-compose logs server`
- Verify PORT environment variable matches docker-compose port mapping
- Check if port 3000 is accessible: `curl http://localhost:3000/api/health`
- Verify MongoDB and Redis connections

### Out of Disk Space

**Problem:** Docker build fails due to disk space

**Solutions:**
```bash
# Clean up Docker
docker system prune -a

# Check disk space
df -h

# Remove old images
docker images | grep ar-13 | tail -n +4 | awk '{print $3}' | xargs docker rmi
```

## Environment Variables Reference

See `.env.example` for all available environment variables.

### Required Variables

- `JWT_SECRET` - Secret key for JWT tokens (min 32 chars)
- `MONGODB_URI` - MongoDB connection string
- `MONGODB_DATABASE` - Database name

### Important for Docker

- `REDIS_ADDR` - Should be `redis:6379` (Docker service name)
- `PORT` - Server port (default: 3000)

## Security Best Practices

1. **Never commit `.env` file** - It contains sensitive information
2. **Use strong JWT_SECRET** - Generate with: `openssl rand -base64 32`
3. **Restrict SSH access** - Only allow your IP in security group
4. **Keep Docker updated** - Regularly update Docker and images
5. **Use HTTPS** - Set up Nginx reverse proxy with SSL (see EC2_DEPLOYMENT_GUIDE.md)

## Next Steps

After successful deployment:

1. **Set up Nginx reverse proxy** (optional but recommended)
   - See `docs/deployment/EC2_DEPLOYMENT_GUIDE.md`

2. **Set up SSL certificate** (optional)
   - Use Let's Encrypt with Certbot

3. **Configure monitoring** (optional)
   - Set up CloudWatch alarms
   - Monitor container logs

4. **Set up backups** (optional)
   - Backup MongoDB data
   - Backup uploaded files

## Support

For issues or questions:
1. Check logs: `docker-compose logs -f`
2. Check container status: `docker-compose ps`
3. Verify environment variables: `cat .env`
4. Review this documentation

