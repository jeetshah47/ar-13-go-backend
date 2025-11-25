# Cloudflare Tunnel Setup Script
# Uses native Windows cloudflared installation

param(
    [string]$TunnelName = "filebrowser-tunnel",
    [string]$Hostname = ""
)

Write-Host "Cloudflare Tunnel Setup Script" -ForegroundColor Cyan
Write-Host "===============================" -ForegroundColor Cyan
Write-Host "Using native Windows cloudflared" -ForegroundColor Yellow
Write-Host ""

# Check if cloudflared is installed
$cloudflaredPath = Get-Command cloudflared -ErrorAction SilentlyContinue
if (-Not $cloudflaredPath) {
    Write-Host "❌ cloudflared is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Installing cloudflared..." -ForegroundColor Yellow
    
    # Try to install using winget
    $wingetCheck = Get-Command winget -ErrorAction SilentlyContinue
    if ($wingetCheck) {
        Write-Host "Using winget to install cloudflared..." -ForegroundColor Yellow
        winget install --id Cloudflare.cloudflared --accept-package-agreements --accept-source-agreements
        if ($LASTEXITCODE -ne 0) {
            Write-Host "❌ Failed to install via winget" -ForegroundColor Red
            Write-Host ""
            Write-Host "Please install cloudflared manually:" -ForegroundColor Yellow
            Write-Host "1. Download from: https://github.com/cloudflare/cloudflared/releases" -ForegroundColor Yellow
            Write-Host "2. Extract cloudflared.exe to a folder in your PATH" -ForegroundColor Yellow
            Write-Host "3. Or add the folder to your PATH environment variable" -ForegroundColor Yellow
            exit 1
        }
        
        # Refresh PATH
        $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
        
        # Check again
        $cloudflaredPath = Get-Command cloudflared -ErrorAction SilentlyContinue
        if (-Not $cloudflaredPath) {
            Write-Host "⚠️  cloudflared installed but not in PATH. Please restart your terminal." -ForegroundColor Yellow
            Write-Host "Or add cloudflared to your PATH and run this script again." -ForegroundColor Yellow
            exit 1
        }
        Write-Host "✅ cloudflared installed successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ winget is not available" -ForegroundColor Red
        Write-Host ""
        Write-Host "Please install cloudflared manually:" -ForegroundColor Yellow
        Write-Host "1. Download from: https://github.com/cloudflare/cloudflared/releases" -ForegroundColor Yellow
        Write-Host "2. Extract cloudflared.exe to a folder in your PATH" -ForegroundColor Yellow
        Write-Host "3. Or add the folder to your PATH environment variable" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Alternatively, install winget from Microsoft Store" -ForegroundColor Yellow
        exit 1
    }
}

Write-Host "✅ cloudflared is installed: $($cloudflaredPath.Source)" -ForegroundColor Green
Write-Host ""

# Get tunnel name if not provided
if ([string]::IsNullOrEmpty($TunnelName)) {
    $TunnelName = Read-Host "Enter tunnel name (default: filebrowser-tunnel)"
    if ([string]::IsNullOrEmpty($TunnelName)) {
        $TunnelName = "filebrowser-tunnel"
    }
}

# Get hostname
if ([string]::IsNullOrEmpty($Hostname)) {
    $Hostname = Read-Host "Enter full hostname (e.g., filebrowser.yourdomain.com)"
    if ([string]::IsNullOrEmpty($Hostname)) {
        Write-Host "❌ Error: Hostname is required" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "Configuration:" -ForegroundColor Cyan
Write-Host "  Tunnel Name: $TunnelName"
Write-Host "  Hostname: $Hostname"
Write-Host ""

$confirm = Read-Host "Continue? (y/n)"
if ($confirm -ne 'y' -and $confirm -ne 'Y') {
    Write-Host "Aborted." -ForegroundColor Yellow
    exit 1
}

# Create cloudflared directory if it doesn't exist
$cloudflaredDir = "cloudflared"
if (-Not (Test-Path $cloudflaredDir)) {
    New-Item -ItemType Directory -Path $cloudflaredDir | Out-Null
    Write-Host "✅ Created cloudflared directory" -ForegroundColor Green
}

# Determine the path for cloudflared data (Windows default location)
$cloudflaredDataPath = "$env:USERPROFILE\.cloudflared"
if (-Not (Test-Path $cloudflaredDataPath)) {
    New-Item -ItemType Directory -Path $cloudflaredDataPath | Out-Null
}

# Step 1: Authenticate with Cloudflare
Write-Host ""
Write-Host "Step 1: Authenticating with Cloudflare..." -ForegroundColor Yellow
Write-Host "This will open a browser window for authentication..." -ForegroundColor Yellow

$certPath = "$cloudflaredDataPath\cert.pem"
if (Test-Path $certPath) {
    Write-Host "✅ Already authenticated (cert.pem exists)" -ForegroundColor Green
} else {
    Write-Host "Running cloudflared tunnel login..." -ForegroundColor Yellow
    
    cloudflared tunnel login
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Error: Authentication failed" -ForegroundColor Red
        exit 1
    }
    
    # Wait a moment and verify cert.pem was created
    Start-Sleep -Seconds 2
    if (-Not (Test-Path $certPath)) {
        Write-Host "❌ Error: cert.pem was not created after login" -ForegroundColor Red
        Write-Host "Please check the authentication process and try again." -ForegroundColor Yellow
        exit 1
    }
    Write-Host "✅ Authentication successful" -ForegroundColor Green
}

# Step 2: Create tunnel
Write-Host ""
Write-Host "Step 2: Creating tunnel '$TunnelName'..." -ForegroundColor Yellow

# Verify cert.pem exists before proceeding
if (-Not (Test-Path $certPath)) {
    Write-Host "❌ Error: cert.pem not found at $certPath" -ForegroundColor Red
    Write-Host "Please run the authentication step again." -ForegroundColor Yellow
    exit 1
}

$tunnelOutput = cloudflared tunnel create $TunnelName 2>&1 | Out-String

# Extract tunnel ID - try multiple patterns
$tunnelId = $null

# Pattern 1: "Created tunnel <id>"
if ($tunnelOutput -match "Created tunnel\s+([a-f0-9-]{36})") {
    $tunnelId = $matches[1]
    Write-Host "✅ Created tunnel: $tunnelId" -ForegroundColor Green
}
# Pattern 2: "Created tunnel <id>" (shorter format)
elseif ($tunnelOutput -match "Created tunnel\s+([a-f0-9-]+)") {
    $tunnelId = $matches[1]
    Write-Host "✅ Created tunnel: $tunnelId" -ForegroundColor Green
}
# Pattern 3: Look for UUID pattern anywhere in output
elseif ($tunnelOutput -match "([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})") {
    $tunnelId = $matches[1]
    Write-Host "✅ Created tunnel: $tunnelId" -ForegroundColor Green
}
else {
    # Try to get existing tunnel
    Write-Host "Checking for existing tunnel..." -ForegroundColor Yellow
    $tunnelList = cloudflared tunnel list 2>&1 | Out-String
    
    # Try to match tunnel name and extract ID
    if ($tunnelList -match "$TunnelName\s+([a-f0-9-]{36})") {
        $tunnelId = $matches[1]
        Write-Host "✅ Using existing tunnel: $tunnelId" -ForegroundColor Green
    }
    elseif ($tunnelList -match "$TunnelName\s+([a-f0-9-]+)") {
        $tunnelId = $matches[1]
        Write-Host "✅ Using existing tunnel: $tunnelId" -ForegroundColor Green
    }
    elseif ($tunnelList -match "([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})") {
        $tunnelId = $matches[1]
        Write-Host "✅ Found tunnel ID: $tunnelId" -ForegroundColor Green
    }
    else {
        Write-Host "❌ Error: Could not create or find tunnel" -ForegroundColor Red
        Write-Host "Tunnel creation output:" -ForegroundColor Yellow
        Write-Host $tunnelOutput -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Tunnel list output:" -ForegroundColor Yellow
        Write-Host $tunnelList -ForegroundColor Yellow
        exit 1
    }
}

if ([string]::IsNullOrEmpty($tunnelId)) {
    Write-Host "❌ Error: Tunnel ID is empty" -ForegroundColor Red
    Write-Host "Output: $tunnelOutput" -ForegroundColor Yellow
    exit 1
}

# Step 3: Copy credentials.json to project directory
Write-Host ""
Write-Host "Step 3: Setting up credentials..." -ForegroundColor Yellow

$credentialsSource = "$cloudflaredDataPath\$tunnelId.json"
$credentialsDest = "$cloudflaredDir\credentials.json"

# Check if credentials file exists
if (-Not (Test-Path $credentialsSource)) {
    Write-Host "⚠️  Credentials file not found at expected location: $credentialsSource" -ForegroundColor Yellow
    Write-Host "Searching for credentials files..." -ForegroundColor Yellow
    
    # Look for any .json files in the cloudflared directory
    $jsonFiles = Get-ChildItem -Path $cloudflaredDataPath -Filter "*.json" -ErrorAction SilentlyContinue
    if ($jsonFiles) {
        Write-Host "Found credential files:" -ForegroundColor Yellow
        foreach ($file in $jsonFiles) {
            Write-Host "  - $($file.Name)" -ForegroundColor Cyan
        }
        
        # Try to find the one matching our tunnel ID
        $matchingFile = $jsonFiles | Where-Object { $_.Name -like "*$tunnelId*" }
        if ($matchingFile) {
            $credentialsSource = $matchingFile.FullName
            Write-Host "✅ Using credentials file: $($matchingFile.Name)" -ForegroundColor Green
        } else {
            # Use the most recent one
            $latestFile = $jsonFiles | Sort-Object LastWriteTime -Descending | Select-Object -First 1
            $credentialsSource = $latestFile.FullName
            Write-Host "✅ Using most recent credentials file: $($latestFile.Name)" -ForegroundColor Green
        }
    } else {
        Write-Host "❌ Error: No credentials files found in $cloudflaredDataPath" -ForegroundColor Red
        Write-Host "Please check if the tunnel was created successfully." -ForegroundColor Yellow
        exit 1
    }
}

Copy-Item $credentialsSource $credentialsDest -Force
Write-Host "✅ Copied credentials to $credentialsDest" -ForegroundColor Green

# Copy cert.pem to cloudflared directory for Docker container
$certSource = "$cloudflaredDataPath\cert.pem"
$certDest = "$cloudflaredDir\cert.pem"
if (Test-Path $certSource) {
    Copy-Item $certSource $certDest -Force
    Write-Host "✅ Copied cert.pem to $certDest" -ForegroundColor Green
} else {
    Write-Host "⚠️  Warning: cert.pem not found at $certSource" -ForegroundColor Yellow
}

# Step 4: Create DNS route
Write-Host ""
Write-Host "Step 4: Creating DNS route..." -ForegroundColor Yellow

$dnsOutput = cloudflared tunnel route dns $TunnelName $Hostname 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Created DNS route for $Hostname" -ForegroundColor Green
} else {
    Write-Host "⚠️  Warning: DNS route creation may have failed" -ForegroundColor Yellow
    Write-Host "Output: $dnsOutput" -ForegroundColor Yellow
}

# Step 5: Generate config.yml
Write-Host ""
Write-Host "Step 5: Generating config.yml..." -ForegroundColor Yellow

$configContent = @"
# Cloudflare Tunnel Configuration
# Auto-generated by setup-tunnel.ps1

tunnel: $tunnelId
credentials-file: /etc/cloudflared/credentials.json
origincert: /etc/cloudflared/cert.pem

# Ingress rules - define how traffic is routed
ingress:
  # Route all traffic to filebrowser service
  - hostname: $Hostname
    service: http://filebrowser-service:8082
  
  # Catch-all rule (must be last)
  - service: http_status:404
"@

$configContent | Set-Content "$cloudflaredDir\config.yml"
Write-Host "✅ Generated config.yml" -ForegroundColor Green

# Step 6: Summary
Write-Host ""
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "✅ Tunnel setup complete!" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tunnel Details:" -ForegroundColor Cyan
Write-Host "  ID: $tunnelId"
Write-Host "  Name: $TunnelName"
Write-Host "  Hostname: $Hostname"
Write-Host ""
Write-Host "Files created:" -ForegroundColor Cyan
Write-Host "  - $cloudflaredDir\credentials.json"
Write-Host "  - $cloudflaredDir\config.yml"
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Start the services:"
Write-Host "   docker-compose -f docker-compose.filebrowser-cloudflared.windows.yml up -d"
Write-Host ""
Write-Host "2. Check tunnel status:"
Write-Host "   docker logs cloudflared-tunnel"
Write-Host ""
Write-Host "3. Access your service:"
Write-Host "   https://$Hostname"
Write-Host ""
