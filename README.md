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
git clone https://github.com/mathd/gcp-switcher.git
cd gcp-switcher
make build
# or
make build-all
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
├── .github/  # GitHub Actions workflow files (e.g., for CI/CD)
├── cmd/  # GCP command execution logic
│   └── gcp/  # GCP command execution logic
├── internal/  # Core application logic
│   └── version/  # Handles application version information
├── types/  # Data structures and types
├── ui/  # UI styling and theme
├── go.mod  # Defines the module's properties, including its dependencies
├── go.sum  # Contains the checksums of direct and indirect dependencies
└── main.go  # Application entry point
```

## Development

### Build Commands

```bash
make build
make build-all
make build-linux
make build-windows
make build-mac
make run
make run-debug
make clean
make test
make fmt
make lint
```

For more commands, run:
```bash
make help
```

## License

See [LICENSE](LICENSE) file.