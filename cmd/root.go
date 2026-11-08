package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "xbel",
	Short:        "xbel manages XBEL bookmark files",
	SilenceUsage: true,
}

// Execute runs the xbel command line.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
