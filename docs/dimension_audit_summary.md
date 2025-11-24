# DIS Dimension Audit — Triage Summary

Summary of findings from `dimension_audit.out`. Hits are grouped by risk (Critical → Very Low) with recommended next steps.

**Critical:**
- **Migrations / Bootstrap SQL:** `internal/testdb/testdb.go`, `internal/bootstrap/schema_bootstrap.go`, `internal/db/postgres_setup.go`, and several `db/migrations/*.sql` insert or set `terra` as a parent or seed `terra` directly — these will break if an intermediate `aether` dimension is introduced. Update migrations/bootstraps to be dimension-agnostic or include `aether` migration.
- **Hard-coded parent relationships:** Code that `SELECT`s or `INSERT`s with `WHERE name IN ('null','void','terra','numen','lima')` or resolves `null` → `terra` parent IDs (e.g. bedrock/bootstrap, null bootstrap logic). Change to resolve parent by a configurable spine or name aliasing.

**High:**
- **Domain creation & resolver logic:** `internal/core/domain/*` (create, resolver, spine_validator) relies on exact spine ordering and numeric `Dimension` offsets. Refactor to accept an insertable middle dimension and validate by name, not numeric offset.
- **Domain parent_id mutations in migrations:** `db/migrations/20251113_gov7_corporeal_lineage.sql` and related migrations set explicit `parent_id` values to terra-related UUIDs. These migrations need review and possible migration scripts to insert `aether` or remap parent IDs.

**Medium:**
- **Policy files & REGO:** `policy/*.rego` and `policy/terra.rego`, `policy/numen.rego`, `policy/lima.rego` import chains and reasoning assume the terra→numen→lima stack. Policy evaluation inputs and the policy loader should be audited and made robust to an added dimension.
- **Identity triad bootstrapping & tests:** GOV-1 triad code (terra/numen/lima) and tests (bootstrapping identity triads) auto-create or expect terra triads — update tests and bootstrapping to tolerate an inserted `aether` or to run under a config flag.

**Low:**
- **UI and CLI references:** `internal/api/*` forms, select options in admin UI and test scripts reference `terra` and `null` explicitly (dropdowns, `DOMAIN_ID` env defaults). Update UI lists to read canonical spine from config.
- **Docs & phases:** Numerous docs and phase notes (docs/, phases/) enumerate the spine; update documentation to show the new canonical chain and note migration implications.

**Very Low:**
- **Comments & examples:** README snippets, comments, and example bootstrap YAMLs referencing `terra` or the spine string arrays — low technical risk but useful for completeness.

**Top-priority next steps (recommended order):**
1. **Lock down migrations & bootstrap scripts**: Review and patch `db/migrations/*`, `internal/bootstrap/schema_bootstrap.go`, and `internal/testdb/testdb.go` so they do not hard-code `terra` as the immediate child of `null`. If the plan is to add `aether`, add a migration that creates `aether` and adjusts parent relationships transactionally.
2. **Make spine configurable at runtime**: Introduce a canonical spine source (YAML/DB table/config) and replace literal spine arrays and numeric dimension offsets with lookups. Update `internal/core/domain/dimensions.go` and `spine_validator` to consult the canonical spine.
3. **Policy compatibility pass**: Update REGO inputs and policy loader so that `terra`/`numen` imports remain valid when a new `aether` appears — or add a translation shim for policy input naming.
4. **Fix tests & test fixtures**: Update `internal/testdb/testdb.go`, unit tests, and bootstrap test helpers to seed the new spine or be resilient to missing layers. Run full test suite early and often.
5. **Create a migration strategy**: Plan a data migration to insert `aether` where appropriate or to remap `parent_id` fields safely. Consider a compatibility flag and a carefully logged transitional migration run.
6. **Update tooling & docs**: Regenerate docs and update `tools/dimension_audit.sh` to output JSON for programmatic triage; update `docs/DIS_DimensionAudit.md` with checklist outcomes.

**Files to review first (extracted examples found in `dimension_audit.out`):**
- `internal/testdb/testdb.go` (seeded `terra` INSERTs)
- `internal/bootstrap/schema_bootstrap.go` (bedrock init receipts, spine arrays)
- `internal/db/postgres_setup.go` (seeded domains)
- `internal/core/bedrock/bootstrap.go` and `internal/core/domain/null_bootstrap.go` (null/terra create logic)
- `db/migrations/20251113_gov7_corporeal_lineage.sql` (parent_id remapping)
- `internal/core/domain/spine_validator.go` and `internal/core/domain/dimensions.go` (dimension logic)
- `policy/terra.rego`, `policy/numen.rego`, `policy/lima.rego` (policy imports & assumptions)

**Quick mitigation patterns:**
- Add defensive checks: `SELECT EXISTS(...)` before relying on exact domain names.
- Replace numeric `Dimension` offsets with name-based checks (e.g., `IsParentAllowed(parentName, childName)` via config lookup).
- Add a compatibility layer that maps legacy spine arrays to the current runtime spine for tests and migrations.

If you want, I can: (a) create a PR skeleton patching the critical migrations and bootstraps to use a configurable spine, (b) generate a JSON-format audit output by updating `tools/dimension_audit.sh`, or (c) start updating the top 3 critical files listed above. Which should I do next?

Generated from `dimension_audit.out` run on repository root; review recommended files before applying any data migrations.
