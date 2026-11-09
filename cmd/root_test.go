package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// withCwd runs f in dir, restoring the previous working directory after.
func withCwd(t *testing.T, dir string, f func()) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("Chdir back: %v", err)
		}
	}()
	f()
}

// resetInput restores the flag and environment variable after a test.
func resetInput() {
	xbelFlag = ""
	os.Unsetenv(envXbelFile)
}

func TestOpenXbelFlag(t *testing.T) {
	defer resetInput()
	xbelFlag = filepath.Join("..", "testdata", "bookmarks.xbel")
	_, name, closeFn, err := openXbel()
	if err != nil {
		t.Fatalf("openXbel: %v", err)
	}
	defer closeFn()
	if name != "bookmarks.xbel" {
		t.Errorf("name = %q, want bookmarks.xbel", name)
	}
}

func TestOpenXbelFlagWinsOverEnvAndDefault(t *testing.T) {
	defer resetInput()
	withCwd(t, filepath.Join("..", "testdata"), func() {
		t.Setenv(envXbelFile, "/nonexistent-from-env.xbel")
		xbelFlag = "bookmarks.xbel"
		if _, _, _, err := openXbel(); err != nil {
			t.Errorf("flag should win over env: %v", err)
		}
	})
}

func TestOpenXbelEnv(t *testing.T) {
	defer resetInput()
	t.Setenv(envXbelFile, filepath.Join("..", "testdata", "bookmarks.xbel"))
	if _, name, closeFn, err := openXbel(); err != nil {
		t.Fatalf("openXbel: %v", err)
	} else {
		defer closeFn()
		if name != "bookmarks.xbel" {
			t.Errorf("name = %q, want bookmarks.xbel", name)
		}
	}
}

func TestOpenXbelDefaultFile(t *testing.T) {
	defer resetInput()
	withCwd(t, filepath.Join("..", "testdata"), func() {
		if _, _, closeFn, err := openXbel(); err != nil {
			t.Fatalf("openXbel: %v", err)
		} else {
			defer closeFn()
		}
	})
}

func TestOpenXbelMissing(t *testing.T) {
	defer resetInput()
	withCwd(t, t.TempDir(), func() {
		_, _, _, err := openXbel()
		if err == nil {
			t.Fatal("expected an error when no file can be resolved")
		}
	})
}
