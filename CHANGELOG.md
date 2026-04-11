# Changelog

## v1.0.0 (2026-04-12)

The first open-source release of **TCL Smart Home CLI** — the official command-line tool for [TCL Smart Home](https://www.tcl.com/).

### Features

#### Core Commands

- **`tcl init`** — Two-step QR code authorization via TCL App for AI Agent compatibility. One-time setup with 30-day auto-refreshing tokens.
- **`tcl doctor`** — Diagnostic checks on CLI configuration and API connectivity.
- **`tcl device list`** — List all smart home devices under your account.
- **`tcl device detail`** — Query real-time device status and properties.
- **`tcl device model`** — Inspect Thing Model (supported capabilities) for any device.
- **`tcl device control`** — Property-based device control with automatic read-only safety filtering.

#### AI Agent Integration

- All commands support `--json` structured output for AI Agent consumption.
- Bundled [AI Agent Skill](https://www.npmjs.com/package/skills) for automatic device control capabilities.
- `NO_COLOR` environment variable support ([no-color.org](https://no-color.org)).

#### Safety

- Read-only properties are automatically filtered before sending control commands.
- Token expiry parsed from JWT `exp` claim for accurate validity tracking.
- Token refresh happens transparently before expiry.

#### Developer Experience

- Cross-platform support: Linux (x64/ARM64), macOS (Intel/Apple Silicon), Windows (x64).
- Static binaries (CGO_ENABLED=0) for maximum Linux compatibility.
- Zero dependencies — single binary, download and run.
- Multiple installation methods: npm, Go install, binary download, build from source.
- Bilingual documentation (English & Chinese).
- CI/CD: multi-platform testing, automated releases via GoReleaser, automated npm publishing.
