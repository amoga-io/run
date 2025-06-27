#!/bin/bash

export DEBIAN_FRONTEND=noninteractive

echo "Installing Python with pip and development tools..."

# Update package list silently
sudo apt-get update -qq >/dev/null 2>&1

# Install Python and essential packages
sudo apt-get install -y -qq python3 python3-pip python3-dev python3-venv python3-setuptools >/dev/null 2>&1

# Create symbolic link for python command if it doesn't exist
if ! command -v python >/dev/null 2>&1; then
    sudo ln -sf /usr/bin/python3 /usr/local/bin/python
fi

# Verify pip installation
if ! command -v pip3 >/dev/null 2>&1; then
    echo "Error: pip3 installation failed"
    exit 1
fi

# Upgrade pip to latest version
python3 -m pip install --upgrade pip >/dev/null 2>&1

echo "✓ Python installed successfully"
python3 --version
pip3 --version