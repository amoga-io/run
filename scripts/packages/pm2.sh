#!/bin/bash

# Detect the user to set up PM2
if [ "$SUDO_USER" ]; then
  TARGET_USER="$SUDO_USER"
else
  TARGET_USER="$(whoami)"
fi
TARGET_HOME="$(eval echo "~$TARGET_USER")"

DEFAULT_VERSION="latest"
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

export PATH="$TARGET_HOME/.npm-global/bin:$PATH"

if [ "$version" = "latest" ]; then
  sudo -u "$TARGET_USER" npm install -g pm2
else
  sudo -u "$TARGET_USER" npm install -g pm2@"$version"
fi

# Check if PM2 was installed successfully
PM2_PATH="$TARGET_HOME/.npm-global/bin/pm2"
if [ -f "$PM2_PATH" ]; then
    echo "PM2 installed successfully at $PM2_PATH"
else
    echo "PM2 installation failed - binary not found at expected location"
    exit 1
fi

# Set up PM2 with proper environment
sudo -u "$TARGET_USER" bash -c "export PATH=\"$TARGET_HOME/.npm-global/bin:\$PATH\" && pm2 save"
sudo chmod 755 "$PM2_PATH"
sudo chmod -R 755 "$(dirname "$PM2_PATH")/../lib/node_modules/pm2"
sudo mkdir -p /var/log/pm2
sudo chmod 777 /var/log/pm2
sudo -u "$TARGET_USER" bash -c "export PATH=\"$TARGET_HOME/.npm-global/bin:\$PATH\" && pm2 startup systemd"