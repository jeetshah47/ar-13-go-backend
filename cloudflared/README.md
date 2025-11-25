# Cloudflare Tunnel Setup for FileBrowser Service

This directory contains the configuration for exposing the filebrowser service through a Cloudflare tunnel.

## Prerequisites

1. **Cloudflare Account**: You need a Cloudflare account with a domain
2. **Cloudflare Tunnel**: Create a tunnel in your Cloudflare dashboard

## Setup Steps

### Quick Setup (Recommended)

The setup script uses Docker to run cloudflared - no local installation required!

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
- Use Docker to run cloudflared (no local installation needed)
- Authenticate with Cloudflare (opens browser)
- Create the tunnel
- Generate credentials.json
- Create DNS route
- Generate config.yml

### Manual Setup (Alternative)

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. Go to **Zero Trust** → **Networks** → **Tunnels**
3. Click **Create a tunnel**
4. Choose **Cloudflared** as the connector
5. Give your tunnel a name (e.g., `filebrowser-tunnel`)
6. Copy the **Tunnel ID** and **Tunnel Secret** (or download credentials.json)

### Step 2: Configure the Tunnel

#### Option A: Using credentials.json (Recommended)

1. Download the `credentials.json` file from Cloudflare dashboard
2. Place it in this directory (`cloudflared/credentials.json`)
3. The file should look like:
   ```json
   {
     "AccountTag": "your-account-tag",
     "TunnelSecret": "your-tunnel-secret",
     "TunnelID": "your-tunnel-id",
     "TunnelName": "filebrowser-tunnel"
   }
   ```

#### Option B: Using Tunnel Token

1. In Cloudflare dashboard, go to your tunnel settings
2. Generate a **Tunnel Token**
3. Add it to your `.env` file:
   ```env
   CLOUDFLARE_TUNNEL_TOKEN=your-tunnel-token-here
   ```
4. Uncomment the `TUNNEL_TOKEN` line in `docker-compose.filebrowser-cloudflared.yml`

### Step 3: Configure Domain and Hostname

1. In Cloudflare dashboard, go to your tunnel → **Public Hostnames**
2. Add a new public hostname:
   - **Subdomain**: e.g., `filebrowser`
   - **Domain**: Your domain (e.g., `example.com`)
   - **Service**: `http://filebrowser-service:8082`
   - **Path**: Leave empty (or set a path if needed)

3. Create a `.env` file in the project root (or update existing):
   ```env
   # Cloudflare Tunnel Configuration
   TUNNEL_ID=your-tunnel-id
   TUNNEL_HOSTNAME=filebrowser.yourdomain.com
   ```

### Step 4: Update config.yml

Update `cloudflared/config.yml` with your values:
- Replace `${TUNNEL_ID}` with your actual tunnel ID
- Replace `${TUNNEL_HOSTNAME}` with your configured hostname

Or use environment variable substitution (recommended).

### Step 5: Start the Services

```bash
# Start both filebrowser and cloudflared
docker-compose -f docker-compose.filebrowser-cloudflared.yml up -d

# View logs
docker-compose -f docker-compose.filebrowser-cloudflared.yml logs -f

# Check cloudflared status
docker logs cloudflared-tunnel
```

## Verification

1. Check that both containers are running:
   ```bash
   docker ps | grep -E "filebrowser-service|cloudflared-tunnel"
   ```

2. Access your filebrowser service via the Cloudflare tunnel:
   ```
   https://filebrowser.yourdomain.com
   ```

3. Check tunnel status in Cloudflare dashboard:
   - Go to **Zero Trust** → **Networks** → **Tunnels**
   - Your tunnel should show as **Active**

## Troubleshooting

### Tunnel Not Connecting

1. **Check credentials**: Ensure `credentials.json` is correct and in the right location
2. **Check config**: Verify `config.yml` has correct tunnel ID and hostname
3. **Check logs**: 
   ```bash
   docker logs cloudflared-tunnel
   ```

### Service Not Accessible

1. **Check filebrowser service**: 
   ```bash
   docker logs filebrowser-service
   docker exec filebrowser-service wget -q -O- http://localhost:8082/health
   ```

2. **Check network**: Ensure both services are on the same Docker network
   ```bash
   docker network inspect filebrowser-network
   ```

3. **Check Cloudflare DNS**: Ensure your hostname has a DNS record pointing to the tunnel

### Permission Issues

If you see permission errors:
1. Check file permissions on `credentials.json`:
   ```bash
   chmod 600 cloudflared/credentials.json
   ```

2. Ensure the mounted volume has correct permissions for filebrowser-service

## Security Considerations

1. **Credentials**: Never commit `credentials.json` to version control (already in `.gitignore`)
2. **Access Control**: Consider adding Cloudflare Access rules to protect your filebrowser
3. **HTTPS**: Cloudflare automatically provides HTTPS for your tunnel
4. **Rate Limiting**: Configure rate limiting in Cloudflare dashboard if needed

## Advanced Configuration

### Multiple Services

You can route multiple services through the same tunnel by updating `config.yml`:

```yaml
ingress:
  - hostname: filebrowser.yourdomain.com
    service: http://filebrowser-service:8082
  - hostname: api.yourdomain.com
    service: http://api-service:3000
  - service: http_status:404
```

### Path-based Routing

```yaml
ingress:
  - hostname: yourdomain.com
    path: /files/*
    service: http://filebrowser-service:8082
  - service: http_status:404
```

### Access Policies

In Cloudflare dashboard:
1. Go to **Zero Trust** → **Access** → **Applications**
2. Add your tunnel hostname as an application
3. Configure access policies (e.g., require email domain, MFA, etc.)

## Maintenance

### Update Cloudflared

```bash
docker-compose -f docker-compose.filebrowser-cloudflared.yml pull cloudflared
docker-compose -f docker-compose.filebrowser-cloudflared.yml up -d cloudflared
```

### Restart Services

```bash
docker-compose -f docker-compose.filebrowser-cloudflared.yml restart
```

### Stop Services

```bash
docker-compose -f docker-compose.filebrowser-cloudflared.yml down
```

## References

- [Cloudflare Tunnel Documentation](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- [Cloudflared Docker Image](https://hub.docker.com/r/cloudflare/cloudflared)

