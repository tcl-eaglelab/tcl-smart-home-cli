package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"
	"github.com/tcl-eaglelab/tcl-smart-home-cli/internal/config"
	"github.com/tcl-eaglelab/tcl-smart-home-cli/internal/output"
)

const (
	maxPollAttempts = 12
	pollInterval    = 5 * time.Second
	qrCodeSize      = "400"
)

// Shorter timeout than api.Client (30s) since auth endpoints are lightweight.
var initHTTPClient = &http.Client{Timeout: 15 * time.Second}

type qrResponse struct {
	Status int `json:"status"`
	Data   struct {
		Code     string `json:"code"`
		ImageURL string `json:"imageUrl"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type authCheckResponse struct {
	Status int `json:"status"`
	Data   struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refreshToken"`
	} `json:"data"`
	Msg string `json:"msg"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Configure TCL account credentials via QR code authorization",
	Long: `Configure TCL account credentials via QR code authorization.

Two-step flow for AI Agent compatibility:

  Step 1 — Generate QR code URL (default):
    tcl init
    Open the URL in a browser and scan with TCL App, then run step 2.

  Step 2 — Verify authorization after scanning:
    tcl init --verify <qr_code>

The QR code value is printed by step 1. AI Agents should run step 1,
send the URL to the user, wait for confirmation, then run step 2.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verifyCode, _ := cmd.Flags().GetString("verify")
		if verifyCode != "" {
			return runInitVerify(verifyCode)
		}
		return runInitGetURL()
	},
}

// runInitGetURL generates a QR code URL and exits immediately.
// Designed for AI Agent use: the agent calls this, sends the URL to the user,
// and waits for the user to confirm scanning before calling runInitVerify.
func runInitGetURL() error {
	fmt.Println()
	output.Bold("  TCL Smart Home CLI Setup")
	output.Gray("  ─────────────────────────────")
	fmt.Println()

	output.Gray("  Requesting QR code...")
	qr, err := getQRCode()
	if err != nil {
		return fmt.Errorf("failed to get QR code: %w", err)
	}

	qrCodeURL := fmt.Sprintf("%s/auth/common/v2/generateQRCodePic?code=%s", config.AuthURL, url.QueryEscape(qr.Data.Code))

	fmt.Println()
	output.Success("QR code generated")
	fmt.Println()
	output.Bold("  Scan with TCL App:")
	output.Gray("  ─────────────────────────────")
	fmt.Printf("  %s\n", output.FormatCyan(qrCodeURL))
	output.Gray("  ─────────────────────────────")
	fmt.Println()
	output.Gray("  After scanning, run:")
	fmt.Println()
	fmt.Printf("  $ tcl init --verify \\\n    %s\n", qr.Data.Code)
	fmt.Println()

	return nil
}

// runInitVerify polls for authorization using the QR code from step 1.
func runInitVerify(code string) error {
	fmt.Println()
	output.Gray("  Waiting for authorization (polling every 5s, timeout 60s)...")
	fmt.Println()

	var token, refreshToken string
	for i := 1; i <= maxPollAttempts; i++ {
		auth, err := checkQRCodeAuth(code)
		if err != nil {
			output.Warn(fmt.Sprintf("  Attempt %d/%d: network error — %v", i, maxPollAttempts, err))
			if i < maxPollAttempts {
				time.Sleep(pollInterval)
			}
			continue
		}
		if auth.Status == 1 && auth.Data.Token != "" {
			token = auth.Data.Token
			refreshToken = auth.Data.RefreshToken
			break
		}

		msg := "Waiting..."
		if auth.Msg != "" {
			msg = auth.Msg
		}

		output.Gray(fmt.Sprintf("  Attempt %d/%d: %s", i, maxPollAttempts, msg))
		if i < maxPollAttempts {
			time.Sleep(pollInterval)
		}
	}

	if token == "" {
		fmt.Println()
		output.Error("Authorization timed out. Run \"tcl init\" to try again.")
		return errSilent
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.CLIConfig{}
	}
	cfg.AccessToken = token
	cfg.RefreshToken = refreshToken
	expiry, err := config.TokenExpiryMs(token)
	if err != nil {
		return fmt.Errorf("failed to parse token expiry: %w", err)
	}
	cfg.TokenExpires = expiry

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	output.Success("Authorization successful!")
	output.Info("Config saved to " + config.FilePath())
	fmt.Println()
	output.Gray("  Try these commands:")
	output.Gray("  $ tcl doctor")
	output.Gray("  $ tcl device list")
	fmt.Println()

	return nil
}

func getQRCode() (*qrResponse, error) {
	u, err := url.Parse(config.AuthURL + "/auth/common/v2/getQRCodePic")
	if err != nil {
		return nil, fmt.Errorf("failed to parse QR code URL: %w", err)
	}
	q := u.Query()
	q.Set("appId", config.AppID)
	q.Set("width", qrCodeSize)
	q.Set("height", qrCodeSize)
	q.Set("to", "qrlogin")
	q.Set("deviceName", "tcl-cli")
	q.Set("from", fmt.Sprintf("tcl-cli_%d", time.Now().UnixMilli()))
	u.RawQuery = q.Encode()

	resp, err := initHTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	const maxResponseSize = 1 * 1024 * 1024 // 1MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read QR code response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("QR code request returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result qrResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Status != 1 {
		msg := result.Msg
		if msg == "" {
			msg = fmt.Sprintf("unknown error (status: %d)", result.Status)
		}
		return nil, fmt.Errorf("%s", msg)
	}
	if result.Data.Code == "" {
		return nil, fmt.Errorf("QR code response missing code")
	}
	return &result, nil
}

func checkQRCodeAuth(code string) (*authCheckResponse, error) {
	u, err := url.Parse(config.AuthURL + "/auth/common/v2/checkQRCodeAuthor")
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth check URL: %w", err)
	}
	q := u.Query()
	q.Set("code", code)
	u.RawQuery = q.Encode()

	resp, err := initHTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	const maxAuthResponseSize = 1 * 1024 * 1024 // 1MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAuthResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read auth check response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("auth check returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result authCheckResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func init() {
	initCmd.Flags().String("verify", "", "Verify authorization using QR code from 'tcl init'")
	rootCmd.AddCommand(initCmd)
}
