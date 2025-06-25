#!/bin/bash

# Function to check if Java is installed
check_java() {
    if java -version &>/dev/null; then
        echo "Java is already installed."
        java -version
        return 0
    else
        echo "Java is not installed."
        return 1
    fi
}

# Function to install Java
install_java() {
    echo "Available Java versions to install: 11, 17, 21"
    read -p "Enter the Java version you want to install: " java_version

    case "$java_version" in
        11)
            sudo apt install -y openjdk-11-jdk
            ;;
        17)
            sudo apt install -y openjdk-17-jdk
            ;;
        21)
            sudo apt install -y openjdk-21-jdk
            ;;
        *)
            echo "Invalid selection. Please choose 11, 17, or 21."
            exit 1
            ;;
    esac
}

# Function to set default Java version
set_java_version() {
    echo "Setting Java version..."
    sudo update-alternatives --config java
}

# Main script execution
echo "Checking if Java is installed..."
check_java || install_java

# Activate the selected Java version
set_java_version

echo "Java installation and setup complete!"
java -version
