# Engineering taste

Keep rePugnant small, deterministic, and editor-agnostic. Build vertical slices across CLI, API, and web UI after shared scaffolding. Use Go for CLI/backend, explicit interfaces around storage and publishing, and a SQLite/PostgreSQL repository implementation without a needless registry pattern elsewhere. No Redis in V1: expected scale does not justify it.

Use YAML for project configuration and `.env`/environment variables for container/backend settings. Validate config early and return actionable errors. Never silently rewrite source except the documented annotation-to-stable-reference normalization; make generated diffs predictable.

Logging is structured and level-selectable (`error`, `warn`, `info`, `debug`). Default test runs minimize output. Log operation IDs, article IDs, paths, hook phases, and failures; do not emit whole source files, secrets, or noisy per-line traces at normal levels.

Avoid clever parsing based solely on regular expressions. Use language comment rules plus a line-oriented state machine that preserves source. Keep IDs opaque and stable. Prefer idempotent commands and portable POSIX-compatible hook scripts where practical.
