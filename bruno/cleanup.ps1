# Clean up test servers via API (no docker reset needed)
# Usage: .\cleanup.ps1
# Reads credentials from environments/local.bru

param([string]$BaseURL = "http://localhost:8000")

$ErrorActionPreference = "Continue"

# Read credentials from Bruno env (same as bru run --env local)
$envFile = "$PSScriptRoot\environments\local.bru"
$adminEmail = "admin@example.com"
$adminPass = "12345678"
if (Test-Path $envFile) {
    $envContent = Get-Content $envFile -Raw
    if ($envContent -match 'adminEmail:\s*(\S+)') { $adminEmail = $matches[1] }
    if ($envContent -match 'adminPassword:\s*(\S+)') { $adminPass = $matches[1] }
}

Write-Host "=== Logging in as admin ===" -ForegroundColor Cyan
$loginBody = @{identifier=$adminEmail;password=$adminPass} | ConvertTo-Json
$login = Invoke-RestMethod -Uri "$BaseURL/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $login.accessToken
$headers = @{Authorization = "Bearer $token"}

Write-Host "=== Finding test servers ===" -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "$BaseURL/api/v1/servers?page=1&limit=100" -Headers $headers
$testServers = $response.servers | Where-Object {
    $_.serverName -like "import-srv-*" -or
    $_.serverName -like "bruno-test-*" -or
    $_.serverName -like "unique-*" -or
    $_.serverName -like "conflict-*"
}

if (-not $testServers) {
    Write-Host "No test servers found, DB is clean." -ForegroundColor Green
    exit 0
}

Write-Host "Deleting $($testServers.Count) test servers..." -ForegroundColor Yellow
foreach ($srv in $testServers) {
    $url = "$BaseURL/api/v1/servers/$($srv.serverId)"
    Invoke-RestMethod -Uri $url -Method Delete -Headers $headers | Out-Null
    Write-Host "  Deleted: $($srv.serverName) ($($srv.ipv4))"
}

Write-Host "Done. DB is clean." -ForegroundColor Green
