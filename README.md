# GCP Switcher

A terminal user interface (TUI) for managing Google Cloud Platform accounts and projects.

## Features

- View and switch between GCP accounts
- View and switch between GCP projects
- Login to new GCP accounts
- Manual project ID entry
- Debug logging support
- Interactive UI with keyboard navigation
- Cross-platform support (Linux, Windows, macOS)

## Prerequisites

- Go 1.21 or later
- Google Cloud SDK (gcloud) installed and in PATH

## Installation

```bash
git clone https://github.com/mathieudupuis/gcp-switcher.git
cd gcp-switcher
make build  # Build for current platform
# or
make build-all  # Build for all platforms
```

The binaries will be created in the `bin` directory:
- Linux: `bin/gcp-switcher-linux-amd64`
- Windows: `bin/gcp-switcher-windows-amd64.exe`
- macOS Intel: `bin/gcp-switcher-darwin-amd64`
- macOS ARM: `bin/gcp-switcher-darwin-arm64`

## Usage

Run the application:

```bash
# On Unix-like systems (Linux/macOS)
./bin/gcp-switcher

# On Windows
.\bin\gcp-switcher.exe
```

With debug logging enabled:

```bash
# On Unix-like systems (Linux/macOS)
./bin/gcp-switcher --debug

# On Windows
.\bin\gcp-switcher.exe --debug
```

### Controls

- `↑/↓` or `j/k`: Navigate through options
- `Enter`: Select option
- `q`: Quit or go back
- `Ctrl+C`: Quit application

## Project Structure

```
.
├── cmd/
│   └── gcp/       # GCP command execution logic
├── internal/      # Core application logic
├── types/         # Data structures and types
├── ui/           # UI styling and theme
└── main.go       # Application entry point
```

## Development

### Build Commands

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build for specific platforms
make build-linux
make build-windows
make build-mac

# Run the application
make run

# Run with debug logging
make run-debug

# Clean build artifacts
make clean

# Run tests
make test

# Format code
make fmt

# Run linter
make lint
```

For more commands, run:
```bash
make help
```

## License

See [LICENSE](LICENSE) file.