package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tcl-eaglelab/tcl-smart-home-cli/internal/api"
	"github.com/tcl-eaglelab/tcl-smart-home-cli/internal/config"
	"github.com/tcl-eaglelab/tcl-smart-home-cli/internal/output"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check CLI configuration and connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		output.Bold("  TCL CLI Doctor")
		output.Gray("  ───────────────")
		fmt.Println()

		cfg, err := config.Load()
		if err != nil || cfg.AccessToken == "" {
			output.Error("Config file: not found or empty (" + config.FilePath() + ")")
			output.Gray("\n  Run \"tcl init\" to set up credentials.\n")
			return errSilent
		}

		output.Success("Config file: " + config.FilePath())

		token := cfg.AccessToken
		const maskPrefixLen, maskSuffixLen = 10, 4
		var masked string
		if len(token) > maskPrefixLen+maskSuffixLen {
			masked = token[:maskPrefixLen] + "****" + token[len(token)-maskSuffixLen:]
		} else {
			masked = "****"
		}
		output.Success("Access Token: " + masked)
		output.Success("IoT URL: " + config.IoTURL)

		now := time.Now().UnixMilli()
		daysLeft := (cfg.TokenExpires - now) / config.MillisPerDay
		if cfg.TokenExpires > now {
			if daysLeft == 0 {
				output.Success("Token expires today")
			} else {
				output.Success(fmt.Sprintf("Token expires in %d day(s)", daysLeft))
			}
		} else {
			expiredDays := (now - cfg.TokenExpires) / config.MillisPerDay
			if expiredDays == 0 {
				output.Error("Token expired today. Run \"tcl init\" to re-authorize.")
			} else {
				output.Error(fmt.Sprintf("Token expired %d day(s) ago. Run \"tcl init\" to re-authorize.", expiredDays))
			}
			return errSilent
		}

		client, err := api.NewClient()
		if err != nil {
			output.Error("API client: " + err.Error())
			return errSilent
		}

		devices, err := client.GetAllDevices()
		if err != nil {
			output.Error("API connection: Failed — " + err.Error())
			return errSilent
		}

		output.Success(fmt.Sprintf("API connection: OK (%d device(s) found)", len(devices)))
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
