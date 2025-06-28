#!/bin/bash

DEFAULT_VERSION=mainline
version="$DEFAULT_VERSION"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      shift
      version="$1"
      shift
      ;;
    *)
      shift
      ;;
  esac
done

# Set non-interactive mode
export DEBIAN_FRONTEND=noninteractive

# Function to retry commands
retry() {
  local n=0
  until [ $n -ge 5 ]; do
    "$@" && break
    n=$((n+1))
    echo "Command failed, retrying... ($n/5)"
    sleep 2
  done
}

# Add Nginx official repository
echo "Adding Nginx repository..."
echo "deb [arch=amd64] http://nginx.org/packages/$version/ubuntu/ $(lsb_release -cs) nginx" | sudo tee /etc/apt/sources.list.d/nginx.list

# Add Nginx signing key (force overwrite)
echo "Adding Nginx signing key..."
curl -fsSL https://nginx.org/keys/nginx_signing.key | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/nginx.gpg

# Install nginx
echo "Installing Nginx..."
retry sudo apt update -qq
retry sudo apt install -y -qq nginx \
  -o Dpkg::Options::="--force-confdef" \
  -o Dpkg::Options::="--force-confold"

# Create required directories
sudo mkdir -p /var/run/nginx
sudo mkdir -p /var/log/nginx

# Set ownership
sudo chown -R $USER:$USER /var/log/nginx
sudo chown -R $USER:$USER /var/run/nginx

# Set directory permissions
sudo chmod 755 /var/run/nginx
sudo chmod 755 /var/log/nginx

# Allow nginx to bind to ports 80/443 without root
sudo setcap cap_net_bind_service=+ep /usr/sbin/nginx

# Backup original nginx.conf
sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup

# Update nginx.conf - remove user directive and update pid path
sudo sed -i "s/user .*;/user $USER;/" /etc/nginx/nginx.conf
sudo sed -i '/http {/a \    client_max_body_size 10M;' /etc/nginx/nginx.conf

# Create minimal site configuration
echo "server { listen 80 default_server; listen [::]:80 default_server; server_name _; location / { return 200 'nginx is working!'; add_header Content-Type text/plain; } }" | sudo tee /etc/nginx/conf.d/test-site.conf

# Set proper ownership for configuration
sudo chown $USER:$USER /etc/nginx/nginx.conf
sudo chown -R $USER:$USER /etc/nginx/conf.d

# Test configuration
nginx -t

# Start nginx
sudo systemctl start nginx
sudo systemctl enable nginx

echo "Nginx installed and running as user $USER"
echo "Test the installation: curl http://localhost"