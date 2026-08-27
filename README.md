# rePugnant

rePugnant (`rpg`) is a code-adjacent documentation system. Developers write short annotations and Markdown comments in source; rpg turns them into stable, revisioned documentation with code quotes, local `/docs` output, and a searchable web workspace.

V1 includes a Go CLI and backend, React web interface, Docker deployment, SQLite by default, and optional PostgreSQL. It is designed to keep documentation in the coding flow, not move people into a separate editor.

## Development

The initial scaffold provides `rpg --version`, a Go health endpoint at `GET /healthz`, and a React/Tailwind web shell.

```sh
go test ./cmd/... ./internal/...
npm --prefix web ci
npm --prefix web run build
```

`Dockerfile` and `compose.yaml` provide the container deployment foundation. Product behavior, style, infrastructure, test expectations, and deferred work are defined in [`rules_of_engagement/`](rules_of_engagement/).
