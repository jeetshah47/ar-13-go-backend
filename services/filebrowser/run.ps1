# Run FileBrowser Service directly with go run
# Windows PowerShell script

$env:PORT = "8082"
$env:DATA_ROOT = "D:/Resumers/"

Write-Host "Starting FileBrowser Service..." -ForegroundColor Cyan
Write-Host "Port: $env:PORT" -ForegroundColor Yellow
Write-Host "Data Root: $env:DATA_ROOT" -ForegroundColor Yellow
Write-Host ""

# Navigate to the filebrowser directory
Set-Location $PSScriptRoot

# Run the service
go run main.go


