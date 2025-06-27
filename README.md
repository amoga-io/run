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

# Show version information
run version

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

This tool is designed for internal enterprise use with Git-based versioning:

1. **Versioning**: Uses Git tags for semantic versioning (v1.0.0, v1.2.3, etc.)
2. **Releases**: GitHub Actions automatically creates releases when tags are pushed
3. **Installation**: Always builds from source with proper version embedding
4. **Updates**: Self-updating mechanism maintains version consistency

The installation process:

1. Verifies system compatibility (Ubuntu/Debian)
2. Checks and installs required dependencies
3. Clones repository to `~/.run/` for persistent updates
4. Builds with version information embedded from Git
5. Installs binary to `/usr/local/bin/` using atomic operations

## Versioning

The CLI uses Git-based versioning:

```bash
# Create and push a new version
git tag v1.0.0
git push origin v1.0.0

# This triggers GitHub Actions to create a release
# Next installation will show the new version
```

Default version is `v0.0.0` when no tags exist.

## License

MIT
