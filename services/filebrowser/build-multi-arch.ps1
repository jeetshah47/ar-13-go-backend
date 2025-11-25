# Build and push multi-arch Docker image for filebrowser service
# Supports both AMD64 and ARM64 architectures
# PowerShell script for Windows

param(
    [string]$Version = "latest"
)

$ImageName = "jeetshah786/ar-13-go-backend-filebrowser-service"
$FullImageName = "$ImageName`:$Version"

Write-Host "Building multi-arch Docker image: $FullImageName" -ForegroundColor Cyan
Write-Host "Architectures: linux/amd64, linux/arm64" -ForegroundColor Yellow
Write-Host ""

# Check if buildx is available
$buildxCheck = docker buildx version 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Error: docker buildx is not available" -ForegroundColor Red
    Write-Host "Please install Docker Desktop with Buildx support" -ForegroundColor Yellow
    exit 1
}

# Create and use a new builder instance (if needed)
$BuilderName = "multiarch-builder"
$builderCheck = docker buildx inspect $BuilderName 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Creating new buildx builder: $BuilderName" -ForegroundColor Yellow
    docker buildx create --name $BuilderName --use
    docker buildx inspect --bootstrap
} else {
    Write-Host "Using existing buildx builder: $BuilderName" -ForegroundColor Green
    docker buildx use $BuilderName
}

# Navigate to the filebrowser directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Build and push multi-arch image
Write-Host ""
Write-Host "Building and pushing multi-arch image..." -ForegroundColor Yellow
Write-Host "This may take several minutes..." -ForegroundColor Yellow
Write-Host ""

docker buildx build `
    --platform linux/amd64,linux/arm64 `
    --tag "$ImageName`:$Version" `
    --tag "$ImageName`:latest" `
    --push `
    .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Successfully built and pushed multi-arch image!" -ForegroundColor Green
    Write-Host "Image: $FullImageName" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "To verify, run:" -ForegroundColor Yellow
    Write-Host "  docker buildx imagetools inspect $FullImageName" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

