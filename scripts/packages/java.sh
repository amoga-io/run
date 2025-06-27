#!/bin/bash

export DEBIAN_FRONTEND=noninteractive

echo "Installing Java 17..."

# Update package list silently
sudo apt-get update -qq >/dev/null 2>&1

# Install Java 17 (most stable for enterprise)
sudo apt-get install -y -qq openjdk-17-jdk >/dev/null 2>&1

# Set as default Java version
if ! sudo update-alternatives --install /usr/bin/java java /usr/lib/jvm/java-17-openjdk-amd64/bin/java 1 >/dev/null 2>&1; then
    echo "Warning: Failed to set java alternative, but Java is installed"
fi
if ! sudo update-alternatives --install /usr/bin/javac javac /usr/lib/jvm/java-17-openjdk-amd64/bin/javac 1 >/dev/null 2>&1; then
    echo "Warning: Failed to set javac alternative, but Java is installed"
fi

# Set JAVA_HOME environment variable
echo 'export JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64' | sudo tee -a /etc/environment >/dev/null

echo "✓ Java 17 installed successfully"
java -version
java -version
