# Delivery progress

Status: **In progress — not ready for review or release.**

| Area | Status | Evidence / outstanding work |
|---|---|---|
| CLI initialization | Complete | `rpg init` is idempotent, installs chained hooks, generates/stages configured local docs, and uses a blocking pre-push default. |
| Local documentation | Partial | Category/tag metadata, Git identity, stable IDs, quote-drift diagnostics, append-only revisions, source ranges, and manifest-aware language selection work. |
| Backend | Partial | Project/API-key/article/revision/short-link/search/organization/source-range APIs work; PostgreSQL passes a real Docker-backed contract test. |
| Frontend | Partial | Built React wiki is served by the backend with project creation, tree/search, reader/source modes, history, and organization controls. Mobile browser rendering passes; keyboard/browser regression coverage remains. A prior JSON-casing bug (see below) meant the tree/search/reader silently rendered blank in every real browser session; this is now fixed and verified with real screenshots. |
| Deployment | Complete | Container defaults serve the web app and persist SQLite; PostgreSQL is selectable and verified with a real Docker engine container. |
| Examples | Partial | Ten fixtures are checked in and an automated temporary-Git integration test validates init, commit hook, generation, manual push, and server persistence. |
| Manual server push | Complete | `rpg push` publishes docs metadata and is idempotent against unchanged server content. |

## Latest verified work

- `go test ./cmd/... ./internal/...`, `go vet ./cmd/... ./internal/...`, and `npm --prefix web run build` pass.
- `go test ./internal/integration -v` exercises all ten fixture repositories using real `rpg init`, pre-commit hook, generated docs, `rpg push`, and an HTTP server.
- A live server served the built homepage and successfully created, published, searched, and organized a test article; Playwright rendered the responsive 390×844 homepage after the empty-array regression fix.
- `RPG_TEST_POSTGRES_DSN=... go test ./internal/store -run TestPostgresStorageContract -v` passed against PostgreSQL 17 in Docker.
- A full code-review pass against `rules_of_engagement/` found and fixed four real bugs, each with regression tests: (1) a parser cross-contamination bug where a standalone article's stable marker could absorb an unrelated later quote's `!rPg`, producing a false drift error and a corrupted manifest source range; (2) tracked-quote manifest entries never refreshed `Path`/`SourceRange` when a quote moved without its content changing; (3) the `/d/{article-id}` and `/d/{article-id}/{revision}` permalinks `SPEC.md` requires were never implemented, so every generated doc's "Web" link silently fell through to the homepage instead of resolving; `AddArticle` also never refreshed `source_range` when only the location changed; (4) **`store.Project`/`store.Article` had no `json` struct tags**, so the API serialized PascalCase Go field names while the web UI's TypeScript reads lowercase/snake_case — every project selector, tree, and article view silently rendered blank in a real browser. Verified end-to-end with real headless Chromium via Playwright against a live server (see the real screenshots in the README).
- A `/p/{slug}/article/{id}/{revision}` deep link now actually restores that project, article, and revision on direct load/refresh instead of always landing on the first project (`web/src/main.tsx`); confirmed with the same Playwright run.
- The server hardcoded `slog.LevelInfo` despite already using `slog.LevelVar` (built for exactly this) and having debug-level request logging that could therefore never be turned on — a direct gap against `TASTE.md`'s "level-selectable" logging requirement. `RPG_LOG_LEVEL` (error/warn/info/debug) now controls it, documented in `.env.example` and wired through `compose.yaml`.

These checks do **not** constitute full delivery. See `ROADMAP.md` for remaining required work.
