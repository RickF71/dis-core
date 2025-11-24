# DIS Dimension Audit — null → terra Detection

This document describes the canonical dimension chain used in TAG/DIS and provides a short checklist for reviewing hits produced by the `tools/dimension_audit.sh` script.

## Canonical Dimension Chain (0–6)
- 0: void
- 1: (time vector / root axis) [currently named `null` in code]
- 2: aether
- 3: terra
- 4: numen
- 5: lima
- 6: corporeal

## Purpose
The goal of this audit is to locate files that may contain implicit or explicit assumptions that `null` maps directly to `terra` (or that the dimension mapping is hard-coded), and to flag places where adding an intermediate dimension (e.g., `aether`) could break behavior.

## Checklist for reviewing each hit
For every search result produced by `tools/dimension_audit.sh`, run through this checklist:

- Context: Is the file part of core logic, migrations, or UI?
  - Core logic: treat with high caution.
  - Migrations: ensure migration ordering and existing DB seeds are considered.
  - UI/docs: likely safe, but note for messaging/UX.
- Explicit parent/child check:
  - Does the code assume `null` is the parent of `terra`?
  - Is there a SQL `parent_id`, `parent_domain`, or hard-coded domain UID referencing this relation?
- Dimension constants:
  - Are there numeric constants mapped to domain names (e.g., `dim == 3` implies `terra`)?
  - Is there any code doing `dimension+1` to find the child? This will break if aether is inserted.
- Hard-coded spine sequences:
  - Any string arrays or comments that enumerate `null, terra, numen, ...`?
  - Any code deriving indices from position in those arrays?
- Impact assessment:
  - Is this UI-only or will changing it affect data/migrations?
  - If data-affecting, does it require a migration plan (backfill, reorder, or seeding)?

## Next Phase (Not Yet Implemented)
Planned future work once audit is complete:

- Insert `aether` as dimension 2 (between `null` and `terra`).
- Shift downstream dimensions accordingly (terra becomes 3, etc.).
- Update database migrations, spine endpoints (`/api/domain/spine`) and any code relying on numeric dimension offsets.
- Add integration tests to ensure the spine behaves correctly after rebase.

## Implementation Notes
- Canonical spine is implemented in Go at `internal/core/domain/spineconfig/spineconfig.go`.
- A migration to insert `aether` has been added: `db/migrations/20251124_insert_aether_between_null_and_terra.sql`.

## How to run the audit
From the repository root:

```sh
chmod +x tools/dimension_audit.sh
tools/dimension_audit.sh
```

Review results and follow the checklist above for each hit.
