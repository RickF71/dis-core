MX Core: internal/core

This directory contains the MX series skeletons for the new `internal/core` tree.

Phase MX-1:
- Created empty package skeletons under `internal/core/*` to establish a parallel core tree.
- These are placeholders only and do not change runtime behavior.

Phase MX-2 / MX-2.1:
- Added `internal/core/authority` engine skeleton (Status, Lineage, Introspect, Freeze stubs).
- API handlers in `internal/api` have thin delegators that call into the new engine.

Developer notes:
- No production logic has been moved yet. Future phases will port code from `internal/api` and
  other packages into `internal/core` slice-by-slice.
- If you need the repo to build cleanly now, add compatibility shims or proceed with staged migrations.

MX: core tree placeholder
