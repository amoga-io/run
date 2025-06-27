#!/bin/bash

# Script to install Python, pip, gunicorn and venv on azureuser
# Exit immediately if a command exits with a non-zero status
set -e

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
  echo "This script must be run as root or with sudo"
  exit 1
fi

# Update package lists
apt-get update

# Install Python and development tools
apt-get install -y python3 python3-pip python3-dev python3-venv python3-full

# Install gunicorn via apt instead of pip
apt-get install -y gunicorn

# Create symbolic link for python command
if ! command -v python &> /dev/null; then
  ln -sf /usr/bin/python3 /usr/local/bin/python
fi

# Create log directories
mkdir -p /var/log/django
mkdir -p /var/log/celery

# Set permissions (adjust user/group as needed)
chown -R azureuser:azureuser /var/log/django
chown -R azureuser:azureuser /var/log/celery
chmod 755 /var/log/django
chmod 755 /var/log/celery

# Print status
echo "Log directories created and permissions set"
echo "Django logs: /var/log/django"
echo "Celery logs: /var/log/celery"

# Print installed Python version
python --version

echo "Installation and setup completed successfully"