# Quick Deployment Guide

## One-Command Deployment

Deploy to EC2 with default settings:

```bash
./scripts/deploy-to-ec2.sh
```

## Custom Deployment

Deploy with custom EC2 IP and SSH key:

```bash
./scripts/deploy-to-ec2.sh \
  --ec2-ip 13.201.254.55 \
  --ec2-user ubuntu \
  --ssh-key ~/.ssh/your-key.pem
```

## First Time Setup on EC2

Before first deployment, SSH into EC2 and run:

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Log out and log back in
exit
```

## Environment Variables

Create `.env` file in project root:

```env
PORT=3000
NODE_ENV=production
MONGODB_URI=your-mongodb-uri
MONGODB_DATABASE=ar13_backend
JWT_SECRET=your-strong-secret-min-32-chars
REDIS_ADDR=redis:6379
FRONTEND_URL=https://your-frontend.com
```

## Common Commands

### View Logs
```bash
ssh -i ~/.ssh/your-key.pem ubuntu@13.201.254.55 \
  'cd ~/ar-13-go-backend && docker-compose logs -f'
```

### Restart Services
```bash
ssh -i ~/.ssh/your-key.pem ubuntu@13.201.254.55 \
  'cd ~/ar-13-go-backend && docker-compose restart'
```

### Check Status
```bash
ssh -i ~/.ssh/your-key.pem ubuntu@13.201.254.55 \
  'cd ~/ar-13-go-backend && docker-compose ps'
```

## Troubleshooting

**Connection failed?**
- Check security group allows SSH (port 22)
- Verify SSH key: `chmod 400 ~/.ssh/your-key.pem`
- Check EC2 instance is running

**Docker permission denied?**
- Add user to docker group: `sudo usermod -aG docker $USER`
- Log out and log back in

**Container not starting?**
- Check logs: `docker-compose logs server`
- Verify `.env` file exists
- Check binaries: `ls -la bin/`

For detailed documentation, see: `docs/deployment/DOCKER_DEPLOYMENT.md`

