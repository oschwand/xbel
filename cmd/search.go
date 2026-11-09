package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/oschwand/xbel/internal/xbel"
	"github.com/spf13/cobra"
)

var (
	searchTitle  bool
	searchURL    bool
	searchFolder bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search bookmarks by title or URL",
	Long: `Search bookmarks by title, URL, or folder name.

Bookmarks match when the query appears, case-insensitively, in their
title, their URL, or the name of a folder on their path. By default
all fields are searched and matching bookmarks are listed. --title
and --url restrict the search to those fields; --folder searches
folder names instead and lists the matching folders.`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&searchTitle, "title", false, "search titles only")
	searchCmd.Flags().BoolVar(&searchURL, "url", false, "search URLs only")
	searchCmd.Flags().BoolVar(&searchFolder, "folder", false, "search folder names only")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	reader, name, closeFn, err := openXbel()
	if err != nil {
		return err
	}
	defer closeFn()

	doc, err := xbel.Parse(reader)
	if err != nil {
		return err
	}
	matches, folders := searchAll(doc, args[0], searchFields{
		title:  searchTitle,
		url:    searchURL,
		folder: searchFolder,
	})
	fmt.Fprintln(cmd.OutOrStdout(), renderSearch(name, matches, folders))
	return nil
}

// match is a bookmark together with the folder path leading to it.
type match struct {
	path []string
	bm   xbel.Bookmark
}

// searchFields selects which fields of a bookmark are searched.
type searchFields struct {
	title  bool
	url    bool
	folder bool
}

// searchAll runs the query and returns the bookmark matches plus the
// folder matches. --folder searches folder names and returns folders;
// combined with --title or --url the bookmark fields are searched too.
// An empty selection searches every field.
func searchAll(doc *xbel.Document, query string, fields searchFields) ([]match, []folderMatch) {
	if fields.folder {
		folders := searchFolders(doc, query)
		if fields.title || fields.url {
			fields.folder = false
			return searchBookmarks(doc, query, fields), folders
		}
		return nil, folders
	}
	return searchBookmarks(doc, query, fields), nil
}

// searchBookmarks returns the bookmarks whose selected fields contain
// the query, case-insensitively. An empty selection searches every
// field: title, URL, and folder names on the path.
func searchBookmarks(doc *xbel.Document, query string, fields searchFields) []match {
	if !fields.title && !fields.url && !fields.folder {
		fields = searchFields{title: true, url: true, folder: true}
	}
	needle := strings.ToLower(query)
	contains := func(s string) bool { return strings.Contains(strings.ToLower(s), needle) }

	var matches []match
	doc.Walk(func(path []string, bm xbel.Bookmark) {
		if (fields.title && contains(bm.Title)) ||
			(fields.url && contains(bm.Href)) ||
			(fields.folder && slices.ContainsFunc(path, contains)) {
			matches = append(matches, match{path: path, bm: bm})
		}
	})
	return matches
}

// folderMatch is a folder together with its path.
type folderMatch struct {
	path []string
	f    *xbel.Folder
}

// searchFolders returns the folders whose own name contains the query,
// case-insensitively.
func searchFolders(doc *xbel.Document, query string) []folderMatch {
	needle := strings.ToLower(query)
	var folders []folderMatch
	doc.WalkFolders(func(path []string, f *xbel.Folder) {
		if strings.Contains(strings.ToLower(path[len(path)-1]), needle) {
			folders = append(folders, folderMatch{path: path, f: f})
		}
	})
	return folders
}

// renderSearch renders the match count header and one line per match:
// folders first, then bookmarks with their folder path, title, and URL.
func renderSearch(name string, matches []match, folders []folderMatch) string {
	total := len(matches) + len(folders)
	var b strings.Builder
	b.WriteString(rootStyle.Render("📑 " + name))
	b.WriteString(statsStyle.Render(fmt.Sprintf("  %d %s for query", total, pluralWord(total, "match", "matches"))))
	b.WriteString("\n\n")
	if total == 0 {
		b.WriteString(dimStyle.Render("no match"))
		b.WriteString("\n")
		return b.String()
	}
	for _, fm := range folders {
		b.WriteString(folderLine(fm))
		b.WriteString("\n")
	}
	for _, m := range matches {
		b.WriteString(matchLine(m))
		b.WriteString("\n")
	}
	return b.String()
}

// folderLine renders a folder match: its slash-separated path, the
// same format ls accepts, and its recursive bookmark count.
func folderLine(fm folderMatch) string {
	b := pathStyle.Render(strings.Join(fm.path, " / "))
	if n := fm.f.CountBookmarks(); n > 0 {
		b += " " + countStyle.Render("· "+plural(n, "link"))
	} else {
		b += " " + dimStyle.Render("(empty)")
	}
	return b
}

// pluralWord returns the singular or plural form of a word for n.
func pluralWord(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// matchLine renders a bookmark match: the slash-separated folder path,
// in the same format ls accepts, then the title and URL.
func matchLine(m match) string {
	var b strings.Builder
	if len(m.path) > 0 {
		b.WriteString(pathStyle.Render(strings.Join(m.path, " / ")))
		b.WriteString(dimStyle.Render(" › "))
	}
	b.WriteString(linkStyle.Render(bookmarkLabel(m.bm)))
	if href := bookmarkURL(m.bm); href != "" {
		b.WriteString("  ")
		b.WriteString(urlStyle.Render(href))
	}
	return b.String()
}
