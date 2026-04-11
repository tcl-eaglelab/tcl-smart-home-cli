package main

import (
	"fmt"
	"os"

	"github.com/tcl-eaglelab/tcl-smart-home-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if err.Error() != "" {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
