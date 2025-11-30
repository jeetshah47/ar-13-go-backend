# Quick Start: Cloudflare Tunnel for API Backend

Get your API backend exposed via Cloudflare tunnel in 5 minutes.

## Prerequisites

- [ ] Cloudflare account with domain
- [ ] Docker and Docker Compose installed
- [ ] API backend built and ready

## 2-Minute Setup

### Step 1: Run Setup Script

```bash
cd cloudflared
chmod +x setup-api-tunnel.sh
./setup-api-tunnel.sh
```

When prompted:
- **Tunnel name**: `ar13-api-tunnel` (or your choice)
- **Hostname**: `api.yourdomain.com` (your API subdomain)

The script will:
- ✅ Authenticate with Cloudflare
- ✅ Create tunnel
- ✅ Generate credentials.json
- ✅ Create DNS route
- ✅ Generate config.yml

### Step 2: Verify Config

Check `cloudflared/config.yml` - it should use `server:3000`:

```yaml
ingress:
  - hostname: api.yourdomain.com
    service: http://server:3000
```

### Step 3: Start Services

```bash
# Start all services (including cloudflared)
docker-compose up -d

# Check logs
docker-compose logs -f cloudflared
```

### Step 4: Verify

```bash
# Check tunnel is running
docker ps | grep cloudflared

# Test API
curl https://api.yourdomain.com/api/health
```

## Done! 🎉

Your API is now accessible at:
- `https://api.yourdomain.com/api/*`
- `https://api.yourdomain.com/ws` (WebSocket)
- `https://api.yourdomain.com/uploads/*` (File uploads)

## Troubleshooting

### Tunnel not connecting?
```bash
# Check credentials
ls -la cloudflared/credentials.json

# Check logs
docker logs ar-13-cloudflared
```

### 502 Bad Gateway?
```bash
# Check backend is running
docker-compose ps
curl http://localhost:3000/api/health

# Verify service name in config.yml matches docker-compose service name
```

### DNS not working?
- Check Cloudflare dashboard → DNS → Records
- Should see CNAME for `api` pointing to tunnel

## Next Steps

- Set up Cloudflare Access for authentication
- Configure rate limiting
- Set up WAF rules
- Monitor tunnel status in Cloudflare dashboard

## Full Documentation

See `API_TUNNEL_SETUP.md` for detailed documentation.

