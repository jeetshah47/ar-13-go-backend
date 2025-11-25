# Build FileBrowser Service as a standalone binary
# PowerShell script for Windows

param(
    [string]$OS = "windows",
    [string]$Arch = "amd64",
    [string]$Output = ""
)

# Navigate to the filebrowser directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Determine output filename
if ($Output -eq "") {
    if ($OS -eq "windows") {
        $Output = "filebrowser-service.exe"
    } elseif ($OS -eq "linux") {
        $Output = "filebrowser-service-linux-$Arch"
    } elseif ($OS -eq "darwin") {
        $Output = "filebrowser-service-darwin-$Arch"
    } else {
        $Output = "filebrowser-service-$OS-$Arch"
    }
}

Write-Host "Building FileBrowser Service..." -ForegroundColor Cyan
Write-Host "OS: $OS" -ForegroundColor Yellow
Write-Host "Architecture: $Arch" -ForegroundColor Yellow
Write-Host "Output: $Output" -ForegroundColor Yellow
Write-Host ""

# Set build environment variables
$env:GOOS = $OS
$env:GOARCH = $Arch

# Build the binary
Write-Host "Running: go build -o $Output ." -ForegroundColor Cyan
go build -o $Output .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Successfully built binary!" -ForegroundColor Green
    Write-Host "Output: $Output" -ForegroundColor Cyan
    
    # Show file info
    if (Test-Path $Output) {
        $fileInfo = Get-Item $Output
        Write-Host "Size: $([math]::Round($fileInfo.Length / 1MB, 2)) MB" -ForegroundColor Yellow
        Write-Host "Location: $($fileInfo.FullName)" -ForegroundColor Yellow
    }
} else {
    Write-Host ""
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

