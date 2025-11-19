# GOV-8: Name Non-Authority & UUID Canonicalization

**DIS-Invariant-001:** No authority, lookup, or receipt may rely on domain names.

## Core Principle

Domain names are **display metadata only**. All internal operations, authorization decisions, receipt generation, and policy evaluation MUST use `domain.id` (UUID) as the single canonical identity.

## Rationale

1. **Stability**: Names can change; UUIDs cannot
2. **Security**: Names are user-controlled; UUIDs are system-assigned
3. **Auditability**: UUID-based lineage is immutable and cryptographically traceable
4. **Performance**: UUID indexes are faster than text-based name lookups
5. **Disambiguation**: Names can collide across contexts; UUIDs are globally unique

## Acceptable Uses of Names

### ✅ Acceptable (Display Only)
- Logging and debugging output
- UI display labels
- Human-readable error messages
- Documentation and comments
- Audit trail annotations (alongside UUID)

### ❌ Prohibited (Authority Operations)
- Domain lookup by name (`WHERE name = ?`)
- Authorization decisions based on name
- Receipt generation using name as identifier
- Policy evaluation keyed by name
- Cross-domain references by name

## Implementation

### UUID-Only Domain Resolution

```go
import "dis-core/internal/domain"

// CORRECT: UUID-based lookup
domainID := uuid.MustParse("a1111111-1111-1111-1111-111111111111")
valid, err := domain.ValidateDomainID(ctx, pool, domainID)

// WRONG: Name-based lookup
domainName := "terra.numen.lima.corporeal"
// DON'T: SELECT id FROM domains WHERE name = ?
```

### Display-Only Name Resolution

```go
// For logging/display ONLY
fdn, err := domain.ResolveDomainFDN(ctx, pool, domainID)
log.Printf("Processing domain %s (%s)", domainID, fdn)
```

### Lineage Tracking

```go
// UUID-based lineage (canonical)
lineage, err := domain.ResolveDomainLineage(ctx, pool, domainID)
// Returns: [uuid1, uuid2, uuid3, ...]

// Human-readable lineage (display only)
fdnLineage, err := domain.ResolveDomainLineageFDN(ctx, pool, domainID)
// Returns: ["terra", "numen", "lima", "corporeal"]
```

## Migration Guide

### Before (GOV-7 and earlier)
```go
// DEPRECATED: Name-based lookup
domainID, err := resolveDomainInternalID(ctx, "user.rick")
```

### After (GOV-8)
```go
// CORRECT: UUID-based lookup
domainID := uuid.MustParse("52e6c580-da99-4acd-b456-6dc8e760e5f7")
exists, err := domain.ValidateDomainID(ctx, pool, domainID)

// For display only
fdn, _ := domain.ResolveDomainFDN(ctx, pool, domainID)
// fdn = "user.rick"
```

## Authority Schema Updates

### Domain Structure
```json
{
  "id": "a1111111-1111-1111-1111-111111111111",
  "name": "terra.numen.lima.corporeal",
  "name_warning": "GOV-8: Name is display metadata only. Use ID for all operations.",
  "parent_id": "4daf928e-e58c-454e-8395-f3dedd103dde",
  "governance": "human_sovereign"
}
```

### Receipt Format
```json
{
  "receipt_id": "uuid-...",
  "event_id": "uuid-...",
  "domain_id": "a1111111-1111-1111-1111-111111111111",
  "domain_fdn": "terra.numen.lima.corporeal",
  "lineage_uuid": [
    "4daf928e-e58c-454e-8395-f3dedd103dde",
    "a1111111-1111-1111-1111-111111111111"
  ],
  "lineage_fdn": ["terra.numen.lima", "terra.numen.lima.corporeal"],
  "note": "lineage_fdn is display metadata only"
}
```

## Verification

### Regression Tests
Run GOV-8 test suite to verify compliance:

```bash
go test -v ./internal/domain/... -run TestNoNameBasedQueries
```

### Code Audit
Search for violations:

```bash
# Find name-based domain lookups (violations)
grep -r "WHERE name\s*=" --include="*.go" internal/

# Find UUID-based lookups (acceptable)
grep -r "WHERE id\s*=" --include="*.go" internal/
```

### Database Queries
```sql
-- VIOLATION: Name-based lookup
SELECT id FROM domains WHERE name = 'user.rick';

-- ACCEPTABLE: UUID-based lookup
SELECT id FROM domains WHERE id = 'a1111111-1111-1111-1111-111111111111';

-- ACCEPTABLE: Display-only name retrieval
SELECT name FROM domains WHERE id = 'a1111111-1111-1111-1111-111111111111';
```

## Exceptions

### Bootstrap Operations
During initial system bootstrap, name-based lookups MAY be used to establish UUID mappings. These operations:

1. MUST log warnings
2. MUST occur only during bootstrap phase
3. MUST NOT be used in runtime authority operations
4. MUST document the temporary violation with GOV-8 comments

Example:
```go
// GOV-8: Bootstrap exception - establishing UUID mapping
var parentID *uuid.UUID
err := pool.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'simula'`).Scan(&parentID)
// This is acceptable ONLY during bootstrap
log.Printf("⚠️  GOV-8: Bootstrap used name lookup for 'simula'")
```

## Future Work

### GOV-9: Name Deprecation Timeline
- Phase 1 (GOV-8): Add deprecation warnings
- Phase 2 (GOV-9): Remove name-based lookup functions
- Phase 3 (GOV-10): Make name column nullable
- Phase 4 (GOV-11): Remove name column entirely (display in payload only)

## References

- DIS-Invariant-001: UUID Canonical Identity
- GOV-7: Corporeal Lineage & Root Seat Establishment
- Phase 10: Receipt Verification & Provenance Continuity

## Compliance Checklist

- [x] All domain lookups use UUID
- [x] Names retrieved only for display
- [x] Lineage tracking uses UUID arrays
- [x] Receipts include UUID-based lineage
- [x] Policy evaluation keyed by UUID
- [x] Regression tests enforce invariant
- [x] Documentation updated
- [x] Bootstrap exceptions documented
