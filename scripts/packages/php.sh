#!/bin/bash
# Simple script to install latest PHP on Ubuntu

# Exit on error
set -euo pipefail

# Set non-interactive mode
export DEBIAN_FRONTEND=noninteractive

DEFAULT_VERSION=8.3
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

# Update package lists
echo "Updating package lists..."
retry sudo apt update -qq

# Install prerequisites
echo "Installing prerequisites..."
retry sudo apt install -y -qq software-properties-common

# Add PHP repository (non-interactive)
echo "Adding PHP repository..."
echo 'deb https://ppa.launchpadcontent.net/ondrej/php/ubuntu noble main' | sudo tee /etc/apt/sources.list.d/ondrej-php.list
retry sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 4F4EA0AAE5267A6C

retry sudo apt update -qq

# Install PHP with specified version
echo "Installing PHP $version..."
retry sudo apt install -y -qq \
  php"$version" \
  php"$version"-fpm \
  php"$version"-common \
  php"$version"-mysql \
  php"$version"-curl \
  php"$version"-gd \
  php"$version"-mbstring \
  php"$version"-xml \
  php"$version"-zip \
  -o Dpkg::Options::="--force-confdef" \
  -o Dpkg::Options::="--force-confold"

# Create www-data user if it doesn't exist
if ! id "www-data" &>/dev/null; then
    echo "Creating www-data user..."
    sudo useradd -r -s /bin/false www-data
fi

# Enable and start PHP-FPM
echo "Starting PHP-FPM service..."
sudo systemctl enable php"$version"-fpm
sudo systemctl start php"$version"-fpm

# Show installed PHP version
php -v

echo "PHP $version installed successfully"