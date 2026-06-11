# Run full Bruno test suite (idempotent - no docker reset needed)
# Usage: .\run_tests.ps1

$ErrorActionPreference = "Continue"

Push-Location "$PSScriptRoot\.."

# ---- Build ----
Write-Host "=== Building server ===" -ForegroundColor Cyan
go build -o server.exe ./cmd/api/
if ($LASTEXITCODE -ne 0) { Write-Host "Build failed!" -ForegroundColor Red; exit 1 }

# ---- Restart ----
Write-Host "=== Restarting server ===" -ForegroundColor Cyan
$p = (Get-NetTCPConnection -LocalPort 8000 -ErrorAction SilentlyContinue).OwningProcess
if ($p) { Stop-Process -Id $p -Force; Start-Sleep 1 }
Start-Process -FilePath ".\server.exe" -WorkingDirectory (Get-Location) -WindowStyle Hidden
Start-Sleep 4

# ---- Cleanup stale test data via API ----
Write-Host "=== Cleaning up stale test data ===" -ForegroundColor Cyan
Push-Location "$PSScriptRoot"
& "$PSScriptRoot\cleanup.ps1"

# ---- Run tests ----
Write-Host "=== Running Bruno tests ===" -ForegroundColor Cyan
bru run auth servers reporting health authorization --env local

Write-Host "`n=== Done ===" -ForegroundColor Green
