#!/bin/bash

# Update package lists and install essential development tools
# build-essential: provides compiler and libraries needed for building packages
# python3: Python programming language interpreter
# g++: GNU C++ compiler
# make: utility to maintain groups of programs
sudo apt-get update
sudo apt-get install -y build-essential python3 g++ make

# Configure system logs to prevent disk space issues
# This limits the maximum size of the systemd journal logs to 512MB
# Prevents logs from consuming too much disk space
echo "SystemMaxUse=512M" >> /etc/systemd/journald.conf
sudo systemctl restart systemd-journald

# Install and configure Redis server
# Redis is an in-memory data structure store used as database, cache, and message broker
sudo apt-get install -y redis-server
sudo systemctl enable redis-server  # Configure Redis to start on boot
sudo systemctl start redis-server   # Start the Redis service

# Install system utility packages
# ncdu: NCurses Disk Usage - interactive disk usage analyzer
# jq: lightweight command-line JSON processor
# curl: tool for transferring data with URLs
# wget: non-interactive network downloader
# git: distributed version control system
sudo apt install ncdu jq curl wget git -y

# Disable core dumps for security
# This prevents the system from creating core dump files when programs crash
# Core dumps can contain sensitive information and consume disk space
grep -q "* hard core 0" /etc/security/limits.conf || echo "* hard core 0" | sudo tee -a /etc/security/limits.conf