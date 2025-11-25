# FileBrowser Setup Script for Windows
# This script sets up FileBrowser on Windows machine

$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "FileBrowser Setup Script for Windows" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$ComposeFile = "docker-compose.filebrowser.windows.yml"
$ContainerName = "filebrowser"
$ConfigDir = ".\filebrowser-config"

# Function to print colored messages
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Stop existing container
function Stop-Container {
    Write-Info "Stopping existing FileBrowser container..."
    docker-compose -f $ComposeFile down 2>$null
    if ($LASTEXITCODE -eq 0 -or $LASTEXITCODE -eq 1) {
        Write-Info "Container stopped."
    }
}

# Setup config directory
function Setup-ConfigDirectory {
    Write-Info "Setting up config directory..."
    
    if (Test-Path $ConfigDir) {
        Write-Info "Config directory already exists: $ConfigDir"
        $response = Read-Host "Remove and recreate? (y/N)"
        if ($response -eq "y" -or $response -eq "Y") {
            Remove-Item -Recurse -Force $ConfigDir -ErrorAction SilentlyContinue
            Write-Info "Removed existing directory."
        }
    }
    
    if (-not (Test-Path $ConfigDir)) {
        Write-Info "Creating directory: $ConfigDir"
        New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
        Write-Info "Directory created."
    }
}

# Start container
function Start-Container {
    Write-Info "Starting FileBrowser container..."
    docker-compose -f $ComposeFile up -d
    
    if ($LASTEXITCODE -eq 0) {
        Write-Info "Container started successfully!"
    } else {
        Write-Error "Failed to start container!"
        exit 1
    }
}

# Wait for container to be ready
function Wait-ForContainer {
    Write-Info "Waiting for container to be ready..."
    Start-Sleep -Seconds 5
    
    $container = docker ps --filter "name=$ContainerName" --format "{{.Names}}" 2>$null
    if ($container -eq $ContainerName) {
        Write-Info "Container is running."
    } else {
        Write-Error "Container is not running!"
        Write-Info "Checking logs..."
        docker logs $ContainerName 2>&1 | Select-Object -Last 20
        exit 1
    }
}

# Check container logs
function Check-Logs {
    Write-Info "Checking container logs..."
    Write-Host ""
    docker logs $ContainerName 2>&1 | Select-Object -Last 30
    Write-Host ""
}

# Verify setup
function Verify-Setup {
    Write-Info "Verifying setup..."
    
    # Check if container is running
    $container = docker ps --filter "name=$ContainerName" --format "{{.Names}}" 2>$null
    if ($container -eq $ContainerName) {
        Write-Info "✓ Container is running" -ForegroundColor Green
    } else {
        Write-Error "✗ Container is not running"
        return $false
    }
    
    # Check if database file exists
    $dbExists = docker exec $ContainerName test -f /filebrowser-config/database.db 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Info "✓ Database file created successfully" -ForegroundColor Green
    } else {
        Write-Warn "⚠ Database file not found (may be initializing)"
    }
    
    # Check if port is accessible
    $portCheck = netstat -an | Select-String ":8081" | Select-Object -First 1
    if ($portCheck) {
        Write-Info "✓ Port 8081 is listening" -ForegroundColor Green
    } else {
        Write-Warn "⚠ Port 8081 may not be accessible"
    }
    
    return $true
}

# Get admin password from logs
function Get-AdminPassword {
    Write-Info "Retrieving admin password from logs..."
    $logs = docker logs $ContainerName 2>&1
    $passwordLine = $logs | Select-String "randomly generated password" | Select-Object -Last 1
    
    if ($passwordLine) {
        $password = ($passwordLine -split ": ")[1].Trim()
        Write-Host ""
        Write-Host "==========================================" -ForegroundColor Cyan
        Write-Host "FileBrowser Admin Credentials:" -ForegroundColor Cyan
        Write-Host "==========================================" -ForegroundColor Cyan
        Write-Host "Username: " -NoNewline
        Write-Host "admin" -ForegroundColor Green
        Write-Host "Password: " -NoNewline
        Write-Host $password -ForegroundColor Green
        Write-Host ""
        Write-Info "Access FileBrowser at: http://localhost:8081"
        Write-Host "==========================================" -ForegroundColor Cyan
    } else {
        Write-Warn "Could not retrieve password from logs."
        Write-Info "Default credentials: admin / admin"
        Write-Info "Check logs with: docker logs $ContainerName"
    }
}

# Main execution
function Main {
    Write-Host ""
    Write-Info "Starting FileBrowser setup..."
    Write-Host ""
    
    # Stop existing container
    Stop-Container
    
    # Setup config directory
    Setup-ConfigDirectory
    
    # Start container
    Start-Container
    
    # Wait for container
    Wait-ForContainer
    
    # Check logs
    Check-Logs
    
    # Verify setup
    $verified = Verify-Setup
    if (-not $verified) {
        Write-Error "Setup verification failed!"
        exit 1
    }
    
    # Get admin password
    Get-AdminPassword
    
    Write-Host ""
    Write-Info "Setup complete!"
    Write-Info "To view logs: docker logs $ContainerName"
    Write-Info "To stop: docker-compose -f $ComposeFile down"
    Write-Info "To restart: docker-compose -f $ComposeFile restart"
    Write-Host ""
}

# Run main function
Main

