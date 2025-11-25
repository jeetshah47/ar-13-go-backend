# Quick Start: Cloudflare Tunnel for FileBrowser

This is a quick reference guide to get your filebrowser service exposed via Cloudflare tunnel.

## Prerequisites Checklist

- [ ] Cloudflare account with a domain
- [ ] Docker and Docker Compose installed
- [ ] Filebrowser service code ready

## 2-Minute Setup (Automated with Docker)

No local installation needed! The script uses Docker to run cloudflared.

### 1. Create Tunnel Automatically

**Windows:**
```powershell
cd cloudflared
.\setup-tunnel.ps1
```

**Linux/macOS:**
```bash
cd cloudflared
chmod +x setup-tunnel.sh
./setup-tunnel.sh
```

Or with parameters:
```bash
./setup-tunnel.sh filebrowser-tunnel filebrowser.yourdomain.com
```

The script will:
- Use Docker to run cloudflared (no installation needed)
- Prompt for tunnel name (default: `filebrowser-tunnel`)
- Prompt for hostname (e.g., `filebrowser.yourdomain.com`)
- Authenticate with Cloudflare (opens browser)
- Create tunnel and DNS route automatically
- Generate all configuration files

### 2. Start Services (1 minute)

**Windows:**
```powershell
docker-compose -f docker-compose.filebrowser-cloudflared.yml up -d
```

**NAS/Linux:**
```bash
docker-compose -f docker-compose.filebrowser-cloudflared.nas.yml up -d
```

### 5. Verify (30 seconds)

```bash
# Check containers are running
docker ps | grep -E "filebrowser|cloudflared"

# Check logs
docker logs cloudflared-tunnel
docker logs filebrowser-service

# Access your service
# https://filebrowser.yourdomain.com
```

## Troubleshooting Quick Fixes

### Tunnel shows "Inactive" in Cloudflare

```bash
# Check credentials file exists and is readable
ls -la cloudflared/credentials.json

# Check config file syntax
docker run --rm -v $(pwd)/cloudflared:/etc/cloudflared cloudflare/cloudflared:latest tunnel --config /etc/cloudflared/config.yml ingress validate
```

### Can't access service

```bash
# Test filebrowser service directly (if you temporarily expose port)
docker exec filebrowser-service wget -q -O- http://localhost:8082/health

# Check network connectivity
docker network inspect filebrowser-network
```

### Permission denied

```bash
# Fix credentials file permissions
chmod 600 cloudflared/credentials.json
```

## Next Steps

- [ ] Set up Cloudflare Access for authentication
- [ ] Configure rate limiting
- [ ] Set up monitoring/alerting
- [ ] Review security settings

## Need Help?

See the full documentation in `cloudflared/README.md`

