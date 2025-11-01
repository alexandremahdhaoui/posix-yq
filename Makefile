.PHONY: build clean generate test-unit-generator test-unit-posix-yq test-unit test-e2e help

# Default target
all: build generate

# Build the Go binary into ./build/
build:
	@echo "Building generator binary..."
	@mkdir -p build
	@go build -o build/generator ./cmd/generator

# Clean build artifacts
clean:
	@echo "Cleaning build directory..."
	@rm -rf ./build/*
	@rm -f ./posix-yq

# Generate the posix-yq script
generate:
	@echo "Generating posix-yq script..."
	@go run cmd/generator/main.go > posix-yq
	@chmod +x posix-yq
	@echo "posix-yq script generated successfully"

# Run unit tests for the Go generator
test-unit-generator:
	@echo "Running Go generator unit tests..."
	@if go list ./pkg/generator/... >/dev/null 2>&1; then \
		go test ./pkg/generator/... -v; \
	else \
		echo "No Go generator tests found, skipping..."; \
	fi

# Run unit tests for the posix-yq script
test-unit-posix-yq:
	@echo "Running posix-yq script unit tests..."
	@./test/unit/run_tests.sh

# Run all unit tests
test-unit: build generate
	@echo "Running all unit tests..."
	@echo ""
	@echo "========================================="
	@echo "Running Edge-CD Real-World Tests"
	@echo "========================================="
	@./test/yq-edge-cd-tests.sh ./posix-yq
	@echo ""
	@echo "========================================="
	@echo "Running Unit Test Scenarios"
	@echo "========================================="
	@./test/unit/run_tests.sh 
	@echo ""
	@echo "Unit tests completed"

# Run E2E tests (depends on build and generate)
test-e2e: build generate
	@echo "Running E2E tests..."
	@./test/e2e/run_tests.sh

# Run all tests
test: build generate
	@echo "Running all tests..."
	@./test/unit/run_tests.sh 
	@./test/e2e/run_tests.sh
	@echo "All tests completed"

# Help target
help:
	@echo "Available targets:"
	@echo "  make build              - Build the Go binary into ./build/"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make generate           - Generate the posix-yq script"
	@echo "  make test-unit-generator - Run Go generator unit tests"
	@echo "  make test-unit-posix-yq - Run posix-yq script unit tests against test scenarios"
	@echo "  make test-unit          - Run all unit tests (edge-cd + scenarios)"
	@echo "  make test-e2e           - Run E2E tests (depends on test-unit)"
	@echo "  make test               - Run all tests (unit + E2E)"
	@echo "  make help               - Show this help message"
