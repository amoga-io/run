#!/bin/bash

export DEBIAN_FRONTEND=noninteractive
set -euo pipefail

status_update() { echo "  $1"; }
step_complete() { echo "✓ $1"; }
step_error() { echo "❌ $1"; exit 1; }

echo "🟢 Installing Node.js..."

# Install Node.js 20
status_update "Adding NodeSource repository..."
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash - >/dev/null 2>&1 || step_error "Failed to add NodeSource repository"
step_complete "NodeSource repository added"

status_update "Installing Node.js..."
sudo apt-get install -y -qq nodejs >/dev/null 2>&1 || step_error "Failed to install Node.js"
step_complete "Node.js installed"

# Create npm global directory in user's home
status_update "Configuring npm global directory..."
mkdir -p ~/.npm-global >/dev/null 2>&1
npm config set prefix ~/.npm-global >/dev/null 2>&1
step_complete "npm global directory configured"

# Add npm global bin to PATH in ~/.profile if not already present
status_update "Updating PATH configuration..."
if ! grep -q "PATH=~/.npm-global/bin:\$PATH" ~/.profile 2>/dev/null; then
    echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.profile
fi
step_complete "PATH updated"

# Source the updated profile
export PATH=~/.npm-global/bin:$PATH

# Install pnpm 9.10.0 using npm
status_update "Installing pnpm package manager..."
npm install -g pnpm@9.10.0 >/dev/null 2>&1 || step_error "Failed to install pnpm"
step_complete "pnpm installed"

# Install pm2
status_update "Installing PM2 process manager..."
npm install -g pm2 >/dev/null 2>&1 || step_error "Failed to install PM2"
step_complete "PM2 installed"

# Setup PM2 startup script
status_update "Configuring PM2 startup..."
pm2 startup >/dev/null 2>&1 || true
step_complete "PM2 startup configured"

# Verify installations
status_update "Verifying installations..."
if node --version >/dev/null 2>&1 && npm --version >/dev/null 2>&1; then
    step_complete "Node.js installation verified"
    echo "  Node.js: $(node --version)"
    echo "  npm: $(npm --version)"
    echo "  pnpm: $(pnpm --version 2>/dev/null || echo 'installed')"
    echo "  PM2: $(pm2 --version)"
else
    step_error "Node.js installation verification failed"
fi

echo "🎉 Node.js installed successfully!"

# Execute the generated PM2 startup command
sudo env PATH=$PATH:/usr/bin /home/azureuser/.npm-global/lib/node_modules/pm2/bin/pm2 startup systemd -u azureuser --hp /home/azureuser