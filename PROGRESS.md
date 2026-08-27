# Delivery progress

Status: **In progress — not ready for review or release.**

| Area | Status | Evidence / outstanding work |
|---|---|---|
| CLI initialization | Complete | `rpg init` is idempotent, installs chained hooks, generates/stages configured local docs, and uses a blocking pre-push default. |
| Local documentation | Partial | Category/tag metadata, Git identity, stable IDs, quote-drift diagnostics, and append-only revisions work. Source ranges and language auto-detection still need implementation. |
| Backend | Partial | Project/API-key/article/revision/short-link/search/organization APIs work; source ranges, author fields, and PostgreSQL runtime verification remain. |
| Frontend | Partial | Built React wiki is served by the backend with project creation, tree/search, reader/source modes, history, and organization controls. Chromium interaction verification remains. |
| Deployment | Partial | Container defaults now serve the web app and persist SQLite; Compose is unavailable in this host environment and PostgreSQL has not been runtime-verified. |
| Examples | Partial | Ten fixtures are checked in and an automated temporary-Git integration test validates init, commit hook, generation, manual push, and server persistence. |
| Manual server push | Complete | `rpg push` publishes docs metadata and is idempotent against unchanged server content. |

## Latest verified work

- `go test ./cmd/... ./internal/...` and `npm --prefix web run build` pass.
- `go test ./internal/integration -v` exercises all ten fixture repositories using real `rpg init`, pre-commit hook, generated docs, `rpg push`, and an HTTP server.
- A live server on port 18082 served the built homepage and successfully created, published, searched, and organized a test article.

These checks do **not** constitute full delivery. See `ROADMAP.md` for remaining required work.
