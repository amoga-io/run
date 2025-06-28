#!/bin/bash

DEFAULT_VERSION=3.10
version="$DEFAULT_VERSION"

# Parse --version flag
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

# Script to install Python, pip, gunicorn and venv
# Exit immediately if a command exits with a non-zero status
set -euo pipefail

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

# Update package lists with retry
echo "Updating package lists..."
retry sudo apt-get update -qq

# Install Python and development tools
echo "Installing Python $version and development tools..."
retry sudo apt-get install -y -qq \
  python"$version" \
  python"$version"-pip \
  python"$version"-dev \
  python"$version"-venv \
  python"$version"-full \
  -o Dpkg::Options::="--force-confdef" \
  -o Dpkg::Options::="--force-confold"

# Install gunicorn via apt instead of pip
echo "Installing gunicorn..."
retry sudo apt-get install -y -qq gunicorn

# Create symbolic link for python command
if ! command -v python &> /dev/null; then
  sudo ln -sf /usr/bin/python3 /usr/local/bin/python
fi

# Create log directories
sudo mkdir -p /var/log/django
sudo mkdir -p /var/log/celery

# Get the current user (who invoked sudo) instead of hardcoded azureuser
CURRENT_USER=${SUDO_USER:-$(who am i | awk '{print $1}')}
if [ -z "$CURRENT_USER" ]; then
  CURRENT_USER=$(logname 2>/dev/null || echo "root")
fi

# Set permissions using current user
sudo chown -R "$CURRENT_USER:$CURRENT_USER" /var/log/django
sudo chown -R "$CURRENT_USER:$CURRENT_USER" /var/log/celery
sudo chmod 755 /var/log/django
sudo chmod 755 /var/log/celery

# Print status
echo "Log directories created and permissions set for user: $CURRENT_USER"
echo "Django logs: /var/log/django"
echo "Celery logs: /var/log/celery"

# Print installed Python version
python --version

echo "Installation and setup completed successfully"