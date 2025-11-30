# Cloudflare Tunnel Quick Config for arback.jsdeveloper.cloud

## Complete Configuration

### 1. config.yml

```yaml
tunnel: YOUR_TUNNEL_ID_HERE
credentials-file: /etc/cloudflared/credentials.json

ingress:
  - hostname: arback.jsdeveloper.cloud
    service: http://server:3000
  - service: http_status:404
```

### 2. docker-compose.yml (cloudflared service)

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
  mem_limit: 128m
  mem_reservation: 50m
```

## Quick Setup

```bash
# 1. Run setup script
cd cloudflared
./setup-api-tunnel.sh ar13-api-tunnel arback.jsdeveloper.cloud

# 2. Uncomment cloudflared service in docker-compose.yml

# 3. Start services
docker-compose up -d

# 4. Test
curl https://arback.jsdeveloper.cloud/api/health
```

## Cloudflare Dashboard Settings

**DNS:**
- Type: CNAME (auto-created by tunnel)
- Name: `arback`
- Target: `<tunnel-id>.cfargotunnel.com`

**SSL/TLS:**
- Mode: Full (strict) or Full

**Tunnel:**
- Status: Should show "Active"
- Public Hostname: `arback.jsdeveloper.cloud` → `http://server:3000`

