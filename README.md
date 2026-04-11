# TCL Smart Home CLI

[![CI](https://github.com/tcl-eaglelab/tcl-smart-home-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/tcl-eaglelab/tcl-smart-home-cli/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/tcl-eaglelab/tcl-smart-home-cli)](https://github.com/tcl-eaglelab/tcl-smart-home-cli)
[![License](https://img.shields.io/github/license/tcl-eaglelab/tcl-smart-home-cli)](LICENSE)

[English](./README.md) | [中文](./README_zh.md)

The official command-line tool for TCL Smart Home — enabling both humans and AI Agents to control TCL smart home devices from the terminal.

A single binary with zero dependencies. Built with Go.

[Installation](#installation) · [AI Agent Skill](#ai-agent-skill) · [Authentication](#authentication) · [Commands](#commands) · [Security](#security--risk-warnings-read-before-use) · [Contributing](#contributing)

## Features

| Capability | Description |
|------------|-------------|
| 🔐 QR Auth | Scan with TCL App, one-time setup, 30-day token |
| 📱 Device Management | List devices, query status, inspect Thing Model |
| 🎮 Device Control | Property-based control with read-only filtering |
| 🤖 AI Friendly | All commands support `--json` structured output, `NO_COLOR` aware |
| 🛡️ Safety Check | Read-only properties auto-filtered before control |
| ⚡ Zero Dependencies | Single binary, download and run |

## Installation

### Option 1 — npm (Recommended)

```bash
npm install -g @tcl-eaglelab-01/smart-home-cli
```

### Option 2 — Download Binary

Download from [GitHub Releases](https://github.com/tcl-eaglelab/tcl-smart-home-cli/releases):

| Platform | File |
|----------|------|
| Linux x64 | `tcl-<version>-linux-amd64.tar.gz` |
| Linux ARM64 | `tcl-<version>-linux-arm64.tar.gz` |
| macOS Intel | `tcl-<version>-darwin-amd64.tar.gz` |
| macOS Apple Silicon | `tcl-<version>-darwin-arm64.tar.gz` |
| Windows x64 | `tcl-<version>-windows-amd64.zip` |

### Option 3 — Go Install

```bash
go install github.com/tcl-eaglelab/tcl-smart-home-cli/cmd/tcl@latest
```

### Option 4 — Build from Source

```bash
git clone https://github.com/tcl-eaglelab/tcl-smart-home-cli.git
cd tcl-smart-home-cli
go build -o tcl ./cmd/tcl
```

## AI Agent Skill

Install the [Skill](https://www.npmjs.com/package/skills) for AI Agents:

```bash
npx skills add tcl-eaglelab/tcl-smart-home-cli -y -g
```

Once installed, AI Agents (OpenClaw, Claude Code, Cursor, etc.) will automatically gain TCL device control capabilities.

## Authentication

Authorization uses a two-step flow designed for both human and AI Agent use:

**Step 1 — Generate QR code URL:**
```bash
tcl init
```
Outputs a QR code URL and exits immediately. Open the URL in a browser and scan with the **TCL App**.

**Step 2 — Verify after scanning:**
```bash
tcl init --verify <qr_code>
```
The `<qr_code>` value is printed by step 1. Polls for authorization and saves credentials on success.

Tokens are valid for 30 days and auto-refresh before expiry.

Verify your setup:

```bash
tcl doctor
```

## Commands

```
tcl <command> [subcommand] [flags]

Commands:
  init           Configure account credentials via QR code
  doctor         Check configuration and connectivity
  device list    List all devices
  device detail  Device detail and current status
  device model   Thing Model (supported capabilities)
  device control Control a device
```

> All commands support the `--json` flag for structured output.

---

### List Devices

```bash
tcl device list
```

### Device Detail

```bash
tcl device detail <device_id>
```

### Thing Model

```bash
tcl device model <device_id>
```

### Control Device

```bash
tcl device control <device_id> '{"powerSwitch":1}'
tcl device control <device_id> '{"targetTemperature":26}'
tcl device control <device_id> '{"powerSwitch":1,"workMode":1,"targetTemperature":26}'
```

**Windows PowerShell** — use doubled double quotes:

```powershell
tcl device control <device_id> "{""powerSwitch"":1}"
```

> Different devices have completely different Thing Models. Always check `tcl device model` before controlling.

---

## Common Workflows

```bash
tcl device list                              # 1. Find your device
tcl device model <id>                        # 2. Check supported properties
tcl device detail <id>                       # 3. Check current state
tcl device control <id> '{"powerSwitch":1}'  # 4. Send command
```

## JSON Output

```bash
tcl device list --json
tcl device list --json | jq '.[].deviceId'
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NO_COLOR` | Set to any value to disable colored output ([no-color.org](https://no-color.org)) |

## Config File

Credentials stored at `~/.tcl-cli/config.json`:

```json
{
  "accessToken": "eyJhbG...",
  "refreshToken": "eyJhbG...",
  "tokenExpiresAt": 1234567890
}
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Config not found | Run `tcl init` |
| Token expired | Run `tcl init` to re-authorize |
| Device not found | Check device ID with `tcl device list` |
| Control failed | Check `tcl device model` — read-only properties cannot be controlled |
| Device offline | Check device WiFi connection |

## Security & Risk Warnings (Read Before Use)

This tool can be invoked by AI Agents to automate control of TCL smart home devices, and carries inherent risks such as model hallucinations, unpredictable execution, and prompt injection. After you authorize via QR code, the AI Agent will act under your user identity within the authorized scope, which may lead to unintended device operations.

To reduce these risks, the tool filters read-only properties before sending control commands. However, risks still exist. Please ensure your devices are in a safe state before allowing automated control.

By using this tool, you acknowledge and accept all related risks and responsibilities.

## Project Structure

```
tcl-smart-home-cli/
├── cmd/
│   ├── tcl/
│   │   └── main.go      # Entry point
│   ├── root.go          # Root command
│   ├── init.go          # QR code authorization
│   ├── doctor.go        # Diagnostics
│   └── device.go        # Device management
├── internal/
│   ├── api/             # TCL IoT API client
│   ├── config/          # Config management
│   └── output/          # Output formatting
├── scripts/
│   ├── install.js       # npm postinstall (downloads binary)
│   └── run.js           # npm bin wrapper
├── skills/              # AI Agent Skill
├── package.json         # npm distribution
├── .goreleaser.yml      # Release automation
├── Makefile             # Local development
└── .github/workflows/   # CI/CD
```

## License

MIT

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).
