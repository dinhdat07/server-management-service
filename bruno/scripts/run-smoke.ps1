param([string]$Env = "local")
$ErrorActionPreference = "Stop"
$collectionDir = Resolve-Path "$PSScriptRoot\.."

Write-Host "=== Bruno API Smoke Tests - SMS ===" -ForegroundColor Cyan
Write-Host "Environment: $Env" -ForegroundColor Cyan

Push-Location $collectionDir
$folders = @("health", "auth", "servers", "reporting", "authorization")
$failed = @()
foreach ($folder in $folders) {
    Write-Host "`n--- $folder ---" -ForegroundColor Yellow
    bru run $folder --env $Env
    if ($LASTEXITCODE -ne 0) {
        $failed += $folder
        Write-Host "  FAILED" -ForegroundColor Red
    } else {
        Write-Host "  PASSED" -ForegroundColor Green
    }
}
Pop-Location

if ($failed.Count -eq 0) {
    Write-Host "`nALL TESTS PASSED" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`nFAILED: $($failed -join ', ')" -ForegroundColor Red
    exit 1
}
