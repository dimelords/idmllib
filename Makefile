.PHONY: help build test test-coverage lint fmt clean install

# Default target
help:
	@echo "Available targets:"
	@echo "  build          - Build the CLI binary"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linters"
	@echo "  fmt            - Format code"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install CLI to GOPATH/bin"

# Build the CLI tool
build:
	@echo "Building idmllib CLI..."
	@go build -v -o bin/idmllib ./cmd/idmllib

# Run all tests
test:
	@echo "Running tests..."
	@go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@goimports -w .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.txt coverage.html
	@rm -rf idml/testdata/output/

# Install CLI to GOPATH/bin
install:
	@echo "Installing idmllib..."
	@go install ./cmd/idmllib
