# devkit

A simple, cross-platform CLI tool written in Go for running custom scripts and installing packages from the terminal. Built with the Cobra CLI framework, `devkit` is designed to be easily extensible and user-friendly.

## Features

- Run custom shell scripts with subcommands
- Install tools like Neofetch via CLI
- Easily extensible for new scripts and commands
- Cross-platform support

## Installation

To install `devkit` globally, run:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/amoga-io/run/main/install.sh)
```

Or, build from source:

```bash
git clone https://github.com/amoga-io/run.git
cd devkit
go build -o devkit
```

## Usage

### Run a script

```bash
devkit run hello -a "World"
```

### Install Neofetch

```bash
devkit install neofetch
```

### Show help

```bash
devkit --help
```

## Project Structure

```
.
├── cmd/         # CLI command implementations
├── scripts/     # Shell scripts
├── install.sh   # Installer script
├── main.go      # Entry point
```

## Requirements

- Go 1.18+
- Bash
- Internet connection (for some scripts)

## Uninstallation

To completely remove `devkit` from your system:

1. Delete the `devkit` binary:

   ```bash
   rm -f $(which devkit)
   # or
    rm -f /usr/local/bin/devkit
   ```

2. (Optional) Remove the cloned repository:

   ```bash
   rm -rf ~/.devkit
   ```

## Contributing

Contributions are welcome! Please open issues or submit pull requests for new features, bug fixes, or improvements.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
