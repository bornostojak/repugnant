# Delivery progress

Status: **In progress — not ready for review or release.**

| Area | Status | Evidence / outstanding work |
|---|---|---|
| CLI initialization | Complete | `rpg init` is idempotent, installs chained hooks, generates/stages configured local docs, and uses a blocking pre-push default. |
| Local documentation | Partial | Category/tag metadata, Git identity, stable IDs, quote-drift diagnostics, append-only revisions, source ranges, and manifest-aware language selection work. |
| Backend | Partial | Project/API-key/article/revision/short-link/search/organization/source-range APIs work; PostgreSQL passes a real Docker-backed contract test. |
| Frontend | Partial | Built React wiki is served by the backend with project creation, tree/search, reader/source modes, history, and organization controls. Mobile browser rendering passes; keyboard/browser regression coverage remains. |
| Deployment | Complete | Container defaults serve the web app and persist SQLite; PostgreSQL is selectable and verified with a real Docker engine container. |
| Examples | Partial | Ten fixtures are checked in and an automated temporary-Git integration test validates init, commit hook, generation, manual push, and server persistence. |
| Manual server push | Complete | `rpg push` publishes docs metadata and is idempotent against unchanged server content. |

## Latest verified work

- `go test ./cmd/... ./internal/...`, `go vet ./cmd/... ./internal/...`, and `npm --prefix web run build` pass.
- `go test ./internal/integration -v` exercises all ten fixture repositories using real `rpg init`, pre-commit hook, generated docs, `rpg push`, and an HTTP server.
- A live server served the built homepage and successfully created, published, searched, and organized a test article; Playwright rendered the responsive 390×844 homepage after the empty-array regression fix.
- `RPG_TEST_POSTGRES_DSN=... go test ./internal/store -run TestPostgresStorageContract -v` passed against PostgreSQL 17 in Docker.

These checks do **not** constitute full delivery. See `ROADMAP.md` for remaining required work.
