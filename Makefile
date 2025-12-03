.PHONY: help test test-coverage test-verbose lint fmt vet build clean run-demo install-deps

help:
	@echo "ChainSystemPro - Makefile Commands"
	@echo "===================================="
	@echo "test             - Run all tests"
	@echo "test-coverage    - Run tests with coverage (requires 90%+)"
	@echo "test-verbose     - Run tests with verbose output"
	@echo "lint             - Run linters (fmt + vet)"
	@echo "fmt              - Format code"
	@echo "vet              - Run go vet"
	@echo "build            - Build server binary"
	@echo "run-demo         - Run HD wallet demo"
	@echo "install-deps     - Install/update dependencies"
	@echo "clean            - Remove build artifacts"

test:
	@echo "Running tests..."
	@go test ./pkg/... -race

test-coverage:
	@echo "Running tests with coverage..."
	@go test ./pkg/... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
	@go tool cover -func=coverage.out | tail -1 | awk '{if ($$3+0 < 90.0) {print "ERROR: Coverage below 90%!"; exit 1}}'

test-verbose:
	@go test -v ./pkg/...

lint: fmt vet
	@echo "Linting complete"

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

build:
	@echo "Building server..."
	@go build -o bin/server cmd/server/main.go

run-demo:
	@echo "Running HD Wallet Demo..."
	@go run examples/wallet_demo/main.go

install-deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

clean:
	@echo "Cleaning..."
	@rm -f bin/server coverage.out
	@go clean -cache
