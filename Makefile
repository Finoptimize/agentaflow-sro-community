# Makefile for AgentaFlow SRO Community

.PHONY: all build test clean run examples help

# Variables
BINARY_NAME=agentaflow
CMD_DIR=./cmd/agentaflow
PKG_DIRS=./pkg/...
EXAMPLES_DIR=./examples

all: build test

# Build the main application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test $(PKG_DIRS) -v
	@echo "Tests complete"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test $(PKG_DIRS) -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run the main application
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

# Run GPU scheduling example
example-gpu:
	@echo "Running GPU scheduling example..."
	@cd $(EXAMPLES_DIR) && go run gpu_scheduling.go

# Run model serving example
example-serving:
	@echo "Running model serving example..."
	@cd $(EXAMPLES_DIR) && go run model_serving.go

# Run observability example
example-observability:
	@echo "Running observability example..."
	@cd $(EXAMPLES_DIR) && go run observability.go

# Run all examples
examples: example-gpu example-serving example-observability

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt $(PKG_DIRS)
	@go fmt $(CMD_DIR)
	@echo "Format complete"

# Vet code
vet:
	@echo "Vetting code..."
	@go vet $(PKG_DIRS)
	@go vet $(CMD_DIR)
	@echo "Vet complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Check code quality
check: fmt vet test
	@echo "Code quality check complete"

# Help
help:
	@echo "Available targets:"
	@echo "  make build                 - Build the main application"
	@echo "  make test                  - Run all tests"
	@echo "  make test-coverage         - Run tests with coverage report"
	@echo "  make clean                 - Remove build artifacts"
	@echo "  make run                   - Build and run the main application"
	@echo "  make example-gpu           - Run GPU scheduling example"
	@echo "  make example-serving       - Run model serving example"
	@echo "  make example-observability - Run observability example"
	@echo "  make examples              - Run all examples"
	@echo "  make fmt                   - Format code"
	@echo "  make vet                   - Vet code"
	@echo "  make deps                  - Install dependencies"
	@echo "  make check                 - Run format, vet, and test"
	@echo "  make help                  - Show this help message"
