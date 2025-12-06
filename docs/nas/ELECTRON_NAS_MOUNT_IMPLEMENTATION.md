# Electron NAS Mount Implementation Guide

This guide provides implementation details for mounting QNAP NAS shares in the Electron app to enable direct file access for large files.

## Overview

The Electron app will mount the QNAP NAS share using SMB protocol, allowing desktop applications (Photoshop, Illustrator, etc.) to access files directly from the mounted network drive without downloading.

## Architecture

```
Electron Renderer
    ↓ IPC: open-remote-file
Electron Main Process
    ↓ GET /api/nas/mount-credentials (with JWT)
Backend
    ↓ Returns mount credentials
Electron Main Process
    ↓ Mounts SMB share
Desktop Apps
    ↓ Access files from mounted path
```

## Implementation Steps

### 1. Add IPC Handlers in Main Process

**File:** `main.js` or `main.ts` (Electron main process)

```javascript
const { ipcMain, shell } = require('electron');
const { exec } = require('child_process');
const { promisify } = require('util');
const os = require('os');
const path = require('path');
const fs = require('fs-extra');

const execAsync = promisify(exec);

// Track mounted drives
const mountedDrives = new Map();

// Get available drive letter (Windows)
async function getAvailableDriveLetter() {
  if (process.platform !== 'win32') {
    return null;
  }

  // Check used drive letters (A-Z)
  const usedLetters = new Set();
  for (let i = 65; i <= 90; i++) { // A-Z
    const letter = String.fromCharCode(i);
    try {
      await fs.access(`${letter}:\\`);
      usedLetters.add(letter);
    } catch {
      // Drive not in use
    }
  }

  // Find first available letter (start from Z and go backwards)
  for (let i = 90; i >= 65; i--) {
    const letter = String.fromCharCode(i);
    if (!usedLetters.has(letter)) {
      return letter;
    }
  }

  throw new Error('No available drive letters');
}

// Mount NAS share (Windows)
async function mountNASShareWindows(credentials) {
  const { nasIP, shareName, username, password } = credentials;
  const driveLetter = await getAvailableDriveLetter();
  
  // Windows: net use Z: \\192.168.1.100\studio-work /user:username password /persistent:no
  const command = `net use ${driveLetter}: \\\\${nasIP}\\${shareName} /user:${username} ${password} /persistent:no`;
  
  try {
    const { stdout, stderr } = await execAsync(command);
    if (stderr && !stderr.includes('successfully')) {
      throw new Error(stderr);
    }
    
    const mountedPath = `${driveLetter}:\\`;
    mountedDrives.set(driveLetter, {
      path: mountedPath,
      credentials,
      mountedAt: new Date()
    });
    
    return { driveLetter, mountedPath };
  } catch (error) {
    throw new Error(`Failed to mount NAS share: ${error.message}`);
  }
}

// Mount NAS share (macOS)
async function mountNASShareMacOS(credentials) {
  const { nasIP, shareName, username, password } = credentials;
  const mountPoint = path.join(os.tmpdir(), 'nas-mount');
  
  await fs.ensureDir(mountPoint);
  
  // macOS: mount_smbfs //username:password@nas-ip/share-name /mount/point
  const smbUrl = `//${username}:${password}@${nasIP}/${shareName}`;
  const command = `mount_smbfs "${smbUrl}" "${mountPoint}"`;
  
  try {
    await execAsync(command);
    mountedDrives.set('macos', {
      path: mountPoint,
      credentials,
      mountedAt: new Date()
    });
    
    return { mountedPath: mountPoint };
  } catch (error) {
    throw new Error(`Failed to mount NAS share: ${error.message}`);
  }
}

// Mount NAS share (Linux)
async function mountNASShareLinux(credentials) {
  const { nasIP, shareName, username, password } = credentials;
  const mountPoint = path.join(os.tmpdir(), 'nas-mount');
  
  await fs.ensureDir(mountPoint);
  
  // Linux: mount -t cifs //nas-ip/share-name /mount/point -o username=user,password=pass
  const command = `mount -t cifs //${nasIP}/${shareName} ${mountPoint} -o username=${username},password=${password}`;
  
  try {
    await execAsync(`sudo ${command}`);
    mountedDrives.set('linux', {
      path: mountPoint,
      credentials,
      mountedAt: new Date()
    });
    
    return { mountedPath: mountPoint };
  } catch (error) {
    throw new Error(`Failed to mount NAS share: ${error.message}`);
  }
}

// IPC Handler: Mount NAS share
ipcMain.handle('mount-nas-share', async (event, credentials) => {
  try {
    let result;
    
    if (process.platform === 'win32') {
      result = await mountNASShareWindows(credentials);
    } else if (process.platform === 'darwin') {
      result = await mountNASShareMacOS(credentials);
    } else if (process.platform === 'linux') {
      result = await mountNASShareLinux(credentials);
    } else {
      throw new Error(`Unsupported platform: ${process.platform}`);
    }
    
    return { success: true, ...result };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

// IPC Handler: Open remote file
ipcMain.handle('open-remote-file', async (event, { fileId, filePath, fileSize, accessToken }) => {
  try {
    const SIZE_THRESHOLD = 100 * 1024 * 1024; // 100MB
    
    // For large files, use mount; for small files, download
    if (fileSize && fileSize > SIZE_THRESHOLD) {
      // Get mount credentials from backend
      const axios = require('axios');
      const response = await axios.get('http://localhost:3000/api/nas/mount-credentials', {
        headers: {
          'Authorization': `Bearer ${accessToken}`
        }
      });
      
      const credentials = response.data;
      
      // Mount NAS share (if not already mounted)
      let mountedPath;
      if (process.platform === 'win32') {
        const driveLetter = Object.keys(mountedDrives)[0];
        if (driveLetter) {
          mountedPath = mountedDrives.get(driveLetter).path;
        } else {
          const mountResult = await mountNASShareWindows(credentials);
          mountedPath = mountResult.mountedPath;
        }
      } else {
        const mountKey = process.platform === 'darwin' ? 'macos' : 'linux';
        if (mountedDrives.has(mountKey)) {
          mountedPath = mountedDrives.get(mountKey).path;
        } else {
          const mountResult = process.platform === 'darwin' 
            ? await mountNASShareMacOS(credentials)
            : await mountNASShareLinux(credentials);
          mountedPath = mountResult.mountedPath;
        }
      }
      
      // Build local path
      // filePath from backend: "/folder/file.pdf"
      // Convert to Windows: "Z:\folder\file.pdf" or Unix: "/tmp/nas-mount/folder/file.pdf"
      let localPath;
      if (process.platform === 'win32') {
        const normalizedPath = filePath.replace(/\//g, '\\');
        localPath = path.join(mountedPath, normalizedPath);
      } else {
        localPath = path.join(mountedPath, filePath);
      }
      
      // Open with default app
      await shell.openPath(localPath);
      
      return { success: true, localPath, method: 'mount' };
    } else {
      // Small file: download to temp and open
      // Implementation for download method...
      return { success: false, error: 'Download method not yet implemented' };
    }
  } catch (error) {
    return { success: false, error: error.message };
  }
});

// IPC Handler: Unmount NAS share
ipcMain.handle('unmount-nas-share', async (event) => {
  try {
    if (process.platform === 'win32') {
      for (const [driveLetter] of mountedDrives) {
        await execAsync(`net use ${driveLetter}: /delete /yes`);
      }
    } else {
      const mountKey = process.platform === 'darwin' ? 'macos' : 'linux';
      if (mountedDrives.has(mountKey)) {
        const mountInfo = mountedDrives.get(mountKey);
        await execAsync(`umount "${mountInfo.path}"`);
      }
    }
    
    mountedDrives.clear();
    return { success: true };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

// Cleanup on app close
app.on('before-quit', async () => {
  // Unmount all drives
  await ipcMain.handle('unmount-nas-share', () => {});
});
```

### 2. Add Renderer Process API

**File:** `preload.js` or renderer process

```javascript
const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  openRemoteFile: (fileId, filePath, fileSize, accessToken) => 
    ipcRenderer.invoke('open-remote-file', { fileId, filePath, fileSize, accessToken }),
  
  mountNASShare: (credentials) => 
    ipcRenderer.invoke('mount-nas-share', credentials),
  
  unmountNASShare: () => 
    ipcRenderer.invoke('unmount-nas-share'),
});
```

### 3. Use in React/Vue Component

```javascript
// In your file browser component
const handleOpenFile = async (file) => {
  const accessToken = getAccessToken(); // Get from your auth system
  
  const result = await window.electronAPI.openRemoteFile(
    file.id,
    file.path,
    file.size,
    accessToken
  );
  
  if (result.success) {
    console.log('File opened:', result.localPath);
  } else {
    console.error('Failed to open file:', result.error);
  }
};
```

## Dependencies

Install required npm packages:

```bash
npm install axios
```

## Testing

1. **Test Mount Credentials Endpoint:**
   ```bash
   curl -H "Authorization: Bearer <jwt_token>" \
        http://localhost:3000/api/nas/mount-credentials
   ```

2. **Test File Opening:**
   - Open Electron app
   - Click on a large file (>100MB)
   - Verify NAS share mounts
   - Verify file opens in default app

## Troubleshooting

### Mount Fails on Windows
- Ensure "Client for Microsoft Networks" is enabled
- Check Windows Firewall settings
- Verify NAS IP and share name are correct

### Permission Denied
- Verify service account has SMB access enabled
- Check share permissions on QNAP

### Drive Letter Already in Use
- The code automatically finds available drive letters
- Manually unmount if needed: `net use Z: /delete`

## Security Notes

- Mount credentials include password - handle securely
- Consider using temporary tokens instead of passwords
- Unmount shares when app closes
- Don't store credentials in plain text

