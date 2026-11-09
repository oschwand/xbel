package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/oschwand/xbel/internal/xbel"
	"github.com/spf13/cobra"
)

var (
	linkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B"))
	urlStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370"))
	pathStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#56B6C2"))
)

var lsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "List folders and bookmarks in a directory",
	Long: `List the folders and bookmarks in a directory of an XBEL document.

The directory is given as a slash-separated path of folder titles,
relative to the document root (e.g. "Bookmarks Menu/Development").
With no path, the root directory is listed.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLs,
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

func runLs(cmd *cobra.Command, args []string) error {
	reader, name, closeFn, err := openXbel()
	if err != nil {
		return err
	}
	defer closeFn()

	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	doc, err := xbel.Parse(reader)
	if err != nil {
		return err
	}
	folder, err := doc.FindFolder(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, renderLs(doc, name, folder, xbel.SplitPath(path), lsWidth(out)))
	return nil
}

// lsWidth returns the available line width for rendering, querying the
// terminal when possible and defaulting to 80 columns otherwise.
func lsWidth(w io.Writer) int {
	if f, ok := w.(*os.File); ok {
		if width, _, err := term.GetSize(f.Fd()); err == nil && width > 0 {
			return width
		}
	}
	return 80
}

// renderLs renders the contents of a folder (or the document root when
// folder is nil), preceded by a breadcrumb header. Each bookmark is
// rendered as "title url" on a single line, unless that would not fit
// in width for every entry, in which case every URL moves to its own
// line under the title.
func renderLs(doc *xbel.Document, name string, folder *xbel.Folder, crumb []string, width int) string {
	var folders []*xbel.Folder
	var bookmarks []xbel.Bookmark
	if folder == nil {
		for i := range doc.Folders {
			folders = append(folders, &doc.Folders[i])
		}
		bookmarks = doc.Bookmarks
	} else {
		for i := range folder.Folders {
			folders = append(folders, &folder.Folders[i])
		}
		bookmarks = folder.Bookmarks
	}

	var b strings.Builder
	b.WriteString(lsHeader(name, crumb))
	b.WriteString("\n")
	b.WriteString(statsStyle.Render(fmt.Sprintf("  %s · %s",
		plural(len(folders), "folder"),
		plural(len(bookmarks), "bookmark"),
	)))
	b.WriteString("\n\n")

	if len(folders) == 0 && len(bookmarks) == 0 {
		b.WriteString(dimStyle.Render("empty directory"))
		b.WriteString("\n")
		return b.String()
	}

	labels := make([]string, 0, len(folders)+len(bookmarks))
	for _, f := range folders {
		labels = append(labels, folderLabel(f))
	}
	for _, bm := range bookmarks {
		labels = append(labels, bookmarkLabel(bm))
	}
	colw := 0
	for _, l := range labels {
		if w := lipgloss.Width(l); w > colw {
			colw = w
		}
	}

	compact := true
	for _, bm := range bookmarks {
		if href := bookmarkURL(bm); href != "" && lipgloss.Width(bookmarkPrefix)+colw+1+lipgloss.Width(href) > width {
			compact = false
			break
		}
	}

	for _, f := range folders {
		line := folderStyle.Render("📁 ") + folderStyle.Render(padRight(folderLabel(f), colw))
		if len(f.Folders) == 0 && len(f.Bookmarks) == 0 {
			line += " " + dimStyle.Render("(empty)")
		} else if n := f.CountBookmarks(); n > 0 {
			line += " " + countStyle.Render("· "+plural(n, "link"))
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	for _, bm := range bookmarks {
		label := bookmarkLabel(bm)
		if compact {
			line := linkStyle.Render(bookmarkPrefix) + linkStyle.Render(padRight(label, colw))
			if href := bookmarkURL(bm); href != "" {
				line += " " + urlStyle.Render(href)
			}
			b.WriteString(line)
		} else {
			b.WriteString(linkStyle.Render(bookmarkPrefix) + linkStyle.Render(label))
			if href := bookmarkURL(bm); href != "" {
				b.WriteString("\n" + urlStyle.Render(strings.Repeat(" ", lipgloss.Width(bookmarkPrefix))+href))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func lsHeader(name string, crumb []string) string {
	var b strings.Builder
	b.WriteString(rootStyle.Render("📑 " + name))
	for _, c := range crumb {
		b.WriteString(dimStyle.Render(" › "))
		b.WriteString(pathStyle.Render(c))
	}
	return b.String()
}

// folderLabel returns the display title of a folder.
func folderLabel(f *xbel.Folder) string {
	if t := strings.TrimSpace(f.Title); t != "" {
		return t
	}
	return "(untitled)"
}

// bookmarkPrefix is the icon shown before every bookmark entry.
const bookmarkPrefix = "🔗 "

// bookmarkLabel returns the display title of a bookmark, falling back
// to its URL when it has no title.
func bookmarkLabel(bm xbel.Bookmark) string {
	if t := strings.TrimSpace(bm.Title); t != "" {
		return t
	}
	return strings.TrimSpace(bm.Href)
}

// bookmarkURL returns the URL to display next to a bookmark title, or
// "" when it would not add information (no URL, or the URL is already
// shown as the label of a title-less bookmark).
func bookmarkURL(bm xbel.Bookmark) string {
	href := strings.TrimSpace(bm.Href)
	if href == "" || href == bookmarkLabel(bm) {
		return ""
	}
	return href
}

func padRight(s string, width int) string {
	if d := width - lipgloss.Width(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}
