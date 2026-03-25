# AGENTS.md

## Project

`not` is a small Go TUI note app inspired by nvALT.

The project should evolve in small, reversible steps. Favor simple implementations that leave room for iteration over premature architecture or polished abstractions.

## Current Direction

- Keep the app keyboard-first and fast.
- Preserve a lightweight local-first model.
- Use plain Markdown files as notes.
- Treat `:`-prefixed input as internal commands.
- Prefer organic UI evolution instead of locking in a full layout too early.

## Tech Stack

- Language: Go
- TUI: Bubble Tea, Bubbles, Lip Gloss
- Entry point: `cmd/not`
- App state/UI: `internal/app`
- Config: `internal/config`

## Important Behavior

- New notes are created from the main input.
- Filenames are sanitized before writing.
- Whitespace-only note names are invalid.
- Note creation must never overwrite an existing file.
- Command dialogs should clear the input and refocus it when closed.
- The config file lives at `~/.config/not/config.json`.
- The default notes directory is `~/not`.

## Workflow

Use the Make targets instead of raw Go commands when practical:

- `make run`
- `make build`
- `make fmt`
- `make tidy`

The repo uses workspace-local `GOCACHE` and `GOMODCACHE` through the `Makefile`. Keep that setup intact unless there is a clear reason to change it.

## Coding Guidelines

- Keep changes small and easy to review.
- Prefer straightforward code over cleverness.
- Do not introduce large frameworks or complex abstractions early.
- Keep UI text concise and practical.
- Preserve cross-platform behavior for macOS and Linux.
- When adding commands, make them discoverable through `:help`.

## Near-Term Priorities

- Grow the command palette gradually.
- Add note search and selection.
- Introduce safer note-opening and existing-note flows.
- Keep config and storage paths explicit and inspectable.
