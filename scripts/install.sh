#!/usr/bin/env bash
set -euo pipefail

# Configuration
BINARY_NAME="run"
INSTALL_DIR="/usr/local/bin"
REPO_URL="https://github.com/amoga-io/run.git"
PERSISTENT_DIR="$HOME/.run"

# Set non-interactive mode for apt
export DEBIAN_FRONTEND=noninteractive

echo "Installing run CLI..."

# Step 1: Check if binary and directory already exist, remove them
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
existing_installation=false

if [ -f "$BINARY_PATH" ] || [ -d "$PERSISTENT_DIR" ]; then
    existing_installation=true
    echo "Existing installation detected - cleaning up..."
    
    if [ -f "$BINARY_PATH" ]; then
        sudo rm -f "$BINARY_PATH"
    fi
    
    if [ -d "$PERSISTENT_DIR" ]; then
        rm -rf "$PERSISTENT_DIR"
    fi
    
    echo "✓ Existing installation cleaned up"
fi

# Cleanup function for failed installations
cleanup_on_error() {
    echo "✗ Installation failed. Cleaning up..."
    sudo rm -f "$INSTALL_DIR/${BINARY_NAME}.new" 2>/dev/null || true
    sudo rm -f "$BINARY_PATH" 2>/dev/null || true
    rm -rf "$PERSISTENT_DIR" 2>/dev/null || true
    exit 1
}

# Set trap for cleanup on error
trap cleanup_on_error ERR

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for apt locks to clear
wait_for_apt() {
    local timeout=60
    local count=0
    
    while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1; do
        if [ $count -ge $timeout ]; then
            echo "✗ Timeout waiting for package manager locks to clear"
            exit 1
        fi
        echo "Waiting for other package managers to finish... ($count/$timeout)"
        sleep 2
        count=$((count + 2))
    done
}

# Step 2: Bootstrap minimal dependencies for CLI build (SILENT)
bootstrap_dependencies() {
    echo "Bootstrapping minimal dependencies..."

    local bootstrap_deps=()

    # Only install what's absolutely needed to build CLI
    if ! command_exists git; then
        bootstrap_deps+=("git")
    fi

    if ! command_exists go; then
        bootstrap_deps+=("golang-go")
    fi

    if ! command_exists sudo; then
        echo "✗ Error: sudo is required but not available"
        exit 1
    fi

    # Install bootstrap dependencies SILENTLY
    if [ ${#bootstrap_deps[@]} -gt 0 ]; then
        echo "Installing bootstrap dependencies: ${bootstrap_deps[*]}"
        
        # Wait for any existing apt processes to finish
        wait_for_apt
        
        # Update package list silently
        echo "Updating package list..."
        sudo apt-get update -qq >/dev/null 2>&1
        
        # Install dependencies silently without user interaction
        echo "Installing packages..."
        sudo apt-get install -y -qq --no-install-recommends "${bootstrap_deps[@]}" >/dev/null 2>&1
        
        echo "✓ Bootstrap dependencies installed successfully"
    else
        echo "✓ Bootstrap dependencies already available"
    fi
}
        
        echo "✓ Dependencies installed successfully"
    else
        echo "✓ All required dependencies are available"
    fi
}

# Check if running on Ubuntu/Debian (don't warn, just note)
if ! grep -q -i "ubuntu\|debian" /etc/os-release 2>/dev/null; then
    echo "Note: Optimized for Ubuntu/Debian systems"
fi

# Step 2: Install dependencies
bootstrap_dependencies

# Step 3: Clone and build CLI
echo "Cloning repository..."
git clone "$REPO_URL" "$PERSISTENT_DIR" >/dev/null 2>&1
cd "$PERSISTENT_DIR"

# Build binary
echo "Building CLI..."
go mod tidy >/dev/null 2>&1

# Get version information for build
VERSION=$(git describe --tags --always 2>/dev/null || echo "v0.0.0-dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "Building version: $VERSION"

# Build with version information embedded (silently)
go build -ldflags "\
  -X 'github.com/amoga-io/run/cmd.Version=$VERSION' \
  -X 'github.com/amoga-io/run/cmd.GitCommit=$COMMIT' \
  -X 'github.com/amoga-io/run/cmd.BuildDate=$BUILD_DATE'" \
  -o "$BINARY_NAME" . >/dev/null 2>&1

# Verify binary was built
if [ ! -f "$BINARY_NAME" ]; then
    echo "✗ Error: Binary was not created successfully"
    exit 1
fi

# Step 4: Install binary and let CLI handle advanced dependency management
echo "Installing binary to $INSTALL_DIR..."
sudo mkdir -p "$INSTALL_DIR"

# Use atomic installation to prevent "text file busy" errors
TEMP_BINARY="$INSTALL_DIR/${BINARY_NAME}.new"

sudo cp "$BINARY_NAME" "$TEMP_BINARY"
sudo chmod +x "$TEMP_BINARY"
sudo mv "$TEMP_BINARY" "$BINARY_PATH"

# Step 5: Let CLI verify its own installation
echo "Finalizing installation..."
"$BINARY_PATH" internal verify-installation || {
    echo "✗ Installation verification failed"
    exit 1
}

# Verify installation (silent check)
if command_exists "$BINARY_NAME"; then
    echo "✓ Installation completed successfully!"
    
    if [ "$existing_installation" = true ]; then
        echo "✓ Existing installation was updated"
    else
        echo "✓ Fresh installation completed"
    fi
    
    echo ""
    echo "run CLI is ready to use!"
    echo ""
    echo "Try: run --help or run version"
else
    echo "⚠ Binary installed but not found in PATH"
    echo "You may need to restart your terminal"
fi
