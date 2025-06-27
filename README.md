# Run CLI

A safe, Git-based CLI tool for managing developer and system packages on Ubuntu servers and VMs.

## Features

- Install, update, and remove packages (Node.js, Python, PHP, Docker, Nginx, Postgres, Java, PM2, and more)
- Safe removal: only user-installed versions are removed, never system/essential packages
- Version selection: install or remove specific versions with `--version`
- Self-updating and self-verifying
- Designed for enterprise VM/server use

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/amoga-io/run/main/scripts/install.sh | bash
```

## Usage

### List available packages

```bash
run install list
```

### Install a package (default version)

```bash
run install node
```

### Install a specific version

```bash
run install node --version 18
run install python --version 3.10
```

### Remove a package (user-installed versions only)

```bash
run remove node
run remove python --version 3.10
```

### Show version info

```bash
run version
```

### Update the CLI

```bash
run update
```

## Safe Removal

- To remove a package, run:
  ```
  run remove node
  run remove python --version 3.10
  ```
- Only user-installed versions are removed. System/essential versions are never touched.
- After removal, the system is ready for a fresh install of any supported version.
- To see all available packages, run:
  ```
  run install list
  ```

## Safety

- The CLI will **never remove system or essential package versions** (e.g., system Python).
- Only user-installed versions (listed in the code) are targeted for removal.
- Always review the list of user versions in the code and update as needed for your environment.

## Help

```bash
run --help
run install --help
run remove --help
```

## License

MIT
