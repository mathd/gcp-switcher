# GCP Switcher

A terminal user interface (TUI) application for easily managing and switching between Google Cloud Platform accounts and projects.

## Features

- 🔄 View and switch between multiple GCP accounts
- 📂 List and switch between GCP projects
- 🔑 Login to new GCP accounts
- 🎯 Manually enter project IDs
- 💻 Interactive terminal interface with keyboard navigation
- 🎨 Beautiful TUI with color-coded information

## Prerequisites

- Go 1.x or higher
- Google Cloud SDK (gcloud) installed and in PATH
- Terminal with support for TUI applications

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/gcp-switcher.git
cd gcp-switcher
```

2. Build the application:
```bash
go build
```

## Usage

Run the application:
```bash
./gcp-switcher
```

With debug logging enabled:
```bash
./gcp-switcher --debug
```

### Navigation

- Use `↑`/`↓` or `j`/`k` to navigate through menus
- `Enter` to select an option
- `q` to quit or go back to the main menu
- Numbers `1-4` for quick menu selection

### Main Menu Options

1. View/Switch Accounts - List all configured accounts and switch between them
2. View/Switch Projects - List all available projects and switch to a different project
3. Login to a New Account - Add a new GCP account
4. Enter Project ID Manually - Switch to a project by entering its ID directly

## Implementation Details

Built using:
- [Bubble Tea](github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lip Gloss](github.com/charmbracelet/lipgloss) - Style definitions
- Google Cloud SDK - GCP interaction

## Debug Mode

When run with the `--debug` flag, the application creates a log file (`gcp-switcher.log`) with detailed operation information, useful for troubleshooting.

## Error Handling

- Automatically checks for Google Cloud SDK installation
- Handles command timeouts gracefully
- Provides clear error messages in the UI
- Fallback timer prevents infinite loading states

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.