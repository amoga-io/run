#!/bin/bash

export DEBIAN_FRONTEND=noninteractive
set -euo pipefail

status_update() { echo "  $1"; }
step_complete() { echo "✓ $1"; }
step_error() { echo "❌ $1"; exit 1; }

echo "☕ Installing Java 17..."

# Update package list silently
status_update "Updating package lists..."
sudo apt-get update -qq >/dev/null 2>&1 || step_error "Failed to update package lists"
step_complete "Package lists updated"

# Install Java 17 (most stable for enterprise)
status_update "Installing OpenJDK 17..."
sudo apt-get install -y -qq openjdk-17-jdk >/dev/null 2>&1 || step_error "Failed to install OpenJDK 17"
step_complete "OpenJDK 17 installed"

# Set as default Java version
status_update "Configuring Java alternatives..."
if ! sudo update-alternatives --install /usr/bin/java java /usr/lib/jvm/java-17-openjdk-amd64/bin/java 1 >/dev/null 2>&1; then
    echo "Warning: Failed to set java alternative, but Java is installed"
fi
if ! sudo update-alternatives --install /usr/bin/javac javac /usr/lib/jvm/java-17-openjdk-amd64/bin/javac 1 >/dev/null 2>&1; then
    echo "Warning: Failed to set javac alternative, but Java is installed"
fi
step_complete "Java alternatives configured"

# Set JAVA_HOME environment variable
status_update "Setting JAVA_HOME environment variable..."
echo 'export JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64' | sudo tee -a /etc/environment >/dev/null 2>&1
step_complete "JAVA_HOME environment variable set"

# Verify installation
status_update "Verifying Java installation..."
if java -version >/dev/null 2>&1; then
    step_complete "Java installation verified"
    echo "  $(java -version 2>&1 | head -n 1)"
else
    step_error "Java installation verification failed"
fi

echo "🎉 Java 17 installed successfully!"
