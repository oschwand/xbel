package cmd

import (
	"strings"
	"testing"

	ansi "github.com/charmbracelet/x/ansi"
	"github.com/oschwand/xbel/internal/xbel"
)

// plain strips ANSI styling so assertions match regardless of the
// detected color profile.
func plain(s string) string { return ansi.Strip(s) }

func lsDoc(t *testing.T) *xbel.Document {
	t.Helper()
	doc, err := xbel.Parse(strings.NewReader(`<?xml version="1.0"?>
	<xbel version="1.0">
		<folder><title>Toolbar</title>
			<bookmark href="https://go.dev"><title>Go</title></bookmark>
			<folder><title>Empty folder</title></folder>
		</folder>
		<folder><title>Bookmarks Menu</title></folder>
		<bookmark href="https://example.com"><title>Example</title></bookmark>
		<bookmark href="https://notitle.example"><title></title></bookmark>
	</xbel>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return doc
}

func TestRenderLsRoot(t *testing.T) {
	out := plain(renderLs(lsDoc(t), "bookmarks.xbel", nil, nil))
	for _, want := range []string{
		"📑 bookmarks.xbel",
		"2 folders · 2 bookmarks",
		"Toolbar",
		"Bookmarks Menu",
		"Example",
		"https://example.com",
		"https://notitle.example",
		"1 link",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("root listing missing %q:\n%s", want, out)
		}
	}
}

func TestRenderLsFolder(t *testing.T) {
	doc := lsDoc(t)
	folder, err := doc.FindFolder("Toolbar")
	if err != nil {
		t.Fatalf("FindFolder: %v", err)
	}
	out := plain(renderLs(doc, "bookmarks.xbel", folder, []string{"Toolbar"}))
	for _, want := range []string{
		"bookmarks.xbel › Toolbar",
		"1 folder · 1 bookmark",
		"Go",
		"https://go.dev",
		"Empty folder",
		"(empty)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("folder listing missing %q:\n%s", want, out)
		}
	}
}

func TestRenderLsEmpty(t *testing.T) {
	doc := lsDoc(t)
	folder, err := doc.FindFolder("Bookmarks Menu")
	if err != nil {
		t.Fatalf("FindFolder: %v", err)
	}
	if out := plain(renderLs(doc, "bookmarks.xbel", folder, []string{"Bookmarks Menu"})); !strings.Contains(out, "empty directory") {
		t.Errorf("empty folder listing should say so:\n%s", out)
	}
}
