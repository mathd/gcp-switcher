.PHONY: build clean run build-all

BINARY_NAME=gcp-switcher
BINARY_PATH=bin/$(BINARY_NAME)
VERSION=1.0.0

# Platform specific extensions and names
ifeq ($(OS),Windows_NT)
	BINARY_EXT=.exe
else
	BINARY_EXT=
endif

# Build flags
BUILD_FLAGS=-ldflags "-X github.com/mathd/gcp-switcher/internal/version.Version=$(VERSION) \
-X github.com/mathd/gcp-switcher/internal/version.Commit=$(shell git rev-parse --short HEAD) \
-X github.com/mathd/gcp-switcher/internal/version.Date=$(shell date -u +"%Y-%m-%d_%H:%M:%S")"

build:
	@echo "Building for current platform..."
	@mkdir -p bin
	@go build $(BUILD_FLAGS) -o $(BINARY_PATH)$(BINARY_EXT)
	@echo "Build complete: $(BINARY_PATH)$(BINARY_EXT)"

# Cross compilation targets
build-all: build-linux build-windows build-mac

build-linux:
	@echo "Building for Linux..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux-amd64
	@echo "Linux build complete"

build-windows:
	@echo "Building for Windows..."
	@mkdir -p bin
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe
	@echo "Windows build complete"

build-mac:
	@echo "Building for macOS..."
	@mkdir -p bin
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-amd64
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-arm64
	@echo "macOS build complete"

run: build
	@./$(BINARY_PATH)$(BINARY_EXT)

run-debug: build
	@./$(BINARY_PATH)$(BINARY_EXT) --debug

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f gcp-switcher.log
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test ./...
	@echo "Tests complete"

# Install development dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

# Run linter
lint:
	@echo "Running linter..."
	@go vet ./...
	@echo "Lint complete"

# Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build for current platform"
	@echo "  build-all       - Build for all platforms (Linux, Windows, macOS)"
	@echo "  build-linux     - Build for Linux"
	@echo "  build-windows   - Build for Windows"
	@echo "  build-mac       - Build for macOS (Intel and ARM)"
	@echo "  run             - Build and run the application"
	@echo "  run-debug       - Build and run with debug logging enabled"
	@echo "  clean           - Remove built binary and log files"
	@echo "  test            - Run tests"
	@echo "  deps            - Install dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run linter"
	@echo "  help            - Show this help"