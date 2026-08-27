# Delivery progress

Status: **In progress — not ready for review or release.**

| Area | Status | Evidence / outstanding work |
|---|---|---|
| CLI initialization | Partial | `rpg init` and hook installation work; hooks need complete staging/publishing behavior. |
| Local documentation | Partial | Initial examples generate Markdown; revision, drift detection, metadata identity, and correct staging are incomplete. |
| Backend | Partial | Project/API-key/article/short-link prototype exists; wiki data model and APIs are incomplete. |
| Frontend | Not started as product UI | A placeholder React shell exists but is not served or navigable. |
| Deployment | Partial | Go server can listen on all interfaces; Compose runtime unavailable here and static frontend serving is not connected. |
| Examples | Partial | Ten source fixtures exist and temporary commits were manually verified; no committed automated integration harness yet. |
| Manual server push | Not implemented | `rpg push` deliberately reports unavailable; it must be implemented before this status can advance. |

## Latest verified work

- Go package tests passed after adding parser/generator foundation and SQLite API tests.
- Ten copied fixtures were initialized as temporary Git repositories and each produced one local Markdown file through its pre-commit hook.
- Live backend smoke-tested project creation, API-key article publishing, rendered article route, and short-link redirect.

These checks do **not** constitute full delivery. See `ROADMAP.md` for remaining required work.
