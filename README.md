# not

`not` is a small Go TUI note app inspired by nvALT.

It currently focuses on a fast keyboard-first flow:

- Type a note name in the launcher to fuzzy-search existing notes.
- Use `Up` and `Down` to select a matching note when you want to open one.
- Press `Enter` with a selected note to open it.
- Run `:list` to browse all existing notes in a dialog, including character count and file size, and open one directly.
- Press `Enter` without selecting a note to create a new one.
- Edit the note immediately in a simple plain-text editor.
- Changes autosave as you type.
- Press `Esc` to return to the launcher.

## Current Features

- Lightweight local-first note creation
- Fuzzy note search in the launcher
- Existing-note list dialog with metadata and scrolling for direct browsing
- Plain Markdown files on disk
- Filename sanitization for new notes
- Protection against overwriting existing notes
- Minimal command palette for internal commands

## Commands

- `:help` shows available commands
- `:info` shows version and path information
- `:list` shows all existing notes in a selectable dialog
- `:quit` exits the app

## Paths

- Config: `~/.config/not/config.json`
- Default notes directory: `~/not`

The config file currently supports:

```json
{
  "notes_path": "~/not"
}
```

## Development

Use the Make targets when practical:

- `make run`
- `make build`
- `make fmt`
- `make tidy`

The `Makefile` sets workspace-local `GOCACHE` and `GOMODCACHE`.

## Status

The app is intentionally early and evolving in small, reversible steps. The current editor is plain text only, with room to grow into richer note browsing and command flows later.
