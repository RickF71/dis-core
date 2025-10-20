## Quick orientation for AI coding agents

This repository is a Go-based implementation of the DIS core (Direct Individual Sovereignty). Use the notes below to be productive quickly — reference the files called out when making changes.

Key directories and entrypoints
- `cmd/dis-core/main.go` — primary CLI used during development. It connects to Postgres (env `DIS_DB_DSN`), loads schemas, validates domains and can `--freeze` core versions.
- `internal/schema/registry.go` — schema registry: `NewRegistry()`, `LoadDir(dir)`, `Verify(id,version)`, and `HashAll()` (deterministic sha256 of all registered schema hashes).
- `internal/receipts/receipt.go` — receipt structure and helpers: `NewReceipt(...)`, `ToJSON()`, `Save(dir)`. Payload ordering is stable and important for signature verification.
- `internal/ledger/postgres_store.go` — ledger persistence used by the CLI: `Open(db *sql.DB) *Store`, `InsertReceipt`, `ListReceipts`, `VerifyReceipt` (Postgres SQL statements here define table expectations).
- `internal/ledger/ledger.go` — higher-level ledger helpers / in-memory registry behavior used elsewhere.

Concrete workflows you can run or emulate
- Build everything quickly:
  - `go build ./...` (project uses Go modules; see `go.mod` for required versions).
- Run the CLI (reads `DIS_DB_DSN` or falls back to an insecure default DSN from code):
  - `DIS_DB_DSN="postgres://user:pw@host:5432/dis_core?sslmode=disable" go run cmd/dis-core/main.go --schemas=schemas --domains=domains`
- List receipts saved in the Postgres ledger (CLI):
  - `go run cmd/dis-core/main.go --list-receipts`
- Verify a specific receipt id (CLI):
  - `go run cmd/dis-core/main.go --verify-receipt=r-xxxxxxxx`
- Freeze a release (creates JSON receipt under `versions/<ver>/receipts` and attempts DB insert):
  - `go run cmd/dis-core/main.go --freeze=v0.9.7`

Important code patterns & conventions (do not change without checking callers)
- Schemas are YAML files that must include `meta.schema_id` and `meta.schema_version` in their frontmatter. `LoadDir` skips YAMLs missing those fields.
- Schema versions must begin with `v` and include a dot (example: `v0.1`). The loader enforces this and will return an error for invalid versions.
- `Registry.HashAll()` sorts keys and hashes ID+version+hash bytes to produce a deterministic fingerprint used in freeze receipts.
- Receipts sign a stable payload string created as: `fmt.Sprintf("%s|%s|%s|%s|%s|%s", by, action, timestamp, frozenCoreHash, consoleID, issuerSeat)` — maintain this ordering for compatibility.
- Domain validation is performed by `internal/domain` helpers (`domain.LoadAndValidate(file, reg)`); the CLI enumerates `domains/*.yaml` and calls that.

Integration & external dependencies to be mindful of
- Postgres: `cmd/dis-core` expects a Postgres server and schema (see SQL in `internal/ledger/postgres_store.go`). Use `DIS_DB_DSN` to point to the DB.
- Crypto/keys: receipts rely on `internal/crypto.EnsureDomainKeys()` which returns a signer; changes to signing must preserve public-key export (`Signer.Pub`) and base64 encoding used in receipts.
- Filesystem layout: freezing writes to `versions/<version>/receipts/<receipt-id>.json` — other components read those files in CI/release flows.

Where to look for examples when editing or adding features
- Adding a new CLI flag or behavior: modify `cmd/dis-core/main.go` (patterns for DB connect, registry load, and freeze are present).
- Persisting receipts: replicate SQL patterns from `internal/ledger/postgres_store.go` (note `ON CONFLICT (id) DO UPDATE` semantics).
- Schema parsing and hashing: follow `internal/schema/registry.go` (use `meta.schema_id` / `meta.schema_version` and `HashAll()` semantics).

Small safety notes for automated edits
- Avoid changing schema version format checks unless you update callers that assume `vX.Y` style.
- Keep the receipt payload ordering and hashing code intact; other subsystems (verifiers, stored receipts) depend on the exact bytes being signed.
- When running the CLI in automated tests, prefer a local Postgres test instance and set `DIS_DB_DSN` explicitly — the repository contains an insecure fallback DSN in code.

If anything above is unclear, tell me which section or file you want expanded and I will iterate on this guidance.
