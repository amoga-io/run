#!/bin/bash
# Simple script to install latest PHP on Ubuntu

# Exit on error
set -e

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

# Update package lists
apt update

# Install prerequisites
apt install -y software-properties-common

# Add PHP repository
add-apt-repository -y ppa:ondrej/php
apt update

# Install PHP 8.3 (latest stable as of April 2025)
apt install -y php"$version" php"$version"-fpm php"$version"-common php"$version"-mysql php"$version"-curl php"$version"-gd \
  php"$version"-mbstring php"$version"-xml php"$version"-zip

# Enable and start PHP-FPM
systemctl enable php"$version"-fpm
systemctl start php"$version"-fpm

# Show installed PHP version
php -v

echo "PHP $version installed successfully"