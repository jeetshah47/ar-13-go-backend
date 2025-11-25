# PowerShell Script to create MinIO Access Keys using MinIO Client (mc)
# Usage: .\create_minio_access_key.ps1

# Configuration - Update these values
$MINIO_ENDPOINT = if ($env:MINIO_ENDPOINT) { $env:MINIO_ENDPOINT } else { "your-nas-ip:32774" }
$MINIO_ROOT_USER = if ($env:MINIO_ROOT_USER) { $env:MINIO_ROOT_USER } else { "ar-web-backend" }
$MINIO_ROOT_PASSWORD = if ($env:MINIO_ROOT_PASSWORD) { $env:MINIO_ROOT_PASSWORD } else { "bIF62kxCm54D" }
$ACCESS_KEY_NAME = if ($env:ACCESS_KEY_NAME) { $env:ACCESS_KEY_NAME } else { "ar-13-backend" }

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "MinIO Access Key Creation Script" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Check if mc is installed
try {
    $null = Get-Command mc -ErrorAction Stop
} catch {
    Write-Host "ERROR: MinIO Client (mc) is not installed." -ForegroundColor Red
    Write-Host ""
    Write-Host "Install it from: https://min.io/docs/minio/linux/reference/minio-mc.html" -ForegroundColor Yellow
    Write-Host "Or on Windows with Chocolatey: choco install minio-client" -ForegroundColor Yellow
    exit 1
}

Write-Host "Step 1: Configuring MinIO alias..." -ForegroundColor Green
& mc alias set myminio "http://$MINIO_ENDPOINT" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

Write-Host ""
Write-Host "Step 2: Creating user for access key..." -ForegroundColor Green
$USER_NAME = "$ACCESS_KEY_NAME-user"
& mc admin user add myminio "$USER_NAME"

Write-Host ""
Write-Host "Step 3: Creating access key..." -ForegroundColor Green
Write-Host "Creating service account for user: $USER_NAME" -ForegroundColor Yellow

# Create service account and capture output
$OUTPUT = & mc admin user svcacct add myminio "$USER_NAME" --name "$ACCESS_KEY_NAME-key" 2>&1

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Access Key Created Successfully!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host $OUTPUT
Write-Host ""
Write-Host "IMPORTANT: Copy the Access Key and Secret Key above!" -ForegroundColor Yellow
Write-Host "Add them to your .env file:" -ForegroundColor Yellow
Write-Host ""
Write-Host "MINIO_ACCESS_KEY=<Access Key from above>" -ForegroundColor White
Write-Host "MINIO_SECRET_KEY=<Secret Key from above>" -ForegroundColor White
Write-Host ""

