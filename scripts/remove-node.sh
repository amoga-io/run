#!/bin/bash

# Stop any running Node processes
echo "Stopping any running Node processes..."
killall -9 node 2>/dev/null || echo "No Node processes were running"

# Remove Node.js using apt
echo "Removing Node.js packages..."
sudo apt-get purge --auto-remove nodejs npm -y

# Remove Node.js installed via NVM if it exists
echo "Checking for NVM installations..."
if [ -d "$HOME/.nvm" ]; then
  echo "Removing NVM and all Node versions installed with it..."
  rm -rf $HOME/.nvm
  # Remove NVM references from profile files
  sed -i '/NVM_DIR/d' ~/.profile ~/.bashrc ~/.zshrc 2>/dev/null
fi

# Remove Node.js installed via N if it exists
echo "Checking for N installations..."
if command -v n &> /dev/null; then
  echo "Removing N and all Node versions installed with it..."
  n uninstall
  sudo npm uninstall -g n
fi

# Remove pnpm
echo "Removing pnpm..."
sudo npm uninstall -g pnpm 2>/dev/null
rm -rf ~/.pnpm-store 2>/dev/null
rm -rf ~/.pnpm 2>/dev/null

# Clean up any remaining Node related directories
echo "Cleaning up remaining Node directories..."
sudo rm -rf /usr/local/lib/node_modules
sudo rm -rf /usr/local/bin/node
sudo rm -rf /usr/local/bin/npm
sudo rm -rf /usr/local/bin/pnpm
sudo rm -rf /usr/local/include/node
sudo rm -rf /opt/local/bin/node
sudo rm -rf /opt/local/include/node
sudo rm -rf /usr/local/bin/node-debug
sudo rm -rf /usr/local/bin/node-gyp

# Remove from PATH
echo "Removing Node related paths from PATH..."
if [ -f ~/.bashrc ]; then
  sed -i '/node/d' ~/.bashrc
  sed -i '/npm/d' ~/.bashrc
  sed -i '/pnpm/d' ~/.bashrc
fi

# Remove any config files
echo "Removing configuration files..."
rm -rf ~/.npm 2>/dev/null
rm -rf ~/.node-gyp 2>/dev/null
rm -rf ~/.node_repl_history 2>/dev/null
rm -rf ~/.npmrc 2>/dev/null

# Update PATH immediately
echo "Updating PATH..."
if [ -f ~/.bashrc ]; then
  source ~/.bashrc 2>/dev/null
fi

echo "Node.js and pnpm have been completely removed from your system."
echo "Please log out and log back in to ensure all changes take effect."