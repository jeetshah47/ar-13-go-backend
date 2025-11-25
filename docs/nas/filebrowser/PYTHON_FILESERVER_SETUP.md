# Python Simple HTTP File Server Setup Guide

A simple, lightweight file browser using Python's built-in HTTP server. No permission issues, no database needed!

## Features

- ✅ **Super Simple** - Just Python, no complex setup
- ✅ **No Permission Issues** - Works out of the box
- ✅ **Shows Existing Files** - Immediately visible
- ✅ **No Database** - No configuration needed
- ✅ **Lightweight** - Minimal resource usage
- ✅ **Download Support** - Click files to download

## Quick Start (Web GUI)

### Step 1: Deploy via Container Station

1. **Open Container Station** from QNAP main menu
2. **Click "Create"** → **"Application"** or **"Compose"**
3. **Click "Create Application"** or **"Create Stack"**

4. **Configure:**
   - **Name:** `python-fileserver`
   - **Source:** Select **"Upload YAML file"** or paste YAML directly
   - **Paste this YAML:**

```yaml
services:
  fileserver:
    image: python:3-alpine
    container_name: python-fileserver
    restart: unless-stopped
    ports:
      - "8081:8000"
    volumes:
      - "/share/studio work/:/shared"
    command: python -m http.server 8000 --directory /shared
    networks:
      - fileserver-network

networks:
  fileserver-network:
    driver: bridge
```

5. **Click "Create"** or **"Deploy"**

6. **Wait for deployment** (about 30 seconds)

### Step 2: Access File Server

1. **Open web browser**
2. **Navigate to:** `http://YOUR-NAS-IP:8081`
   - Replace `YOUR-NAS-IP` with your NAS IP address
   - Example: `http://192.168.1.100:8081`

3. **You'll see:**
   - Directory listing of all files and folders
   - Click folders to navigate
   - Click files to download

## Features

### What You Can Do

- ✅ **Browse folders** - Navigate through directory structure
- ✅ **Download files** - Click any file to download
- ✅ **View file sizes** - See file sizes and dates
- ✅ **No authentication** - Direct access (add nginx reverse proxy if needed)

### Limitations

- ❌ **No upload** - Read-only (Python's http.server doesn't support upload)
- ❌ **No authentication** - Anyone with URL can access (use nginx reverse proxy for security)
- ❌ **Basic interface** - Simple directory listing (not fancy UI)

## Windows Setup

For Windows machine, use `docker-compose.python-fileserver.windows.yml`:

```powershell
cd docs\nas\filebrowser
docker-compose -f docker-compose.python-fileserver.windows.yml up -d
```

Access at: http://localhost:8081

## Management via Container Station

### View Logs

1. **Container Station** → **Containers**
2. **Click** `python-fileserver` container
3. **Click "Logs"** tab

### Restart Container

1. **Container Station** → **Containers**
2. **Find** `python-fileserver` container
3. **Click "Restart"** button

### Stop Container

1. **Container Station** → **Containers**
2. **Find** `python-fileserver` container
3. **Click "Stop"** button

### Start Container

1. **Container Station** → **Containers**
2. **Find** `python-fileserver` container
3. **Click "Start"** button

## Troubleshooting

### Container Won't Start

1. **Check logs:**
   - Container Station → Containers → python-fileserver → Logs
   - Look for error messages

2. **Check port:**
   - Verify port 8081 is not in use
   - Try changing to 8082:8000 in docker-compose

3. **Check volume mount:**
   - Verify `/share/studio work/` exists
   - Check folder permissions in File Station

### Can't Access Web Interface

1. **Check container is running:**
   - Container Station → Containers → Status should be "Running"

2. **Check port mapping:**
   - Container Station → Containers → python-fileserver → Details
   - Verify: `8081:8000`

3. **Try different URL:**
   - `http://YOUR-NAS-IP:8081`
   - `http://localhost:8081` (if accessing from NAS itself)

### Files Not Showing

1. **Check volume mount:**
   - Container Station → Containers → python-fileserver → Details
   - Verify volume: `/share/studio work/:/shared`

2. **Check folder exists:**
   - File Station → Navigate to `/share/studio work/`
   - Verify files are there

3. **Restart container:**
   - Container Station → Containers → python-fileserver → Restart

## Adding Authentication (Optional)

If you need authentication, add nginx reverse proxy in front:

```yaml
services:
  nginx:
    image: nginx:alpine
    container_name: nginx-auth
    ports:
      - "8081:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./htpasswd:/etc/nginx/htpasswd
    depends_on:
      - fileserver

  fileserver:
    image: python:3-alpine
    container_name: python-fileserver
    volumes:
      - "/share/studio work/:/shared"
    command: python -m http.server 8000 --directory /shared
```

## Comparison with FileBrowser

| Feature | Python HTTP Server | FileBrowser |
|---------|-------------------|-------------|
| Setup Complexity | ⭐ Very Simple | ⭐⭐⭐ Complex |
| Permission Issues | ✅ None | ❌ Many |
| Shows Existing Files | ✅ Yes | ✅ Yes |
| Upload Files | ❌ No | ✅ Yes |
| Authentication | ❌ No | ✅ Yes |
| UI Quality | ⭐ Basic | ⭐⭐⭐ Modern |
| Resource Usage | ⭐ Very Low | ⭐⭐ Medium |

## Advantages

1. **No Permission Issues** - Works immediately
2. **Simple Setup** - Just deploy and go
3. **Lightweight** - Minimal resources
4. **Reliable** - Python's built-in server is stable
5. **No Configuration** - Works out of the box

## Use Cases

Perfect for:
- ✅ Browsing and downloading existing files
- ✅ Quick file access without complex setup
- ✅ When you just need to view files
- ✅ Low-resource environments

Not ideal for:
- ❌ File uploads
- ❌ User authentication
- ❌ Fancy UI requirements

## Next Steps

1. ✅ Deploy the container via Container Station
2. ✅ Access at http://YOUR-NAS-IP:8081
3. ✅ Browse your files
4. ✅ Download files as needed

If you need upload or authentication features later, we can add nginx reverse proxy or switch to another solution.

