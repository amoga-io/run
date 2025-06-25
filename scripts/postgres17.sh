#!/bin/bash

# Generate random 20 character password
POSTGRES_PASSWORD=$(openssl rand -base64 20 | tr -dc 'a-zA-Z0-9' | head -c 20)

# Add PostgreSQL repository and key
echo "Adding PostgreSQL repository and key..."
sudo sh -c 'echo "deb https://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo gpg --dearmor -o /usr/share/keyrings/postgresql-keyring.gpg
sudo sh -c 'echo "deb [signed-by=/usr/share/keyrings/postgresql-keyring.gpg] https://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'

# Update package lists
echo "Updating package lists..."
sudo apt update

# Install PostgreSQL 17
echo "Installing PostgreSQL 17..."
sudo apt install -y postgresql-17

# Check PostgreSQL service status
echo "Checking PostgreSQL service status..."
sudo systemctl status postgresql@17-main

# Configure PostgreSQL to listen on all interfaces
echo "Configuring PostgreSQL to listen on all interfaces..."
sudo cp /etc/postgresql/17/main/postgresql.conf /etc/postgresql/17/main/postgresql.conf.backup
sudo sh -c "echo \"listen_addresses = '*'\" >> /etc/postgresql/17/main/postgresql.conf"

# Configure PostgreSQL to allow remote connections
echo "Configuring PostgreSQL to allow remote connections..."
sudo cp /etc/postgresql/17/main/pg_hba.conf /etc/postgresql/17/main/pg_hba.conf.backup
sudo sh -c "echo \"host    all             all             0.0.0.0/0               scram-sha-256\" >> /etc/postgresql/17/main/pg_hba.conf"

# Set postgres user password
echo "Setting postgres user password..."
sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD '$POSTGRES_PASSWORD';"

# Restart PostgreSQL to apply changes
echo "Restarting PostgreSQL service..."
sudo systemctl restart postgresql@17-main

echo "PostgreSQL 17 installation complete."
echo "Generated postgres user password: $POSTGRES_PASSWORD"