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
	matches := searchBookmarks(searchDoc(t), "go", searchFields{})
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
	matches := searchBookmarks(searchDoc(t), "go", searchFields{title: true})
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
	matches := searchBookmarks(searchDoc(t), "go", searchFields{url: true})
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
	matches := searchBookmarks(searchDoc(t), "LWN", searchFields{})
	if len(matches) != 1 || matches[0].bm.Title != "LWN Kernel News" {
		t.Fatalf("query LWN should match exactly one bookmark: %+v", matches)
	}
}

func TestSearchFlagsSelectFields(t *testing.T) {
	// "toolbar" only appears in the Toolbar folder name: it matches the
	// default bookmark search but not --title or --url.
	if got := searchBookmarks(searchDoc(t), "toolbar", searchFields{}); len(got) != 2 {
		t.Errorf("default search should match folder names, got %d matches", len(got))
	}
	if got := searchBookmarks(searchDoc(t), "toolbar", searchFields{title: true}); len(got) != 0 {
		t.Errorf("--title should not match folder names, got %d matches", len(got))
	}
	if got := searchBookmarks(searchDoc(t), "toolbar", searchFields{url: true}); len(got) != 0 {
		t.Errorf("--url should not match folder names, got %d matches", len(got))
	}
}

func TestSearchFolderFlag(t *testing.T) {
	matches, folders := searchAll(searchDoc(t), "news", searchFields{folder: true})
	if len(matches) != 0 {
		t.Errorf("--folder should not list bookmarks, got %d", len(matches))
	}
	if len(folders) != 1 {
		t.Fatalf("--folder news should match the News folder, got %+v", folders)
	}
	if got := strings.Join(folders[0].path, "/"); got != "Toolbar/News" {
		t.Errorf("folder path = %q, want Toolbar/News", got)
	}
	if got := folders[0].f.CountBookmarks(); got != 1 {
		t.Errorf("News folder should hold 1 link, got %d", got)
	}
}

func TestSearchFolderFlagCaseInsensitive(t *testing.T) {
	if _, folders := searchAll(searchDoc(t), "TOOLBAR", searchFields{folder: true}); len(folders) != 1 {
		t.Errorf("--folder TOOLBAR should match the Toolbar folder, got %+v", folders)
	}
}

func TestSearchFolderFlagWithBookmarkFields(t *testing.T) {
	matches, folders := searchAll(searchDoc(t), "go", searchFields{folder: true, title: true})
	if len(folders) != 0 {
		t.Errorf("no folder named go, got %+v", folders)
	}
	if len(matches) != 2 {
		t.Errorf("--folder --title should still search titles, got %d matches", len(matches))
	}
}

func TestSearchFlagUnion(t *testing.T) {
	// --title --url selects those two fields; folder names are left
	// out, so a query that only appears in a folder name finds nothing.
	fields := searchFields{title: true, url: true}
	if got := len(searchBookmarks(searchDoc(t), "go", fields)); got != 2 {
		t.Errorf("--title --url should search titles and URLs, got %d matches", got)
	}
	if got := len(searchBookmarks(searchDoc(t), "toolbar", fields)); got != 0 {
		t.Errorf("--title --url should not match folder names, got %d matches", got)
	}
}

func TestSearchFolderPath(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "news", searchFields{})
	if len(matches) != 1 || matches[0].bm.Title != "LWN Kernel News" {
		t.Fatalf("query news should match via the News folder name: %+v", matches)
	}
}

func TestSearchFolderPathCoversWholeFolder(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "toolbar", searchFields{})
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
	if got := searchBookmarks(searchDoc(t), "toolbar", searchFields{title: true}); len(got) != 0 {
		t.Errorf("--title should not match folder names, got %d matches", len(got))
	}
}

func TestRenderSearch(t *testing.T) {
	matches := searchBookmarks(searchDoc(t), "lwn", searchFields{})
	out := plain(renderSearch("bookmarks.xbel", matches, nil))
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

func TestRenderSearchFolder(t *testing.T) {
	_, folders := searchAll(searchDoc(t), "news", searchFields{folder: true})
	out := plain(renderSearch("bookmarks.xbel", nil, folders))
	for _, want := range []string{
		"1 match for query",
		"Toolbar › 📁 News",
		"1 link",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("folder search output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderSearchEmpty(t *testing.T) {
	if out := plain(renderSearch("bookmarks.xbel", nil, nil)); !strings.Contains(out, "0 matches for query") || !strings.Contains(out, "no match") {
		t.Errorf("empty search should report zero matches:\n%s", out)
	}
}
