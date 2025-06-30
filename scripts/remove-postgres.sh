#!/bin/bash

# Stop PostgreSQL service
echo "Stopping PostgreSQL service..."
sudo systemctl stop postgresql

# Remove PostgreSQL and its dependencies
echo "Removing PostgreSQL completely..."
sudo apt-get --purge remove postgresql\* -y
sudo rm -rf /etc/postgresql/
sudo rm -rf /var/lib/postgresql/
sudo rm -rf /var/log/postgresql/
sudo userdel -r postgres
sudo groupdel postgres

# Clean up any remaining packages
echo "Cleaning up remaining packages..."
sudo apt-get autoremove -y
sudo apt-get autoclean
