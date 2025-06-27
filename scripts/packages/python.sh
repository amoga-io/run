#!/bin/bash

export DEBIAN_FRONTEND=noninteractive
set -euo pipefail

status_update() { echo "  $1"; }
step_complete() { echo "✓ $1"; }
step_error() { echo "❌ $1"; exit 1; }

echo "🐍 Installing Python..."

# Update package list silently
status_update "Updating package lists..."
sudo apt-get update -qq >/dev/null 2>&1 || step_error "Failed to update package lists"
step_complete "Package lists updated"

# Install Python and essential packages
status_update "Installing Python and development tools..."
sudo apt-get install -y -qq python3 python3-pip python3-dev python3-venv python3-setuptools >/dev/null 2>&1 || step_error "Failed to install Python packages"
step_complete "Python and development tools installed"

# Create symbolic link for python command if it doesn't exist
status_update "Creating python symlink..."
if ! command -v python >/dev/null 2>&1; then
    sudo ln -sf /usr/bin/python3 /usr/local/bin/python >/dev/null 2>&1
fi
step_complete "Python symlink created"

# Verify pip installation
status_update "Verifying pip installation..."
if ! command -v pip3 >/dev/null 2>&1; then
    step_error "pip3 installation failed"
fi
step_complete "pip3 installation verified"

# Upgrade pip to latest version
status_update "Upgrading pip to latest version..."
python3 -m pip install --upgrade pip >/dev/null 2>&1 || step_error "Failed to upgrade pip"
step_complete "pip upgraded to latest version"

# Verify installations
status_update "Verifying Python installation..."
if python3 --version >/dev/null 2>&1 && pip3 --version >/dev/null 2>&1; then
    step_complete "Python installation verified"
    echo "  $(python3 --version)"
    echo "  $(pip3 --version)"
else
    step_error "Python installation verification failed"
fi

echo "🎉 Python installed successfully!"