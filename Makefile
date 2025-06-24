# Makefile for gocli

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -ldflags "-X 'github.com/amoga-io/run/cmd.Version=$(VERSION)' \
                   -X 'github.com/amoga-io/run/cmd.GitCommit=$(COMMIT)' \
                   -X 'github.com/amoga-io/run/cmd.BuildDate=$(BUILD_DATE)'"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building gocli $(VERSION)..."
	go build $(LDFLAGS) -o gocli

# Build and install to /usr/local/bin
.PHONY: install
install: build
	@echo "Installing gocli to /usr/local/bin..."
	sudo cp gocli /usr/local/bin/gocli
	sudo chmod +x /usr/local/bin/gocli

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f gocli

# Show version information
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Create a new release tag
.PHONY: tag
tag:
	@if [ -z "$(v)" ]; then \
		echo "Usage: make tag v=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating tag $(v)..."
	git tag -a $(v) -m "Release $(v)"
	@echo "Tag created. Push with: git push origin $(v)"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build     - Build the gocli binary"
	@echo "  install   - Build and install to /usr/local/bin"
	@echo "  clean     - Remove build artifacts"
	@echo "  version   - Show version information"
	@echo "  tag       - Create a release tag (usage: make tag v=v1.0.0)"
	@echo "  help      - Show this help message"
