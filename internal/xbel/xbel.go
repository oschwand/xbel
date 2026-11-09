// Package xbel parses and manipulates XBEL (XML Bookmark Exchange
// Language) documents.
package xbel

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Document is the root of a parsed XBEL file.
type Document struct {
	XMLName   xml.Name   `xml:"xbel"`
	Version   string     `xml:"version,attr"`
	Folders   []Folder   `xml:"folder"`
	Bookmarks []Bookmark `xml:"bookmark"`
}

// Folder is an XBEL folder. It can hold bookmarks and nest other folders.
type Folder struct {
	Title     string     `xml:"title"`
	Folders   []Folder   `xml:"folder"`
	Bookmarks []Bookmark `xml:"bookmark"`
}

// Bookmark is a single XBEL bookmark.
type Bookmark struct {
	Href  string `xml:"href,attr"`
	Title string `xml:"title"`
}

// Parse reads an XBEL document from r. It returns an error if the input
// is not well-formed XML or if its root element is not <xbel>.
func Parse(r io.Reader) (*Document, error) {
	var doc Document
	err := xml.NewDecoder(r).Decode(&doc)
	if errors.Is(err, io.EOF) {
		return nil, errors.New("empty XBEL document")
	}
	if err != nil {
		return nil, fmt.Errorf("invalid XBEL: %w", err)
	}
	return &doc, nil
}

// SplitPath splits an XBEL folder path into its components, ignoring
// empty components so that "", "/", and "a//b" behave sensibly.
func SplitPath(path string) []string {
	parts := strings.Split(path, "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// FindFolder resolves a slash-separated path of folder titles (e.g.
// "Bookmarks Menu/Development") relative to the document root. The
// empty path denotes the root itself and returns a nil folder.
// Components match folder titles exactly, with a fallback to a unique
// case-insensitive match.
func (d *Document) FindFolder(path string) (*Folder, error) {
	var f *Folder
	for _, name := range SplitPath(path) {
		folders := d.Folders
		if f != nil {
			folders = f.Folders
		}
		next, err := findSubfolder(folders, name)
		if err != nil {
			return nil, err
		}
		f = next
	}
	return f, nil
}

func findSubfolder(folders []Folder, name string) (*Folder, error) {
	var matches []*Folder
	for i := range folders {
		title := strings.TrimSpace(folders[i].Title)
		if title == name {
			return &folders[i], nil
		}
		if strings.EqualFold(title, name) {
			matches = append(matches, &folders[i])
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("no folder named %q (%s)", name, describeSubfolders(folders))
	default:
		return nil, fmt.Errorf("ambiguous folder %q: %d folders match case-insensitively", name, len(matches))
	}
}

func describeSubfolders(folders []Folder) string {
	if len(folders) == 0 {
		return "no subfolders"
	}
	names := make([]string, len(folders))
	for i := range folders {
		names[i] = strings.TrimSpace(folders[i].Title)
	}
	return "available: " + strings.Join(names, ", ")
}

// CountBookmarks returns the number of bookmarks in the folder and in
// all of its subfolders.
func (f *Folder) CountBookmarks() int {
	n := len(f.Bookmarks)
	for i := range f.Folders {
		n += f.Folders[i].CountBookmarks()
	}
	return n
}

// CountFolders returns the number of subfolders in the folder, counted
// recursively.
func (f *Folder) CountFolders() int {
	n := len(f.Folders)
	for i := range f.Folders {
		n += f.Folders[i].CountFolders()
	}
	return n
}

// CountBookmarks returns the total number of bookmarks in the document.
func (d *Document) CountBookmarks() int {
	n := len(d.Bookmarks)
	for i := range d.Folders {
		n += d.Folders[i].CountBookmarks()
	}
	return n
}

// CountFolders returns the total number of folders in the document,
// counted recursively.
func (d *Document) CountFolders() int {
	n := len(d.Folders)
	for i := range d.Folders {
		n += d.Folders[i].CountFolders()
	}
	return n
}

// Walk calls fn for every bookmark in the document, with the folder
// path leading to it. The path is empty for top-level bookmarks.
func (d *Document) Walk(fn func(path []string, bm Bookmark)) {
	var walk func(folders []Folder, prefix []string)
	walk = func(folders []Folder, prefix []string) {
		for i := range folders {
			f := &folders[i]
			path := append(append([]string{}, prefix...), strings.TrimSpace(f.Title))
			for j := range f.Bookmarks {
				fn(path, f.Bookmarks[j])
			}
			walk(f.Folders, path)
		}
	}
	walk(d.Folders, nil)
	for i := range d.Bookmarks {
		fn(nil, d.Bookmarks[i])
	}
}

// WalkFolders calls fn for every folder in the document, with the path
// leading to it, its own title included.
func (d *Document) WalkFolders(fn func(path []string, f *Folder)) {
	var walk func(folders []Folder, prefix []string)
	walk = func(folders []Folder, prefix []string) {
		for i := range folders {
			path := append(append([]string{}, prefix...), strings.TrimSpace(folders[i].Title))
			fn(path, &folders[i])
			walk(folders[i].Folders, path)
		}
	}
	walk(d.Folders, nil)
}
