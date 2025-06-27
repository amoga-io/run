#!/bin/bash

export DEBIAN_FRONTEND=noninteractive
set -euo pipefail

status_update() { echo "  $1"; }
step_complete() { echo "✓ $1"; }
step_error() { echo "❌ $1"; exit 1; }

echo "🌐 Installing Nginx..."

# Add Nginx official repository
status_update "Adding Nginx repository..."
echo "deb [arch=amd64] http://nginx.org/packages/mainline/ubuntu/ $(lsb_release -cs) nginx" | sudo tee /etc/apt/sources.list.d/nginx.list >/dev/null 2>&1

# Add Nginx signing key
curl -fsSL https://nginx.org/keys/nginx_signing.key | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/nginx.gpg >/dev/null 2>&1 || step_error "Failed to add Nginx signing key"
step_complete "Nginx repository added"

# Install nginx
status_update "Updating package lists..."
sudo apt-get update -qq >/dev/null 2>&1 || step_error "Failed to update package lists"
step_complete "Package lists updated"

status_update "Installing Nginx..."
sudo apt-get install -y -qq nginx >/dev/null 2>&1 || step_error "Failed to install Nginx"
step_complete "Nginx installed"

# Create required directories
status_update "Creating Nginx directories..."
sudo mkdir -p /var/run/nginx >/dev/null 2>&1
sudo mkdir -p /var/log/nginx >/dev/null 2>&1
step_complete "Nginx directories created"

# Set ownership
status_update "Setting directory permissions..."
sudo chown -R $USER:$USER /var/log/nginx >/dev/null 2>&1
sudo chown -R $USER:$USER /var/run/nginx >/dev/null 2>&1
sudo chmod 755 /var/run/nginx >/dev/null 2>&1
sudo chmod 755 /var/log/nginx >/dev/null 2>&1
step_complete "Directory permissions set"

# Allow nginx to bind to ports 80/443 without root
status_update "Configuring port binding capabilities..."
sudo setcap cap_net_bind_service=+ep /usr/sbin/nginx >/dev/null 2>&1 || step_error "Failed to set capabilities"
step_complete "Port binding capabilities configured"

# Backup original nginx.conf
status_update "Backing up original configuration..."
sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup >/dev/null 2>&1
step_complete "Original configuration backed up"

# Update nginx.conf - remove user directive and update pid path
status_update "Updating Nginx configuration..."
sudo sed -i "s/user .*;/user $USER;/" /etc/nginx/nginx.conf >/dev/null 2>&1
sudo sed -i '/http {/a \    client_max_body_size 10M;' /etc/nginx/nginx.conf >/dev/null 2>&1
step_complete "Nginx configuration updated"

# Create minimal site configuration
status_update "Creating test site configuration..."
echo "server { listen 80 default_server; listen [::]:80 default_server; server_name _; location / { return 200 'nginx is working!'; add_header Content-Type text/plain; } }" | sudo tee /etc/nginx/conf.d/test-site.conf >/dev/null 2>&1
step_complete "Test site configuration created"

# Set proper ownership for configuration
status_update "Setting configuration permissions..."
sudo chown $USER:$USER /etc/nginx/nginx.conf >/dev/null 2>&1
sudo chown -R $USER:$USER /etc/nginx/conf.d >/dev/null 2>&1
step_complete "Configuration permissions set"

# Test configuration
status_update "Testing Nginx configuration..."
nginx -t >/dev/null 2>&1 || step_error "Nginx configuration test failed"
step_complete "Nginx configuration test passed"

# Start nginx
status_update "Starting Nginx service..."
sudo systemctl start nginx >/dev/null 2>&1 || step_error "Failed to start Nginx"
sudo systemctl enable nginx >/dev/null 2>&1 || step_error "Failed to enable Nginx"
step_complete "Nginx service started and enabled"

# Verify installation
status_update "Verifying Nginx installation..."
if sudo systemctl is-active nginx >/dev/null 2>&1; then
    step_complete "Nginx installation verified"
    echo "  Service: Active"
    echo "  Configuration: Valid"
else
    step_error "Nginx installation verification failed"
fi

echo "🎉 Nginx installed successfully!"
echo "📝 Running as user: $USER"
echo "📝 Test the installation: curl http://localhost"