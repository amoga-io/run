# Run CLI

A safe, Git-based CLI tool for managing developer and system packages on Ubuntu servers and VMs.

---

## 🚀 Quick Installation

```bash
curl -fsSL https://raw.githubusercontent.com/amoga-io/run/main/scripts/install.sh | bash
```

---

## 🧹 Uninstall

```bash
sudo rm -f /usr/local/bin/run
rm -rf ~/.run
```

---

## 📖 Commands & Flags Usage Guide

### General Help

```bash
run --help
```

---

### Install Packages

Install one or more packages (Node.js, Python, Docker, etc.):

```bash
run install <package> [<package> ...] [flags]
```

#### Flags
- `--version <ver>`: Install a specific version (if supported)
- `--all`: Install all available packages

#### Examples
```bash
run install node
run install python --version 3.10
run install node python docker
run install --all
```

---

### Remove Packages

Remove one or more packages:

```bash
run remove <package> [<package> ...] [flags]
```

#### Flags
- `--version <ver>`: Remove a specific version (if supported)
- `--all`: Remove all installed packages

#### Examples
```bash
run remove node
run remove python --version 3.10
run remove node python
run remove --all
```

---

### List Available Packages

```bash
run install list
```

---

### Check System & Packages

Check installed packages, dependencies, or all:

```bash
run check packages
run check deps
run check all
```

---

### Show Version

```bash
run version
```

---

### Update CLI

Update the CLI to the latest version from Git:

```bash
run update
```

---

### Command-Specific Help

```bash
run install --help
run remove --help
run check --help
```

---

## 📝 Notes
- Only user-installed versions are removed; system/essential packages are never touched.
- For a full list of supported packages and versions, run `run install list`.
- This CLI is designed for Ubuntu/Debian systems.
