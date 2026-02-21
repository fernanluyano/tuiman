# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**tuiman** is a terminal UI REST API client written in Go — a TUI equivalent of Postman. It lets users compose and send HTTP requests and inspect responses from the terminal.

## Stack

- **Language**: Go (standard toolchain, `go mod`)
- **TUI framework**: [tview](https://github.com/rivo/tview)
- **Terminal library**: [tcell](https://github.com/gdamore/tcell)

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
- All "services" (HTTP client, config, collections) are stubs that return hardcoded data.
- No code should live outside `internal/ui/` except `main.go` and any stub files needed to satisfy interfaces.
- Do NOT implement `internal/api/` or `internal/config/` with real logic yet.

When the user asks to implement a feature, build the UI interaction and wire it to a mock/stub — do not write real HTTP or file I/O code.

## Architecture

The app is built with tview, using a `tview.Application` with a `tview.Pages` root to layer the main layout and modal overlays (e.g. help screen).

Package layout (mocked phase):

```
main.go           # Entry point, initialises and runs the tview application
internal/
  ui/             # All TUI code
                  # root.go    — wires all panes together, global key handling
                  # sidebar.go — collections tree pane
                  # styles.go  — shared styles/colors
```

Key conventions:
- Each major pane (sidebar, request, response) is its own constructor returning a tview primitive.
- Global keybindings are handled in `app.SetInputCapture` in `root.go`.
- Per-pane keybindings are handled in each pane's own `SetInputCapture`.
- All data displayed in the UI comes from hardcoded structs inside `internal/ui/` during this phase.
