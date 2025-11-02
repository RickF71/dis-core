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
