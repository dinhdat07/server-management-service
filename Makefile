.PHONY: infra-up infra-down run-api run-monitor run-scheduler dev test-coverage test-coverage-html build-linux docker-up

# Chỉ khởi động các DB/Cache (Dành cho việc dev local)
infra-up:
	docker-compose up -d postgres redis elasticsearch mailhog

# Tắt toàn bộ hạ tầng
infra-down:
	docker-compose down -v

# ============================================
# Local Dev Commands (Chạy trên máy thật)
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

# Build ra file thực thi (.exe trên Windows) để chạy nhanh hơn
build:
	if not exist bin mkdir bin
	go build -o bin/api.exe cmd/api/main.go
	go build -o bin/monitoring-worker.exe cmd/monitoring-worker/main.go
	go build -o bin/daily-scheduler.exe cmd/daily-scheduler/main.go

# Chạy bản đã build (tốc độ khởi động siêu nhanh)
dev-fast: build
	cmd /c start cmd /k bin\api.exe
	cmd /c start cmd /k bin\monitoring-worker.exe
	cmd /c start cmd /k bin\daily-scheduler.exe

# Lệnh này dành cho trường hợp bạn đã có Monitor chạy trong Docker
dev-no-monitor:
	cmd /c start cmd /k go run cmd/api/main.go
	cmd /c start cmd /k go run cmd/daily-scheduler/main.go

# ============================================
# Full Docker Commands (Mô phỏng production)
# ============================================
# Build nhị phân Linux để đưa vào Docker container
build-linux:
	set GOOS=linux&& set GOARCH=amd64&& go build -o bin/simulator-linux cmd/simulator/main.go
	if not exist bin mkdir bin
	set GOOS=linux&& set GOARCH=amd64&& go build -o bin/monitoring-worker-linux cmd/monitoring-worker/main.go

# Chạy TẤT CẢ dịch vụ trong Docker (bao gồm cả Worker và Simulator để chúng nằm chung mạng)
docker-up: build-linux
	docker-compose up -d --build

# ============================================
# Testing
# ============================================
test-coverage:
	go test ./... -coverprofile=coverage.out

test-coverage-html: test-coverage
	go tool cover -html=coverage.out
