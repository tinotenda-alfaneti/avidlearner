param(
    [string]$Port = '8081'
)

# Build
Write-Host "Building backend..."
Push-Location (Join-Path $PSScriptRoot "..\backend")
if(-not (Test-Path -Path "backend.exe")){
    go build -o backend.exe .
}
# Run
Write-Host "Starting backend on port $Port..."
$env:PORT = $Port
Start-Process -NoNewWindow -FilePath (Join-Path $PSScriptRoot "..\backend\backend.exe") -WorkingDirectory (Join-Path $PSScriptRoot "..\backend")
Write-Host "Backend started; open http://localhost:$Port/"
Pop-Location
