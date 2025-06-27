#!/usr/bin/env bash
set -euo pipefail

# Configuration
BINARY_NAME="run"
INSTALL_DIR="/usr/local/bin"
REPO_URL="https://github.com/amoga-io/run.git"
PERSISTENT_DIR="$HOME/.run"

echo "Installing run CLI..."

# Cleanup function for failed installations
cleanup_on_error() {
    echo "Installation failed. Cleaning up..."
    rm -f "$INSTALL_DIR/${BINARY_NAME}.new" 2>/dev/null || true
    exit 1
}

# Set trap for cleanup on error
trap cleanup_on_error ERR

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install dependencies
install_dependencies() {
    echo "Checking dependencies..."
    
    local missing_deps=()
    
    # Check required dependencies
    if ! command_exists git; then
        missing_deps+=("git")
    fi
    
    if ! command_exists go; then
        missing_deps+=("golang-go")
    fi
    
    if ! command_exists sudo; then
        echo "Error: sudo is required but not available"
        exit 1
    fi
    
    # Install missing dependencies
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo "Installing missing dependencies: ${missing_deps[*]}"
        
        # Update package list
        sudo apt update
        
        # Install each dependency
        for dep in "${missing_deps[@]}"; do
            echo "Installing $dep..."
            sudo apt install -y "$dep"
        done
        
        echo "Dependencies installed successfully"
    else
        echo "All dependencies are available"
    fi
}

# Check if running on Ubuntu/Debian
if ! grep -q -i "ubuntu\|debian" /etc/os-release 2>/dev/null; then
    echo "Warning: This CLI is optimized for Ubuntu/Debian systems"
    echo "Proceeding anyway, but some features may not work as expected"
fi

# Install dependencies
install_dependencies

# Clone or update repository
if [ -d "$PERSISTENT_DIR" ]; then
    echo "Updating existing installation..."
    cd "$PERSISTENT_DIR"
    git pull origin main || {
        echo "Git pull failed. Trying fresh clone..."
        cd /
        rm -rf "$PERSISTENT_DIR"
        git clone "$REPO_URL" "$PERSISTENT_DIR"
        cd "$PERSISTENT_DIR"
    }
else
    echo "Cloning repository..."
    git clone "$REPO_URL" "$PERSISTENT_DIR"
    cd "$PERSISTENT_DIR"
fi

# Build binary
echo "Building binary..."
go mod tidy
go build -o "$BINARY_NAME" .

# Verify binary was built
if [ ! -f "$BINARY_NAME" ]; then
    echo "Error: Binary was not created successfully"
    exit 1
fi

# Install binary
echo "Installing to $INSTALL_DIR..."
sudo mkdir -p "$INSTALL_DIR"

# Use atomic installation
TEMP_BINARY="$INSTALL_DIR/${BINARY_NAME}.new"
FINAL_BINARY="$INSTALL_DIR/$BINARY_NAME"

sudo cp "$BINARY_NAME" "$TEMP_BINARY"
sudo chmod +x "$TEMP_BINARY"
sudo mv "$TEMP_BINARY" "$FINAL_BINARY"

# Verify installation
if command_exists "$BINARY_NAME"; then
    echo "✓ Installation successful!"
    echo "✓ run CLI is ready to use"
    echo ""
    echo "Try: run --help"
else
    echo "⚠ Binary installed but not in PATH"
    echo "You may need to restart your terminal or add $INSTALL_DIR to PATH"
fi

echo ""
echo "Usage:"
echo "  run --help    # Show help"
echo "  run update    # Update to latest version"
echo ""
echo "Files:"
echo "  Binary: $INSTALL_DIR/$BINARY_NAME"
echo "  Repository: $PERSISTENT_DIR"
