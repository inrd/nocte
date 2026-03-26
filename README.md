# nocte

`nocte` is a small keyboard-first terminal note app that stores notes as plain Markdown files on disk.

## Install

- Run `make install` to build and install `nocte` to `~/.local/bin/nocte`
- Set a custom install location with `make install BINDIR=/your/bin/path`
- Make sure your install directory is on your `PATH`

For example, in `~/.zshrc`:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

## What It Does

- Create a new note from the main input
- Fuzzy-search existing notes as you type
- Browse all notes with `:list`
- Edit notes in a plain text editor with an optional live Markdown preview
- Open note links from the editor with a keyboard shortcut
- Delete the current note from the editor with an in-app confirmation dialog
- Notes are saved on editor exit when content changed, and brand-new untouched empty notes are discarded

## How To Use It

- Type a note name to search for existing notes
- Press `Up` or `Down` to move through matches
- Press `Enter` on a selected match to open it
- Press `Enter` without selecting a match to create a new note
- Press `Ctrl+P` in the editor to toggle the live Markdown preview
- Press `Ctrl+L` in the editor to open a dialog of links in the current note, then press `Enter` to open one in your default browser
- Press `Ctrl+D` in the editor to delete the current note after confirming, then return to the launcher
- Use `:list` to browse every note sorted by last update

## Commands

- `:help` shows available commands
- `:info` shows version and path information
- `:list` shows all existing notes in a selectable dialog sorted by last update
- `:files` opens the notes directory in the system file manager
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
