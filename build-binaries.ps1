# PowerShell script to build binaries for Docker containers
# This script builds the server and worker binaries for Linux x86_64

Write-Host "Building AR-13 backend binaries for Linux x86_64..." -ForegroundColor Cyan

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build configuration for x86_64 (amd64)
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"  # amd64 = x86_64 architecture

# Build flags for optimized binary
$buildFlags = "-a -installsuffix cgo -ldflags `"-w -s`""

# Build server binary
Write-Host "Building server binary (x86_64)..." -ForegroundColor Yellow
go build $buildFlags -o bin/server.exe ./cmd/server/main.go
# Rename to remove .exe for Linux
if (Test-Path "bin/server.exe") {
    Move-Item -Path "bin/server.exe" -Destination "bin/server" -Force
}

# Build worker binary
Write-Host "Building worker binary (x86_64)..." -ForegroundColor Yellow
go build $buildFlags -o bin/worker.exe ./cmd/worker/main.go
# Rename to remove .exe for Linux
if (Test-Path "bin/worker.exe") {
    Move-Item -Path "bin/worker.exe" -Destination "bin/worker" -Force
}

Write-Host ""
Write-Host "✅ Binaries built successfully!" -ForegroundColor Green
Write-Host "Location: bin/" -ForegroundColor Green
Write-Host ""
$serverSize = (Get-Item "bin/server").Length / 1MB
$workerSize = (Get-Item "bin/worker").Length / 1MB
Write-Host "Server binary: $([math]::Round($serverSize, 2)) MB"
Write-Host "Worker binary: $([math]::Round($workerSize, 2)) MB"
Write-Host ""
Write-Host "You can now build Docker images:" -ForegroundColor Cyan
Write-Host "  docker build -f docker/Dockerfile.server -t ar-13-server ."
Write-Host "  docker build -f docker/Dockerfile.worker -t ar-13-worker ."
Write-Host "  Or use: docker-compose -f docker/docker-compose.yml build"

