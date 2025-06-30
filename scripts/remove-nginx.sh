#!/bin/bash

# Clean script to remove Nginx from Ubuntu

# Stop and disable Nginx service
sudo systemctl stop nginx
sudo systemctl disable nginx

# Remove Nginx packages
sudo apt-get purge nginx nginx-common nginx-full nginx-core -y
sudo apt-get autoremove -y

# Clean up configuration files
[ -d "/etc/nginx" ] && sudo rm -rf /etc/nginx
[ -d "/var/log/nginx" ] && sudo rm -rf /var/log/nginx
[ -d "/var/cache/nginx" ] && sudo rm -rf /var/cache/nginx