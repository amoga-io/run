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

## 📖 CLI Usage Guide

### **General Help**

```bash
run --help
```

## CLI Flags Summary Table

| Command | Flag            | Alias | Description                                       |
| ------- | --------------- | ----- | ------------------------------------------------- |
| install | --version       | -v    | Install specific version                          |
|         | --set-active    |       | Set installed version as active (version manager) |
|         | --all           | -a    | Install all packages                              |
|         | --clean         | -c    | Force clean reinstallation                        |
|         | --dry-run       | -d    | Preview install                                   |
|         | --replace       | -r    | Remove existing version before install            |
| remove  | --all           | -a    | Remove all packages                               |
|         | --force         | -f    | Force removal of critical packages                |
|         | --dry-run       | -d    | Preview removal                                   |
| check   | --system        | -s    | Check system health                               |
|         | --all           | -a    | Check all packages                                |
|         | --list-versions | -l    | List all installed versions                       |

### **Install Packages**

```bash
run install <package> [<package> ...] [flags]
```

#### Flags

- `--version <ver>`: Install a specific version (e.g., 18 for node, 3.10 for python)
- `--set-active`: Set the installed version as the active/default version in the version manager (for python, node, java, php)
- `--all`: Install all available packages
- `--clean`: Force clean reinstallation (remove existing first)
- `--dry-run`: Show what would be installed, but do not actually install anything

#### Version Manager Auto-Install

When installing a version-managed package (python, node, java, php), the required version manager (`pyenv`, `nvm`, `sdkman`, `phpenv`) will be automatically installed if missing.

#### Examples

```bash
run install node
run install python --version 3.10
run install node python docker
run install --all
run install node --clean  # Remove existing and install fresh
run install node --dry-run # Preview installation
run install python --version 3.10.5 --set-active # Install and set as active
run install node --version 18.20.4 --set-active # Install and set as active
```

### **Remove Packages**

```bash
run remove <package> [<package> ...] [flags]
```

#### Flags

- `--force`: Force removal (bypass critical package protection)
- `--dry-run`: Show what would be removed, but do not actually remove anything

#### Examples

```bash
run remove node
run remove python --force
run remove node python
run remove node --dry-run # Preview removal
```

### **Check Package Status**

```bash
run check <package> [<package> ...]
```

#### Examples

```bash
run check node python
run check all
```

### **List Available Packages**

```bash
run install list
```

### **Update System**

```bash
run update
```

### **Show Version**

```bash
run version
```

### **Command-Specific Help**

```bash
run install --help
run remove --help
run check --help
```

### **Adding New Packages**

1. **Add Package Definition** in `internal/package/registry.go`
2. **Create Installation Script** in `scripts/packages/`
3. **Add Tests** in `internal/package/manager_test.go`
4. **Update Documentation** in this README

### **Adding New Commands**

1. **Create Command File** in `cmd/` directory
2. **Register Command** in `cmd/root.go`
3. **Add Tests** for the new command
4. **Update Help Documentation**
