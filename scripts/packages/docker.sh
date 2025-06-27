#!/bin/bash

# Install dependencies
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg

# Add Docker's GPG key
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Add Docker repository
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker packages
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Setup user permissions
sudo groupadd -f docker
sudo usermod -aG docker $USER

# Ensure docker.sock has correct permissions
sudo chmod 666 /var/run/docker.sock

# Create default Docker config directory
sudo mkdir -p /etc/docker
sudo chown -R $USER:docker /etc/docker

# Start and enable Docker service
sudo systemctl enable docker
sudo systemctl start docker

# Print versions and verification message
docker --version
docker compose version