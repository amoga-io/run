#!/bin/bash

# Make all operations non-interactive and silent
export DEBIAN_FRONTEND=noninteractive
set -euo pipefail

# Functions for status updates
status_update() { echo "  $1"; }
step_complete() { echo "✓ $1"; }
step_error() { echo "❌ $1"; exit 1; }

echo "🐳 Installing Docker..."

# Install dependencies
status_update "Updating package lists..."
sudo apt-get update -qq >/dev/null 2>&1 || step_error "Failed to update package lists"
step_complete "Package lists updated"

status_update "Installing prerequisites..."
sudo apt-get install -y -qq ca-certificates curl gnupg >/dev/null 2>&1 || step_error "Failed to install prerequisites"
step_complete "Prerequisites installed"

# Add Docker's GPG key
status_update "Adding Docker GPG key..."
sudo install -m 0755 -d /etc/apt/keyrings >/dev/null 2>&1
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg 2>/dev/null || step_error "Failed to add GPG key"
sudo chmod a+r /etc/apt/keyrings/docker.gpg >/dev/null 2>&1
step_complete "Docker GPG key added"

# Add Docker repository
status_update "Adding Docker repository..."
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list >/dev/null 2>&1 || step_error "Failed to add repository"
step_complete "Docker repository added"

# Install Docker packages
status_update "Updating package lists with Docker repository..."
sudo apt-get update -qq >/dev/null 2>&1 || step_error "Failed to update with Docker repository"
step_complete "Package lists updated"

status_update "Installing Docker Engine (this may take 2-3 minutes)..."
sudo apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin >/dev/null 2>&1 || step_error "Failed to install Docker packages"
step_complete "Docker Engine installed"

# Setup user permissions
status_update "Setting up user permissions..."
sudo groupadd -f docker >/dev/null 2>&1
sudo usermod -aG docker $USER >/dev/null 2>&1
step_complete "User permissions configured"

# Ensure docker.sock has correct permissions
status_update "Configuring Docker socket permissions..."
sudo chmod 666 /var/run/docker.sock >/dev/null 2>&1 || true
step_complete "Docker socket permissions configured"

# Create default Docker config directory
status_update "Creating Docker configuration directory..."
sudo mkdir -p /etc/docker >/dev/null 2>&1
sudo chown -R $USER:docker /etc/docker >/dev/null 2>&1 || true
step_complete "Docker configuration directory created"

# Start and enable Docker service
status_update "Starting Docker service..."
sudo systemctl enable docker >/dev/null 2>&1 || step_error "Failed to enable Docker service"
sudo systemctl start docker >/dev/null 2>&1 || step_error "Failed to start Docker service"
step_complete "Docker service started"

# Print versions and verification message
status_update "Verifying installation..."
if docker --version >/dev/null 2>&1; then
    step_complete "Docker installation verified"
    echo "  $(docker --version)"
    echo "  $(docker compose version)"
else
    step_error "Docker installation verification failed"
fi

echo "🎉 Docker installed successfully!"
echo "📝 Note: You may need to log out and back in for group changes to take effect"