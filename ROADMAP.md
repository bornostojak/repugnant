# Implementation roadmap

This file is the delivery checklist. A feature is not complete merely because a handler, page, or command exists: its CLI/API/UI path, configuration, documentation, tests, and relevant example coverage must work together.

## 1. Foundation and governance

- [x] Monorepo scaffold, Go CLI/backend, React/TypeScript/Tailwind workspace.
- [x] Logging/config conventions, Docker foundation, documentation governance.
- [ ] Docker Compose runtime validation and production static-web serving.

## 2. Local documentation workflow

- [x] Parse baseline standalone/quote markers for supported extensions.
- [x] `rpg init`, config creation, `.rpg` workspace, and hook installation.
- [x] Generate initial local Markdown and stable opaque source IDs.
- [x] Parse category/title/tags deterministically and render them in Markdown.
- [x] Persist source quote snapshots; block changed quotes until an explanatory revision is supplied.
- [x] Append revisions without overwriting user local edits; preserve metadata/history.
- [x] Stage generated docs correctly in pre-commit.

## 3. Project-scoped publishing backend

- [x] SQLite storage and preliminary PostgreSQL driver/migration support.
- [x] Project creation API, generated project API key, authenticated article create API.
- [x] Project-scoped article URL and opaque short-link redirect.
- [ ] Project listing/detail, categories, tags, source-file, search, and revision APIs.
- [x] Article revision persistence and idempotent publishing.
- [ ] Verify PostgreSQL contract with an actual database.

## 4. CLI publishing workflow

- [x] `rpg project create` creates a server project and prints safe config guidance.
- [ ] `rpg generate` fully supports current/revision annotations and output selection.
- [x] `rpg push` manually publishes generated documents to configured project API.
- [x] Pre-push uses the same publisher, retains pending diagnostics, prompts interactively, and blocks by default non-interactively.

## 5. Web wiki

- [x] Serve the built React application from the backend; provide a navigable homepage.
- [x] Project creation/listing and project workspace UI.
- [x] Documentation tree/category navigation, title/body/tag search, and article reading view.
- [ ] Documented-source browser with source ranges (source path is present; ranges need persistence).
- [x] Revision/history sidebar and organization management.
- [ ] Responsive, keyboard-accessible Chromium-verified UI.

## 6. Deployment, documentation, and validation

- [x] Complete `.env`, Dockerfile, Compose, SQLite volume, and PostgreSQL configuration.
- [ ] Update all relevant governance docs and user README/configuration documentation.
- [x] Keep ten versioned language examples under `tests/examples`.
- [ ] Convert the ten examples into automated temporary-Git integration tests including commit, generation, manual push, server validation, and browser UI checks (Git/generation/push are automated; browser check remains).
- [ ] Run full test/vet/build suite and live end-to-end smoke tests before declaring delivery complete.

## Cycle review

At the end of each feature cycle, answer twice: “Did I fulfill every requested behavior across backend, frontend, CLI, hooks, docs, and tests?” Then identify one scoped improvement that reduces a real usability, reliability, or consistency gap. Update `PROGRESS.md` with evidence.
