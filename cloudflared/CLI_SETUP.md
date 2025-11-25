# Cloudflare Tunnel CLI Setup Guide

This guide explains how to create and manage Cloudflare tunnels using Docker to run `cloudflared`.

## Prerequisites

1. **Docker** (required)
   - Docker Desktop for Windows/Mac
   - Docker Engine for Linux
   - No need to install cloudflared locally!

2. **Cloudflare Account**
   - You need a Cloudflare account with a domain added
   - The domain should be using Cloudflare DNS

## Quick Setup

Use the automated scripts (uses Docker):
- **Windows**: `.\setup-tunnel.ps1`
- **Linux/macOS**: `./setup-tunnel.sh`

## Manual CLI Commands (Using Docker)

If you prefer to run commands manually using Docker:

### 1. Authenticate with Cloudflare

```bash
# Windows
docker run --rm -it -v "%USERPROFILE%\.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel login

# Linux/macOS
docker run --rm -it -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel login
```

This opens your browser to authenticate. After authentication, a certificate is saved to:
- Windows: `%USERPROFILE%\.cloudflared\cert.pem`
- Linux/macOS: `~/.cloudflared/cert.pem`

### 2. Create a Tunnel

```bash
# Windows
docker run --rm -v "%USERPROFILE%\.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel create filebrowser-tunnel

# Linux/macOS
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel create filebrowser-tunnel
```

This creates a tunnel and saves credentials to:
- Windows: `%USERPROFILE%\.cloudflared\<tunnel-id>.json`
- Linux/macOS: `~/.cloudflared/<tunnel-id>.json`

**Output:**
```
Created tunnel filebrowser-tunnel with id abc12345-6789-0123-4567-890123456789
```

### 3. Copy Credentials

```bash
# Windows
copy %USERPROFILE%\.cloudflared\<tunnel-id>.json cloudflared\credentials.json

# Linux/macOS
cp ~/.cloudflared/<tunnel-id>.json cloudflared/credentials.json
```

### 4. Create DNS Route

```bash
# Windows
docker run --rm -v "%USERPROFILE%\.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel route dns filebrowser-tunnel filebrowser.yourdomain.com

# Linux/macOS
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel route dns filebrowser-tunnel filebrowser.yourdomain.com
```

This creates a CNAME record in Cloudflare DNS pointing to your tunnel.

### 5. Generate Config File

Create `cloudflared/config.yml`:

```yaml
tunnel: <tunnel-id>
credentials-file: /etc/cloudflared/credentials.json

ingress:
  - hostname: filebrowser.yourdomain.com
    service: http://filebrowser-service:8082
  - service: http_status:404
```

## Useful CLI Commands (Using Docker)

### List All Tunnels

```bash
# Windows
docker run --rm -v "%USERPROFILE%\.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel list

# Linux/macOS
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel list
```

### View Tunnel Details

```bash
# Windows
docker run --rm -v "%USERPROFILE%\.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel info filebrowser-tunnel

# Linux/macOS
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel info filebrowser-tunnel
```

### Delete a Tunnel

```bash
# Windows
docker run --rm -v "%USERPROFILE%\.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel delete filebrowser-tunnel

# Linux/macOS
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel delete filebrowser-tunnel
```

### Route DNS (Alternative Method)

If you need to route DNS manually:

```bash
# Create CNAME record
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel route dns filebrowser-tunnel filebrowser.yourdomain.com
```

### Validate Config

```bash
# Windows
docker run --rm -v "%CD%\cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel ingress validate

# Linux/macOS
docker run --rm -v "$(pwd)/cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel ingress validate
```

This validates your `config.yml` file.

### Test Tunnel (Local)

```bash
# Windows
docker run --rm -v "%CD%\cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel --config /etc/cloudflared/config.yml run

# Linux/macOS
docker run --rm -v "$(pwd)/cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel --config /etc/cloudflared/config.yml run
```

This runs the tunnel in Docker for testing.

## Advanced Configuration

### Multiple Hostnames

```yaml
ingress:
  - hostname: filebrowser.yourdomain.com
    service: http://filebrowser-service:8082
  - hostname: files.yourdomain.com
    service: http://filebrowser-service:8082
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

### Multiple Services

```yaml
ingress:
  - hostname: filebrowser.yourdomain.com
    service: http://filebrowser-service:8082
  - hostname: api.yourdomain.com
    service: http://api-service:3000
  - service: http_status:404
```

### Using Tunnel Token (Alternative Authentication)

Instead of credentials.json, you can use a tunnel token:

1. In Cloudflare dashboard, go to your tunnel settings
2. Generate a **Tunnel Token**
3. Use it in docker-compose:

```yaml
environment:
  - TUNNEL_TOKEN=your-tunnel-token-here
```

Then update config.yml:

```yaml
tunnel: <tunnel-id>
# No credentials-file needed when using token
```

## Troubleshooting

### Authentication Issues

```bash
# Re-authenticate
docker run --rm -it -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel login

# Check certificate exists
# Windows
dir %USERPROFILE%\.cloudflared\cert.pem

# Linux/macOS
ls ~/.cloudflared/cert.pem
```

### Tunnel Not Found

```bash
# List all tunnels
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel list

# Verify tunnel name matches
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel info <tunnel-name>
```

### DNS Route Issues

```bash
# Check DNS records
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel route dns list

# Recreate DNS route
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel route dns delete filebrowser.yourdomain.com
docker run --rm -v "$HOME/.cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel route dns filebrowser-tunnel filebrowser.yourdomain.com
```

### Validate Configuration

```bash
# Test config file
docker run --rm -v "$(pwd)/cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel ingress validate

# Run tunnel in Docker to test
docker run --rm -v "$(pwd)/cloudflared:/etc/cloudflared" cloudflare/cloudflared:latest tunnel --config /etc/cloudflared/config.yml run
```

## File Locations

### Windows
- Certificate: `%USERPROFILE%\.cloudflared\cert.pem`
- Credentials: `%USERPROFILE%\.cloudflared\<tunnel-id>.json`
- Config: `cloudflared\config.yml` (project directory)

### Linux/macOS
- Certificate: `~/.cloudflared/cert.pem`
- Credentials: `~/.cloudflared/<tunnel-id>.json`
- Config: `cloudflared/config.yml` (project directory)

## References

- [Cloudflare Tunnel Documentation](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- [cloudflared CLI Reference](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/tunnel-guide/)
- [Ingress Rules](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/configuration/ingress/)

