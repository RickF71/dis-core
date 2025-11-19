# GOV-10: Identity Provenance & Alias Receipt Integration

**Implementation Date**: November 14, 2024
**Status**: ✅ Complete
**Tests**: 7/7 PASS

## Overview

GOV-10 extends the immutable receipt ledger from GOV-9 (Authority Continuity) to cover identity operations, creating comprehensive auditability for both authority decisions and identity governance actions.

## Architecture

### Core Components

1. **Identity Receipt Store** (`internal/identity/identity_receipts.go`)
   - Immutable receipt recording with SHA-256 hash chains
   - Automatic prev_id chaining for chronological integrity
   - Support for 5 identity action types

2. **Identity Lineage Retrieval** (`internal/identity/identity_lineage.go`)
   - Chronological lineage reconstruction per actor_id
   - Hash chain integrity verification
   - Aggregated integrity status (valid/broken/root counts)

3. **Database Schema** (`db/migrations/20251114_gov10_identity_receipts.sql`)
   - identity_receipts table with UUID primary key
   - Foreign key constraints for referential integrity
   - Indexes for efficient lineage queries
   - identity_lineage_view for easy querying

4. **API Endpoints** (`internal/api/handlers_identity.go`)
   - GET /api/identity/lineage/{actor_id} - Full lineage with payloads
   - GET /api/identity/lineage/{actor_id}/verify - Integrity summary only

5. **UI Component** (`finagler/src/components/IdentityLineage.jsx`)
   - Visual integrity badges (✓ valid, ✖ broken, ⚑ root, ⚠ missing)
   - JSON export for audit reports
   - Real-time integrity statistics

6. **Mutation Engine Integration** (`internal/authority/mutation_engine.go`)
   - Automatic receipt emission on seat transitions
   - identity.root.v1 for new identity bindings (unbound → bound)
   - identity.binding.update.v1 for state changes

## Identity Actions

| Action | Description | Use Case |
|--------|-------------|----------|
| identity.root.v1 | Root identity establishment | First receipt for an actor |
| identity.alias.add.v1 | Alias addition | Adding alias to existing identity |
| identity.alias.remove.v1 | Alias removal | Removing alias from identity |
| identity.convert.v1 | notech→dis conversion | Legacy system migration |
| identity.binding.update.v1 | Binding update | Domain association changes |

## Database Schema

```sql
CREATE TABLE identity_receipts (
    id UUID PRIMARY KEY,
    domain_id UUID NOT NULL REFERENCES domains(id),
    actor_id UUID NOT NULL,
    action TEXT NOT NULL,
    payload JSONB NOT NULL,
    prev_id UUID REFERENCES identity_receipts(id),
    hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    consent_by UUID NOT NULL REFERENCES domains(id),
    alias_scope TEXT
);

-- Indexes for efficient queries
CREATE INDEX idx_identity_receipts_actor_created ON identity_receipts(actor_id, created_at ASC);
CREATE INDEX idx_identity_receipts_domain ON identity_receipts(domain_id);
CREATE INDEX idx_identity_receipts_action ON identity_receipts(action);
CREATE INDEX idx_identity_receipts_prev ON identity_receipts(prev_id);
```

## Hash Chain Integrity

### Hash Computation
```go
hash = SHA-256(JSON(payload) + prev_id)
```

### Chain Verification Rules
1. **Root Receipt**: First receipt must have `prev_id = NULL`
2. **Chained Receipt**: Subsequent receipts must have `prev_id` pointing to previous receipt's ID
3. **FK Protection**: Database enforces prev_id references valid receipt IDs

### Integrity Status
- **TotalReceipts**: Count of all receipts for actor
- **ValidChains**: Count of receipts with correct prev_id links
- **BrokenChains**: Count of receipts with broken prev_id links (prevented by FK)
- **RootReceipts**: Count of receipts with NULL prev_id (should be 1)
- **Issues**: Array of integrity violation descriptions

## API Examples

### Get Full Lineage
```bash
curl -X GET http://localhost:8080/api/identity/lineage/{actor_id}
```

Response:
```json
{
  "actor_id": "123e4567-e89b-12d3-a456-426614174000",
  "entries": [
    {
      "id": "...",
      "prev_id": null,
      "hash": "abc123...",
      "valid": true,
      "action": "identity.root.v1",
      "created_at": "2024-11-14T10:00:00Z",
      "domain_id": "...",
      "consent_by": "...",
      "payload": {...}
    }
  ],
  "integrity": {
    "total_receipts": 3,
    "valid_chains": 2,
    "broken_chains": 0,
    "root_receipts": 1,
    "issues": []
  }
}
```

### Verify Integrity Only
```bash
curl -X GET http://localhost:8080/api/identity/lineage/{actor_id}/verify
```

Response:
```json
{
  "status": "ok",
  "actor_id": "123e4567-e89b-12d3-a456-426614174000",
  "integrity": {
    "total_receipts": 3,
    "valid_chains": 2,
    "broken_chains": 0,
    "root_receipts": 1,
    "issues": []
  }
}
```

## Mutation Engine Integration

When seat transitions occur, the mutation engine automatically records:

1. **Authority Receipt** (GOV-9):
   - Domain-level governance action
   - seat.transition.v1 action type
   - Includes CAS version, decision ID

2. **Identity Receipt** (GOV-10):
   - Actor-level identity action
   - identity.root.v1 (unbound → bound)
   - identity.binding.update.v1 (state changes)
   - Chains to previous identity receipt

Example log output:
```
✅ GOV-9: Authority receipt recorded: abc123 (hash: def45678)
✅ GOV-10: Identity receipt recorded: identity.root.v1
```

## Test Coverage

### Test Suite (`internal/identity/identity_receipts_test.go`)

1. **TestRecordIdentityReceipt**: Basic receipt creation and hash computation
2. **TestIdentityReceiptChaining**: Automatic prev_id chaining
3. **TestGetIdentityLineage**: Lineage retrieval and chronological ordering
4. **TestIdentityHashChainIntegrity**: Hash chain validation across 3 receipts
5. **TestIdentityLineageIntegrityStatus**: FK constraint protection verification
6. **TestIdentityAliasScope**: Alias scope field handling
7. **TestIdentityHashDeterminism**: Deterministic hash computation

**Results**: 7/7 PASS (all tests passing)

## Security Features

1. **Immutability**: Receipts are INSERT-only (no UPDATE/DELETE)
2. **Hash Chains**: Cryptographic linking prevents tampering
3. **FK Constraints**: Database-level prevention of broken chains
4. **Consent Tracking**: Records which domain authorized each action
5. **Audit Trail**: Complete chronological history per actor

## Comparison with GOV-9

| Feature | GOV-9 (Authority) | GOV-10 (Identity) |
|---------|-------------------|-------------------|
| **Scope** | Domain-level governance | Actor-level identity |
| **Keyed By** | domain_id | actor_id |
| **Actions** | seat.transition, seat.freeze, seat.unfreeze | identity.root, identity.alias.*, identity.convert, identity.binding.update |
| **Extra Fields** | policy_digest | consent_by, alias_scope |
| **Use Case** | Authority decisions | Identity provenance |

## Files Created

1. `internal/identity/identity_receipts.go` (179 lines)
2. `internal/identity/identity_lineage.go` (224 lines)
3. `internal/api/handlers_identity.go` (60 lines)
4. `db/migrations/20251114_gov10_identity_receipts.sql` (79 lines)
5. `finagler/src/components/IdentityLineage.jsx` (410 lines)
6. `internal/identity/identity_receipts_test.go` (422 lines)

## Files Modified

1. `internal/api/routes.go` - Added identity lineage endpoints
2. `internal/authority/mutation_engine.go` - Added identity receipt recording

## Migration Applied

```bash
psql "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable" \
  -f db/migrations/20251114_gov10_identity_receipts.sql
```

Output:
```
✅ GOV-10 — Identity Provenance & Alias Receipt Integration
    Identity receipts table created with hash chain integrity
    Existing receipts: 0 across 0 actors
    Lineage view: identity_lineage_view (root/valid/broken status)
    Covers: alias management, notech→dis conversion, binding updates
```

## Build Verification

```bash
go build -o dis-core-gov10 cmd/dis-core/main.go
```

**Status**: ✅ Build successful (no errors)

## Future Enhancements

1. **Alias Management API**: Dedicated endpoints for alias add/remove operations
2. **Notech Conversion**: Migration tool for legacy system identities
3. **Multi-Actor Queries**: Batch lineage retrieval for multiple actors
4. **Lineage Export**: PDF/CSV export for compliance reporting
5. **Real-time Monitoring**: WebSocket streaming of identity events
6. **Receipt Pruning**: Archival strategy for old receipts (with hash preservation)

## Related Work

- **GOV-9**: Authority Continuity & Receipt Lineage (domain-level governance)
- **GOV-8**: UUID Canonicalization (identity normalization)
- **GOV-7**: Corporeal Lineage (ownership tracking)
- **GOV-6**: Synthetic Domain Detection (auto.corporeal routing)

## Documentation

- Phase 9C: Receipt Verification & Provenance Continuity
- Receipt ledger pattern established in GOV-9
- Hash chain integrity model from blockchain/git commit chains
- Actor-centric vs domain-centric governance separation

## Notes

- Foreign key constraint on `prev_id` prevents database-level tampering
- Tests verify FK protection rather than simulating broken chains
- Hash computation uses JSON serialization determinism
- Lineage queries ordered by `created_at ASC` for chronological display
- CreatedAt stored as TIMESTAMPTZ but returned as RFC3339 string in API
