## Quick orientation for AI coding agents

This repository is a Go-based implementation of the DIS core (Direct Individual Sovereignty). Use the notes below to be productive quickly — reference the files called out when making changes.

Key directories and entrypoints
- `cmd/dis-core/main.go` → `internal/app/run.go` — primary service entrypoint. Connects to Postgres (env `DIS_DB_DSN`), initializes schema registry, ledger, policy engine, and HTTP API server.
- `internal/api/server.go` + `internal/api/routes.go` — REST API server with endpoints like `/api/ping`, `/api/domain/*`, `/api/status`. Uses `http.ServeMux` for routing.
- `internal/schema/registry.go` — schema registry: `NewRegistry()`, `LoadDir(dir)`, `Verify(id,version)`, and `HashAll()` (deterministic sha256 of all registered schema hashes). Loads from `disyaml/schemas/`.
- `internal/ledger/ledger.go` + `postgres_store.go` — ledger with PostgreSQL persistence. Main methods: `Open()`, `StoreCanon()`, `BootstrapDomains()`, receipt management.
- `internal/bootstrap/` — table creation and YAML import logic. `BootstrapAllTables()` creates database schema, `ImportYAML()` loads initial data.
- `internal/domain/loader.go` — domain YAML parsing and validation against schema registry.

Multiple executable entrypoints
- `cmd/dis-core/` — main HTTP API server (port 8080 by default)
- `cmd/dis-webd/` — web server with freeze/receipt CLI flags + API server
- `cmd/console_server/` — console management server with action logging
- `cmd/dis-netd/` — peer-to-peer network daemon
- `dis-core` root executable (`main_legacy2.go`) — runs `cmd/dis-core` by default

Concrete workflows you can run or emulate
- Build and run main server:
  - `go run ./cmd/dis-core` (reads `DIS_DB_DSN` or uses default postgres://dis_user:card567@localhost:5432/dis_core)
- Run web server with CLI features:
  - `go run ./cmd/dis-webd --schemas=disyaml/schemas --domains=disyaml/domains --freeze=v0.9.7`
- List/verify receipts:
  - `go run ./cmd/dis-webd --list-receipts --verify-receipt=r-xxxxxxxx`
- Test API endpoints:
  - `curl http://localhost:8080/api/ping`, `curl http://localhost:8080/api/status`

Important code patterns & conventions (do not change without checking callers)
- Schemas are YAML files in `disyaml/schemas/` that must include `meta.schema_id` and `meta.schema_version` in their frontmatter. `LoadDir` skips YAMLs missing those fields.
- Domains are YAML files in `disyaml/domains/` following the pattern `domain.<subject>.yaml` with schema references and validation.
- Schema versions must begin with `v` and include a dot (example: `v0.1`). The loader enforces this and will return an error for invalid versions.
- `Registry.HashAll()` sorts keys and hashes ID+version+hash bytes to produce a deterministic fingerprint used in freeze receipts.
- Canon table stores JSON documents with `type` and `content` columns. API queries use PostgreSQL JSON operators like `content->'meta'->>'domain_id'`.
- Bootstrap process: tables → schema loading → domain loading → policy engine → API server startup (see `internal/app/run.go`).
- HTTP API uses `http.ServeMux` with function-based handlers. Routes are registered in `internal/api/routes.go`, handlers in separate files like `domain_get.go`.

Integration & external dependencies to be mindful of
- PostgreSQL: All commands expect a Postgres server. Use `DIS_DB_DSN` env var or fallback DSN `postgres://dis_user:card567@localhost:5432/dis_core?sslmode=disable`.
- Database schema: Bootstrap creates tables for `canon`, `domains`, `schemas`, `policies`, `mirror_events`, `peers`, etc. See `internal/bootstrap/schema_bootstrap.go`.
- Multiple servers: `dis-netd` (peer networking), `console_server` (management), `dis-webd` (web+CLI), `dis-core` (main API) can run simultaneously on different ports.
- Policy engine: Uses Open Policy Agent (OPA) for authorization. Policies loaded from `./policies` directory.
- File structure: `disyaml/` contains schemas and domains; `versions/` contains release receipts; config in `config.yaml`.

Where to look for examples when editing or adding features
- Adding new API endpoints: follow pattern in `internal/api/routes.go` for registration, create handler files like `domain_get.go`.
- Database operations: see patterns in `internal/ledger/postgres_store.go` and bootstrap table creation in `internal/bootstrap/`.
- Schema/domain processing: see `internal/schema/registry.go` and `internal/domain/loader.go` for YAML parsing and validation patterns.
- Config management: extend `internal/config/config.go` for new settings, used throughout via dependency injection.

Small safety notes for automated edits
- The bootstrap sequence in `internal/app/run.go` has dependencies: config → DB → schemas → ledger → domains → policy → API. Don't reorder.
- Canon table stores raw JSON - preserve exact JSON structure when writing queries with PostgreSQL JSON operators.
- Multiple commands share the same database schema but have different entrypoints - changes to table structure affect all commands.
- Schema version format checks are enforced in multiple places - keep `vX.Y` format consistent across the codebase.

If anything above is unclear, tell me which section or file you want expanded and I will iterate on this guidance.
