package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/oschwand/xbel/internal/xbel"
	"github.com/spf13/cobra"
)

var (
	linkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B"))
	urlStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370"))
	pathStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#56B6C2"))
)

var lsCmd = &cobra.Command{
	Use:   "ls [file] [path]",
	Short: "List folders and bookmarks in a directory",
	Long: `List the folders and bookmarks in a directory of an XBEL document.

The directory is given as a slash-separated path of folder titles,
relative to the document root (e.g. "Bookmarks Menu/Development").
With no path, the root directory is listed. With no file, XBEL is read
from stdin ("-" also means stdin).`,
	Args: cobra.MaximumNArgs(2),
	RunE: runLs,
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

func runLs(cmd *cobra.Command, args []string) error {
	reader, name, closeFn, path, err := lsInput(args)
	if err != nil {
		return err
	}
	defer closeFn()

	doc, err := xbel.Parse(reader)
	if err != nil {
		return err
	}
	folder, err := doc.FindFolder(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), renderLs(doc, name, folder, xbel.SplitPath(path)))
	return nil
}

// lsInput resolves the file and directory path arguments. The first
// argument is the XBEL file (or "-" for piped stdin), the second the
// folder path to list.
func lsInput(args []string) (io.Reader, string, func(), string, error) {
	file, path := "", ""
	if len(args) > 0 {
		file = args[0]
	}
	if len(args) > 1 {
		path = args[1]
	}

	if file == "" || file == "-" {
		if st, err := os.Stdin.Stat(); err == nil && st.Mode()&os.ModeCharDevice == 0 {
			return os.Stdin, "stdin", func() {}, path, nil
		}
		return nil, "", nil, path, fmt.Errorf("provide an XBEL file, or pipe one on stdin")
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, "", nil, path, err
	}
	return f, filepath.Base(file), func() { f.Close() }, path, nil
}

// renderLs renders the contents of a folder (or the document root when
// folder is nil), preceded by a breadcrumb header.
func renderLs(doc *xbel.Document, name string, folder *xbel.Folder, crumb []string) string {
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
	width := 0
	for _, l := range labels {
		if w := lipgloss.Width(l); w > width {
			width = w
		}
	}

	for _, f := range folders {
		line := folderStyle.Render("📁 ") + folderStyle.Render(padRight(folderLabel(f), width))
		if len(f.Folders) == 0 && len(f.Bookmarks) == 0 {
			line += " " + dimStyle.Render("(empty)")
		} else if n := f.CountBookmarks(); n > 0 {
			line += " " + countStyle.Render("· "+plural(n, "link"))
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	for _, bm := range bookmarks {
		line := linkStyle.Render("🔗 ") + linkStyle.Render(padRight(bookmarkLabel(bm), width))
		if href := strings.TrimSpace(bm.Href); href != "" && href != bookmarkLabel(bm) {
			line += " " + urlStyle.Render(href)
		}
		b.WriteString(line)
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

// bookmarkLabel returns the display title of a bookmark, falling back
// to its URL when it has no title.
func bookmarkLabel(bm xbel.Bookmark) string {
	if t := strings.TrimSpace(bm.Title); t != "" {
		return t
	}
	return strings.TrimSpace(bm.Href)
}

func padRight(s string, width int) string {
	if d := width - lipgloss.Width(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}
