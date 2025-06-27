#!/bin/bash

export DEBIAN_FRONTEND=noninteractive
set -euo pipefail

status_update() { echo "  $1"; }
step_complete() { echo "✓ $1"; }
step_error() { echo "❌ $1"; exit 1; }

echo "🚀 Installing PM2..."

# Install and configure pm2
status_update "Installing PM2 globally..."
sudo npm install -g pm2 >/dev/null 2>&1 || step_error "Failed to install PM2"
step_complete "PM2 installed"

status_update "Configuring PM2 permissions..."
sudo chmod 755 $(which pm2) >/dev/null 2>&1
sudo chmod -R 755 $(dirname $(which pm2))/../lib/node_modules/pm2 >/dev/null 2>&1
step_complete "PM2 permissions configured"

status_update "Creating PM2 log directory..."
sudo mkdir -p /var/log/pm2 >/dev/null 2>&1
sudo chmod 777 /var/log/pm2 >/dev/null 2>&1
step_complete "PM2 log directory created"

status_update "Configuring PM2 startup..."
pm2 startup systemd >/dev/null 2>&1 || true
pm2 save >/dev/null 2>&1 || true
step_complete "PM2 startup configured"

# Verify installation
status_update "Verifying PM2 installation..."
if pm2 --version >/dev/null 2>&1; then
    step_complete "PM2 installation verified"
    echo "  $(pm2 --version)"
else
    step_error "PM2 installation verification failed"
fi

echo "🎉 PM2 installed successfully!"
echo "📝 Run 'pm2 startup' command shown above to complete startup configuration"