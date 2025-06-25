#!/bin/bash

# Install Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Create npm global directory in user's home
mkdir -p ~/.npm-global
npm config set prefix ~/.npm-global

# Add npm global bin to PATH in ~/.profile if not already present
if ! grep -q "PATH=~/.npm-global/bin:\$PATH" ~/.profile; then
    echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.profile
fi

# Source the updated profile
source ~/.profile

# Install pnpm 9.10.0 using npm
npm install -g pnpm@9.10.0

# Add pnpm to PATH
export PATH=~/.npm-global/bin:$PATH

# Install pm2
npm install -g pm2

# Setup PM2 startup script
pm2 startup

# Execute the generated PM2 startup command
sudo env PATH=$PATH:/usr/bin /home/azureuser/.npm-global/lib/node_modules/pm2/bin/pm2 startup systemd -u azureuser --hp /home/azureuser