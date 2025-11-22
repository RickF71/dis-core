# internal/schema — DIS-Core schema structs & validators (v0.9.5)

This package defines Go structs and validation logic for **Authorization Domains (DomAuth)** and the **Shared Core Triad** used to instantiate new domains.

## Contents
- `domauth_definition.go` — Base DomAuthDefinition
- `domauth_single.go` — DomAuthSingle
- `domauth_collective.go` — DomAuthCollective
- `domauth_delegated.go` — DomAuthDelegated
- `triad_shared_core.go` — Jikka Shared Core Triad
- `validate.go` — Common validation helpers
- `register.go` — Optional registry hook (no-op if unused)

All structs expose `Validate() error` and carry both `json` and `yaml` tags.

## Testing notes

- Use a dedicated test database for integration tests. Set the environment variable `DIS_TEST_DB_DSN` to a Postgres DSN pointing at the test database before running `go test` so destructive operations run in an isolated DB. Example:

```
export DIS_TEST_DB_DSN='postgres://username:password@localhost:5432/dis_test?sslmode=disable'
go test ./... -v -count=1
```

- Defensive integration tests: some integration tests perform schema-destructive operations (DROP TABLE). When these operations fail because other objects depend on the table (Postgres SQLSTATE `2BP01`, or error text containing "depends on it"), the test suite will skip that specific destructive test rather than fail the whole run. This allows running the suite safely against shared or constrained test environments, but for full validation use an isolated DB.

If you'd like, I can add a short `docs/TESTING.md` with these notes and a few examples (creating a local test DB, reset commands, and a CI snippet).
