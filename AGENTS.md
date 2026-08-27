# rePugnant agent instructions

Read this file before working in this repository. The project is a monorepo; do not make assumptions that a change is isolated to one package.

## Source of truth

- `rules_of_engagement/VISION.md` explains the user outcome.
- `rules_of_engagement/SPEC.md` is the functional and deterministic product contract.
- `rules_of_engagement/TASTE.md` governs architecture, configuration, logging, and consistency.
- `rules_of_engagement/STYLE.md` governs the React web UI.
- `rules_of_engagement/INFRASTRUCTURE.md` governs runtime boundaries and deployment.
- `rules_of_engagement/TESTS.md` is the required verification matrix.
- `rules_of_engagement/DOCS.md` governs user and maintainer documentation.
- `rules_of_engagement/FUTURE_IDEAS.md` records deliberately deferred work.

Resolve conflicts in that order, except that explicit user instructions always win. Update **every relevant document above** when a design choice, behavior, interface, architecture boundary, or operational assumption changes. Do not implement code until the user explicitly authorizes implementation after reviewing prep.

## Working conventions

Use feature branches once Git is initialized; test a feature before committing it, then merge it into `main` using `git merge --no-ff`. Keep logs useful and level-controlled. Prefer simple, explicit code over speculative abstractions. Respect generated/local documentation ownership rules in `SPEC.md`.
