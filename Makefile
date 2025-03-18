# Binary names
BINARY_NAME=gcp-switcher
BINARY_WINDOWS=$(BINARY_NAME).exe

# Build flags
BUILD_FLAGS=-v

.PHONY: all build build-mac build-windows clean

all: build-mac build-windows

build-mac:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)

build-windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_WINDOWS)

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_WINDOWS)

.DEFAULT_GOAL := all