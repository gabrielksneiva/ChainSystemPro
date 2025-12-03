.PHONY: help build test coverage lint run clean install-tools swagger

help:
	@echo "ChainSystemPro - Unified Multi-Chain Connector"
	@echo ""
	@echo "Available targets:"
	@echo "  make build         - Build the application"
	@echo "  make test          - Run all tests"
	@echo "  make coverage      - Run tests with coverage report"
	@echo "  make lint          - Run linters"
	@echo "  make run           - Run the server"
	@echo "  make swagger       - Generate Swagger documentation"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install-tools - Install development tools"

build:
	@echo "Building..."
	go build -o bin/server ./cmd/server

test:
	@echo "Running tests..."
	go test ./... -v -race

coverage:
	@echo "Generating coverage report..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	go tool cover -func=coverage.out

lint:
	@echo "Running linters..."
	golangci-lint run --timeout 5m

run: build
	@echo "Starting server..."
	./bin/server

swagger:
	@echo "Generating Swagger documentation..."
	swag init -g internal/api/server.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation generated at docs/"
	@echo "Access at: http://localhost:8080/swagger/index.html"

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/swaggo/swag/cmd/swag@latest
