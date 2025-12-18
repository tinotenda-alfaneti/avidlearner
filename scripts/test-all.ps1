#!/usr/bin/env pwsh
# Test script for running all tests (backend + frontend)

Write-Host "Running AvidLearner Test Suite" -ForegroundColor Cyan
Write-Host ""

$failed = $false

# Backend tests
Write-Host "Running Backend Tests..." -ForegroundColor Yellow
Push-Location "$PSScriptRoot\..\backend"
go test ./... -v
if ($LASTEXITCODE -ne 0) {
    $failed = $true
    Write-Host "❌ Backend tests failed" -ForegroundColor Red
} else {
    Write-Host "✅ Backend tests passed" -ForegroundColor Green
}
Pop-Location

Write-Host ""

# Frontend tests
Write-Host "Running Frontend Tests..." -ForegroundColor Yellow
Push-Location "$PSScriptRoot\..\frontend"
npm test -- --run
if ($LASTEXITCODE -ne 0) {
    $failed = $true
    Write-Host "❌ Frontend tests failed" -ForegroundColor Red
} else {
    Write-Host "✅ Frontend tests passed" -ForegroundColor Green
}
Pop-Location

Write-Host ""

if ($failed) {
    Write-Host "❌ Some tests failed. Please fix before committing." -ForegroundColor Red
    exit 1
} else {
    Write-Host "✅ All tests passed!" -ForegroundColor Green
    exit 0
}
