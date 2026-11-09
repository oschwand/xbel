# xbel

A CLI for XBEL bookmark files. Reads XBEL (XML Bookmark Exchange
Language), validates it, and renders its contents in the terminal.

## Build

    go build

Produces an `xbel` binary. Requires Go (see `go.mod`).

## Commands

The XBEL file is resolved in this order:

1. the `--xbel` flag (`xbel --xbel /path/to/bookmarks.xbel tree`);
2. the `XBEL_FILE` environment variable;
3. `bookmarks.xbel` in the current directory, when it exists.

### xbel tree

Prints the folder tree of the document, with a bookmark count per
folder and a summary header (XBEL version, folder and bookmark totals).
Empty folders are marked `(empty)`, untitled folders `(untitled)`.

    xbel tree

### xbel ls [path]

Lists the folders and bookmarks of one directory:

- `path` is a slash-separated list of folder titles, relative to the
  document root (`xbel ls "Bookmarks Menu/Development"`). Omit it to
  list the root.
- Folder titles match exactly first, then by unique case-insensitive
  match. Errors list the folder names available at that level.

Output: folders first (with bookmark counts), then bookmarks. Titles
and URLs share a line when every entry fits the terminal width;
otherwise each URL is printed on its own line under its title.
Non-terminal output assumes 80 columns.

### xbel search <query>

Searches every bookmark of the document: a bookmark matches when the
query appears, case-insensitively, in its title, its URL, or the name
of a folder on its path (the default searches all of them).

    xbel search charm
    xbel search --title charm
    xbel search --url charm
    xbel search --folder charm

Results show the folder path leading to each match, then the title
and URL. Paths are slash-separated, the same format `xbel ls`
accepts, so a search result can be copied straight into an ls call.
The flags select the fields to search: `--title` and `--url`
restrict bookmarks to those fields; `--folder` searches folder names
and lists the matching folders instead (with their bookmark count).
Combining flags searches the union of the selected fields; with no
flag all fields are searched.

## Exit status

0 on success, 1 on error (unreadable file, invalid XML, unknown path).
