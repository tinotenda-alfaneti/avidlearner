param(
    [string]$BackendPort = "8081",
    [string]$FrontendPort = "5173"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Definition)
$backendDir  = Join-Path $root "backend"
$frontendDir = Join-Path $root "frontend"
$backendExe  = Join-Path $backendDir "backend.exe"

function Stop-Port($port) {
    $conn = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    if ($conn) {
        $procid = $conn.OwningProcess
        Write-Host ("Port {0} in use by PID {1}. Killing it..." -f $port, $procid) -ForegroundColor Yellow
        Stop-Process -Id $procid -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
    }
}

Write-Host "========== AVID LEARNER DEV RUNNER ==========" -ForegroundColor Cyan
Stop-Port $BackendPort
Stop-Port $FrontendPort

# --- Build and start backend ---
Push-Location $backendDir
Write-Host "Building Go backend..." -ForegroundColor Cyan
if (Test-Path $backendExe) { Remove-Item $backendExe -Force }
go build -o backend.exe .
Pop-Location

Write-Host "Starting backend on port $BackendPort..." -ForegroundColor Green
$env:PORT = $BackendPort
$backendProcess = Start-Process -FilePath $backendExe -WorkingDirectory $backendDir -PassThru

# --- Start frontend ---
Write-Host "Starting frontend on port $FrontendPort..." -ForegroundColor Green
Push-Location $frontendDir
if (-not (Test-Path -Path "node_modules")) {
    Write-Host "Installing npm dependencies..." -ForegroundColor Cyan
    npm install
}
$frontendProcess = Start-Process "npm" -ArgumentList "run", "dev" -WorkingDirectory $frontendDir -PassThru
Pop-Location

Write-Host ""
Write-Host ("Backend:  http://localhost:{0}" -f $BackendPort)
Write-Host ("Frontend: http://localhost:{0}" -f $FrontendPort)
Write-Host "Press Ctrl+C to stop both." -ForegroundColor Cyan

try {
    while ($true) {
        Start-Sleep -Seconds 1
    }
}
finally {
    Write-Host "`nStopping backend and frontend..." -ForegroundColor Yellow
    Stop-Process -Id $backendProcess.Id -Force -ErrorAction SilentlyContinue
    Stop-Process -Id $frontendProcess.Id -Force -ErrorAction SilentlyContinue
}
