# Agent Instructions for XSC

This is a Go-based SSH session manager CLI tool using Bubble Tea for TUI.

## Build Commands

```bash
# Build the binary
make build

# Run the application
make run
# Or specific subcommands:
make tui      # Run TUI mode
make list     # List sessions

# Install to /usr/local/bin
make install

# Clean build artifacts
make clean
```

## Test Commands

```bash
# Run all tests
make test
# Equivalent to: go test -v ./...

# Run a single test file
go test -v ./internal/session/... -run TestLoadSession

# Run tests for a specific package
go test -v ./internal/ssh/...

# Run with coverage
go test -v -cover ./...
```

## Lint/Format Commands

```bash
# Format all Go code (uses go fmt)
make fmt

# Run go vet for static analysis
make vet

# Full check (run all of the above)
make fmt && make vet && make test
```

## Dependency Management

```bash
# Download and tidy dependencies
make deps
# Equivalent to: go mod download && go mod tidy
```

## Code Style Guidelines

### Imports
- Group imports: stdlib first, then third-party, then local
- Local imports use module path: `github.com/user/xsc/internal/session`
- Example:
```go
import (
    "fmt"
    "os"

    "github.com/charmbracelet/bubbletea"
    "golang.org/x/crypto/ssh"

    "github.com/user/xsc/internal/session"
    "github.com/user/xsc/pkg/config"
)
```

### Naming Conventions
- Exported: PascalCase (`AuthType`, `LoadSession`)
- Unexported: camelCase (`connectWithPassword`, `handleWindowResize`)
- Constants: CamelCase or PascalCase (`AuthTypePassword`)
- Interface names: -er suffix (`Reader`, `Writer`)
- Acronyms: all caps (`SSH`, `TUI`, `CRT`)

### Types
- Use struct tags for YAML: `` `yaml:"field_name,omitempty"` ``
- Internal fields use `yaml:"-"` to exclude from serialization
- Use custom types for enums: `type AuthType string`

### Error Handling
- Wrap errors with context: `fmt.Errorf("failed to load: %w", err)`
- Use `%w` verb for error wrapping (not `%v` in error paths)
- Return errors rather than logging and continuing
- Validate input early and return descriptive errors

### Comments
- Comments in Chinese OK (existing code uses Chinese)
- Document exported functions with `// FunctionName ...`
- Use `// TODO:` for incomplete implementations

### Project Structure
```
cmd/xsc/          # Entry point (main.go)
internal/         # Private packages
  session/        # Session management
  ssh/            # SSH connection logic
  tui/            # Bubble Tea TUI implementation
  securecrt/      # SecureCRT integration
  tree/           # Tree data structures
pkg/config/       # Public configuration package
```

### Patterns
- Use functional options pattern for configuration when needed
- Prefer composition over inheritance
- Use defer for resource cleanup
- Context cancellation for long-running operations

### Testing
- Test files: `*_test.go` alongside source files
- No tests currently exist - add when modifying existing code
- Use table-driven tests
- Mock external dependencies

## IDE/Editor Settings

No specific IDE configuration files exist. Use standard Go tooling:
- `gofmt` for formatting
- `go vet` for static analysis
- `goimports` for import management

## Development Workflow

1. Run `make fmt` before committing
2. Run `make vet` to catch issues
3. Run `make test` to verify functionality
4. Build with `make build` to test compilation
5. Manual testing: `go run ./cmd/xsc <command>`
