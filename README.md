# tuiman

> **Work in progress** — this project is under active construction and not yet ready for use.

A terminal UI REST API client written in Go — a TUI equivalent of Postman.

## Stack

- **Go** — standard toolchain
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — Elm-style MVU TUI framework
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — layout and styling

## Usage

```sh
make run      # Run the app
make build    # Build binary to build/tuiman
make test     # Run tests
make vet      # Static analysis
make clean    # Remove build artifacts
```

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `tab` / `j` / `k` | Cycle panes |
| `c` | Open collections picker |
| `?` | Toggle help |
| `q` | Quit |

### Request Pane

| Key | Action |
|-----|--------|
| `m` | Open method picker (GET, POST, PUT, PATCH, DELETE) |
| `e` | Edit URL — `enter` or `esc` to stop |
| `s` | Send request |
| `[` / `]` | Previous / next tab |
| `p` | Jump to Params tab |
| `a` | Jump to Auth tab |
| `h` | Jump to Headers tab |
| `b` | Jump to Body tab |

### Collections Picker

| Key | Action |
|-----|--------|
| Type | Fuzzy search |
| `j` / `k` | Navigate list |
| `enter` | Select request |
| `esc` | Close |
