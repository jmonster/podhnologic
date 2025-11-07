# Makefile for podhnologic

# Ensure homebrew Go is in PATH
export PATH := /opt/homebrew/bin:/usr/local/bin:$(PATH)

# Binary name
BINARY_NAME=podhnologic

# Version
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

.PHONY: all build clean test deps run install build-all download-binaries

all: test build

# Download ffmpeg binaries for embedding
download-binaries:
	@echo "Downloading ffmpeg binaries for all platforms..."
	@./scripts/download-ffmpeg.sh

# Build for current platform (with embedded binaries)
build: check-binaries
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms (with embedded binaries)
build-all: check-binaries build-linux build-darwin build-windows

# Check if binaries directory exists
check-binaries:
	@if [ ! -d "binaries" ]; then \
		echo "Error: binaries/ directory not found."; \
		echo "Run 'make download-binaries' first to download ffmpeg binaries."; \
		exit 1; \
	fi

build-linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .

build-darwin:
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .

build-windows:
	@echo "Building for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Install to system
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Show help
help:
	@echo "Available targets:"
	@echo "  make download-binaries - Download ffmpeg binaries for embedding"
	@echo "  make build             - Build for current platform (with embedded binaries)"
	@echo "  make build-all         - Build for all platforms (with embedded binaries)"
	@echo "  make deps              - Download Go dependencies"
	@echo "  make test              - Run tests"
	@echo "  make clean             - Remove build artifacts"
	@echo "  make install           - Install to /usr/local/bin"
	@echo "  make run               - Build and run"
