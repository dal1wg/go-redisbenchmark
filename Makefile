# Simple Makefile for a Go project

# Build the application
all: build test

# Detect OS-specific executable suffix
define exe_ext
$(if $(filter Windows_NT,$(OS)),.exe,)
endef

build:
	@echo "Building..."
	
	
	@go build -o go-redisbenchmark$(exe_ext) cmd/redis-benchmark/main.go

# Run the application
run:
	@go run cmd/redis-benchmark/main.go

# Create DB container
docker-run:
	@docker compose up --build

# Shutdown DB container
docker-down:
	@docker compose down

# Redis 6.0+ with TLS and ACL
redis6-up:
	@echo "启动支持TLS和ACL的Redis 6.0+..."
	@docker-compose -f docker-compose.redis6.yml up -d

redis6-down:
	@echo "停止Redis 6.0+容器..."
	@docker-compose -f docker-compose.redis6.yml down

# Generate TLS certificates
certs:
	@echo "生成TLS证书..."
	@chmod +x scripts/generate-certs.sh
	@./scripts/generate-certs.sh

# Test Redis 6.0+ features
test-redis6:
	@echo "测试Redis 6.0+ ACL和TLS功能..."
	@go run test/test-redis6-features.go

# Test Redis URI parsing
test-uri:
	@echo "测试Redis URI解析功能..."
	@go run test/test-redis-uri.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Cross-compile to dist/
dist:
	@mkdir -p dist

build-all: dist
	@echo "Building cross-platform binaries into dist/ ..."
	@GOOS=windows GOARCH=amd64 go build -o dist/go-redisbenchmark-windows-amd64.exe cmd/redis-benchmark/main.go
	@GOOS=windows GOARCH=arm64 go build -o dist/go-redisbenchmark-windows-arm64.exe cmd/redis-benchmark/main.go
	@GOOS=linux GOARCH=amd64 go build -o dist/go-redisbenchmark-linux-amd64 cmd/redis-benchmark/main.go
	@GOOS=linux GOARCH=arm64 go build -o dist/go-redisbenchmark-linux-arm64 cmd/redis-benchmark/main.go
	@GOOS=darwin GOARCH=amd64 go build -o dist/go-redisbenchmark-darwin-amd64 cmd/redis-benchmark/main.go
	@GOOS=darwin GOARCH=arm64 go build -o dist/go-redisbenchmark-darwin-arm64 cmd/redis-benchmark/main.go

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main
	@rm -f go-redisbenchmark.exe
	@rm -rf dist

# Live Reload
watch:
	@powershell -ExecutionPolicy Bypass -Command "if (Get-Command air -ErrorAction SilentlyContinue) { \
		air; \
		Write-Output 'Watching...'; \
	} else { \
		Write-Output 'Installing air...'; \
		go install github.com/air-verse/air@latest; \
		air; \
		Write-Output 'Watching...'; \
	}"

.PHONY: all build run test clean watch docker-run docker-down itest build-all dist redis6-up redis6-down certs test-redis6 test-uri
