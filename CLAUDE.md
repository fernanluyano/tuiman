# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**tuiman** is a terminal UI REST API client written in Go — a TUI equivalent of Postman. It lets users compose and send HTTP requests and inspect responses from the terminal.

## Stack

- **Language**: Go (standard toolchain, `go mod`)
- **TUI framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) — Elm-style MVU
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss) — layout and color

## Commands

```bash
make run              # Run the app
make build            # Build binary to build/tuiman
make test             # Run all tests
make vet              # Static analysis
make clean            # Remove build artifacts

# Or directly via go tool:
go test ./... -run TestName   # Run a single test
```

## Current Phase: Mocked TUI

**We are building the UI first with mocked/hardcoded data. No real HTTP calls, no persistence, no config loading.**

Goals for this phase:
- Build and iterate on the full TUI layout and UX using static/mocked data.
- All "services" (HTTP client, config, folders) are stubs that return hardcoded data.
- No code should live outside `internal/ui/` except `main.go` and any stub files needed to satisfy interfaces.
- Do NOT implement `internal/api/` or `internal/config/` with real logic yet.

When asked to implement a feature, build the UI interaction and wire it to a mock/stub — do not write real HTTP or file I/O code.

## Architecture

The app uses the Bubble Tea MVU pattern: `Model` → `Update(msg)` → `View()`. All state lives in a single `Model` struct; rendering is pure string composition via Lip Gloss.

Package layout (mocked phase):

```
main.go           # Entry point — tea.NewProgram(ui.New(), tea.WithAltScreen())
internal/
  ui/             # All TUI code
                  # root.go    — Model struct, New(), Init(), Update(), key handlers
                  # view.go    — View() and all rendering helpers (request pane, picker, overlay)
                  # sidebar.go — folder/request data types and mock data
                  # styles.go  — shared Lip Gloss styles and colors
```

Key conventions:
- **Model** holds all state: focused pane, active request, tab index, URL input, folder picker state (level, cursor, query, insert mode, add/confirm-delete sub-states), terminal size, help toggle.
- **Update** handles `tea.WindowSizeMsg` and `tea.KeyMsg`; returns a new Model value (no mutation).
- **View** composes the layout with `lipgloss.JoinHorizontal/JoinVertical`; overlays (picker, help) are rendered via `lipgloss.Place` or line-by-line compositing (`placeOverlay`).
- All data displayed in the UI comes from hardcoded structs inside `internal/ui/` during this phase.

## UX Design Decisions

### Vim-oriented keybindings
All navigation follows vim conventions. Do not introduce letter shortcuts that conflict with vim muscle memory:
- `j` / `k` — move down / up (cycle panes in normal mode, navigate lists)
- `[` / `]` — cycle tabs within a pane (arrow keys are aliases)
- `esc` / `enter` — confirm or cancel an editing mode
- When choosing a letter shortcut for a new action, pick the letter that most naturally represents the action (e.g. `e` for edit, `m` for method, `f` for folders). Avoid letters already in use.

### Folder picker modal modes
The folder picker uses a vim-style normal/insert mode split:
- **Normal mode** (default): `j/k` navigate, `enter` opens/selects, `n` adds, `d` deletes (with `y/n` confirmation), `esc` closes or goes back a level.
- **Insert mode** (`i` to enter): typing filters the list live; `esc` returns to normal mode; `enter` exits insert mode and immediately performs the enter action.
- The query line shows `-- NORMAL --` or `-- INSERT --` as a mode indicator.
- Sub-states (adding, confirm-delete) overlay both modes and are cleared on `esc`.

### Label-oriented UI
Interactive elements must display their key binding inline as a dim label so the UI is self-documenting:
- Format: `(key)` in a dimmed color adjacent to the element.
- Do not rely solely on a footer or help screen to communicate keybindings — the label should be visible at the point of interaction.
- When a mode is active, the label updates to show the exit key — e.g. `(e)` becomes `(esc)` while editing.
