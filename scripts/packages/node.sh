#!/bin/bash

DEFAULT_VERSION=20
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

# Install Node.js with the specified or default version
curl -fsSL https://deb.nodesource.com/setup_"$version".x | sudo -E bash -
sudo apt-get install -y nodejs

# Detect the user to set up PM2 and npm global directory
if [ "$SUDO_USER" ]; then
  TARGET_USER="$SUDO_USER"
else
  TARGET_USER="$(whoami)"
fi
TARGET_HOME="$(eval echo "~$TARGET_USER")"

# Create npm global directory in user's home
mkdir -p "$TARGET_HOME/.npm-global"
sudo -u "$TARGET_USER" npm config set prefix "$TARGET_HOME/.npm-global"

# Add npm global bin to PATH in ~/.profile if not already present
if ! grep -q "PATH=$TARGET_HOME/.npm-global/bin:\$PATH" "$TARGET_HOME/.profile"; then
    echo 'export PATH=$TARGET_HOME/.npm-global/bin:$PATH' >> "$TARGET_HOME/.profile"
fi

# Source the updated profile
source "$TARGET_HOME/.profile"

# Install pnpm 9.10.0 using npm
sudo -u "$TARGET_USER" npm install -g pnpm@9.10.0

# Add pnpm to PATH
export PATH="$TARGET_HOME/.npm-global/bin:$PATH"

# Install pm2
sudo -u "$TARGET_USER" npm install -g pm2

# Setup PM2 startup script
sudo -u "$TARGET_USER" pm2 startup

# Execute the generated PM2 startup command
sudo env PATH=$PATH:/usr/bin "$TARGET_HOME/.npm-global/lib/node_modules/pm2/bin/pm2" startup systemd -u "$TARGET_USER" --hp "$TARGET_HOME"