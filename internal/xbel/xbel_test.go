package xbel

import (
	"slices"
	"strings"
	"testing"
)

func TestSplitPath(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"/", nil},
		{"a//b", []string{"a", "b"}},
		{" Bookmarks Menu / Development ", []string{"Bookmarks Menu", "Development"}},
		{"a/", []string{"a"}},
	}
	for _, tt := range tests {
		if got := SplitPath(tt.in); !slices.Equal(got, tt.want) {
			t.Errorf("SplitPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func testDoc(t *testing.T) *Document {
	t.Helper()
	doc, err := Parse(strings.NewReader(`<?xml version="1.0"?>
	<xbel version="1.0">
		<folder><title>Toolbar</title>
			<folder><title>Go</title>
				<bookmark href="https://go.dev"><title>Go</title></bookmark>
			</folder>
		</folder>
		<folder><title>Bookmarks Menu</title>
			<bookmark href="https://example.com"><title>Example</title></bookmark>
		</folder>
		<bookmark href="https://root.example"><title>Root</title></bookmark>
	</xbel>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return doc
}

func TestFindFolder(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantNil bool
		want    string
		wantErr string
	}{
		{name: "root", path: "", wantNil: true},
		{name: "top level", path: "Toolbar", want: "Toolbar"},
		{name: "nested", path: "Toolbar/Go", want: "Go"},
		{name: "case-insensitive fallback", path: "toolbar/go", want: "Go"},
		{name: "missing component", path: "Nope", wantErr: `no folder named "Nope" (available: Toolbar, Bookmarks Menu)`},
		{name: "missing nested component", path: "Bookmarks Menu/Go", wantErr: `no folder named "Go" (no subfolders)`},
		{name: "trailing slash", path: "Toolbar/", want: "Toolbar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := testDoc(t).FindFolder(tt.path)
			switch {
			case tt.wantErr != "":
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("FindFolder(%q) error = %v, want %q", tt.path, err, tt.wantErr)
				}
			case err != nil:
				t.Fatalf("FindFolder(%q): %v", tt.path, err)
			case tt.wantNil && f != nil:
				t.Fatalf("FindFolder(%q) = %v, want nil", tt.path, f)
			case !tt.wantNil && f.Title != tt.want:
				t.Fatalf("FindFolder(%q).Title = %q, want %q", tt.path, f.Title, tt.want)
			}
		})
	}
}

func TestFindFolderAmbiguous(t *testing.T) {
	doc, err := Parse(strings.NewReader(`<?xml version="1.0"?>
	<xbel version="1.0">
		<folder><title>news</title></folder>
		<folder><title>News</title></folder>
	</xbel>`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := doc.FindFolder("NEWS"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("FindFolder(\"NEWS\") error = %v, want ambiguous", err)
	}
	if f, err := doc.FindFolder("News"); err != nil || f.Title != "News" {
		t.Fatalf("exact match should win: %v, %v", f, err)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *Document)
	}{
		{
			name: "nested folders and counts",
			input: `<?xml version="1.0" encoding="UTF-8"?>
			<xbel version="1.0">
				<folder>
					<title>Toolbar</title>
					<bookmark href="https://go.dev"><title>Go</title></bookmark>
					<folder>
						<title>News</title>
						<bookmark href="https://news.ycombinator.com"><title>HN</title></bookmark>
						<bookmark href="https://lwn.net"><title>LWN</title></bookmark>
					</folder>
				</folder>
				<bookmark href="https://example.com"><title>Example</title></bookmark>
			</xbel>`,
			check: func(t *testing.T, doc *Document) {
				if doc.Version != "1.0" {
					t.Errorf("Version = %q, want %q", doc.Version, "1.0")
				}
				if len(doc.Folders) != 1 {
					t.Fatalf("len(Folders) = %d, want 1", len(doc.Folders))
				}
				f := doc.Folders[0]
				if f.Title != "Toolbar" {
					t.Errorf("Folders[0].Title = %q, want %q", f.Title, "Toolbar")
				}
				if len(f.Folders) != 1 || f.Folders[0].Title != "News" {
					t.Errorf("nested folder News not found")
				}
				if got := doc.CountBookmarks(); got != 4 {
					t.Errorf("CountBookmarks = %d, want 4", got)
				}
				if got := doc.CountFolders(); got != 2 {
					t.Errorf("CountFolders = %d, want 2", got)
				}
				if got := f.CountBookmarks(); got != 3 {
					t.Errorf("folder CountBookmarks = %d, want 3", got)
				}
				if len(doc.Bookmarks) != 1 || doc.Bookmarks[0].Href != "https://example.com" {
					t.Errorf("top-level bookmark not parsed")
				}
			},
		},
		{
			name: "doctype and attributes",
			input: `<?xml version="1.0"?>
			<!DOCTYPE xbel PUBLIC "+//IDN python.org//DTD XML Bookmark Exchange Language 1.0//EN//XML" "http://pyxml.sourceforge.net/topics/dtds/xbel-1.0.dtd">
			<xbel version="1.0"><folder><title>Empty</title></folder></xbel>`,
			check: func(t *testing.T, doc *Document) {
				if len(doc.Folders) != 1 || doc.Folders[0].Title != "Empty" {
					t.Errorf("folder not parsed")
				}
				if got := doc.Folders[0].CountBookmarks(); got != 0 {
					t.Errorf("CountBookmarks = %d, want 0", got)
				}
			},
		},
		{
			name:    "wrong root element",
			input:   `<html><body>nope</body></html>`,
			wantErr: true,
		},
		{
			name:    "malformed xml",
			input:   `<xbel><folder><title>Broken`,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(t, doc)
			}
		})
	}
}
