package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	AuthURL = "https://cn.account.tcl.com"
	IoTURL  = "https://io.zx.tcljd.com"

	// Public OAuth 2.0 client credentials for TCL Smart Home CLI (QR code authorization flow).
	// These are public client identifiers — NOT private secrets. They can only be used to
	// display a QR code authorization page and cannot access any user data without explicit
	// user consent via QR code scanning. Safe to embed in open-source code.
	AppID     = "23301773132314850"
	AppSecret = "504a9b7062c85d7444e2e60a5cf70775769d14c2eda70cef263719d254359b18"
	TenantID  = "tcl"

	MillisPerDay int64 = 24 * 60 * 60 * 1000
)

// CLIConfig holds the stored credentials.
type CLIConfig struct {
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	TokenExpires int64  `json:"tokenExpiresAt,omitempty"`
}

// Dir returns the config directory path (~/.tcl-cli/).
func Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".tcl-cli"
	}
	return filepath.Join(home, ".tcl-cli")
}

// FilePath returns the config file path (~/.tcl-cli/config.json).
func FilePath() string {
	return filepath.Join(Dir(), "config.json")
}

// Load reads the config from disk. Returns empty config if file doesn't exist.
func Load() (*CLIConfig, error) {
	data, err := os.ReadFile(FilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &CLIConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var cfg CLIConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *CLIConfig) error {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(FilePath(), data, 0o600)
}

// MustLoad loads config and ensures accessToken is present.
func MustLoad() (*CLIConfig, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("not configured. Run \"tcl init\" to authorize")
	}
	return cfg, nil
}

// TokenExpiryMs parses the exp claim from a JWT and returns it as Unix milliseconds.
func TokenExpiryMs(jwtToken string) (int64, error) {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid JWT format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, fmt.Errorf("failed to decode JWT payload: %w", err)
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, fmt.Errorf("failed to parse JWT claims: %w", err)
	}
	if claims.Exp == 0 {
		return 0, fmt.Errorf("JWT missing exp claim")
	}
	return claims.Exp * 1000, nil
}
