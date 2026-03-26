# AGENTS.md

## Project

`nocte` is a small Go TUI note app for fast local note-taking.

The project should evolve in small, reversible steps. Favor simple implementations that leave room for iteration over premature architecture or polished abstractions.
Keep this file up to date on each iteration of the codebase so it remains a reliable snapshot of current behavior and expectations.
Keep `README.md` updated on each iteration as user-facing behavior, setup, and workflow change.

## Current Direction

- Keep the app keyboard-first and fast.
- Preserve a lightweight local-first model.
- Use plain Markdown files as notes.
- Treat `:`-prefixed input as internal commands.
- Prefer organic UI evolution instead of locking in a full layout too early.

## Tech Stack

- Language: Go
- TUI: Bubble Tea, Bubbles, Lip Gloss
- Entry point: `cmd/nocte`
- App state/UI: `internal/app`
- Config: `internal/config`

## Important Behavior

- New notes are created from the main input.
- Typing a note name in the main input should show a fuzzy note-search palette for existing notes.
- Running `:list` should show all existing notes in a dialog sorted by last updated, with the updated date/time in green alongside word count and file size metadata.
- Existing note matches should not be selected by default.
- Pressing `Up` or `Down` should move through note matches, and pressing `Enter` with a selected match should open that note in the editor.
- In the `:list` dialog, `Up` and `Down` should move through the full note list, scrolling when needed, and `Enter` should open the selected note in the editor.
- Pressing Enter on a new note name should create the file and immediately open an editor view for that note.
- Pressing `Enter` without a selected note match should keep the create-new-note behavior.
- The initial note editor is plain text only and saves when the editor closes.
- If saving on editor exit fails, the app should warn before discarding unsaved changes.
- Pressing `Esc` in the editor should return to the main launcher input.
- Filenames are sanitized before writing.
- Whitespace-only note names are invalid.
- Note creation must never overwrite an existing file.
- Command dialogs should clear the input and refocus it when closed.
- The config file lives at `~/.config/nocte/config.json`.
- The default notes directory is `~/nocte`.

## Workflow

Use the Make targets instead of raw Go commands when practical:

- `make run`
- `make build`
- `make test`
- `make fmt`
- `make tidy`

The repo uses workspace-local `GOCACHE` and `GOMODCACHE` through the `Makefile`. Keep that setup intact unless there is a clear reason to change it.

## Coding Guidelines

- Keep changes small and easy to review.
- Prefer straightforward code over cleverness.
- Do not introduce large frameworks or complex abstractions early.
- Keep UI text concise and practical.
- Preserve cross-platform behavior for macOS and Linux.
- Add or update tests when introducing new behavior or changing existing behavior.
- When adding commands, make them discoverable through `:help`.

## Near-Term Priorities

- Grow the command palette gradually.
- Introduce safer note-opening and existing-note flows.
- Keep config and storage paths explicit and inspectable.
- Expand test coverage in small steps around current helper logic and note workflows.
