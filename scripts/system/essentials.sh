#!/bin/bash

export DEBIAN_FRONTEND=noninteractive

echo "Installing system essentials..."

# Update package lists and install essential development tools silently
sudo apt-get update -qq >/dev/null 2>&1
sudo apt-get install -y -qq build-essential python3 g++ make >/dev/null 2>&1

# Install system utility packages silently
sudo apt-get install -y -qq ncdu jq curl wget git >/dev/null 2>&1

# Install and configure Redis server silently
sudo apt-get install -y -qq redis-server >/dev/null 2>&1
sudo systemctl enable redis-server >/dev/null 2>&1
sudo systemctl start redis-server

# Configure system logs to prevent disk space issues
echo "SystemMaxUse=512M" | sudo tee -a /etc/systemd/journald.conf >/dev/null
sudo systemctl restart systemd-journald

# Disable core dumps for security
grep -q "* hard core 0" /etc/security/limits.conf || echo "* hard core 0" | sudo tee -a /etc/security/limits.conf >/dev/null

echo "✓ System essentials installed successfully"