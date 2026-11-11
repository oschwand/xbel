---
name: xbel
description: Read and navigate XBEL bookmark files with the xbel CLI. Use when listing, browsing, or searching bookmarks and folders in an XBEL document — tree overview, directory listings, or search by title, URL, or folder name.
---

# xbel

CLI for XBEL (XML Bookmark Exchange Language) bookmark files. It reads
one XBEL document, validates it, and renders its contents in the
terminal.

### Which file is read

No file argument: the XBEL file is resolved in this order:

1. the `--xbel` flag: `xbel --xbel /path/to/bookmarks.xbel tree`;
2. the `XBEL_FILE` environment variable;
3. `bookmarks.xbel` in the current directory, when it exists.

```bash
xbel tree                          # uses ./bookmarks.xbel or $XBEL_FILE
xbel --xbel ~/bookmarks.xbel tree  # explicit file
```

### Overview: tree

```bash
xbel tree
```

Prints the folder tree with a bookmark count per folder and a summary
header (XBEL version, totals). Empty folders are marked `(empty)`,
untitled folders `(untitled)`.

### Browsing: ls

```bash
xbel ls                    # list the root directory
xbel ls "Bookmarks Menu / Development"
```

`path` is a slash-separated list of folder titles relative to the
document root. Spaces around the slashes are fine. Folder titles match
exactly first, then by unique case-insensitive match; errors list the
folder names available at that level. Folders come first (with
bookmark counts), then bookmarks; titles and URLs share a line when
everything fits the terminal width, otherwise each URL moves to its
own line.

### Finding things: search

```bash
xbel search charm             # titles, URLs, and folder names
xbel search --title charm     # titles only
xbel search --url charm       # URLs only
xbel search --folder charm    # folder names only
xbel search --folder --title charm   # union of the selected fields
```

Matching is a case-insensitive substring test. Bookmark results show
the folder path, the title, and the URL; `--folder` results show the
matching folders with their bookmark count.

### Copy-paste workflow: search into ls

Search result paths use the exact input format of `ls` — folder titles
joined by ` / `. Copy the path portion of any result line straight
into an `ls` call:

```bash
xbel search --folder charm
# Bookmarks Toolbar / Charm · 2 links
xbel ls "Bookmarks Toolbar / Charm"
```

### Version

```bash
xbel version      # or: xbel --version
```

Prints the module version (`go install`) or the VCS tag/revision of
a local `go build`, plus the Go toolchain version.

### Exit status

0 on success, 1 on error (no readable file, invalid XML, unknown path).
