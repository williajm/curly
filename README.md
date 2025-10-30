# Curly

A modern, terminal-based API client built with Go. Curly provides a fast, keyboard-driven workflow for constructing, executing, and managing HTTP requests - like Postman, but for the terminal.

## Features

### Current (Phase 1 - MVP)
- Create and send HTTP requests (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS)
- Custom headers and query parameters
- JSON request bodies
- Formatted response display (status, headers, body)
- Multiple authentication methods:
  - Basic Authentication (username/password)
  - Bearer Token
  - API Key (header or query parameter)
- Request persistence with SQLite
- Request history tracking
- Intuitive terminal UI powered by Bubble Tea

### Planned Features
- **Phase 2**: Collections, environment variables, import/export, syntax highlighting
- **Phase 3**: Pre-request scripts, response assertions, GraphQL support, WebSocket connections
- **Phase 4**: OAuth 2.0, certificate authentication, performance testing, CI/CD integration

## Installation

### Prerequisites
- Go 1.22 or later
- golangci-lint (optional, for development)

### From Source

```bash
# Clone the repository
git clone https://github.com/williajm/curly.git
cd curly

# Build the binary
make build

# Install to $GOPATH/bin
make install
```

### Using Go Install

```bash
go install github.com/williajm/curly/cmd/curly@latest
```

## Usage

### Starting Curly

```bash
curly
```

### Keyboard Shortcuts

**Global:**
- `Tab` / `Shift+Tab` - Switch between views (Request, Response, History)
- `1` / `2` / `3` - Jump directly to Request / Response / History tab
- `?` - Show/hide help screen
- `Ctrl+C` / `q` - Quit application

**Request Tab:**
- `Ctrl+R` / `Ctrl+Enter` - Execute request
- `Tab` - Navigate between fields
- `←` / `→` - Change HTTP method or auth type

**Response Tab:**
- `h` - Toggle between headers and body view
- `↑` / `↓` - Scroll response content

**History Tab:**
- `↑` / `↓` - Navigate history entries
- `r` - Refresh history list
- `d` - Delete selected entry

### Basic Workflow

1. **Build Request**: Enter URL, select HTTP method, add headers and body
2. **Execute**: Press `Ctrl+R` to send the request
3. **View Response**: See formatted status code, headers, and body (automatically switches to Response tab)
4. **History**: All executed requests are automatically saved to history
5. **Browse History**: Switch to History tab to view and manage past requests

## Development

### Project Structure

```
curly/
├── cmd/curly/              # Application entry point
├── internal/
│   ├── app/                # Application services
│   ├── domain/             # Domain models
│   ├── infrastructure/     # HTTP client, database, config
│   └── presentation/       # TUI components (Bubble Tea)
├── pkg/                    # Public packages
├── migrations/             # Database migrations
├── docs/                   # Documentation
└── scripts/                # Build scripts
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run without building
make run
```

### Testing

```bash
# Run all tests with coverage
make test

# View coverage report in browser
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Run all checks (fmt + vet + lint + test)
make check
```

### Available Make Targets

```bash
make help
```

## Architecture

Curly follows clean architecture principles with clear separation of concerns:

- **Domain Layer**: Core business logic and models
- **Application Layer**: Use cases and service orchestration
- **Infrastructure Layer**: HTTP client, database, file I/O
- **Presentation Layer**: Terminal UI using Bubble Tea framework

## Configuration

Curly supports configuration via YAML file, environment variables, and command-line flags.

### Configuration File

Location: `~/.config/curly/config.yaml`

See `config.example.yaml` for a complete example. Copy it to the config location and customize:

```bash
mkdir -p ~/.config/curly
cp config.example.yaml ~/.config/curly/config.yaml
```

Example configuration:

```yaml
database:
  path: ~/.local/share/curly/curly.db

http:
  timeout: 30s
  max_redirects: 10
  follow_redirects: true
  insecure_skip_tls: false

ui:
  # The following UI options are planned for Phase 2:
  theme: dark                    # Not yet implemented
  syntax_highlighting: true      # Not yet implemented
  show_response_time: true       # Not yet implemented
  default_tab: request           # Not yet implemented

history:
  # History management features planned for Phase 2:
  max_entries: 1000              # Not yet enforced
  auto_cleanup: true             # Not yet implemented
  cleanup_after_days: 90         # Not yet implemented

logging:
  enabled: true
  path: ~/.cache/curly/curly.log
  level: info  # Options: debug, info, warn, error
```

**Note**: Some configuration options are loaded but not yet active in Phase 1. They are documented here for future use and will be fully implemented in Phase 2.

### Environment Variables

All configuration options can be set via environment variables with the `CURLY_` prefix:

```bash
export CURLY_HTTP_TIMEOUT=60s
export CURLY_LOGGING_LEVEL=debug
export CURLY_DATABASE_PATH=/custom/path/curly.db
```

### Command-Line Flags

```bash
# Use custom config file
curly --config /path/to/config.yaml

# Override database path
curly --db /path/to/database.db

# Show version
curly --version
```

## Data Storage

- **Configuration:** `~/.config/curly/config.yaml`
- **Database:** `~/.local/share/curly/curly.db` (SQLite)
- **Logs:** `~/.cache/curly/curly.log`

All paths follow the XDG Base Directory specification and can be customized via configuration. **All directories are automatically created on first run.**

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting (`make check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Standards

- Follow Go best practices and idioms
- Maintain test coverage above 75%
- Run `make check` before committing
- Write clear commit messages
- Document exported functions and types

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

See [development-plan-curly.md](development-plan-curly.md) for detailed development plan and roadmap.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Support

For bug reports and feature requests, please use the [GitHub Issues](https://github.com/williajm/curly/issues) page.
