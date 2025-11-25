# Build and push ARM64 Docker image for filebrowser service
# PowerShell script for Windows

param(
    [string]$Version = "latest"
)

$ImageName = "jeetshah786/ar-13-go-backend-filebrowser-service"
$FullImageName = "$ImageName`:$Version"

Write-Host "Building ARM64 Docker image: $FullImageName" -ForegroundColor Cyan
Write-Host "Architecture: linux/arm64" -ForegroundColor Yellow
Write-Host ""

# Navigate to the filebrowser directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Build and push ARM64 image
Write-Host "Building and pushing ARM64 image..." -ForegroundColor Yellow
Write-Host "This may take a few minutes..." -ForegroundColor Yellow
Write-Host ""

docker buildx build `
    --platform linux/arm64 `
    --tag "$ImageName`:$Version" `
    --tag "$ImageName`:latest" `
    --push `
    .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Successfully built and pushed ARM64 image!" -ForegroundColor Green
    Write-Host "Image: $FullImageName" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "On your NAS, pull with:" -ForegroundColor Yellow
    Write-Host "  docker pull $FullImageName" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

