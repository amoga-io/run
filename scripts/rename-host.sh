#!/bin/bash

# Simple script to rename Ubuntu host
# Usage: sudo ./rename_simple.sh <new-hostname>

# Check root
if [[ $EUID -ne 0 ]]; then
   echo "Error: Must be run as root." >&2
   exit 1
fi

# Check argument count
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <new-hostname>" >&2
    exit 1
fi

NEW_HOSTNAME="$1"
CURRENT_HOSTNAME=$(hostname)

# Only proceed if the name is actually changing
if [ "$CURRENT_HOSTNAME" != "$NEW_HOSTNAME" ]; then
    # Set hostname using the modern tool (handles /etc/hostname too)
    hostnamectl set-hostname "$NEW_HOSTNAME" || exit 1

    # Update the standard 127.0.1.1 entry in /etc/hosts
    # This assumes the common Debian/Ubuntu pattern.
    # If the pattern or old hostname isn't found, sed does nothing here.
    sed -i "s/^\(127\.0\.1\.1\s\+\)$CURRENT_HOSTNAME\$/\1$NEW_HOSTNAME/" /etc/hosts || exit 1
fi

exit 0