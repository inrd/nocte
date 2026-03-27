# nocte

`nocte` is a small keyboard-first terminal note app that stores notes as plain Markdown files on disk.

![nocte demo](docs/demo/editor-demo.gif)

## Requirements

- Go
- `chafa` if you want Markdown image previews in the editor

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
- Full-text search across note contents with `/`
- Browse all notes with `:list`
- Browse open Markdown tasks across all notes with `:todo`
- Render all notes to HTML with `:export-all`
- Edit notes in a plain text editor with an optional live Markdown preview for headings, lists, task lists, links, inline code, bold, italics, and strikethrough, including terminal image rendering through `chafa` when available
- Export the current note from the editor to an HTML file in `~/nocte/html` or your configured notes directory equivalent, then open it in your default browser
- Open note links from the editor with a keyboard shortcut
- Delete the current note from the editor with an in-app confirmation dialog
- Notes are saved on editor exit when content changed, and brand-new untouched empty notes are discarded

## How To Use It

- Type a note name to search for existing notes
- Type `/` followed by text to search inside your notes, then use `Up` and `Down` to choose a match
- Press `Up` or `Down` to move through matches
- Press `Enter` on a selected match to open it
- Press `Enter` on a selected `/` search result to open the note at the matching line
- Press `Enter` without selecting a match to create a new note
- Press `Ctrl+H` in the editor to open a dialog listing all editor shortcuts
- Press `Ctrl+P` in the editor to toggle the live Markdown preview
- Press `Ctrl+T` in the editor to toggle the current line between an open and checked Markdown task, or turn a non-task line into an open task
- Press `Ctrl+E` in the editor to render the current note to `html/<note>.html` under your notes directory and open it in your default browser
- Add a Markdown image like `![alt](./image.png)` to preview local images beside the editor when `chafa` is installed
- Press `Ctrl+L` in the editor to open a dialog of links in the current note, then press `Enter` to open one in your default browser
- Press `Ctrl+D` in the editor to delete the current note after confirming, then return to the launcher
- Use `:list` to browse every note sorted by last update

If `chafa` is not installed or an image cannot be rendered, the preview falls back to showing the image label and path.

## Commands

- `:help` shows available commands
- `:export-all` renders all notes to HTML in the notes directory `html` folder
- `:info` shows version and path information
- `:list` shows all existing notes in a selectable dialog sorted by last update
- `:todo` shows open Markdown tasks across notes in a searchable results palette
- `:files` opens the notes directory in the system file manager
- `:quit` exits the app

## Paths

- Config: `~/.config/nocte/config.json`
- Default notes directory: `~/nocte`

The config file currently supports:

```json
{
  "notes_path": "~/nocte",
  "tab_width": 4
}
```
