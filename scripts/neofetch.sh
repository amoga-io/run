#!/usr/bin/env bash

# Define the package name
package="neofetch"

echo "insatlling neofetch"

# Detect the package manager
if command -v yay &> /dev/null; then
    yay -S "$package" --noconfirm
elif command -v apt-get &> /dev/null; then
    sudo apt-get update
    sudo apt-get install -y "$package"
elif command -v yum &> /dev/null; then
    sudo yum install -y "$package"
elif command -v dnf &> /dev/null; then
    sudo dnf install -y "$package"
elif command -v pacman &> /dev/null; then
    sudo pacman -S "$package" --noconfirm
elif command -v brew &> /dev/null; then
    brew install "$package"
else
    echo "Unsupported package manager. Please install neofetch manually."
    exit 1
fi

# Check if neofetch was installed successfully
if command -v "$package" &> /dev/null; then
    echo "Neofetch installed successfully."
else
    echo "Failed to install Neofetch."
    exit 1
fi

echo "Neofetch installation complete."
echo "You can run it by typing 'neofetch' in your terminal."
