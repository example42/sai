# SAI - Software Action Interface

SAI is a lightweight CLI tool for executing software management actions using provider-based configurations. The core philosophy is "Do everything on every software on every OS" through a unified interface.

## Project Structure

```
├── cmd/sai/           # Main application entry point
├── internal/cli/      # CLI command implementations
├── pkg/              # Reusable packages (future)
├── build/            # Build artifacts for current platform
├── dist/             # Cross-platform build artifacts
├── docs/             # Documentation and examples
├── providers/        # Provider implementation files
├── schemas/          # JSON Schema validation files
├── Makefile          # Build configuration
└── go.mod            # Go module definition
```

## Building

### Build for current platform
```bash
make build
```

### Build for all platforms
```bash
make build-all
```

### Development build with race detection
```bash
make build-dev
```

## Usage

```bash
# Install software
./build/sai install nginx

# Show version information
./build/sai version nginx

# List installed software
./build/sai list

# Show help
./build/sai --help
```

## Global Flags

- `--config/-c`: Custom configuration file
- `--provider/-p`: Force specific provider
- `--verbose/-v`: Enable verbose output
- `--dry-run`: Show commands without executing
- `--yes/-y`: Auto-confirm prompts
- `--quiet/-q`: Suppress non-essential output
- `--json`: Output in JSON format

## Development

This project follows Go best practices with a clear separation of concerns:

- `cmd/`: Application entry points
- `internal/`: Private application code
- `pkg/`: Public reusable packages

The CLI is built using the Cobra framework for robust command-line interface handling.