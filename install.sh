#!/usr/bin/env bash
set -euo pipefail

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[1;36m'
RED='\033[0;31m'
NC='\033[0m' # No Color

INSTALL_DIR="/usr/local/bin"
REPO_URL="https://github.com/amoga-io/run.git"
CLONE_DIR="$HOME/.devkit"

# Dependency Checks
for cmd in git go sudo; do
    if ! command -v $cmd &> /dev/null; then
        echo -e "${RED}Error: $cmd is not installed. Please install it first.${NC}"
        exit 1
    fi
done

echo -e "${CYAN}Creating install directory...${NC}"
mkdir -p "$INSTALL_DIR"

if [ -d "$CLONE_DIR" ]; then
    echo -e "${YELLOW}devkit is already installed. Pulling latest changes...${NC}"
    cd "$CLONE_DIR"
    git pull origin main
else
    echo -e "${CYAN}Cloning devkit...${NC}"
    git clone "$REPO_URL" "$CLONE_DIR"
    cd "$CLONE_DIR"
fi

echo -e "${CYAN}Building devkit for your system...${NC}"
# Use make if available for better version handling, otherwise fallback to go build
if command -v make &> /dev/null && [ -f "Makefile" ]; then
    make build
else
    go build -o devkit
fi

echo -e "${CYAN}Installing to $INSTALL_DIR...${NC}"

# Use atomic replacement to avoid "text file busy" error
TEMP_BINARY="$INSTALL_DIR/devkit.new"
FINAL_BINARY="$INSTALL_DIR/devkit"

# Copy to temporary location first
sudo cp devkit "$TEMP_BINARY"
sudo chmod +x "$TEMP_BINARY"

# Atomically move to final location (this avoids "text file busy" error)
sudo mv "$TEMP_BINARY" "$FINAL_BINARY"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo -e "${YELLOW}WARNING: $INSTALL_DIR is not in your PATH.${NC}"
    echo "Add this to your shell config:"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\""
fi

echo -e "${GREEN}Installation complete! You can now run devkit 🚀${NC}"
