# Full simulation workflow: build → seed → run → report
# Usage: .\run.ps1 [-Count 10000] [-Rounds 10] [-TogglePct 5]

param(
    [int]$Count = 10000,
    [int]$Rounds = 10,
    [int]$TogglePct = 5
)

$ErrorActionPreference = "Continue"

$env:SIMULATOR_IP_COUNT = $Count
$env:SIMULATION_ROUNDS = $Rounds
$env:SIMULATION_TOGGLE_PCT = $TogglePct

Push-Location $PSScriptRoot\..\..

Write-Host "=== Building ===" -ForegroundColor Cyan
go build -o cmd/simulator/simulator.exe ./cmd/simulator/
go build -o cmd/simulator/seed/seed.exe ./cmd/simulator/seed/
go build -o cmd/simulator/run/simulation.exe ./cmd/simulator/run/

Write-Host "=== Starting simulator container ===" -ForegroundColor Cyan
docker-compose up -d simulator
Start-Sleep 5

Write-Host "=== Seeding $Count servers ===" -ForegroundColor Cyan
Push-Location cmd/simulator/seed
.\seed.exe
if ($LASTEXITCODE -ne 0) { throw "Seed failed" }

Write-Host "=== Running $Rounds simulation rounds ===" -ForegroundColor Cyan
Push-Location ..\run
.\simulation.exe

Write-Host "`n=== Done ===" -ForegroundColor Green
