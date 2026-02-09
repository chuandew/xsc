# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

XSC (XShell CLI) is a Go-based SSH session manager with a TUI (Bubble Tea) and CLI interface. Sessions are YAML files stored in `~/.xsc/sessions/`; the directory hierarchy becomes the tree structure in the TUI. It also supports importing and decrypting SecureCRT sessions.

## Build & Development Commands

```bash
make build          # Build binary to ./build/xsc
make test           # Run all tests: go test -v ./...
make fmt            # Format code: go fmt ./...
make vet            # Static analysis: go vet ./...
make run            # Run TUI mode
make install        # Install to /usr/local/bin
make deps           # Download and tidy dependencies
```

Run a specific test:
```bash
go test -v ./internal/securecrt/... -run TestDecryptPasswordV2Real
```

Full quality check: `make fmt && make vet && make test`

## Architecture

**Entry point**: `cmd/xsc/main.go` — command dispatcher routing to `tui`, `list`, `connect`, `import-securecrt`, `help`.

**Core packages** (all under `internal/`):

- **tui/** — Bubble Tea model-view-update loop. Single file `tui.go` (~1300 LOC) containing the full TUI: multiple modes (normal, search, command, help, error), Vim-style keybindings defined in `KeyMap`/`DefaultKeyMap()`, virtual scrolling, tree rendering, details panel. Styles are defined at the top of the file. SSH connections launch via `tea.Exec()` which pauses the TUI.

- **session/** — `session.go` defines the `Session` struct (YAML-serialized) with three `AuthType` values: `password`, `key`, `agent`. `tree.go` implements `SessionNode` for hierarchical tree organization with expand/collapse and filtering.

- **ssh/** — Pure Go SSH client (`golang.org/x/crypto/ssh`) with three auth strategies (password, key, SSH Agent). Handles terminal raw mode and SIGWINCH for window resize.

- **securecrt/** — Parses SecureCRT `.ini` session files. Decrypts V2 passwords (prefix `02` uses SHA256+AES-256-CBC, prefix `03` uses bcrypt_pbkdf). Lazy decryption for performance. `bcrypt_pbkdf.go` has the custom key derivation.

**Public package**: `pkg/config/` — global config singleton loaded from `~/.xsc/config.yaml`. Manages paths and settings for SecureCRT integration and SSH host key verification.

## Code Conventions

- **Language**: Go 1.21+. Documentation and code comments are written in Chinese.
- **Import order**: stdlib, then third-party, then local (`github.com/user/xsc/...`)
- **Naming**: PascalCase for exported, camelCase for unexported. Acronyms stay uppercase (`SSH`, `TUI`, `CRT`).
- **YAML tags**: `yaml:"field_name,omitempty"` on struct fields; internal fields use `yaml:"-"`
- **Error handling**: wrap with `fmt.Errorf("context: %w", err)` — use `%w` not `%v`
- **File permissions**: session files saved as 0600
- **Specs**: Gherkin format in `specs/xsc.feature` — keep tests in sync with specs

## Extending the Codebase

- **New auth method**: add `AuthType` constant in `session.go` → add validation in `Session.Validate()` → implement in `ssh/client.go` → update TUI details panel
- **New CLI command**: add case in `cmd/xsc/main.go` switch → implement handler → update `showHelp()`
- **TUI changes**: keybindings in `KeyMap`/`DefaultKeyMap()`, rendering in `View()` and `renderXxx()` methods, styles at top of `tui.go`
