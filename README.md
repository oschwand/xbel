# xbel

A CLI for XBEL bookmark files. Reads XBEL (XML Bookmark Exchange
Language), validates it, and renders its contents in the terminal.

## Build

    go build

Produces an `xbel` binary. Requires Go (see `go.mod`).

## Commands

### xbel tree [file]

Prints the folder tree of the document, with a bookmark count per
folder and a summary header (XBEL version, folder and bookmark totals).
Empty folders are marked `(empty)`, untitled folders `(untitled)`.
With no file, XBEL is read from stdin.

    xbel tree bookmarks.xbel
    xbel tree < bookmarks.xbel

### xbel ls [file] [path]

Lists the folders and bookmarks of one directory:

- `path` is a slash-separated list of folder titles, relative to the
  document root (`xbel ls bookmarks.xbel "Bookmarks Menu/Development"`).
  Omit it to list the root.
- With no file, XBEL is read from stdin; `-` also means stdin
  (`cat bookmarks.xbel | xbel ls - Toolbar`).
- Folder titles match exactly first, then by unique case-insensitive
  match. Errors list the folder names available at that level.

Output: folders first (with bookmark counts), then bookmarks. Titles
and URLs share a line when every entry fits the terminal width;
otherwise each URL is printed on its own line under its title.
Non-terminal output assumes 80 columns.

## Exit status

0 on success, 1 on error (unreadable file, invalid XML, unknown path).
