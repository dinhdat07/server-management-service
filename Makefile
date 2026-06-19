.PHONY: infra-up infra-down run-api run-monitor run-scheduler dev test-coverage test-coverage-summary test-coverage-html build-linux docker-up seed

# Only start DB/Cache (For local development)
infra-up:
	docker-compose up -d postgres redis elasticsearch mailhog

# Shut down all infrastructure
infra-down:
	docker-compose down -v

# ============================================
# Local Dev Commands (Run on host machine)
# ============================================
run-api:
	go run cmd/api/main.go

run-monitor:
	go run cmd/monitoring-worker/main.go

run-scheduler:
	go run cmd/daily-scheduler/main.go

dev:
	cmd /k start cmd /k go run cmd/api/main.go
	cmd /k start cmd /k go run cmd/monitoring-worker/main.go
	cmd /k start cmd /k go run cmd/daily-scheduler/main.go

# This command is for when Monitor is already running in Docker
dev-no-monitor:
	cmd /k start cmd /k go run cmd/api/main.go
	cmd /k start cmd /k go run cmd/daily-scheduler/main.go

# ============================================
# Full Docker Commands (Production simulation)
# ============================================
# Build Linux binaries for Docker containers
build-linux:
	set GOOS=linux&& set GOARCH=amd64&& go build -o bin/simulator-linux cmd/simulator/main.go
	if not exist bin mkdir bin
	set GOOS=linux&& set GOARCH=amd64&& go build -o bin/monitoring-worker-linux cmd/monitoring-worker/main.go

# Run ALL services in Docker (including Worker and Simulator in the same network)
docker-up: build-linux
	docker-compose up -d --build

# Seed 10,000 simulated servers into Postgres and Redis
seed:
	go run cmd/simulator/seed/main.go

# ============================================
# Testing
# ============================================
test-coverage:
	go test ./... -coverprofile=coverage.out

test-coverage-summary: test-coverage
	go run scripts/calc_cov.go

test-coverage-html: test-coverage
	go tool cover -html=coverage.out
