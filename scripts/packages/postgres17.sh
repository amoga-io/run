#!/bin/bash

export DEBIAN_FRONTEND=noninteractive
set -euo pipefail

status_update() { echo "  $1"; }
step_complete() { echo "✓ $1"; }
step_error() { echo "❌ $1"; exit 1; }

echo "🐘 Installing PostgreSQL 17..."

# Generate random 20 character password
POSTGRES_PASSWORD=$(openssl rand -base64 20 | tr -dc 'a-zA-Z0-9' | head -c 20)

# Add PostgreSQL repository and key
status_update "Adding PostgreSQL repository..."
curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo gpg --dearmor -o /usr/share/keyrings/postgresql-keyring.gpg >/dev/null 2>&1 || step_error "Failed to add GPG key"
echo "deb [signed-by=/usr/share/keyrings/postgresql-keyring.gpg] https://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" | sudo tee /etc/apt/sources.list.d/pgdg.list >/dev/null 2>&1
step_complete "PostgreSQL repository added"

# Update package lists
status_update "Updating package lists..."
sudo apt-get update -qq >/dev/null 2>&1 || step_error "Failed to update package lists"
step_complete "Package lists updated"

# Install PostgreSQL 17
status_update "Installing PostgreSQL 17..."
sudo apt-get install -y -qq postgresql-17 >/dev/null 2>&1 || step_error "Failed to install PostgreSQL"
step_complete "PostgreSQL 17 installed"

# Configure PostgreSQL to listen on all interfaces
status_update "Configuring PostgreSQL..."
sudo cp /etc/postgresql/17/main/postgresql.conf /etc/postgresql/17/main/postgresql.conf.backup >/dev/null 2>&1
sudo sh -c "echo \"listen_addresses = '*'\" >> /etc/postgresql/17/main/postgresql.conf" >/dev/null 2>&1
step_complete "PostgreSQL configuration updated"

# Configure PostgreSQL to allow remote connections
status_update "Configuring remote connections..."
sudo cp /etc/postgresql/17/main/pg_hba.conf /etc/postgresql/17/main/pg_hba.conf.backup >/dev/null 2>&1
sudo sh -c "echo \"host    all             all             0.0.0.0/0               scram-sha-256\" >> /etc/postgresql/17/main/pg_hba.conf" >/dev/null 2>&1
step_complete "Remote connections configured"

# Set postgres user password
status_update "Setting postgres user password..."
sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD '$POSTGRES_PASSWORD';" >/dev/null 2>&1 || step_error "Failed to set password"
step_complete "postgres user password set"

# Restart PostgreSQL to apply changes
status_update "Restarting PostgreSQL service..."
sudo systemctl restart postgresql@17-main >/dev/null 2>&1 || step_error "Failed to restart PostgreSQL"
step_complete "PostgreSQL service restarted"

# Verify installation
status_update "Verifying PostgreSQL installation..."
if sudo systemctl is-active postgresql@17-main >/dev/null 2>&1; then
    step_complete "PostgreSQL installation verified"
    echo "  Service: Active"
    echo "  Password: $POSTGRES_PASSWORD"
else
    step_error "PostgreSQL installation verification failed"
fi

echo "🎉 PostgreSQL 17 installed successfully!"
echo "📝 Generated postgres user password: $POSTGRES_PASSWORD"
echo "📝 Save this password securely!"