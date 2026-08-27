# Documentation rules

Every implemented feature must be documented after implementation: purpose, user workflow, configuration, API/CLI behavior, operational considerations, failure modes, tests, and maintenance notes. Keep root `README.md` concise enough to onboard a user and link to detailed package/configuration documentation.

The scaffold establishes `cmd/rpg`, `cmd/rpg-server`, `internal/httpapi`, and `web`. `rpg init` creates the commented configuration, ignored staging directory, and portable hooks; `rpg --init-hooks` repairs only hooks and chains a preserved existing hook through its `.rpg-backup` copy. Current verification commands are recorded in the root README; subsequent feature documentation must replace scaffold-only instructions with user-facing setup and configuration guidance.

When behavior or design changes, update all relevant documents: specification, taste, style, infrastructure, tests, future ideas, README, and feature docs. Examples must be copyable and use the current `rpg`/`rpg.conf.yaml` vocabulary. Clearly label unauthenticated web editing as trusted-network-only.
