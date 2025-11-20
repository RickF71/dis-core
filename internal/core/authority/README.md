MX-2 / MX-3 authority core notes

MX-3.1: Migrated Authority Status logic into core engine. Legacy status paths will be removed in MX-3.4.

MX-3.2: Migrated Introspect logic into authority core engine.

Notes:
- This directory contains lightweight placeholders and delegators for the phased
  migration from `internal/authority` into `internal/core/authority`.

MX-3.3: Added lineage engine structure + API wiring. 
Real ancestry/freeze/policy lineage logic to be migrated in MX-3.4.
# internal/core/authority

This package contains the new core authority engine used during the MOAR-X migration.

## MX-3

MX-3.1: Migrated Authority Status logic into core engine. Legacy status paths will be removed in MX-3.4.

Further MX-3.x work will migrate freeze/lineage/introspection internals and remove compatibility shims.
