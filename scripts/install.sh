#!/usr/bin/env bash
set -euo pipefail

# Configuration
BINARY_NAME="run"
INSTALL_DIR="/usr/local/bin"
REPO_URL="https://github.com/amoga-io/run.git"
PERSISTENT_DIR="$HOME/.run"

echo "Installing run CLI..."

# Install dependencies
echo "Checking dependencies..."
if ! command -v git &> /dev/null; then
    echo "Installing Git..."
    sudo apt update && sudo apt install -y git
fi

if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    sudo apt update && sudo apt install -y golang-go
fi

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

# Install binary
echo "Installing to $INSTALL_DIR..."
sudo mkdir -p "$INSTALL_DIR"
sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Verify installation
if command -v "$BINARY_NAME" &>/dev/null; then
    echo "Installation successful!"
else
    echo "Binary installed but not in PATH. You may need to restart your terminal."
fi

echo ""
echo "run CLI is ready!"
echo "Usage: run --help"
