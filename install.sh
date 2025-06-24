#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="/usr/local/bin"
REPO_URL="https://github.com/amoga-io/run.git"
CLONE_DIR="$HOME/.gocli"

# Dependency Checks
for cmd in git go sudo; do
    if ! command -v $cmd &> /dev/null; then
        echo "Error: $cmd is not installed. Please install it first."
        exit 1
    fi
done

mkdir -p "$INSTALL_DIR"

if [ -d "$CLONE_DIR" ]; then
    echo "gocli is already installed. Pulling latest changes..."
    cd "$CLONE_DIR"
    git pull origin main
else
    echo "Cloning gocli..."
    git clone "$REPO_URL" "$CLONE_DIR"
    cd "$CLONE_DIR"
fi

echo "Building gocli for your system..."
# Use make if available for better version handling, otherwise fallback to go build
if command -v make &> /dev/null && [ -f "Makefile" ]; then
    make build
else
    go build -o gocli
fi

echo "Installing to $INSTALL_DIR..."

# Use atomic replacement to avoid "text file busy" error
TEMP_BINARY="$INSTALL_DIR/gocli.new"
FINAL_BINARY="$INSTALL_DIR/gocli"

# Copy to temporary location first
sudo cp gocli "$TEMP_BINARY"
sudo chmod +x "$TEMP_BINARY"

# Atomically move to final location (this avoids "text file busy" error)
sudo mv "$TEMP_BINARY" "$FINAL_BINARY"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "WARNING: $INSTALL_DIR is not in your PATH."
    echo "Add this to your shell config:"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\""
fi

echo "Installation complete! You can now run gocli 🚀"
