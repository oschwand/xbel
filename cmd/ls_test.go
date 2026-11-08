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
	out := plain(renderLs(lsDoc(t), "bookmarks.xbel", nil, nil, 200))
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
	out := plain(renderLs(doc, "bookmarks.xbel", folder, []string{"Toolbar"}, 200))
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

func TestRenderLsCompact(t *testing.T) {
	doc := lsDoc(t)
	folder, err := doc.FindFolder("Toolbar")
	if err != nil {
		t.Fatalf("FindFolder: %v", err)
	}
	lines := splitLines(plain(renderLs(doc, "b.xbel", folder, nil, 200)))
	var goLine string
	for _, l := range lines {
		if strings.Contains(l, "Go") && !strings.Contains(l, "📁") {
			goLine = l
		}
	}
	if goLine == "" {
		t.Fatalf("bookmark line not found:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.Contains(goLine, "https://go.dev") {
		t.Errorf("title and url should share a line when everything fits: %q", goLine)
	}
}

func TestRenderLsMultiline(t *testing.T) {
	doc, err := xbel.Parse(strings.NewReader(`<?xml version="1.0"?>
	<xbel version="1.0">
		<bookmark href="https://long.example.com/some/long/path"><title>A fairly long title</title></bookmark>
	</xbel>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	lines := splitLines(plain(renderLs(doc, "b.xbel", nil, nil, 40)))
	if len(lines) < 3 {
		t.Fatalf("expected title and url on separate lines, got %q", lines)
	}
	if !strings.Contains(lines[3], "A fairly long title") || strings.Contains(lines[3], "https://") {
		t.Errorf("title line should not carry the url: %q", lines[3])
	}
	if !strings.HasPrefix(strings.TrimLeft(lines[4], " "), "https://long.example.com") || lines[4] == strings.TrimLeft(lines[4], " ") {
		t.Errorf("url should be on its own indented line: %q", lines[4])
	}
}

func splitLines(s string) []string {
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}

func TestRenderLsEmpty(t *testing.T) {
	doc := lsDoc(t)
	folder, err := doc.FindFolder("Bookmarks Menu")
	if err != nil {
		t.Fatalf("FindFolder: %v", err)
	}
	if out := plain(renderLs(doc, "bookmarks.xbel", folder, []string{"Bookmarks Menu"}, 200)); !strings.Contains(out, "empty directory") {
		t.Errorf("empty folder listing should say so:\n%s", out)
	}
}
