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
                  # styles.go  — Theme methods (paneStyle, overlayStyle, dim, accent, etc.)
                  # theme.go   — Theme struct, 5 presets, methodStyle
```

Key conventions:
- **Model** holds all state: focused pane, split direction, active theme, active request, tab index, URL input, folder picker state (level, cursor, query, insert mode, add/confirm-delete sub-states), command palette state, terminal size, help/cmd-help toggles, quit confirmation flag.
- **Update** handles `tea.WindowSizeMsg` and `tea.KeyMsg`; returns a new Model value (no mutation).
- **View** composes the layout with `lipgloss.JoinHorizontal/JoinVertical`; overlays (picker, help) are rendered via `lipgloss.Place` or line-by-line compositing (`placeOverlay`).
- All data displayed in the UI comes from hardcoded structs inside `internal/ui/` during this phase.
- **Theming**: all colors must go through `Theme` methods (e.g. `m.theme.accent()`, `m.theme.overlayStyle()`). Never use hardcoded hex strings or `lipgloss.Color(...)` outside of `theme.go`.

## UX Design Decisions

### Vim-oriented keybindings
All navigation follows vim conventions. Do not introduce letter shortcuts that conflict with vim muscle memory:
- `j` / `k` — navigate lists inside overlays (folder picker, method picker). NOT used for pane cycling.
- `tab` / `shift+tab` — cycle between panes
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

## Completed Improvements

### ✓ Split direction toggle
Default layout is request left (60%) and response right (40%). `:orient` toggles between side-by-side and stacked (top/bottom). The bottom bar shows the current mode; `tab`/`shift+tab` cycle panes.

### ✓ Command palette
`:` opens a vim-style command bar replacing the footer while active. `esc` closes, `enter` executes. Unknown commands show an inline red error without closing. `:help` opens a dedicated commands overlay.
Current commands: `:orient`, `:theme <name>`, `:help`.

### ✓ Color themes
5 themes switchable at runtime via `:theme <name>`. Each has a distinct accent color for the active pane border. Default is `tokyonight`.
- `rosepine` — pink `#EB6F92`
- `xcode` — cyan `#6BDFFF`
- `catppuccin` — mauve `#CBA6F7`
- `tokyonight` — blue `#7AA2F7`
- `sonokai` — green `#9ED072`

All colors are defined in `theme.go` and accessed through `Theme` methods in `styles.go`. No hardcoded hex values anywhere outside `theme.go`.

## Planned Improvements

### 1. Environments and variables
Variables like `{{base_url}}` rendered in a distinct highlight color inline in the URL bar and body. An environment switcher (small dropdown overlay, same style as the method picker) lets the user flip between named environments (e.g. `dev`, `staging`, `prod`). Palette command: `:env <name>`. This is a first-class feature — the primary way to avoid hardcoding hosts.

### 2. Response pane information density
When a response is received the response pane shows:
- **Top line**: status code badge (green 2xx, yellow 3xx, red 4xx/5xx) · response time · response size.
- **Tabs**: Body / Headers / Timing — cycled with `[`/`]` like the request pane.
- **Body tab**: scrollable, syntax-highlighted JSON/XML/plain text.
- **Timing tab**: ASCII bar breakdown of DNS, connect, TLS, TTFB, and download phases.

### 3. Request history
Ephemeral in-memory history of sent requests, separate from saved collections.
- Accessible via `H` key as a list overlay (same style as the folder picker).
- Each entry shows: method (color-coded), URL, response status code, and timestamp.
- Selecting a history entry loads it into the request pane for re-use or editing.
- Never written to disk automatically; the user can explicitly promote an entry into a folder.
- Rationale: developers send the same URL dozens of times while debugging and shouldn't need to save it first.
