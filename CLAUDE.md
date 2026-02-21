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
- All "services" (HTTP client, config, collections) are stubs that return hardcoded data.
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
                  # root.go    — Model struct, New(), Init(), Update()
                  # view.go    — View() and all rendering helpers
                  # sidebar.go — sidebar data types and buildItems()
                  # styles.go  — shared Lip Gloss styles and colors
```

Key conventions:
- **Model** holds all state: focused pane, sidebar cursor, expand map, terminal size, help toggle.
- **Update** handles `tea.WindowSizeMsg` and `tea.KeyMsg`; returns a new Model value (no mutation).
- **View** composes the layout with `lipgloss.JoinHorizontal/JoinVertical`; help is rendered via `lipgloss.Place`.
- All data displayed in the UI comes from hardcoded structs inside `internal/ui/` during this phase.
