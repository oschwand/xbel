package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print xbel version and build information",
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	// Cobra prints "xbel version <Version>" for --version, matching the
	// output of the version command.
	info, _ := debug.ReadBuildInfo()
	rootCmd.Version = buildVersion(info)
}

func runVersion(cmd *cobra.Command, args []string) error {
	// The ok flag is false when the binary carries no build metadata; the
	// returned info is then a zero value rendered as "(unknown)".
	info, _ := debug.ReadBuildInfo()
	fmt.Fprintln(cmd.OutOrStdout(), versionString(info))
	return nil
}

// versionString renders the build information: the module version stamped
// by `go install`, otherwise the VCS revision of a local `go build`. A nil
// or zero info means the binary carries no build metadata.
func versionString(info *debug.BuildInfo) string {
	return "xbel version " + buildVersion(info)
}

func buildVersion(info *debug.BuildInfo) string {
	if info == nil || info.GoVersion == "" {
		return "(unknown)"
	}
	version := info.Main.Version
	if version == "" || version == "(devel)" {
		version = develVersion(info)
	}
	return fmt.Sprintf("%s (%s)", version, info.GoVersion)
}

// develVersion describes a locally built binary from the VCS settings the
// toolchain stamps in when building inside a repository. It falls back to
// the plain "(devel)" of a build outside version control.
func develVersion(info *debug.BuildInfo) string {
	rev := buildSetting(info, "vcs.revision")
	if rev == "" {
		return "(devel)"
	}
	version := "devel+" + rev[:min(7, len(rev))]
	if buildSetting(info, "vcs.modified") == "true" {
		version += "-dirty"
	}
	return version
}

func buildSetting(info *debug.BuildInfo, key string) string {
	for _, s := range info.Settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}
