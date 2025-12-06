# Cloudflare Tunnel Setup for NAS (Container Station)

This guide explains how to set up Cloudflare tunnel for filebrowser service on your NAS using Container Station.

## Prerequisites

1. **NAS with Container Station** (QNAP or Synology)
2. **Cloudflare Account** with a domain
3. **SSH access** to your NAS (for running setup script)

## Step 1: Setup Cloudflare Tunnel on Your Computer

First, set up the tunnel on your local computer (Windows/Mac/Linux):

1. **Install cloudflared** on your computer:
   - Windows: `winget install Cloudflare.cloudflared`
   - Mac: `brew install cloudflared`
   - Linux: Download from [GitHub](https://github.com/cloudflare/cloudflared/releases)

2. **Run the setup script**:
   ```powershell
   # Windows
   cd cloudflared
   .\setup-tunnel.ps1
   ```
   
   Or manually:
   ```bash
   # Authenticate
   cloudflared tunnel login
   
   # Create tunnel
   cloudflared tunnel create filebrowser-tunnel
   
   # Create DNS route
   cloudflared tunnel route dns filebrowser-tunnel filebrowser.yourdomain.com
   ```

3. **Copy files to NAS**:
   After setup, copy these files from your computer to your NAS:
   - `~/.cloudflared/cert.pem` → NAS `cloudflared/cert.pem`
   - `~/.cloudflared/<tunnel-id>.json` → NAS `cloudflared/credentials.json`
   - Create `cloudflared/config.yml` (see example below)

## Step 2: Prepare Files on NAS

1. **Create cloudflared directory** on your NAS:
   ```bash
   mkdir -p /share/Container/cloudflared
   ```

2. **Copy required files**:
   - `cert.pem` - Origin certificate
   - `credentials.json` - Tunnel credentials
   - `config.yml` - Tunnel configuration

3. **Create config.yml**:
   ```yaml
   tunnel: YOUR_TUNNEL_ID
   credentials-file: /etc/cloudflared/credentials.json
   origincert: /etc/cloudflared/cert.pem
   
   ingress:
     - hostname: filebrowser.yourdomain.com
       service: http://filebrowser-service:8082
     - service: http_status:404
   ```

## Step 3: Deploy in Container Station

### Option A: Using Docker Compose (Recommended)

1. **Upload docker-compose file** to your NAS:
   - Upload `docker-compose.filebrowser-cloudflared.nas.yml` to your NAS
   - Place it in the same directory as the `cloudflared/` folder

2. **In Container Station**:
   - Go to **Container Station** → **Compose**
   - Click **Create** → **Upload YAML**
   - Select `docker-compose.filebrowser-cloudflared.nas.yml`
   - Click **Create**

### Option B: Manual Container Creation

1. **Create filebrowser-service container**:
   - Image: Build from `../ar-13-nas-filebrowser/Dockerfile`
   - Container name: `filebrowser-service`
   - Network: Create new network `filebrowser-network`
   - Volume: Mount `/share/studio work/` to `/data`
   - Environment:
     - `DATA_ROOT=/data`
     - `PORT=8082`
     - `TZ=Asia/Kolkata`

2. **Create cloudflared-tunnel container**:
   - Image: `cloudflare/cloudflared:latest`
   - Container name: `cloudflared-tunnel`
   - Network: Same `filebrowser-network`
   - Volume: Mount `./cloudflared` to `/etc/cloudflared`
   - Command: `tunnel run`
   - Depends on: `filebrowser-service`

## Step 4: Verify Setup

1. **Check containers are running**:
   ```bash
   docker ps | grep -E "filebrowser|cloudflared"
   ```

2. **Check cloudflared logs**:
   ```bash
   docker logs cloudflared-tunnel
   ```
   Look for: "Registered tunnel connection"

3. **Check filebrowser logs**:
   ```bash
   docker logs filebrowser-service
   ```
   Look for: "File Browser Service is running on port 8082"

4. **Test your service**:
   Open browser: `https://filebrowser.yourdomain.com/health`

## File Structure on NAS

```
/share/Container/
├── docker-compose.filebrowser-cloudflared.nas.yml
└── cloudflared/
    ├── config.yml
    ├── credentials.json
    └── cert.pem
```

## Troubleshooting

### Tunnel Not Connecting

1. **Check cert.pem exists**:
   ```bash
   ls -la /share/Container/cloudflared/cert.pem
   ```

2. **Check credentials.json**:
   ```bash
   cat /share/Container/cloudflared/credentials.json
   ```

3. **Check config.yml**:
   ```bash
   cat /share/Container/cloudflared/config.yml
   ```

4. **Verify tunnel ID matches**:
   - Tunnel ID in `config.yml` should match the one in `credentials.json`

### Service Not Accessible

1. **Check DNS route**:
   - Verify DNS route was created: `cloudflared tunnel route dns list`
   - Check Cloudflare dashboard → DNS → Records

2. **Check network connectivity**:
   ```bash
   docker exec cloudflared-tunnel ping filebrowser-service
   ```

3. **Test filebrowser directly**:
   ```bash
   docker exec filebrowser-service wget -q -O- http://localhost:8082/health
   ```

### Permission Issues

If you see permission errors:

1. **Check file permissions**:
   ```bash
   chmod 600 /share/Container/cloudflared/credentials.json
   chmod 600 /share/Container/cloudflared/cert.pem
   ```

2. **Check volume mount permissions**:
   - Ensure the mounted volume has correct permissions
   - Container runs as UID 1000, GID 1000

## Updating Configuration

1. **Edit config.yml** on NAS
2. **Restart cloudflared container**:
   ```bash
   docker restart cloudflared-tunnel
   ```

## Maintenance

### Update Cloudflared

```bash
docker pull cloudflare/cloudflared:latest
docker-compose -f docker-compose.filebrowser-cloudflared.nas.yml up -d cloudflared
```

### View Logs

```bash
# Cloudflared logs
docker logs -f cloudflared-tunnel

# Filebrowser logs
docker logs -f filebrowser-service
```

### Stop Services

```bash
docker-compose -f docker-compose.filebrowser-cloudflared.nas.yml down
```

## Notes

- The filebrowser service is **not exposed** to the host - it's only accessible via Cloudflare tunnel
- All traffic goes through Cloudflare, providing HTTPS automatically
- The tunnel provides secure access without opening ports on your NAS
- Update the volume path `/share/studio work/` to match your NAS shared folder structure

