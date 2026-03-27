# AGENTS.md

## Project

`nocte` is a small Go TUI note app for fast local note-taking.

The project should evolve in small, reversible steps. Favor simple implementations that leave room for iteration over premature architecture or polished abstractions.
Keep this file up to date on each iteration of the codebase so it remains a reliable snapshot of current behavior and expectations.
Keep `README.md` updated on each iteration as user-facing behavior, setup, and workflow change.
Keep `README.md` focused on meaningful user-facing capabilities and workflow. Avoid documenting low-level implementation details, bugfixes, or internal limits unless they change how a user should use the app.

## Current Direction

- Keep the app keyboard-first and fast.
- Favor keyboard shortcuts over mouse interaction in the core workflow.
- Preserve a lightweight local-first model.
- Use plain Markdown files as notes.
- Treat `:`-prefixed input as internal commands.
- Prefer organic UI evolution instead of locking in a full layout too early.

## Tech Stack

- Language: Go
- TUI: Bubble Tea, Bubbles, Lip Gloss
- Entry point: `cmd/nocte`
- App state/UI: `internal/app`, split into focused files by responsibility while keeping the single `Model` as the central app state
- Tests: `internal/app` tests are split into focused `*_test.go` files by responsibility, with shared test helpers kept separately
- Config: `internal/config`

## Important Behavior

- New notes are created from the main input.
- Typing a note name in the main input should show a fuzzy note-search palette for existing notes.
- Typing `/` followed by a query in the main input should show a selectable full-text search palette with one row per match, including the note name and a dimmed multi-line snippet around the match.
- Running `:list` should show all existing notes in a dialog sorted by last updated, with the updated date/time in green alongside word count and file size metadata.
- Running `:export-all` should rebuild the `html` subdirectory of the main notes directory and render every Markdown note there as HTML without opening a browser.
- Running `:files` should open the notes directory in the system file manager and create the directory first when it does not exist yet.
- Typing after `:` should keep command-name prefix matches ahead of looser fuzzy or description-only command matches in the command palette.
- Existing note matches should not be selected by default.
- Existing full-text search matches should not be selected by default.
- Pressing `Up` or `Down` should move through note matches, and pressing `Enter` with a selected match should open that note in the editor.
- Pressing `Up` or `Down` should move through full-text search matches, and pressing `Enter` with a selected match should open that note in the editor at the matching line.
- In the `:list` dialog, `Up` and `Down` should move through the full note list, scrolling when needed, and `Enter` should open the selected note in the editor.
- Pressing Enter on a new note name should create the file and immediately open an editor view for that note.
- Pressing `Enter` in `/` search mode without a selected match should not create a new note.
- Pressing `Enter` without a selected note match should keep the create-new-note behavior.
- Leaving a newly created note empty and exiting the editor should delete that note instead of saving an empty file.
- The initial note editor is plain text only, saves on close when content changed, and otherwise closes without rewriting the file.
- The editor can show a toggleable live Markdown preview, controlled by keyboard shortcut and documented below the editor, including headings, links, inline code, bold, italics, strikethrough, task lists, and nested Markdown list indentation.
- Pressing `Ctrl+E` in the editor should render the current note to an HTML file inside an `html` subdirectory of the main notes directory and open that rendered file in the default web browser.
- The HTML export should stay broadly aligned with the editor's Markdown preview for the currently supported Markdown subset, but it does not need to behave like a full publishing-grade Markdown renderer.
- Markdown image lines like `![alt](./image.png)` should render local image previews through `chafa` when it is available, and otherwise fall back to readable image label/path text.
- Pressing `Ctrl+L` in the editor should open a dialog listing links found in the current note, with visually distinct labels and URLs, and pressing `Enter` on a selected link should open it in the default web browser.
- Pressing `Ctrl+D` in the editor should open a confirmation alert for deleting the current note, prevent terminal EOF handling, and on confirm delete the note and return to the launcher.
- The editor must support long notes without truncating content.
- The editor footer should show the current note size and warn when a very large note may slow editing or saving.
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
- `make install`
- `make test`
- `make fmt`
- `make tidy`
- `make release VERSION=0.3.1`

The repo uses workspace-local `GOCACHE` and `GOMODCACHE` through the `Makefile`. Keep that setup intact unless there is a clear reason to change it.
Use `make release VERSION=...` for version bumps so the version update, test run, commit, and tag creation stay consistent. Add `PUSH=1` when the release flow should also push the current branch and tag to `origin`.
Keep small developer demo assets under `scripts/` when they help explain the current workflow. The committed README demo asset lives at `docs/demo/editor-demo.gif` and is regenerated with `make demo-gif`, which seeds `/tmp/nocte-vhs-demo` from `scripts/vhs/fixtures/notes` and records `scripts/vhs/editor-demo.tape`.

## Coding Guidelines

- Keep changes small and easy to review.
- Prefer straightforward code over cleverness.
- Do not introduce large frameworks or complex abstractions early.
- Keep UI text concise and practical.
- Preserve cross-platform behavior for macOS and Linux.
- Keep repo content free of machine-specific absolute paths and other personal local environment details.
- Add or update tests when introducing new behavior or changing existing behavior.
- When adding commands, make them discoverable through `:help`.

## Public Repo Hygiene

- Keep the Go module path aligned with the published GitHub repository path.
- Prefer `~`-based or other portable path examples in docs and user-facing text instead of machine-specific absolute paths.
- Avoid committing personal local environment details unless they are intentionally part of project ownership metadata such as the license copyright.

## Near-Term Priorities

- Grow the command palette gradually.
- Introduce safer note-opening and existing-note flows.
- Keep config and storage paths explicit and inspectable.
- Expand test coverage in small steps around current helper logic and note workflows.
