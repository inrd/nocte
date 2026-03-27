# Development

Use the Make targets instead of raw Go commands when practical:

- `make run`
- `make build`
- `make install`
- `make test`
- `make fmt`
- `make tidy`

Release the app with:

- `make release VERSION=0.3.1`

That release flow updates the version, runs tests, creates the release commit, and tags it locally.
Add `PUSH=1` to also push the current branch and tag to `origin`.

Other development workflow:

- Run `make test` before pushing changes
- Run `make demo-gif` to rebuild `docs/demo/editor-demo.gif` from checked-in VHS fixtures and the current UI

The repo uses workspace-local `GOCACHE` and `GOMODCACHE` through the `Makefile`. Keep that setup intact unless there is a clear reason to change it.
