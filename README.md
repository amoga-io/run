# Run CLI

A minimal Git-based CLI tool for Ubuntu systems with self-install and update capabilities.

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/amoga-io/run/main/scripts/install.sh | bash
```

## Usage

```bash
# Show help
run --help

# Update to latest version
run update
```

## Requirements

- Ubuntu/Debian system
- sudo privileges
- Internet connection

The installer automatically handles Git and Go dependencies.

## Update

The CLI can update itself from the Git repository:

```bash
run update
```

This pulls the latest changes, rebuilds the binary, and atomically replaces the installed version.

## Uninstallation

To completely remove the CLI:

```bash
# Remove binary
sudo rm -f /usr/local/bin/run

# Remove repository and cache
rm -rf ~/.run

# Optional: Remove auto-installed dependencies
# sudo apt remove git golang-go
```

## Installation Details

- **Binary location**: `/usr/local/bin/run`
- **Repository cache**: `~/.run/`
- **Dependencies**: git, golang-go (auto-installed if missing)

## Enterprise Notes

This tool is designed for internal enterprise use. The installation process:

1. Verifies system compatibility (Ubuntu/Debian)
2. Checks and installs required dependencies
3. Clones repository to `~/.run/` for persistent updates
4. Builds and installs binary to `/usr/local/bin/`
5. Uses atomic file operations to prevent corruption

The update mechanism ensures the CLI stays current with the enterprise repository while maintaining system stability.

## License

MIT
