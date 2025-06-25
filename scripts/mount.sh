#!/bin/bash

# =============================================================================
# Disk Mounting Script for Docker
# =============================================================================

# Configuration Variables
# Examples:
# DISK_NAME="/dev/nvme0n1"     # For NVMe drives
# DISK_NAME="/dev/sdb"         # For SATA drives  
# DISK_NAME="/dev/sdc"         # For additional SATA drives
DISK_NAME="/dev/nvme0n1"

# Examples:
# MOUNT_POINT="/var/lib/docker"        # For Docker data
# MOUNT_POINT="/var/www"               # For web server files
# MOUNT_POINT="/home/data"             # For user data
MOUNT_POINT="/var/lib/docker"

# =============================================================================
# Script Functions
# =============================================================================

check_prerequisites() {
    echo "=== Checking Prerequisites ==="
    
    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        echo "Error: This script must be run as root (use sudo)"
        exit 1
    fi
    
    # Check if disk exists
    if [[ ! -b "$DISK_NAME" ]]; then
        echo "Error: Disk $DISK_NAME not found"
        echo "Available disks:"
        lsblk -o NAME,SIZE,TYPE | grep disk
        exit 1
    fi
    
    echo "✓ Prerequisites checked"
}

verify_disk() {
    echo "=== Verifying Disk Detection ==="
    lsblk -o NAME,FSTYPE,SIZE,MOUNTPOINT,LABEL
    echo ""
    read -p "Is $DISK_NAME the correct disk to use? (y/N): " confirm
    if [[ ! $confirm =~ ^[Yy]$ ]]; then
        echo "Operation cancelled"
        exit 1
    fi
}

stop_services() {
    echo "=== Stopping Services ==="
    
    # Stop Docker if mount point is Docker-related
    if [[ "$MOUNT_POINT" == *"docker"* ]]; then
        echo "Stopping Docker services..."
        systemctl stop docker
        systemctl stop docker.socket
        echo "✓ Docker services stopped"
    fi
}

backup_existing_data() {
    echo "=== Backing Up Existing Data ==="
    
    if [[ -d "$MOUNT_POINT" ]]; then
        backup_dir="${MOUNT_POINT}.old"
        echo "Moving existing $MOUNT_POINT to $backup_dir"
        mv "$MOUNT_POINT" "$backup_dir"
        echo "✓ Data backed up to $backup_dir"
    else
        echo "✓ No existing data to backup"
    fi
}

create_partition() {
    echo "=== Creating Partition ==="
    
    # Determine partition name based on disk type
    if [[ "$DISK_NAME" == *"nvme"* ]]; then
        PARTITION_NAME="${DISK_NAME}p1"
    else
        PARTITION_NAME="${DISK_NAME}1"
    fi
    
    echo "Creating GPT partition table on $DISK_NAME"
    parted -s "$DISK_NAME" mklabel gpt
    
    echo "Creating primary partition using entire disk"
    parted -s "$DISK_NAME" mkpart primary ext4 0% 100%
    
    # Wait for partition to be recognized
    sleep 2
    partprobe "$DISK_NAME"
    
    echo "✓ Partition created: $PARTITION_NAME"
}

format_partition() {
    echo "=== Formatting Partition ==="
    
    echo "Formatting $PARTITION_NAME with ext4 filesystem"
    mkfs.ext4 "$PARTITION_NAME"
    
    echo "✓ Partition formatted"
}

backup_fstab() {
    echo "=== Backing Up fstab ==="
    
    backup_file="/etc/fstab.$(date +%Y-%m-%d-%H%M%S)"
    cp /etc/fstab "$backup_file"
    
    echo "✓ fstab backed up to $backup_file"
}

update_fstab() {
    echo "=== Updating fstab ==="
    
    # Get UUID of the partition
    UUID=$(blkid -s UUID -o value "$PARTITION_NAME")
    
    if [[ -z "$UUID" ]]; then
        echo "Error: Could not get UUID for $PARTITION_NAME"
        exit 1
    fi
    
    echo "UUID: $UUID"
    
    # Create mount point
    mkdir -p "$MOUNT_POINT"
    
    # Add to fstab
    echo "UUID=$UUID $MOUNT_POINT ext4 defaults 0 2" >> /etc/fstab
    
    echo "✓ fstab updated"
}

verify_mount() {
    echo "=== Verifying Mount Configuration ==="
    
    # Reload systemd
    systemctl daemon-reload
    
    # Test mount
    echo "Testing mount configuration..."
    if mount -a; then
        echo "✓ Mount test successful"
    else
        echo "Error: Mount test failed"
        echo "Check fstab syntax and try again"
        exit 1
    fi
    
    # Verify mount point
    echo "Mount point status:"
    df -h "$MOUNT_POINT"
    
    # Set permissions
    chown root:root "$MOUNT_POINT"
    
    echo "✓ Mount verified and permissions set"
}

start_services() {
    echo "=== Starting Services ==="
    
    # Start Docker if mount point is Docker-related
    if [[ "$MOUNT_POINT" == *"docker"* ]]; then
        echo "Starting Docker services..."
        systemctl start docker
        
        # Verify Docker
        if systemctl is-active --quiet docker; then
            echo "✓ Docker started successfully"
            echo "Docker Root Directory:"
            docker info 2>/dev/null | grep "Docker Root Dir" || echo "Docker info not available yet"
        else
            echo "Warning: Docker failed to start"
        fi
    fi
}

show_summary() {
    echo ""
    echo "=== Summary ==="
    echo "Disk: $DISK_NAME"
    echo "Partition: $PARTITION_NAME"
    echo "Mount Point: $MOUNT_POINT"
    echo "Filesystem: ext4"
    echo ""
    echo "Current disk usage:"
    df -h "$MOUNT_POINT"
    echo ""
    echo "✓ Disk mounting completed successfully!"
}

# =============================================================================
# Main Script Execution
# =============================================================================

main() {
    echo "Starting disk mounting process..."
    echo "Disk: $DISK_NAME"
    echo "Mount Point: $MOUNT_POINT"
    echo ""
    
    check_prerequisites
    verify_disk
    stop_services
    backup_existing_data
    create_partition
    format_partition
    backup_fstab
    update_fstab
    verify_mount
    start_services
    show_summary
}

# Run main function
main "$@"