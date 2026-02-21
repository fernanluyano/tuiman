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

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle through panes |
| `S` | Jump to Sidebar |
| `R` | Jump to Request pane |
| `P` | Jump to Response pane |
| `j` / `↓` | Move down (sidebar) |
| `k` / `↑` | Move up (sidebar) |
| `Space` | Expand / collapse collection |
| `?` | Toggle help |
| `q` | Quit |
