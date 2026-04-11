# Contributing to TCL Smart Home CLI

Thanks for your interest in contributing! Here's how to get started.

## Development Setup

Requires **Go 1.22+**.

```bash
git clone https://github.com/tcl-eaglelab/tcl-smart-home-cli.git
cd tcl-smart-home-cli
make build
```

## Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary |
| `make vet` | Run static analysis |
| `make test` | Run tests |
| `make clean` | Remove build artifacts |

## Workflow

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Run `gofmt -w .` to format code
4. Run `golangci-lint run ./...` to check for lint issues (install: https://golangci-lint.run/docs/install/)
5. Run `make build` to ensure it compiles
6. Run `make vet` to check for issues
7. Run `make test` to verify tests pass
8. Submit a pull request

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `output.Error` / `output.Success` / `output.Warn` for user-facing messages
- All query commands should support `--json` for structured output
- Errors should be returned, not printed directly

## Reporting Issues

- Use GitHub Issues
- Include the output of `tcl doctor` if relevant
- Include the command you ran and the error message

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
