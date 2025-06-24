#!/usr/bin/env bash

package="cmatrix"

# Check if the package is installed
if ! command -v cmatrix &> /dev/null; then
    echo "Installing $package..."
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
        echo "Unsupported package manager. Please install "$package" manually."
        exit 1
    fi
else
    echo "$package is already installed."
fi

# Check if cmatrix was installed successfully
if command -v cmatrix &> /dev/null; then
    echo "$package installed successfully."
else
    echo "Failed to install $package."
    exit 1
fi

echo "$package installation complete."
echo "You can run it by typing 'cmatrix' in your terminal."
