# Run CLI - Ubuntu Server Package Manager

A **safe and intelligent** CLI tool for managing developer and system packages

## 🚀 Quick Installation

```bash
# One-line installation (recommended)
curl -fsSL https://raw.githubusercontent.com/amoga-io/run/main/scripts/install.sh | bash
```

**Alternative installation methods:**

```bash
# Clone and build from source
git clone https://github.com/amoga-io/run.git
cd run
go build -o run .
sudo cp run /usr/local/bin/

# Or install globally
sudo cp run /usr/local/bin/
```

## 🧹 Uninstall

```bash
# Remove the CLI binary
sudo rm -f /usr/local/bin/run

# Remove configuration and cache
rm -rf ~/.run
```

## 📁 Project Structure

```
run/
├── cmd/                          # CLI commands
│   ├── install.go               # Install command implementation
│   ├── remove.go                # Remove command implementation
│   ├── list.go                  # List command implementation
│   ├── update.go                # Update command implementation
│   └── root.go                  # Root CLI setup
├── internal/                     # Internal packages
│   ├── registry.go              # Package registry and definitions
│   ├── scriptPath.go            # Script path resolution
│   └── utils.go                 # Utility functions
├── scripts/                     # Installation scripts
│   ├── docker.sh                # Docker installation
│   ├── essentials.sh            # Essential tools installation
│   ├── install.sh               # CLI installation script
│   ├── java.sh                  # Java installation
│   ├── nginx.sh                 # Nginx installation
│   ├── node.sh                  # Node.js installation
│   ├── php.sh                   # PHP installation
│   ├── pm2.sh                   # PM2 installation
│   ├── postgres17.sh            # PostgreSQL 17 installation
│   ├── python.sh                # Python installation
│   ├── remove-nginx.sh          # Nginx removal
│   ├── remove-node.sh           # Node.js removal
│   └── remove-postgres.sh       # PostgreSQL removal
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── main.go                      # Application entry point
├── README.md                    # Project documentation
└── run                          # Compiled binary
```

## How to Add New Packages

### 1. Create Installation Script
Create a new script in `scripts/` directory:
```bash
# scripts/redis.sh
#!/bin/bash
echo "Installing Redis..."
# Installation logic here
```

### 2. Map Script in Registry
Add package mapping in `internal/registry.go`:
```go
var InstallPackageRegistry = map[string]string{
    "redis": "redis.sh",
    // ...existing mappings
}
```

### 3. Add Removal Script (Optional)
Create removal script:
```bash
# scripts/remove-redis.sh
#!/bin/bash
echo "Removing Redis..."
# Removal logic here
```

### 4. Map Removal Script
Add to removal registry in `internal/registry.go`:
```go
var RemovePackageRegistry = map[string]string{
    "redis": "remove-redis.sh",
    // ...existing mappings
}
```

