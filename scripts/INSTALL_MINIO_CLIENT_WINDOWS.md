# Installing MinIO Client (mc) on Windows

## Method 1: Using Chocolatey (Recommended)

1. **Install Chocolatey** (if not already installed):
   ```powershell
   # Run PowerShell as Administrator
   Set-ExecutionPolicy Bypass -Scope Process -Force
   [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
   iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
   ```

2. **Install MinIO Client:**
   ```powershell
   choco install minio-client
   ```

3. **Verify installation:**
   ```powershell
   mc --version
   ```

## Method 2: Manual Download

1. **Download MinIO Client:**
   - Go to: https://dl.min.io/client/mc/release/windows-amd64/mc.exe
   - Or visit: https://min.io/docs/minio/linux/reference/minio-mc.html#install-mc

2. **Add to PATH:**
   - Create a folder (e.g., `C:\Tools\minio`)
   - Place `mc.exe` in that folder
   - Add folder to Windows PATH:
     - Right-click "This PC" → Properties → Advanced System Settings
     - Click "Environment Variables"
     - Under "System Variables", find "Path" and click "Edit"
     - Click "New" and add `C:\Tools\minio`
     - Click OK on all dialogs

3. **Verify installation:**
   - Open a new PowerShell window
   ```powershell
   mc --version
   ```

## Method 3: Use Scoop (Alternative Package Manager)

```powershell
# Install Scoop (if not installed)
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex

# Install MinIO Client
scoop install minio-client
```

## After Installation

Once `mc` is installed, you can use the scripts or commands from the setup guide.

