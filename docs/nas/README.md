# NAS Setup Documentation

This directory contains documentation for setting up MinIO on your NAS with SSL support.

## Files

- **MINIO_SETUP_GUIDE.md** - Complete guide for setting up MinIO on NAS with SSL
- **SSL_CERTIFICATE_SETUP.md** - Detailed guide for SSL certificate configuration
- **QUICK_ACCESS_KEY_SETUP.md** - Quick guide to create access keys using Container Station terminal
- **USE_EXECUTE_BUTTON.md** - Use "Execute" button if terminal won't let you type (Recommended if terminal has issues)
- **FIX_HTTP_HTTPS_ERROR.md** - Fix "HTTP response to HTTPS client" error - use HTTP for localhost
- **FIX_ACCESS_DENIED.md** - Fix "Access Denied" error - create and attach policy to service account
- **CREATE_ACCESS_KEY_CONTAINER_STATION.md** - Step-by-step guide using Container Station with your current config
- **CONNECT_FROM_WINDOWS.md** - Guide to connect to MinIO from Windows computer using mc client
- **FIX_TIME_SYNC_ERROR.md** - How to fix "time difference too large" error
- **ENABLE_TERMINAL_ACCESS.md** - How to enable terminal access if you see the interactive/TTY error
- **TROUBLESHOOT_IDENTITY_MENU.md** - Troubleshooting guide if you can't find Identity/Access Keys menu
- **docker-compose.minio.nas.yml** - Docker Compose file for NAS deployment with SSL

## Quick Start

1. Read [MINIO_SETUP_GUIDE.md](./MINIO_SETUP_GUIDE.md) for step-by-step instructions
2. Set up SSL certificates using [SSL_CERTIFICATE_SETUP.md](./SSL_CERTIFICATE_SETUP.md)
3. Deploy using `docker-compose.minio.nas.yml`

## Prerequisites

- NAS with Docker/Container Station
- SSL certificates (or ability to generate self-signed)
- SSH access to NAS (recommended)

## Configuration

After setup, configure your backend `.env` file:

```env
MINIO_ENDPOINT=your-nas-domain:9000
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
MINIO_BUCKET=ar-13-uploads
MINIO_USE_SSL=true
MINIO_INSECURE_SSL=false  # Set to true for self-signed certificates
```

