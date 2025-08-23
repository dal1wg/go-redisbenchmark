# go-redisbenchmark - A high performance Redis benchmark tool written in Go.
# Supports Redis 6.0+ features including ACL and TLS.
#
# Copyright (c) 2025, dal1wg <15072697283@139.com>
# All rights reserved.
#
# This source code is licensed under the BSD 3-Clause License.
#
# Simple Makefile for a Go project

# Build the application
all: build test

# Detect OS-specific executable suffix
define exe_ext
$(if $(filter Windows_NT,$(OS)),.exe,)
endef

build:
	@echo "Building..."
	
	
	@go build -o goRedisbenchmark$(exe_ext) main.go

# Run the application
run:
	@go run main.go

# Create DB container
docker-run:
	@docker compose up --build

# Shutdown DB container
docker-down:
	@docker compose down

# Redis 6.0+ with TLS and ACL
redis6-up:
	@echo "启动支持TLS和ACL的Redis 6.0+..."
	@docker-compose -f docker-compose.yml up -d

redis6-down:
	@echo "停止Redis 6.0+容器..."
	@docker-compose -f docker-compose.yml down

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

# Cross-compile to dist/
dist:
	@mkdir -p dist

build-all: dist
	@echo "Building cross-platform binaries into dist/ ..."
	@echo "Building Windows binaries..."
	@GOOS=windows GOARCH=amd64 go build -o dist/goRedisbenchmark-windows-amd64.exe main.go
	@GOOS=windows GOARCH=arm64 go build -o dist/goRedisbenchmark-windows-arm64.exe main.go
	@echo "Building Linux binaries (statically linked for old GLIBC compatibility)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -extldflags=-static" -o dist/goRedisbenchmark-linux-amd64 main.go
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -extldflags=-static" -o dist/goRedisbenchmark-linux-arm64 main.go
	@echo "Building macOS binaries..."
	@GOOS=darwin GOARCH=amd64 go build -o dist/goRedisbenchmark-darwin-amd64 main.go
	@GOOS=darwin GOARCH=arm64 go build -o dist/goRedisbenchmark-darwin-arm64 main.go
	@echo "All binaries built successfully!"
	@echo "Linux binaries are statically linked for maximum compatibility with old systems."

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main
	@rm -f goRedisbenchmark.exe
	@rm -rf dist


.PHONY: all build run test clean  docker-run docker-down  build-all dist redis6-up redis6-down certs test-redis6 test-uri
