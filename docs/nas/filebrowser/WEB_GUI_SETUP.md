# FileBrowser Setup Using Web GUI Only

Complete guide for setting up FileBrowser using only QNAP's web interface (no terminal/SSH needed).

## Prerequisites

- QNAP NAS web interface access
- Container Station app installed
- File Station app installed

---

## Step 1: Prepare Config Folder (File Station)

### Create filebrowser-config Folder

1. **Open File Station** from QNAP main menu
2. **Navigate to** `/share/Container/` folder
3. **Create new folder:**
   - Right-click in empty area → **"Create Folder"**
   - Name: `filebrowser-config`
   - Click **"Create"**

### Set Permissions (Optional - for bind mount)

1. **Right-click** `filebrowser-config` folder
2. **Select "Properties"** → **"Permissions" tab**
3. **Click "Edit"**
4. **Set permissions:**
   - Owner: ✅ Read, ✅ Write, ✅ Execute
   - Group: ✅ Read, ❌ Write, ✅ Execute
   - Others: ✅ Read, ❌ Write, ✅ Execute
5. ✅ Check **"Apply to all subfolders and files"**
6. **Click "Apply"** → **"OK"**

**Note:** If using named volume (recommended), you can skip permission setup - Docker handles it automatically.

---

## Step 2: Upload Docker Compose File (File Station)

### Option A: Create File Directly in File Station

1. **Open File Station**
2. **Navigate to** `/share/Container/` or any accessible folder
3. **Right-click** → **"Create"** → **"Text File"**
4. **Name it:** `docker-compose.filebrowser.nas.yml`
5. **Right-click** the file → **"Edit"**
6. **Copy and paste** the docker-compose content (see below)
7. **Save** the file

### Option B: Upload from Computer

1. **Open File Station**
2. **Navigate to** `/share/Container/` folder
3. **Click "Upload"** button
4. **Select** `docker-compose.filebrowser.nas.yml` from your computer
5. **Upload** the file

---

## Step 3: Deploy via Container Station

### Method 1: Using Compose File (Recommended)

1. **Open Container Station** from QNAP main menu
2. **Click "Create"** button (usually top right)
3. **Select "Application"** or **"Compose"** tab
4. **Click "Create Application"** or **"Create Stack"**

5. **Configure:**
   - **Name:** `filebrowser`
   - **Source:** Select **"Upload YAML file"** or **"From file"**
   - **Browse** and select `docker-compose.filebrowser.nas.yml`
   - Or **paste YAML** directly into the text area

6. **Click "Create"** or **"Deploy"**

7. **Wait for deployment** - Container Station will:
   - Pull the image
   - Create volumes
   - Start the container

### Method 2: Manual Container Creation

If Compose doesn't work, create container manually:

1. **Open Container Station**
2. **Click "Create"** → **"Container"**
3. **Search for:** `filebrowser/filebrowser`
4. **Click "Install"** or **"Create"**

5. **Configure Container:**
   - **Name:** `filebrowser`
   - **Image:** `filebrowser/filebrowser:latest`
   - **Port:**
     - Host: `8081`
     - Container: `3000`
   - **Volumes:**
     - Add volume: `/share/studio work/` → `/srv`
     - Add volume: `filebrowser-config` (named volume) → `/filebrowser-config`
   - **Environment Variables:**
     - `FB_DATABASE=/filebrowser-config/database.db`
     - `FB_ROOT=/srv`
     - `FB_PORT=3000`
     - `FB_DEFAULT_USER=admin`
     - `FB_DEFAULT_PASS=admin`
     - `TZ=Asia/Kolkata`

6. **Click "Create"** or **"Deploy"**

---

## Step 4: Fix Permissions (Container Station Web Terminal)

If you get permission errors:

1. **Open Container Station**
2. **Find** `filebrowser` container in the list
3. **Click on** the container name
4. **Click "Terminal"** or **"Console"** tab
5. **Run these commands:**
   ```
   chown -R 1000:1000 /filebrowser-config
   chmod -R 755 /filebrowser-config
   ls -la /filebrowser-config
   ```

6. **Restart container:**
   - Go back to container list
   - Click **"Restart"** button on filebrowser container

---

## Step 5: Access FileBrowser

1. **Open web browser**
2. **Navigate to:** `http://YOUR-NAS-IP:8081`
   - Replace `YOUR-NAS-IP` with your NAS IP address
   - Example: `http://192.168.1.100:8081`

3. **Login:**
   - **Username:** `admin`
   - **Password:** 
     - Check Container Station logs for generated password
     - Or use `admin` if using default

### How to Check Password in Container Station:

1. **Open Container Station**
2. **Click on** `filebrowser` container
3. **Click "Logs"** tab
4. **Look for:** `User 'admin' initialized with randomly generated password: xxxxx`
5. **Copy the password**

---

## Step 6: Manage Container (Container Station)

### View Logs

1. **Container Station** → **Containers**
2. **Click** `filebrowser` container
3. **Click "Logs"** tab

### Restart Container

1. **Container Station** → **Containers**
2. **Find** `filebrowser` container
3. **Click "Restart"** button (circular arrow icon)

### Stop Container

1. **Container Station** → **Containers**
2. **Find** `filebrowser` container
3. **Click "Stop"** button (square icon)

### Start Container

1. **Container Station** → **Containers**
2. **Find** `filebrowser` container
3. **Click "Start"** button (play icon)

### Remove Container

1. **Container Station** → **Containers**
2. **Find** `filebrowser` container
3. **Click "Remove"** button (trash icon)
4. ✅ Check **"Remove volumes"** if you want to delete config

---

## Troubleshooting via Web GUI

### Container Won't Start

1. **Container Station** → **Containers**
2. **Click** `filebrowser` container
3. **Check "Logs"** tab for errors
4. **Check "Details"** tab for configuration

### Permission Denied Error

1. **Container Station** → **Containers**
2. **Click** `filebrowser` container
3. **Click "Terminal"** tab
4. **Run:**
   ```
   chown -R 1000:1000 /filebrowser-config
   chmod -R 755 /filebrowser-config
   ```
5. **Restart** container

### Can't Access Web UI

1. **Check container is running:**
   - Container Station → Containers → filebrowser should show "Running"

2. **Check port:**
   - Container Station → Containers → filebrowser → Details
   - Verify port mapping: `8081:3000`

3. **Try accessing:**
   - `http://YOUR-NAS-IP:8081`
   - `http://localhost:8081` (if accessing from NAS itself)

### Remove and Start Fresh

1. **Container Station** → **Containers**
2. **Find** `filebrowser` container
3. **Click "Remove"**
4. ✅ Check **"Remove volumes"**
5. **Click "Remove"**
6. **Start over** from Step 3

---

## Docker Compose File Content

Copy this content into your `docker-compose.filebrowser.nas.yml` file:

```yaml
services:
  filebrowser:
    image: filebrowser/filebrowser:latest
    container_name: filebrowser
    restart: unless-stopped
    ports:
      - "8081:3000"
    volumes:
      - "/share/studio work/:/srv"
      - "filebrowser-config:/filebrowser-config"
    environment:
      - FB_DATABASE=/filebrowser-config/database.db
      - FB_ROOT=/srv
      - FB_PORT=3000
      - FB_DEFAULT_USER=admin
      - FB_DEFAULT_PASS=admin
      - TZ=Asia/Kolkata
    networks:
      - filebrowser-network

networks:
  filebrowser-network:
    driver: bridge

volumes:
  filebrowser-config:
```

---

## Quick Checklist

- [ ] Created `filebrowser-config` folder in File Station
- [ ] Set permissions (optional, if using bind mount)
- [ ] Uploaded/created docker-compose file
- [ ] Deployed via Container Station
- [ ] Container is running (check Container Station)
- [ ] Fixed permissions via Container Station Terminal (if needed)
- [ ] Accessed FileBrowser at http://YOUR-NAS-IP:8081
- [ ] Logged in successfully

---

## Summary

**Everything can be done via web GUI:**
1. **File Station** - Create folders, set permissions
2. **Container Station** - Deploy and manage containers
3. **Container Station Terminal** - Fix permissions if needed
4. **Web Browser** - Access FileBrowser interface

No SSH or command line access required! 🎉

