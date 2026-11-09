package cmd

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"
	"github.com/oschwand/xbel/internal/xbel"
	"github.com/spf13/cobra"
)

var (
	rootStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#61AFEF"))
	statsStyle  = lipgloss.NewStyle().Faint(true)
	folderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#56B6C2"))
	countStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379"))
	branchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370"))
	dimStyle    = lipgloss.NewStyle().Faint(true).Italic(true)
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display the folder tree of an XBEL document",
	Long:  "Display the folder tree of an XBEL document, with the number of bookmarks in each folder.",
	RunE:  runTree,
}

func init() {
	rootCmd.AddCommand(treeCmd)
}

func runTree(cmd *cobra.Command, args []string) error {
	reader, name, closeFn, err := openXbel()
	if err != nil {
		return err
	}
	defer closeFn()

	doc, err := xbel.Parse(reader)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), renderTree(doc, name))
	return nil
}

// renderTree renders the document folder tree, rooted at the file name.
func renderTree(doc *xbel.Document, name string) string {
	root := tree.Root(rootLabel(doc, name)).
		Enumerator(tree.RoundedEnumerator).
		EnumeratorStyle(branchStyle)

	for i := range doc.Folders {
		root.Child(folderTree(&doc.Folders[i]))
	}
	if n := len(doc.Bookmarks); n > 0 {
		root.Child(dimStyle.Render(plural(n, "top-level bookmark")))
	}
	return root.String()
}

func plural(n int, what string) string {
	if n == 1 {
		return "1 " + what
	}
	return fmt.Sprintf("%d %ss", n, what)
}

func rootLabel(doc *xbel.Document, name string) string {
	parts := []string{}
	if doc.Version != "" {
		parts = append(parts, "XBEL "+doc.Version)
	}
	parts = append(parts,
		plural(doc.CountFolders(), "folder"),
		plural(doc.CountBookmarks(), "link"),
	)
	label := rootStyle.Render("📑 " + name)
	label += statsStyle.Render("  " + strings.Join(parts, " · "))
	return label
}

// folderTree renders one folder and its subfolders. Folders with no
// children at all are marked as empty; folders holding bookmarks show
// the recursive bookmark count.
func folderTree(f *xbel.Folder) *tree.Tree {
	title := strings.TrimSpace(f.Title)
	label := folderStyle.Render("📁") + " "
	if title == "" {
		label += dimStyle.Render("(untitled)")
	} else {
		label += folderStyle.Render(title)
	}

	switch {
	case len(f.Folders) == 0 && len(f.Bookmarks) == 0:
		label += " " + dimStyle.Render("(empty)")
	case f.CountBookmarks() > 0:
		label += " " + countStyle.Render("· "+plural(f.CountBookmarks(), "link"))
	}

	t := tree.Root(label)
	for i := range f.Folders {
		t.Child(folderTree(&f.Folders[i]))
	}
	return t
}
