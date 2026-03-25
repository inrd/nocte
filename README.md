# not

`not` is a small Go TUI note app inspired by nvALT.

It currently focuses on a fast keyboard-first flow:

- Type a new note name in the launcher.
- Press `Enter` to create the note.
- Edit the note immediately in a simple plain-text editor.
- Changes autosave as you type.
- Press `Esc` to return to the launcher.

## Current Features

- Lightweight local-first note creation
- Plain Markdown files on disk
- Filename sanitization for new notes
- Protection against overwriting existing notes
- Minimal command palette for internal commands

## Commands

- `:help` shows available commands
- `:info` shows version and path information
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

The app is intentionally early and evolving in small, reversible steps. The current editor is plain text only, with room to grow into richer note search and open flows later.
