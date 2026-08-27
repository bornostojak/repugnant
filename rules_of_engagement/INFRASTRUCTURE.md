# Infrastructure

The monorepo will have independently runnable CLI, backend, and web packages with shared contracts only where needed. The Go backend exposes the document, revision, organization, search, and documented-source APIs. The React app consumes those APIs. The CLI writes local docs and publishes staged revision payloads to the API.

Docker provides a backend/web deployment path through a Dockerfile and `compose.yaml`. SQLite persists in a mounted volume by default. PostgreSQL is an optional compose/configured service. Configuration comes from `rpg.conf.yaml` for project behavior and `.env`/environment variables for runtime secrets and connection configuration.

No authentication, public exposure hardening, Redis, background queue, or webhook processing is included in V1. The backend must log and document this trusted-network constraint. Design API and persistence boundaries so authentication, webhooks, and background publishing can be added without replacing core document identity/revision semantics.
