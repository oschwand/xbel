package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/oschwand/xbel/internal/xbel"
	"github.com/spf13/cobra"
)

var (
	searchTitle bool
	searchURL   bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search bookmarks by title or URL",
	Long: `Search bookmarks by title, URL, or folder name.

Bookmarks match when the query appears, case-insensitively, in their
title, their URL, or the name of a folder on their path. By default
all fields are searched; --title and --url restrict the search to one
field. Results show the folder path leading to each match.`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&searchTitle, "title", false, "search titles only")
	searchCmd.Flags().BoolVar(&searchURL, "url", false, "search URLs only")
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
	matches := searchBookmarks(doc, args[0], searchTitle, searchURL)
	fmt.Fprintln(cmd.OutOrStdout(), renderSearch(name, matches))
	return nil
}

// match is a bookmark together with the folder path leading to it.
type match struct {
	path []string
	bm   xbel.Bookmark
}

// searchBookmarks returns the bookmarks whose title, URL, or folder path
// contains the query, case-insensitively. With titleOnly the query must
// appear in the title, with urlOnly in the URL; both set is the same as
// neither.
func searchBookmarks(doc *xbel.Document, query string, titleOnly, urlOnly bool) []match {
	titleOnly = titleOnly && !urlOnly
	urlOnly = urlOnly && !titleOnly
	needle := strings.ToLower(query)
	contains := func(s string) bool { return strings.Contains(strings.ToLower(s), needle) }

	var matches []match
	doc.Walk(func(path []string, bm xbel.Bookmark) {
		switch {
		case titleOnly:
			if !contains(bm.Title) {
				return
			}
		case urlOnly:
			if !contains(bm.Href) {
				return
			}
		case !contains(bm.Title) && !contains(bm.Href) && !slices.ContainsFunc(path, contains):
			return
		}
		matches = append(matches, match{path: path, bm: bm})
	})
	return matches
}

// renderSearch renders the match count header and one line per match:
// folder path, bookmark title, URL.
func renderSearch(name string, matches []match) string {
	var b strings.Builder
	b.WriteString(rootStyle.Render("📑 " + name))
	b.WriteString(statsStyle.Render(fmt.Sprintf("  %d %s for query", len(matches), pluralWord(len(matches), "match", "matches"))))
	b.WriteString("\n\n")
	if len(matches) == 0 {
		b.WriteString(dimStyle.Render("no match"))
		b.WriteString("\n")
		return b.String()
	}
	for _, m := range matches {
		b.WriteString(matchLine(m))
		b.WriteString("\n")
	}
	return b.String()
}

// pluralWord returns the singular or plural form of a word for n.
func pluralWord(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

func matchLine(m match) string {
	var b strings.Builder
	for _, seg := range m.path {
		b.WriteString(pathStyle.Render(seg))
		b.WriteString(dimStyle.Render(" › "))
	}
	b.WriteString(linkStyle.Render(bookmarkLabel(m.bm)))
	if href := bookmarkURL(m.bm); href != "" {
		b.WriteString("  ")
		b.WriteString(urlStyle.Render(href))
	}
	return b.String()
}
