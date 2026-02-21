# tuiman

> **Work in progress** — this project is under active construction and not yet ready for use.

A terminal UI REST API client written in Go — a TUI equivalent of Postman.

## Stack

- **Go** — standard toolchain
- **[tview](https://github.com/rivo/tview)** — TUI framework
- **[tcell](https://github.com/gdamore/tcell)** — terminal cell library

## Usage

```sh
make run      # Run the app
make build    # Build binary to build/tuiman
make test     # Run tests
make vet      # Static analysis
make clean    # Remove build artifacts
```

## Keybindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle through panes |
| `S` | Jump to Sidebar |
| `R` | Jump to Request pane |
| `P` | Jump to Response pane |
| `?` | Toggle help |
| `q` | Quit |
