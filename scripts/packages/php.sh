#!/bin/bash
# Simple script to install latest PHP on Ubuntu

# Exit on error
set -e

# Update package lists
apt update

# Install prerequisites
apt install -y software-properties-common

# Add PHP repository
add-apt-repository -y ppa:ondrej/php
apt update

# Install PHP 8.3 (latest stable as of April 2025)
apt install -y php8.3 php8.3-fpm php8.3-common php8.3-mysql php8.3-curl php8.3-gd \
  php8.3-mbstring php8.3-xml php8.3-zip

# Enable and start PHP-FPM
systemctl enable php8.3-fpm
systemctl start php8.3-fpm

# Show installed PHP version
php -v

echo "PHP 8.3 installed successfully"