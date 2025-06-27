#!/bin/bash

export DEBIAN_FRONTEND=noninteractive
set -euo pipefail

status_update() { echo "  $1"; }
step_complete() { echo "✓ $1"; }
step_error() { echo "❌ $1"; exit 1; }

echo "🐘 Installing PHP..."

# Update package lists
status_update "Updating package lists..."
sudo apt-get update -qq >/dev/null 2>&1 || step_error "Failed to update package lists"
step_complete "Package lists updated"

# Install prerequisites
status_update "Installing prerequisites..."
sudo apt-get install -y -qq software-properties-common >/dev/null 2>&1 || step_error "Failed to install prerequisites"
step_complete "Prerequisites installed"

# Add PHP repository
status_update "Adding PHP repository..."
sudo add-apt-repository -y ppa:ondrej/php >/dev/null 2>&1 || step_error "Failed to add PHP repository"
sudo apt-get update -qq >/dev/null 2>&1 || step_error "Failed to update after adding repository"
step_complete "PHP repository added"

# Install PHP 8.3
status_update "Installing PHP 8.3 and extensions..."
sudo apt-get install -y -qq php8.3 php8.3-fpm php8.3-common php8.3-mysql php8.3-curl php8.3-gd php8.3-mbstring php8.3-xml php8.3-zip >/dev/null 2>&1 || step_error "Failed to install PHP packages"
step_complete "PHP 8.3 installed"

# Enable and start PHP-FPM
status_update "Starting PHP-FPM service..."
sudo systemctl enable php8.3-fpm >/dev/null 2>&1 || step_error "Failed to enable PHP-FPM"
sudo systemctl start php8.3-fpm >/dev/null 2>&1 || step_error "Failed to start PHP-FPM"
step_complete "PHP-FPM service started"

# Verify installation
status_update "Verifying PHP installation..."
if php -v >/dev/null 2>&1; then
    step_complete "PHP installation verified"
    echo "  $(php -v | head -n 1)"
else
    step_error "PHP installation verification failed"
fi

echo "🎉 PHP 8.3 installed successfully!"