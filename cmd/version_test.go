package cmd

import (
	"runtime/debug"
	"testing"
)

func TestVersionString(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		want string
	}{
		{
			name: "nil info",
			info: nil,
			want: "xbel version (unknown)",
		},
		{
			name: "zero info",
			info: &debug.BuildInfo{},
			want: "xbel version (unknown)",
		},
		{
			name: "go install module version",
			info: &debug.BuildInfo{
				Main:      debug.Module{Version: "v1.2.3"},
				GoVersion: "go1.27.3",
			},
			want: "xbel version v1.2.3 (go1.27.3)",
		},
		{
			name: "devel build with vcs revision",
			info: &debug.BuildInfo{
				Main:      debug.Module{Version: "(devel)"},
				GoVersion: "go1.27.3",
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "f1a0c48deadbeef"},
				},
			},
			want: "xbel version devel+f1a0c48 (go1.27.3)",
		},
		{
			name: "devel build with dirty worktree",
			info: &debug.BuildInfo{
				Main:      debug.Module{Version: "(devel)"},
				GoVersion: "go1.27.3",
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "f1a0c48deadbeef"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			want: "xbel version devel+f1a0c48-dirty (go1.27.3)",
		},
		{
			name: "devel build without vcs",
			info: &debug.BuildInfo{
				Main:      debug.Module{Version: "(devel)"},
				GoVersion: "go1.27.3",
			},
			want: "xbel version (devel) (go1.27.3)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := versionString(tt.info); got != tt.want {
				t.Errorf("versionString() = %q, want %q", got, tt.want)
			}
		})
	}
}
