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

# Set non-interactive mode
export DEBIAN_FRONTEND=noninteractive

# Function to retry commands
retry() {
  local n=0
  until [ $n -ge 5 ]; do
    "$@" && break
    n=$((n+1))
    echo "Command failed, retrying... ($n/5)"
    sleep 2
  done
}

# Install the selected Java version
echo "Installing Java $java_version..."
case "$java_version" in
  11)
    retry sudo apt install -y -qq openjdk-11-jdk
    ;;
  17)
    retry sudo apt install -y -qq openjdk-17-jdk
    ;;
  21)
    retry sudo apt install -y -qq openjdk-21-jdk
    ;;
  *)
    echo "Invalid selection. Please choose 11, 17, or 21."
    exit 1
    ;;
esac

# Set the default Java version automatically
echo "Configuring Java alternatives..."
JAVA_PATH=$(sudo update-alternatives --list java | grep "$java_version" | head -n1)
if [ -n "$JAVA_PATH" ]; then
  echo "$JAVA_PATH" | sudo update-alternatives --set java /dev/stdin
fi

# Also set javac
JAVAC_PATH=$(sudo update-alternatives --list javac | grep "$java_version" | head -n1)
if [ -n "$JAVAC_PATH" ]; then
  echo "$JAVAC_PATH" | sudo update-alternatives --set javac /dev/stdin
fi

echo "Java $java_version installation and setup complete!"
java -version
