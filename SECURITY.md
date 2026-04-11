# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please send an email to **ex_song.guo@tcl.com** with the following information:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will acknowledge your report within 48 hours and work with you to understand and address the issue before any public disclosure.

## Credential Storage

Access tokens are stored in plain text at `~/.tcl-cli/config.json` with restricted file permissions (`0600` owner-only). This is consistent with industry-standard CLI tools (e.g., GitHub CLI, AWS CLI). Users should protect their home directory accordingly.

**Windows note:** Go's file permission bits (`0600`) are not enforced on Windows. The config file may be readable by other users on the same machine. This is a known Go + Windows limitation shared by most CLI tools. Windows users should ensure their user profile directory is properly secured.

## Scope

This project controls physical smart home devices (air conditioners, refrigerators, washing machines, etc.). Security vulnerabilities that could lead to unauthorized device control are treated with the highest priority.

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.0.x   | Yes       |
