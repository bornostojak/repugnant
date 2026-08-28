# Product specification

## V1 scope

The monorepo will contain a Go CLI (`rpg`), Go HTTP backend, React/TypeScript/Tailwind web application, Dockerfile, and `compose.yaml`. A mobile client is deferred. SQLite is the default database; PostgreSQL is an immediate selectable alternative.

`rpg init` creates a commented `rpg.conf.yaml`, an ignored project-local `.rpg/` workspace, and portable Git pre-commit/pre-push hooks. `rpg --init-hooks` performs only hook installation/repair.

## Configuration

`rpg.conf.yaml` is the project configuration. It controls local `/docs` output, optional web endpoint/output, enabled languages (or auto-detection), database/backend settings where relevant, and pre-push failure policy. Environment variables override backend/container behavior using standard `.env` syntax. The default is SQLite and block-on-publish-failure.

Language auto-detection uses project evidence such as `go.mod`/`go.sum`, `package.json`/`tsconfig.json`, `Cargo.toml`, and equivalent language manifests. Explicit `langs: [go, ruby]` overrides it. V1 supports Go, TypeScript, JavaScript/React, Svelte, Python, Rust, C/C++, Bash, Dart, Groovy/Jenkinsfiles, YAML, QML, and Ruby through configurable comment conventions.

## Annotation grammar

Comment prefixes are language-specific. Examples below use `//`; the marker appears after the comment prefix and optional whitespace.

```text
// $rPg: Article title
// $~ Markdown belonging to the standalone article

// ?rPg: Article title for a code quote
// ?~ Markdown belonging to the quote article
... exact code to quote ...
// !rPg
```

`$~` only continues a standalone `$rPg` article. `?~` only continues a quote article. A quote begins at `?rPg` and ends at `!rPg`; both are required. The quoted code is the exact intervening source, excluding rpg annotation lines. Its relative indentation is preserved; only common leading indentation that is safe to remove is removed. Parsing must not corrupt heredocs, YAML indentation, QML, or language-specific block structure.

Generation assigns a permanent opaque, URL-safe alphanumeric article ID (for example `s9Aa3A3al`), never a title slug. After generation the marker collapses to a clean, minimal reference: the display title on the `rPg:` line, followed by one `~ {backlink}` line per configured output — the web permalink (`{endpoint}/a/{id}`) first when web publishing is enabled, then the article's Markdown file as a path relative to the source file (so `gf` opens it). The `$~`/`?~` prose is **moved** into the article, not left in the source, and for a quoted range the terminating `!rPg` is removed while the quoted code stays in place. The ID is recoverable from either backlink (`.../a/{id}` or `.../{id}.md`).

```text
// rPg: Cache resolution strategy
// ~ http://127.0.0.1:8080/a/s9Aa3A3al
// ~ ../../docs/s9Aa3A3al.md
```

Because `!rPg` is removed, a generated quote is a one-time snapshot: rpg no longer detects when that code later drifts from its documentation. Documentation is kept current by re-running generation and by explicit revisions, not by a drift block.

To append a revision, keep the generated marker in place and write `$#`/`$~` lines under its backlink — the article ID is recovered from the backlink, so it never has to be retyped:

```text
// rPg: Cache resolution strategy
// ~ ../../docs/s9Aa3A3al.md
// $# striped locks          (optional subtitle for this revision)
// $~ New explanatory Markdown
```

Generation appends a new `# Revision N — {subtitle}` section to the article (keeping earlier revisions), then restores the source to the clean marker. `$#` is the revision's subtitle only; it labels Revision N and does **not** change the article's permanent title. The equivalent explicit form `$rPg@{id}: subtitle` + `$~ …` is still accepted, where the text after the ID is likewise the revision subtitle. The ID never changes; prior revisions remain history.

## Generation and hooks

Pre-commit parses enabled source files, regenerates configured local Markdown, stages generated `/docs` changes when applicable, and writes pending web revisions to `.rpg/`. It blocks a commit when a tracked quote changed without a matching `$rPg@{id}` update. Its diagnostic identifies file, line/range, article ID, reason, and the exact annotation to add.

Pre-push publishes pending revisions to the configured backend. If publishing fails, an interactive hook asks whether to continue and keep the item pending/pre-approved for a later hook run. Without interactive input, the default is to block the push. Configuration may explicitly allow the push while retaining pending data.

Local docs are official output when selected and should be committed. They remain manually editable in V1; rpg appends later generated revisions and does not reconcile edits with web content. When web output is selected, source references are committed but server content is not. Local and web content are intentionally not synchronized after manual editing in V1.

## Document model and web behavior

Generated Markdown begins with title and a compact metadata table: author name/email from Git config by default, generation timestamp, revision number, and web link if configured. It contains revision sections appended in order. Do not use commit hashes.

The backend stores article identity, current organization, revisions, source file/range information, tags, author identity, timestamps, and quoted-code snapshots. URLs are `/a/{article-id}` for latest and `/a/{article-id}/{revision}` for a historical revision. Each article revision is retained.

V1 web access is unauthenticated and must be documented as suitable only for trusted local/private networks; it is not safe to expose publicly. The web UI provides a category tree, title/full-text/tag search, and a VS Code-like browser limited to files containing rpg documentation. It supports organization of categories/groups/tags and a readable revision history. Later web-only manual revision branching is deferred.
