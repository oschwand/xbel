package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// defaultXbelFile is used when neither the --xbel flag nor the XBEL_FILE
// environment variable is set.
const defaultXbelFile = "bookmarks.xbel"

// envXbelFile names the environment variable holding the XBEL file path.
const envXbelFile = "XBEL_FILE"

var xbelFlag string

var rootCmd = &cobra.Command{
	Use:          "xbel",
	Short:        "xbel manages XBEL bookmark files",
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&xbelFlag, "xbel", "", "XBEL file to read (default $XBEL_FILE, or bookmarks.xbel in the current directory)")
}

// Execute runs the xbel command line.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// openXbel opens the XBEL file to operate on: the --xbel flag if set, then
// the XBEL_FILE environment variable, then bookmarks.xbel in the current
// directory when it exists.
func openXbel() (io.Reader, string, func(), error) {
	path := xbelFlag
	if path == "" {
		path = os.Getenv(envXbelFile)
	}
	if path == "" {
		if _, err := os.Stat(defaultXbelFile); err != nil {
			return nil, "", nil, fmt.Errorf("no XBEL file: pass --xbel, set %s, or place %s in the current directory", envXbelFile, defaultXbelFile)
		}
		path = defaultXbelFile
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, "", nil, err
	}
	return f, filepath.Base(path), func() { f.Close() }, nil
}
