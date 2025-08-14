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
	@GOOS=windows GOARCH=amd64 go build -o dist/redis-benchmark-windows-amd64.exe cmd/redis-benchmark/main.go
	@GOOS=windows GOARCH=arm64 go build -o dist/redis-benchmark-windows-arm64.exe cmd/redis-benchmark/main.go
	@GOOS=linux GOARCH=amd64 go build -o dist/redis-benchmark-linux-amd64 cmd/redis-benchmark/main.go
	@GOOS=linux GOARCH=arm64 go build -o dist/redis-benchmark-linux-arm64 cmd/redis-benchmark/main.go
	@GOOS=darwin GOARCH=amd64 go build -o dist/redis-benchmark-darwin-amd64 cmd/redis-benchmark/main.go
	@GOOS=darwin GOARCH=arm64 go build -o dist/redis-benchmark-darwin-arm64 cmd/redis-benchmark/main.go

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main
	@rm -f redis-benchmark.exe
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

.PHONY: all build run test clean watch docker-run docker-down itest build-all dist
