package cmd

import (
	"strings"
	"testing"

	"github.com/oschwand/xbel/internal/xbel"
)

func searchDoc(t *testing.T) *xbel.Document {
	t.Helper()
	doc, err := xbel.Parse(strings.NewReader(`<?xml version="1.0"?>
	<xbel version="1.0">
		<folder><title>Toolbar</title>
			<bookmark href="https://go.dev"><title>Go</title></bookmark>
			<folder><title>News</title>
				<bookmark href="https://lwn.net"><title>LWN Kernel News</title></bookmark>
			</folder>
		</folder>
		<bookmark href="https://blog.golang.org/rss"><title>Go Blog</title></bookmark>
		<bookmark href="https://example.com"><title></title></bookmark>
	</xbel>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return doc
}

func TestSearchBothFields(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "go", false, false)
	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2: %+v", len(matches), matches)
	}
	if got := matches[0].bm.Title; got != "Go" {
		t.Errorf("first match title = %q, want Go", got)
	}
	if got := strings.Join(matches[0].path, "/"); got != "Toolbar" {
		t.Errorf("first match path = %q, want Toolbar", got)
	}
}

func TestSearchTitleOnly(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "go", true, false)
	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2 (Go, Go Blog): %+v", len(matches), matches)
	}
	for _, m := range matches {
		if !strings.Contains(strings.ToLower(m.bm.Title), "go") {
			t.Errorf("match %q does not carry the query in its title", m.bm.Title)
		}
	}
}

func TestSearchURLOnly(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "go", false, true)
	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2 (go.dev, blog.golang.org): %+v", len(matches), matches)
	}
	for _, m := range matches {
		if !strings.Contains(strings.ToLower(m.bm.Href), "go") {
			t.Errorf("match %q does not carry the query in its URL", m.bm.Href)
		}
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "LWN", false, false)
	if len(matches) != 1 || matches[0].bm.Title != "LWN Kernel News" {
		t.Fatalf("query LWN should match exactly one bookmark: %+v", matches)
	}
}

func TestSearchBothFlagsSearchBothFields(t *testing.T) {
	if got := len(searchBookmarks(searchDoc(t), "go", true, true)); got != 2 {
		t.Errorf("--title --url should behave like the default, got %d matches", got)
	}
}

func TestSearchFolderPath(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "news", false, false)
	if len(matches) != 1 || matches[0].bm.Title != "LWN Kernel News" {
		t.Fatalf("query news should match via the News folder name: %+v", matches)
	}
}

func TestSearchFolderPathCoversWholeFolder(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "toolbar", false, false)
	if len(matches) != 2 {
		t.Fatalf("query toolbar should match both bookmarks under Toolbar: %+v", matches)
	}
	for _, m := range matches {
		if m.path[0] != "Toolbar" {
			t.Errorf("match %q is not under Toolbar: %v", m.bm.Title, m.path)
		}
	}
}

func TestSearchTitleOnlyIgnoresPath(t *testing.T) {
	if got := searchBookmarks(searchDoc(t), "toolbar", true, false); len(got) != 0 {
		t.Errorf("--title should not match folder names, got %d matches", len(got))
	}
}

func TestRenderSearch(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "lwn", false, false)
	out := plain(renderSearch("bookmarks.xbel", matches))
	for _, want := range []string{
		"📑 bookmarks.xbel",
		"1 match for query",
		"Toolbar › News › LWN Kernel News",
		"https://lwn.net",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("search output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderSearchEmpty(t *testing.T) {
	if out := plain(renderSearch("bookmarks.xbel", nil)); !strings.Contains(out, "0 matches for query") || !strings.Contains(out, "no match") {
		t.Errorf("empty search should report zero matches:\n%s", out)
	}
}
