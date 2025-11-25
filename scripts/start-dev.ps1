# PowerShell script to start frontend and backend with interactive restart capability
# Press 'r' to restart both processes, 'q' to quit

# Get script directory and set paths relative to it
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$backendPath = $scriptDir
$frontendPath = Join-Path (Split-Path -Parent $scriptDir) "ar-13-ui"

$frontendProcess = $null
$backendProcess = $null

function Start-Processes {
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "Starting Frontend and Backend..." -ForegroundColor Cyan
    Write-Host "========================================`n" -ForegroundColor Cyan
    
    # Start Frontend
    Write-Host "[FRONTEND] Starting npm run dev..." -ForegroundColor Green
    $frontendProcess = Start-Process -FilePath "npm" -ArgumentList "run", "dev" -WorkingDirectory $frontendPath -PassThru
    
    # Start Backend
    Write-Host "[BACKEND] Starting go run cmd/server/main.go..." -ForegroundColor Yellow
    $backendProcess = Start-Process -FilePath "go" -ArgumentList "run", "cmd/server/main.go" -WorkingDirectory $backendPath -PassThru
    
    Write-Host "`n[INFO] Both processes started!" -ForegroundColor Cyan
    Write-Host "[INFO] Frontend PID: $($frontendProcess.Id)" -ForegroundColor Gray
    Write-Host "[INFO] Backend PID: $($backendProcess.Id)" -ForegroundColor Gray
    Write-Host "[INFO] Press 'r' to restart both processes" -ForegroundColor Cyan
    Write-Host "[INFO] Press 'q' to quit`n" -ForegroundColor Cyan
    
    return @{
        FrontendProcess = $frontendProcess
        BackendProcess = $backendProcess
    }
}

function Stop-Processes {
    param($Processes)
    
    Write-Host "`n[INFO] Stopping processes..." -ForegroundColor Yellow
    
    if ($Processes.FrontendProcess) {
        try {
            # Stop the main npm process and its children
            if (-not $Processes.FrontendProcess.HasExited) {
                Stop-Process -Id $Processes.FrontendProcess.Id -Force -ErrorAction SilentlyContinue
            }
            # Stop child node processes (vite, etc.)
            Get-Process | Where-Object { 
                $_.ProcessName -eq "node" -and $_.Path -like "*$frontendPath*"
            } | Stop-Process -Force -ErrorAction SilentlyContinue
        } catch {}
    }
    
    if ($Processes.BackendProcess) {
        try {
            # Stop the main go process and its children
            if (-not $Processes.BackendProcess.HasExited) {
                Stop-Process -Id $Processes.BackendProcess.Id -Force -ErrorAction SilentlyContinue
            }
            # Stop the compiled go server process if it exists
            Get-Process | Where-Object { 
                $_.ProcessName -like "*server*" -or 
                ($_.ProcessName -eq "go" -and $_.Path -like "*$backendPath*")
            } | Stop-Process -Force -ErrorAction SilentlyContinue
        } catch {}
    }
    
    Start-Sleep -Milliseconds 500
}

# Cleanup function
function Cleanup {
    param($Processes)
    Write-Host "`n[INFO] Cleaning up..." -ForegroundColor Yellow
    Stop-Processes -Processes $Processes
    Write-Host "[INFO] Exiting..." -ForegroundColor Yellow
    exit
}

# Register cleanup on script exit
Register-EngineEvent PowerShell.Exiting -Action {
    if ($script:processes) {
        Stop-Processes -Processes $script:processes
    }
} | Out-Null

# Main loop
$processes = Start-Processes
$script:processes = $processes

try {
    while ($true) {
        # Check for key input (non-blocking)
        if ([Console]::KeyAvailable) {
            $key = [Console]::ReadKey($true)
            
            if ($key.KeyChar -eq 'r' -or $key.KeyChar -eq 'R') {
                Write-Host "`n[RESTART] Restarting both processes..." -ForegroundColor Magenta
                Stop-Processes -Processes $processes
                Start-Sleep -Seconds 1
                $processes = Start-Processes
                $script:processes = $processes
            }
            elseif ($key.KeyChar -eq 'q' -or $key.KeyChar -eq 'Q') {
                Cleanup -Processes $processes
            }
        }
        
        # Check if processes are still running
        if (($processes.FrontendProcess -and $processes.FrontendProcess.HasExited) -or 
            ($processes.BackendProcess -and $processes.BackendProcess.HasExited)) {
            Write-Host "`n[WARNING] One or more processes exited!" -ForegroundColor Red
            Write-Host "[INFO] Press 'r' to restart or 'q' to quit" -ForegroundColor Yellow
        }
        
        Start-Sleep -Milliseconds 100
    }
}
catch {
    Write-Host "`n[ERROR] An error occurred: $_" -ForegroundColor Red
    Cleanup -Processes $processes
}
finally {
    Cleanup -Processes $processes
}

