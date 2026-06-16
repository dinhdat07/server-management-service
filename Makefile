.PHONY: infra-up infra-down run-api run-monitor run-scheduler dev test-coverage test-coverage-html build-linux docker-up

# Start only DB/Cache infrastructure (For local development)
infra-up:
	docker-compose up -d postgres redis elasticsearch mailhog

# Shut down all infrastructure
infra-down:
	docker-compose down -v

# ============================================
# Local Dev Commands (Run directly on host machine)
# ============================================
run-api:
	go run cmd/api/main.go

run-monitor:
	go run cmd/monitoring-worker/main.go

run-scheduler:
	go run cmd/daily-scheduler/main.go

dev:
	cmd /c start cmd /k go run cmd/api/main.go
	cmd /c start cmd /k go run cmd/monitoring-worker/main.go
	cmd /c start cmd /k go run cmd/daily-scheduler/main.go

# Build executables (.exe on Windows) for faster startup
build:
	if not exist bin mkdir bin
	go build -o bin/api.exe cmd/api/main.go
	go build -o bin/monitoring-worker.exe cmd/monitoring-worker/main.go
	go build -o bin/daily-scheduler.exe cmd/daily-scheduler/main.go

# Run compiled binaries (super fast startup)
dev-fast: build
	cmd /c start cmd /k bin\api.exe
	cmd /c start cmd /k bin\monitoring-worker.exe
	cmd /c start cmd /k bin\daily-scheduler.exe

# Use this command if Monitoring Worker is already running in Docker
dev-no-monitor:
	cmd /c start cmd /k go run cmd/api/main.go
	cmd /c start cmd /k go run cmd/daily-scheduler/main.go

# Run compiled binaries super fast but WITHOUT Monitoring Worker
dev-fast-no-monitor: build
	cmd /c start cmd /k bin\api.exe
	cmd /c start cmd /k bin\daily-scheduler.exe

# ============================================
# Full Docker Commands (Production simulation)
# ============================================
# Build Linux binaries to run inside Docker containers
build-linux:
	set GOOS=linux&& set GOARCH=amd64&& go build -o bin/simulator-linux cmd/simulator/main.go
	if not exist bin mkdir bin
	set GOOS=linux&& set GOARCH=amd64&& go build -o bin/monitoring-worker-linux cmd/monitoring-worker/main.go

# Run ALL services in Docker (including Worker and Simulator for shared network)
docker-up: build-linux
	docker-compose up -d --build

# ============================================
# Testing
# ============================================
test-coverage:
	go test ./... -coverprofile=coverage.out

test-coverage-html: test-coverage
	go tool cover -html=coverage.out
