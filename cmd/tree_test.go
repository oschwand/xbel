package cmd

import (
	"strings"
	"testing"

	"github.com/oschwand/xbel/internal/xbel"
)

func TestRenderTree(t *testing.T) {
	doc, err := xbel.Parse(strings.NewReader(`<?xml version="1.0"?>
		<xbel version="1.0">
			<folder>
				<title>Toolbar</title>
				<bookmark href="https://go.dev"><title>Go</title></bookmark>
				<folder>
					<title>News</title>
					<bookmark href="https://lwn.net"><title>LWN</title></bookmark>
				</folder>
			</folder>
			<folder><title></title></folder>
			<bookmark href="https://example.com"><title>Example</title></bookmark>
		</xbel>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	out := renderTree(doc, "bookmarks.xbel")
	for _, want := range []string{
		"bookmarks.xbel",
		"Toolbar",
		"News",
		"(untitled)",
		"2 links",
		"1 top-level bookmark",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("renderTree output missing %q:\n%s", want, out)
		}
	}
}
