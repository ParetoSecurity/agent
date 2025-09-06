# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Pareto Security Agent Development Guide

## Build & Test Commands
- Build main agent: `goreleaser release --snapshot --clean`
- Build tray app: `goreleaser release --snapshot --clean`
- Build installer: `goreleaser release --snapshot --clean`
- Test all: `go test ./...`
- Test specific package: `go test github.com/ParetoSecurity/agent/checks/linux`
- Test single test: `go test -run TestApplicationUpdates_Run ./checks/linux`
- Coverage: `go test -coverprofile=coverage.txt ./...`
- Lint: Uses pre-commit hooks with Go and Nix linters: `pre-commit run --all-files`
- Run checks: `./paretosecurity check`
- Run specific check: `./paretosecurity check --only <UUID>`

## Architecture Overview

### Check System
The core of Pareto Security is the check system. Each security check implements the Check interface:
- `Name()`, `PassedMessage()`, `FailedMessage()` - Display messages
- `Run()` - Executes check logic, sets internal state
- `Passed()` - Returns whether check passed
- `IsRunnable()` - Determines if check applies to current system
- `UUID()` - Unique identifier
- `Status()` - Detailed status beyond pass/fail
- `RequiresRoot()` - Whether elevated privileges needed

Checks are organized by platform in `/checks/{darwin,linux,windows,shared}/` and grouped into security claims in `/claims/`.

### Runner Architecture
- Standard runner executes checks concurrently
- Root runner handles privileged checks via Unix socket
- Check states: pass, fail, off (disabled), error
- Results persisted between runs

### Platform Support
- Linux: apt, dnf, pacman, snap, flatpak package managers; LUKS, firewall, secure boot checks
- Windows: Defender, updates, firewall checks; system tray integration
- macOS: Application-specific security checks
- Shared: SSH keys, remote login, port security

## Code Style Guidelines
- Imports: Standard library first, third-party packages second, project-specific last
- Formatting: Use `gofmt` standard formatting; tabs for indentation
- Naming: CamelCase for exported identifiers, camelCase for unexported
- Interfaces: Prefer small, focused interfaces with clear purpose
- Tests: Table-driven tests with descriptive names, using assert package
- Error handling: Return errors up the stack, use early returns with `if err != nil`
- Logging: Use `log.WithField/WithError` for structured logging
- Mocks: Use dependency injection patterns with interface mocks for testing
- Documentation: Add descriptive comments for all exported functions/methods
- File layout: Related types and their implementations together

## Testing Patterns
- Mock external commands using `RunCommandMock` pattern
- Test checks in isolation with mocked dependencies
- Minimum 45% test coverage enforced
- Integration tests use NixOS VMs in `/test/integration/`

## Development Workflow
- Use `devenv shell` for consistent environment
- Pre-commit hooks enforce formatting
- GoReleaser handles multi-platform builds
- CGO disabled for all builds
- Version info injected via ldflags

## Adding New Checks
1. Create struct implementing Check interface
2. Place in appropriate platform directory
3. Add to relevant claim in `/claims/`
4. Implement `IsRunnable()` to detect applicability
5. Set `RequiresRoot()` if privileged access needed
6. Write unit tests with mocked dependencies
7. Test manually with `./paretosecurity check --only <UUID>`