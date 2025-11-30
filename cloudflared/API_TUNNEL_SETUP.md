# Cloudflare Tunnel Setup for AR-13 API Backend

This guide explains how to set up a Cloudflare tunnel to expose your AR-13 API backend through your domain.

## Prerequisites

1. **Cloudflare Account** with a domain
2. **Docker and Docker Compose** installed
3. **API backend running** in Docker containers

## Quick Setup (Automated)

### Step 1: Run Setup Script

```bash
cd cloudflared
chmod +x setup-api-tunnel.sh
./setup-api-tunnel.sh
```

Or with parameters:

```bash
./setup-api-tunnel.sh ar13-api-tunnel api.yourdomain.com
```

The script will:
- Authenticate with Cloudflare (opens browser)
- Create a tunnel
- Generate `credentials.json`
- Create DNS route
- Generate `config.yml`

### Step 2: Update Config for Docker Network

After running the setup script, update `cloudflared/config.yml` to use the Docker service name:

```yaml
tunnel: YOUR_TUNNEL_ID
credentials-file: /etc/cloudflared/credentials.json

ingress:
  # Route all API traffic to Docker backend service
  - hostname: api.yourdomain.com
    service: http://server:3000
  
  # Catch-all rule (must be last)
  - service: http_status:404
```

**Important**: Use `server:3000` (Docker service name) instead of `localhost:3000` when running in Docker Compose.

### Step 3: Start Services

```bash
# Start all services including cloudflared
docker-compose up -d

# View logs
docker-compose logs -f cloudflared
```

## Manual Setup

### Step 1: Create Tunnel in Cloudflare Dashboard

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. Navigate to **Zero Trust** → **Networks** → **Tunnels**
3. Click **Create a tunnel**
4. Choose **Cloudflared** connector
5. Name it (e.g., `ar13-api-tunnel`)
6. Download `credentials.json` or copy the tunnel token

### Step 2: Configure Tunnel

#### Option A: Using credentials.json

1. Place `credentials.json` in `cloudflared/` directory
2. Create `cloudflared/config.yml`:

```yaml
tunnel: YOUR_TUNNEL_ID
credentials-file: /etc/cloudflared/credentials.json

ingress:
  - hostname: api.yourdomain.com
    service: http://server:3000
  - service: http_status:404
```

#### Option B: Using Tunnel Token

1. Get tunnel token from Cloudflare dashboard
2. Update `docker-compose.yml`:

```yaml
cloudflared:
  # ... other config ...
  command: tunnel --token ${CLOUDFLARE_TUNNEL_TOKEN} run
  # Remove volumes for credentials.json
```

3. Add to `.env`:
```env
CLOUDFLARE_TUNNEL_TOKEN=your-tunnel-token-here
```

### Step 3: Configure DNS Route

In Cloudflare dashboard:
1. Go to your tunnel → **Public Hostnames**
2. Add hostname:
   - **Subdomain**: `api`
   - **Domain**: `yourdomain.com`
   - **Service**: `http://server:3000` (or `http://localhost:3000` if using host network)
   - **Path**: Leave empty

Or use CLI:
```bash
docker run --rm \
  -v ~/.cloudflared:/etc/cloudflared \
  cloudflare/cloudflared:latest \
  tunnel route dns ar13-api-tunnel api.yourdomain.com
```

## Configuration Options

### Multiple Hostnames

```yaml
ingress:
  - hostname: api.yourdomain.com
    service: http://server:3000
  - hostname: api-staging.yourdomain.com
    service: http://server:3000
  - service: http_status:404
```

### Path-based Routing

```yaml
ingress:
  - hostname: yourdomain.com
    path: /api/*
    service: http://server:3000
  - hostname: yourdomain.com
    path: /uploads/*
    service: http://server:3000
  - service: http_status:404
```

### WebSocket Support

WebSocket is automatically supported. No special configuration needed.

## Docker Compose Integration

The `docker-compose.yml` includes a cloudflared service:

```yaml
cloudflared:
  image: cloudflare/cloudflared:latest
  container_name: ar-13-cloudflared
  restart: unless-stopped
  command: tunnel --config /etc/cloudflared/config.yml run
  volumes:
    - ./cloudflared/config.yml:/etc/cloudflared/config.yml:ro
    - ./cloudflared/credentials.json:/etc/cloudflared/credentials.json:ro
  networks:
    - ar-13-network
  depends_on:
    - server
```

## Verification

### Check Tunnel Status

```bash
# Check container is running
docker ps | grep cloudflared

# Check logs
docker logs ar-13-cloudflared

# Check in Cloudflare dashboard
# Zero Trust → Networks → Tunnels → Your tunnel should show "Active"
```

### Test API Access

```bash
# Test health endpoint
curl https://api.yourdomain.com/api/health

# Test API endpoint
curl https://api.yourdomain.com/api/auth/login
```

## Troubleshooting

### Tunnel Not Connecting

1. **Check credentials**:
   ```bash
   ls -la cloudflared/credentials.json
   cat cloudflared/credentials.json
   ```

2. **Check config syntax**:
   ```bash
   docker run --rm \
     -v $(pwd)/cloudflared:/etc/cloudflared \
     cloudflare/cloudflared:latest \
     tunnel --config /etc/cloudflared/config.yml ingress validate
   ```

3. **Check logs**:
   ```bash
   docker logs ar-13-cloudflared
   ```

### Service Not Accessible

1. **Check backend is running**:
   ```bash
   docker-compose ps
   curl http://localhost:3000/api/health
   ```

2. **Check network connectivity**:
   ```bash
   docker network inspect ar-13-network
   docker exec ar-13-cloudflared ping server
   ```

3. **Verify config service name**:
   - Use `server:3000` (Docker service name) when in Docker network
   - Use `localhost:3000` if using `network_mode: host`

### DNS Issues

1. **Check DNS record**:
   ```bash
   dig api.yourdomain.com
   nslookup api.yourdomain.com
   ```

2. **Verify in Cloudflare dashboard**:
   - DNS → Records → Should see CNAME for `api` pointing to tunnel

### 502 Bad Gateway

1. **Backend not accessible**:
   - Check if server container is running
   - Verify service name in config.yml matches docker-compose service name
   - Check if port 3000 is correct

2. **Network issues**:
   - Ensure cloudflared is on same network as server
   - Verify `depends_on: server` in docker-compose

## Security Considerations

1. **Credentials**: Never commit `credentials.json` to git (already in `.gitignore`)
2. **Access Control**: Consider Cloudflare Access for authentication
3. **Rate Limiting**: Configure in Cloudflare dashboard
4. **WAF Rules**: Set up Web Application Firewall rules
5. **HTTPS**: Automatically provided by Cloudflare

## Environment Variables

Add to `.env` if using tunnel token:

```env
# Cloudflare Tunnel
CLOUDFLARE_TUNNEL_TOKEN=your-tunnel-token-here
CLOUDFLARE_TUNNEL_ID=your-tunnel-id
CLOUDFLARE_API_HOSTNAME=api.yourdomain.com
```

## Updating Configuration

After updating `config.yml`:

```bash
# Restart cloudflared
docker-compose restart cloudflared

# Or rebuild
docker-compose up -d --force-recreate cloudflared
```

## Multiple Environments

You can run multiple tunnels for different environments:

```yaml
# config-production.yml
ingress:
  - hostname: api.yourdomain.com
    service: http://server:3000

# config-staging.yml
ingress:
  - hostname: api-staging.yourdomain.com
    service: http://server:3000
```

Then use different compose files or override config:

```bash
docker run -v $(pwd)/cloudflared/config-production.yml:/etc/cloudflared/config.yml ...
```

## References

- [Cloudflare Tunnel Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- [Cloudflared Docker Image](https://hub.docker.com/r/cloudflare/cloudflared)
- [Ingress Rules](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/configuration/ingress/)

