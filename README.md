# nocte

`nocte` is a small keyboard-first terminal note app that stores notes as plain Markdown files on disk.

## What It Does

- Create a new note from the main input
- Fuzzy-search existing notes as you type
- Open a selected note with `Enter`
- Browse all notes with `:list`
- Notes are automatically saved on editor exit
- Return to the launcher with `Esc`

## How To Use It

- Type a note name to search for existing notes
- Press `Up` or `Down` to move through matches
- Press `Enter` on a selected match to open it
- Press `Enter` without selecting a match to create a new note
- Use `:list` to browse every note sorted by last update, with a green updated timestamp plus word count and file size metadata

## Commands

- `:help` shows available commands
- `:info` shows version and path information
- `:list` shows all existing notes in a selectable dialog sorted by last update
- `:quit` exits the app

## Paths

- Config: `~/.config/nocte/config.json`
- Default notes directory: `~/nocte`

The config file currently supports:

```json
{
  "notes_path": "~/nocte"
}
```

## Disclaimer

This project is developed using iterative LLM-assisted workflows, with incremental changes and continuous testing.
