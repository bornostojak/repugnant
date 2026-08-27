# rePugnant

> **Status: beta.** rePugnant is under active development. Core CLI, backend, and web UI paths work end-to-end and are covered by tests, but interfaces, config format, and stored data may still change without a migration path. Not yet recommended for production use. See [`PROGRESS.md`](PROGRESS.md) and [`ROADMAP.md`](ROADMAP.md) for exact current state.

**rePugnant** (`rpg`) turns comments you were already going to write into stable, revisioned, searchable documentation — without ever leaving your editor.

Programmers like writing code. They tolerate writing comments. They actively avoid opening Confluence, Notion, or any tool that isn't a terminal in the middle of a flow state. `rpg` bets that if writing docs feels exactly like writing a comment, people will actually write good docs — and keep them up to date, because the tool notices when the code they describe has changed.

```c
// $rPg(Platform/Caching/Resolve cache, performance)
// $~ Reads hit the local cache first and fall back to origin on a miss.
```

commit, and that becomes a page. Edit the code it's pinned to and forget to explain the change, and your next `git commit` refuses to proceed until you do.

<p>
  <a href="rules_of_engagement/VISION.md"><strong>Vision</strong></a> ·
  <a href="rules_of_engagement/SPEC.md"><strong>Full spec</strong></a> ·
  <a href="#configuration"><strong>Configuration</strong></a> ·
  <a href="#screenshots"><strong>Screenshots</strong></a> ·
  <a href="rules_of_engagement/"><strong>All governance docs</strong></a>
</p>

---

## Table of contents

- [What it does](#what-it-does)
- [Why it exists](#why-it-exists)
- [How it works](#how-it-works)
- [Screenshots](#screenshots)
- [Install & quick start](#install--quick-start)
- [Configuration](#configuration)
- [CLI reference](#cli-reference)
- [Deployment](#deployment)
- [Development](#development)
- [Project status](#project-status)
- [Documentation map](#documentation-map)

---

## What it does

`rpg` is a small Go CLI plus an optional Go/SQLite (or PostgreSQL) backend and a React web wiki. Together they:

- **Parse annotations directly out of your source comments** — no separate doc files to remember to update, in any of Go, TypeScript, JavaScript/React, Svelte, Python, Rust, C/C++, Bash, Dart, Groovy/Jenkinsfiles, YAML, and Ruby.
- **Generate stable Markdown**, either locally under a `docs/` directory you commit alongside your code, or published to a small self-hosted web wiki, or both.
- **Quote exact code ranges** into the generated docs, so an explanation always sits next to the real implementation it describes — not a paraphrase that quietly rots.
- **Detect documentation drift.** If a quoted block of code changes and nobody explained *why* in a matching revision annotation, the Git pre-commit hook blocks the commit and tells you exactly what to add, in which file, on which line.
- **Keep a full revision history per article**, appended forever, never overwritten — so "why did we do it this way" survives long after the original author has forgotten.
- **Assign every article a permanent, opaque, URL-safe ID** the moment it's generated. Titles can change. Categories can change. Links never break.

It is explicitly *not* a general-purpose wiki, a public documentation host, or a replacement for architecture decision records that live outside the code. V1's web UI ships with no authentication and is meant for a trusted LAN or VPN — see [Deployment](#deployment).

## Why it exists

Writing a `// TODO: explain this later` comment is easy. Writing a paragraph that actually explains *why* a cache uses striped locking instead of a single mutex is hard, and doing it in a separate tool is harder still, because by the time you switch windows the reason is already fading. `rpg` removes both kinds of friction at once:

- The syntax is a comment. It lives in the file you're already editing, in the language you're already writing.
- The output isn't a wiki page you have to remember to update by hand later — it's regenerated from source on every commit, and the tool actively stops you from letting quoted code drift silently out of sync with its explanation.
- The result is still real documentation: categorized, tagged, searchable, linkable, with full history — not just scattered code comments.

See [`rules_of_engagement/VISION.md`](rules_of_engagement/VISION.md) for the fuller pitch and [`rules_of_engagement/SPEC.md`](rules_of_engagement/SPEC.md) for the deterministic rules the tool follows.

## How it works

Two marker families, both written as ordinary language comments:

| Marker | Meaning |
|---|---|
| `$rPg: Title` or `$rPg(Category/Title, tag, tag)` | Start a **standalone article**. Continuation lines use `$~`. |
| `?rPg: Title` or `?rPg(Category/Title, tag, tag)` … `!rPg` | Start a **quoted-code article**. Everything between the marker and the closing `!rPg` is captured verbatim (with safe common indentation stripped) and rendered under the article as `## Documented code`. Continuation lines use `?~`. |
| `rPg: {id}` | What the tool rewrites your marker into after first generation. Stable forever; never edit by hand. |
| `$rPg@{id}: New title, tag` + `$~` lines | Append a **new revision** to an existing article. Required whenever a tracked quote's code changes. |

A worked example, straight out of this repository's own demo fixtures:

```go
// $rPg(Services/Cache/Cache resolution strategy, cache, performance)
// $~ Reads hit the local cache first and fall back to origin on a miss.
// $~ Misses are written back so the next read is warm.
func Resolve(key string, cache map[string]string, origin func(string) string) string {
	if v, ok := cache[key]; ok {
		return v
	}
	v := origin(key)
	cache[key] = v
	return v
}
```

After the pre-commit hook runs `rpg generate`, that marker is rewritten in place to a permanent reference:

```go
// rPg: ICh1mDM-7Ej5
```

Now watch what happens with a **quoted** block when the implementation actually changes. This is the real annotation, edited a second time to explain a real change:

```go
// $rPg@VSVEZVLG1ipa: Stampede guard, cache, reliability
// $~ Switched from one global lock to 16 striped locks keyed by hash so
// $~ unrelated cache keys no longer contend on the same mutex under load.
func guarded(key string, cache map[string]string, origin func(string) string) string {
	lock := stripe(key)
	lock.Lock()
	defer lock.Unlock()
	if v, ok := cache[key]; ok {
		return v
	}
	v := origin(key)
	cache[key] = v
	return v
}
// !rPg
```

Generation appends a **new revision** rather than overwriting the old one. Here is the actual `docs/VSVEZVLG1ipa.md` this produced — every word below is real generator output, not hand-written prose:

```markdown
# Stampede guard

| Metadata | Info |
| :- | :- |
| By | Jane Doe |
| Email | jane.doe@example.com |
| Generated on | 2026-08-27T19:15:21Z |
| Revision | 2 |
| Web | [Open article](http://127.0.0.1:8099/d/VSVEZVLG1ipa) |
| Category | Services/Cache |
| Tags | cache, reliability |

# Revision 1

Only one goroutine may repopulate a key at a time; concurrent misses
wait on the same lock instead of hammering the origin simultaneously.

## Documented code

    func guarded(key string, cache map[string]string, origin func(string) string) string {
        mu.Lock()
        defer mu.Unlock()
        if v, ok := cache[key]; ok {
            return v
        }
        v := origin(key)
        cache[key] = v
        return v
    }

# Revision 2

Switched from one global lock to 16 striped locks keyed by hash so
unrelated cache keys no longer contend on the same mutex under load.

## Documented code

    func guarded(key string, cache map[string]string, origin func(string) string) string {
        lock := stripe(key)
        lock.Lock()
        defer lock.Unlock()
        if v, ok := cache[key]; ok {
            return v
        }
        v := origin(key)
        cache[key] = v
        return v
    }
```

If you'd changed that code *without* adding the `$rPg@{id}:` revision annotation, `git commit` would have stopped with:

```
rpg: cmd/api/stampede.go:18: documented quote changed for VSVEZVLG1ipa; add $rPg@VSVEZVLG1ipa:
explain the change above the quote, followed by $~ Markdown and !rPg
```

file, line, article ID, and the exact fix — every time.

## Screenshots

> Real captures from a locally running build (`go run ./cmd/rpg-server` + the built `web/`), taken with Playwright against live data — not mockups.

**The web wiki** is a single page: a project switcher and full-text search on the left, the selected article in the center, and organization/history controls on the right. No separate "homepage" — the whole workspace is one screen, live-searchable and keyboard-navigable.

![rePugnant web workspace: topbar, project intro, category tree, article reader, and history sidebar](assets/readme/web-overview.png)

**The documentation tree** groups articles by the `Category/Subcategory` path you gave them in the annotation, with tags and full-text search layered on top:

![Sidebar documentation tree grouped by category, with the active article highlighted](assets/readme/web-tree.png)

**Revision history** is never destructive. Every push appends; the reader can jump to any past revision, and a banner offers a one-click way back to latest:

![Article reader showing revision 2 of "Stampede guard" with its revision-history sidebar](assets/readme/web-revision2.png)

## Install & quick start

Requires Go 1.26+ (CLI/backend) and Node.js (only if you're building the web UI yourself). Run everything below from the Git worktree you want to document.

```sh
# From this repository, build the CLI:
go build -o rpg ./cmd/rpg

# In the repository you want to document:
rpg init
# → writes rpg.conf.yaml, an ignored .rpg/ workspace, and portable
#   pre-commit/pre-push hooks (existing hooks are chained, never clobbered)

# Add annotations to your source, then just commit as normal:
git add .
git commit -m "..."
# → pre-commit runs `rpg generate`, writes/stages docs/, and blocks
#   the commit if a quoted region drifted without an explanation

# Or generate on demand without committing:
rpg generate
```

Want the web wiki too? Start the server (SQLite by default, zero setup) and register your project:

```sh
docker compose up --build
# serves the wiki on :8080 — LAN/VPN only, V1 has no authentication

rpg project create --server http://your-server:8080 --name "My project"
# prints an output.web / project block — paste it into rpg.conf.yaml,
# keep the printed api_key private (it is shown exactly once)

rpg push          # publish once, manually
# or just `git push` — pre-push publishes automatically when web output is enabled
```

`rpg --init-hooks` installs or repairs *only* the Git hooks, useful after a fresh clone where `rpg.conf.yaml` already exists.

## Configuration

Everything lives in one commented `rpg.conf.yaml`, written by `rpg init` and safe to hand-edit afterward. Full deterministic rules: [`rules_of_engagement/SPEC.md`](rules_of_engagement/SPEC.md#configuration).

```yaml
# rePugnant project configuration
version: 1
langs: []                 # e.g. [go, typescript] — empty means auto-detect

output:
  docs:
    enabled: true
    dir: docs
  web:
    enabled: false
    # endpoint: http://127.0.0.1:8080

hooks:
  on_publish_failure: block   # or allow_pending

# project:                    # only required when output.web.enabled: true
#   slug: my-project
#   api_url: http://127.0.0.1:8080/api/projects/my-project/articles
#   api_key: ...               # printed once by `rpg project create` — keep private
```

| Key | Type | Default | Meaning |
|---|---|---|---|
| `version` | int | `1` | Config schema version. Must currently be `1`. |
| `langs` | list of strings | `[]` (auto-detect) | Explicit language allowlist, e.g. `[go, ruby]`. Empty means rpg infers languages from root manifests (`go.mod`, `package.json`, `tsconfig.json`, `Cargo.toml`, `pyproject.toml`/`requirements.txt`, `Gemfile`, `pubspec.yaml`, `CMakeLists.txt`); a repo with none of those stays permissive. |
| `output.docs.enabled` | bool | `true` | Generate local Markdown under `output.docs.dir`. Local docs are official output when enabled and are meant to be committed. |
| `output.docs.dir` | string | `docs` | Directory local Markdown is written to. Required (non-empty) when `output.docs.enabled` is `true`. |
| `output.web.enabled` | bool | `false` | Whether the pre-push hook (and `rpg push`) publish to a server. |
| `output.web.endpoint` | string | *(none)* | Base URL of the `rpg-server` instance, e.g. `http://127.0.0.1:8080`. Required when `output.web.enabled: true`. |
| `hooks.on_publish_failure` | `block` \| `allow_pending` | `block` | `block`: a failed `rpg push` fails the git push. `allow_pending`: the push proceeds, the failure is recorded in `.rpg/pending.json`, and (when run interactively) you're asked to confirm. |
| `project.slug` | string | *(none)* | Project slug on the server. Printed by `rpg project create`. |
| `project.api_url` | string | *(none)* | Full article-ingest API URL for this project. Printed by `rpg project create`. |
| `project.api_key` | string | *(none)* | Per-project publish credential. **Shown only once**, at creation time — store it like a secret. |

`project.*` fields are only validated as required when `output.web.enabled: true`; a docs-only project can omit the whole block.

Runtime/container behavior (database driver, listen address, PostgreSQL DSN) is configured separately via `.env` / environment variables — see [`.env.example`](.env.example) and [Deployment](#deployment) below.

## CLI reference

| Command | What it does |
|---|---|
| `rpg init` | Writes `rpg.conf.yaml`, creates the ignored `.rpg/` workspace, installs Git hooks. Idempotent — safe to re-run. |
| `rpg --init-hooks` | Installs/repairs only the hooks; leaves config and `.rpg/` untouched. An existing hook is preserved as a `.rpg-backup` and still runs. |
| `rpg generate` | Parses enabled source files, writes/updates local Markdown, updates `.rpg/manifest.json`, rewrites markers to stable `rPg: {id}` form. Fails loudly on drifted quotes or unknown IDs. |
| `rpg push` | Publishes every file currently in `output.docs.dir` to the configured project API. Idempotent against unchanged server content. |
| `rpg project create --server <url> --name "<name>" [--slug <slug>]` | Creates a project on a running server and prints the `output.web` / `project` config block to paste into `rpg.conf.yaml`. |
| `rpg hook pre-commit` / `rpg hook pre-push` | What the installed Git hooks actually invoke. You normally never call these directly. |
| `rpg --version` | Prints the CLI version. |

## Deployment

```sh
cp .env.example .env    # then edit as needed
docker compose up --build
```

| Variable | Default | Meaning |
|---|---|---|
| `RPG_HTTP_ADDR` | `0.0.0.0:8080` | Listen address. Binds all interfaces — intended for a trusted LAN/VPN only, **not** public exposure. |
| `RPG_DB_DRIVER` | `sqlite` | `sqlite` or `postgres`. |
| `RPG_DB_DSN` | `file:/app/data/rpg.db?_pragma=busy_timeout(5000)` | SQLite file path, or a `postgres://user:pass@host:5432/db?sslmode=disable` DSN when `RPG_DB_DRIVER=postgres`. |

`compose.yaml` mounts a named volume for the SQLite file by default, and includes an opt-in `postgres` profile (`docker compose --profile postgres up`) for the PostgreSQL alternative. There is **no authentication in V1** — see [`rules_of_engagement/INFRASTRUCTURE.md`](rules_of_engagement/INFRASTRUCTURE.md) for the exact trust boundary and what's deliberately deferred.

## Development

```sh
go test ./cmd/... ./internal/...
go vet ./cmd/... ./internal/...
npm --prefix web ci
npm --prefix web run build
```

Ten versioned language fixtures under [`tests/examples/`](tests/examples/) drive an end-to-end integration test (`go test ./internal/integration -v`) that runs real `rpg init` → commit → generate → push → server-verify cycles per language.

## Project status

**Beta — actively developed, not yet released.** Backend, CLI, and web UI are all functional against the current spec; remaining gaps (source-file browser UI, keyboard-navigation coverage, full end-to-end smoke suite) are tracked with checkboxes in [`ROADMAP.md`](ROADMAP.md), and current verified state is in [`PROGRESS.md`](PROGRESS.md).

## Documentation map

This README is intentionally the short version. Everything else lives in [`rules_of_engagement/`](rules_of_engagement/), each file scoped to one concern so an LLM or a human only has to load what's relevant:

| File | Covers |
|---|---|
| [`VISION.md`](rules_of_engagement/VISION.md) | The problem this solves and who it's for. |
| [`SPEC.md`](rules_of_engagement/SPEC.md) | The full deterministic functional contract — annotation grammar, generation/hook behavior, document model, web behavior. |
| [`TASTE.md`](rules_of_engagement/TASTE.md) | Backend architecture, config, and logging conventions. |
| [`STYLE.md`](rules_of_engagement/STYLE.md) | React/Tailwind UI rules — layout, responsiveness, accessibility. |
| [`INFRASTRUCTURE.md`](rules_of_engagement/INFRASTRUCTURE.md) | Runtime boundaries, deployment, and the V1 trust model. |
| [`TESTS.md`](rules_of_engagement/TESTS.md) | The required verification matrix. |
| [`DOCS.md`](rules_of_engagement/DOCS.md) | How every feature must itself be documented. |
| [`FUTURE_IDEAS.md`](rules_of_engagement/FUTURE_IDEAS.md) | Deliberately deferred work (mobile client, auth, webhooks, editor integrations, …). |

[`AGENTS.md`](AGENTS.md) explains how these documents relate and take precedence for anyone (human or agent) working in this repository.
