#!/bin/bash

# Default Java version
DEFAULT_VERSION=17
java_version=""

# Parse arguments for --version flag
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      shift
      java_version="$1"
      shift
      ;;
    *)
      shift
      ;;
  esac
done

# If no version specified, use default
if [ -z "$java_version" ]; then
  java_version="$DEFAULT_VERSION"
fi

# Install the selected Java version
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

# Set the default Java version automatically
JAVA_PATH=$(update-alternatives --list java | grep "$java_version" | head -n1)
if [ -n "$JAVA_PATH" ]; then
  sudo update-alternatives --set java "$JAVA_PATH"
fi

echo "Java $java_version installation and setup complete!"
java -version
