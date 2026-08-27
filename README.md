# rePugnant

rePugnant (`rpg`) is a code-adjacent documentation system. Developers write short annotations and Markdown comments in source; rpg turns them into stable, revisioned documentation with code quotes, local `/docs` output, and a searchable web workspace.

V1 includes a Go CLI and backend, React web interface, Docker deployment, SQLite by default, and optional PostgreSQL. It is designed to keep documentation in the coding flow, not move people into a separate editor.

## Quick start

Run this from the Git worktree you want to document:

```sh
go run ./cmd/rpg init
# Add source annotations, then commit. The hook writes and stages docs/.
go run ./cmd/rpg generate
```

Use `$rPg(Platform/Caching/Resolve cache, performance)` for a standalone page or `?rPg(...)` plus `!rPg` for a quoted source region. The CLI replaces a generated marker with a permanent `rPg: opaqueID` reference. To record a changed quote, add `$rPg@opaqueID: explanation` and `$~` continuation lines; rpg appends a revision without replacing manual edits in the Markdown file.

Start the server with `docker compose up --build` (or build/run the Go server directly). It listens on all interfaces and serves the wiki on port 8080. Copy `.env.example` to `.env` to override the SQLite defaults or select PostgreSQL. This V1 server has no user authentication: deploy it only on a trusted LAN or VPN.

Create a remote project with `rpg project create --server http://server:8080 --name "My project"`, copy its one-time API key into `rpg.conf.yaml`, then run `rpg push` or let pre-push publish it. `rpg --init-hooks` installs or repairs only hooks and preserves pre-existing hooks as executable `.rpg-backup` files.

## Development

```sh
go test ./cmd/... ./internal/...
npm --prefix web ci
npm --prefix web run build
```

`Dockerfile` and `compose.yaml` provide the container deployment foundation. Product behavior, style, infrastructure, test expectations, and deferred work are defined in [`rules_of_engagement/`](rules_of_engagement/).
