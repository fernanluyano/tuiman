# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**tuiman** is a terminal UI REST API client written in Go — a TUI equivalent of Postman. It lets users compose and send HTTP requests and inspect responses from the terminal.

## Stack

- **Language**: Go (standard toolchain, `go mod`)
- **TUI framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-architecture: `Model`, `Update`, `View`)
- **Companion libs** (charmbracelet ecosystem): Lip Gloss (styling), Bubbles (reusable components)

## Commands

```bash
go run .              # Run the app
go build -o tuiman .  # Build binary
go test ./...         # Run all tests
go test ./... -run TestName   # Run a single test
go vet ./...          # Static analysis
```

## Architecture

Bubble Tea apps follow the Elm architecture — every screen/component is a `Model` with three functions:

```
Init()  → initial Cmd
Update(msg) → (Model, Cmd)
View()  → string
```

Expected package layout:

```
main.go           # Entry point, starts the Bubble Tea program
internal/
  ui/             # TUI models and views (one file/package per screen or pane)
  api/            # HTTP client logic (building requests, parsing responses)
  config/         # Persistence: saved collections, environments, history
```

Key conventions:
- Keep HTTP logic (`internal/api`) fully decoupled from TUI code; it should be testable without a terminal.
- Use `tea.Cmd` (async commands) for all I/O — never block inside `Update`.
- Each major pane (request editor, response viewer, sidebar) should be its own `Model` that the root model composes.
- Lip Gloss styles should be defined as package-level `var` in a `styles.go` file within the relevant package.
