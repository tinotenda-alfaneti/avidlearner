# Setup script for AvidLearner after updates

Write-Host " Setting up AvidLearner with new features..." -ForegroundColor Cyan

# Navigate to backend
Set-Location -Path "$PSScriptRoot\..\backend"

Write-Host "`n Updating Go module dependencies..." -ForegroundColor Yellow
go mod tidy

if ($LASTEXITCODE -eq 0) {
    Write-Host " Go modules updated successfully" -ForegroundColor Green
} else {
    Write-Host " Failed to update Go modules" -ForegroundColor Red
    exit 1
}

# Navigate to frontend
Set-Location -Path "$PSScriptRoot\..\frontend"

Write-Host "`n Installing frontend dependencies..." -ForegroundColor Yellow
npm install

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Frontend dependencies installed successfully" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to install frontend dependencies" -ForegroundColor Red
    exit 1
}

Write-Host "`n Setup complete! You can now:" -ForegroundColor Cyan
Write-Host "  1. Copy .env.example to .env and configure AI (optional)" -ForegroundColor White
Write-Host "  2. Run: pwsh scripts/run.ps1" -ForegroundColor White

Set-Location -Path "$PSScriptRoot\.."
