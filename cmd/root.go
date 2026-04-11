package cmd

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/tcl-eaglelab/tcl-smart-home-cli/internal/output"
)

// errSilent indicates the error has already been printed to the user.
var errSilent = errors.New("")

var version = "dev"

var rootCmd = &cobra.Command{
	Use:           "tcl",
	Short:         "The official CLI for TCL Smart Home",
	Version:       version,
	SilenceErrors: true,
	SilenceUsage:  true,
	Long: fmt.Sprintf("\n  %s\n  %s",
		output.FormatCyanBold("TCL Smart Home CLI"),
		output.FormatGray("Manage your TCL smart home devices from the terminal")),
}

func init() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			version = info.Main.Version
		}
	}
	rootCmd.Version = version
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func Execute() error {
	return rootCmd.Execute()
}
