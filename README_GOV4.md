# GOV-4: Authority Console — Policy Evaluation & Audit Trail

## Overview

GOV-4 integrates **OPA policy evaluation** and **database audit logging** into the seat mutation pipeline, completing the write-API governance stack established in GOV-3.

**Key Features:**
- ✅ OPA-based policy evaluation (REGO policies from filesystem)
- ✅ PostgreSQL audit logging (authority_decisions table)
- ✅ Receipt generation for provenance continuity (ci.call.v1)
- ✅ Development mode with null adapters (policy/audit disabled)
- ✅ Environment-based configuration (no code changes needed)
- ✅ CAS version tracking for race condition analysis

## Architecture

```
┌──────────────────┐
│  POST /api/      │
│  authority/seat/ │
│  transition      │
└────────┬─────────┘
         │
         ▼
┌────────────────────────────────────────┐
│  SeatMutationEngine                    │
│  ┌──────────────────────────────────┐  │
│  │ 1. Fetch Actor Triad (consent)   │  │
│  │ 2. PolicyEvaluator.Evaluate()    │◄─┼── OPA (data.dis.seat)
│  │ 3. CAS Update (if allowed)       │  │
│  │ 4. AuditLogger.LogDecision()     │◄─┼── DB (authority_decisions)
│  │ 5. EventBus.PublishSeatChange()  │  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
         │
         ▼
┌────────────────────────────────────────┐
│  SSE /api/authority/events             │
│  → decision_id, receipt_id in payload  │
└────────────────────────────────────────┘
```

## Quick Start

### 1. Database Migration

Apply the GOV-4 schema:

```bash
psql "$DIS_DB_DSN" < db/migrations/20251113_gov4_audit_indexes.sql
```

This creates:
- `authority_decisions` table (15 columns)
- 5 indexes (domain, actor, seat, policy, CAS)
- 2 views (denied transitions, CAS conflicts)

### 2. Environment Configuration

```bash
# Policy Configuration
export DIS_POLICY_ENABLED=true              # Enable OPA policy evaluation
export DIS_POLICY_LOAD=fs                   # Load policies from filesystem
export DIS_POLICY_BUNDLE_DIR=./policies     # REGO bundle directory

# Audit Configuration
export DIS_AUDIT_ENABLED=true               # Enable DB audit logging
export DIS_AUDIT_RECEIPT_KIND=ci.call.v1    # Receipt type for provenance
export DIS_DECISIONS_TABLE=authority_decisions  # Audit table name
```

**Development Mode (Policy + Audit Disabled):**

```bash
export DIS_POLICY_ENABLED=false
export DIS_AUDIT_ENABLED=false
# Null adapters return allow=true and dry-run IDs
```

### 3. REGO Policy Bundle

Place REGO policies in `./policies/`:

```
policies/
├── seat.rego          # Required: seat transition rules
├── freeze.rego        # Optional: freeze constraints
├── risk.rego          # Optional: risk analysis
└── gates.rego         # Optional: gate requirements
```

**Example seat.rego:**

```rego
package dis

# Allow guest → user transition
allow {
    input.request.from == "guest"
    input.request.to == "user"
}

# Deny any transition to admin without explicit permission
deny {
    input.request.to == "admin"
    not has_permission(input.actor, "admin.grant")
}

has_permission(actor, perm) {
    actor.permissions[_] == perm
}
```

### 4. Build and Run

```bash
go build -o dis-core cmd/dis-core/main.go
./dis-core --schemas=schemas --domains=domains
```

Check logs for wiring confirmation:

```
🔧 Wiring GOV-4 mutation engine components...
✅ Policy evaluator: OPA (bundle: ./policies)
✅ Audit logger: DB (table: authority_decisions, receipt: ci.call.v1)
✅ Event bus: In-memory (SSE)
✅ Mutation engine: Wired and ready
```

### 5. Test API

```bash
export IDENTITY_ID="identity-123"
export ACTOR_ID="identity-123"
export DOMAIN_ID="terra"

curl -X POST http://localhost:8080/api/authority/seat/transition \
  -H 'Content-Type: application/json' \
  -d '{
    "domain_id": "terra",
    "identity_id": "identity-123",
    "layer": "terra",
    "from": "guest",
    "to": "user",
    "actor_id": "identity-123",
    "reason": "onboarding",
    "context": {
      "permissions": ["seat.transition"]
    }
  }'
```

**Expected Response:**

```json
{
  "ok": true,
  "decision_id": "abc-123-def-456",
  "receipt_id": "xyz-789-uvw-012",
  "before_state": "guest",
  "after_state": "user"
}
```

## Database Schema

### authority_decisions Table

| Column             | Type        | Description                          |
|--------------------|-------------|--------------------------------------|
| id                 | UUID        | Primary key                          |
| domain_id          | TEXT        | Target domain                        |
| actor_id           | TEXT        | Who authorized the transition        |
| seat               | TEXT        | terra \| numen \| lima               |
| from_state         | TEXT        | Previous state                       |
| to_state           | TEXT        | New state (or attempted)             |
| allow              | BOOLEAN     | Policy decision (true=allow)         |
| reason             | TEXT        | Policy reason / explanation          |
| policy_ref         | TEXT        | Policy module reference              |
| cas_applied        | BOOLEAN     | Whether CAS update succeeded         |
| cas_version_prev   | BIGINT      | Previous version (0 if N/A)          |
| cas_version_new    | BIGINT      | New version (0 if N/A)               |
| request_id         | TEXT        | Trace ID for correlation             |
| extra              | JSONB       | Additional metadata                  |
| created_at         | TIMESTAMPTZ | Decision timestamp                   |

### Monitoring Views

**1. Denied Transitions (Security Monitoring):**

```sql
SELECT * FROM authority_decisions_denied ORDER BY created_at DESC LIMIT 20;
```

**2. CAS Conflicts (Race Condition Detection):**

```sql
SELECT * FROM authority_decisions_cas_conflicts ORDER BY created_at DESC LIMIT 20;
```

## Troubleshooting

### Policy Bundle Not Found

**Error:**
```
failed to create OPA policy evaluator: failed to read seat.rego: no such file or directory
```

**Solution:**
1. Verify `DIS_POLICY_BUNDLE_DIR` points to correct directory
2. Ensure `seat.rego` exists in that directory
3. Check file permissions (readable by process)

**Temporary Fix (Development):**
```bash
export DIS_POLICY_ENABLED=false  # Use null policy evaluator
```

### Audit Table Missing

**Error:**
```
pq: relation "authority_decisions" does not exist
```

**Solution:**
Run migration:
```bash
psql "$DIS_DB_DSN" < db/migrations/20251113_gov4_audit_indexes.sql
```

### Policy Always Denies

**Debug:**
1. Check policy input format in logs
2. Test REGO policy with OPA CLI:

```bash
cd policies/
opa eval -d seat.rego 'data.dis.seat.allow' --input input.json
```

**Example input.json:**

```json
{
  "request": {
    "from": "guest",
    "to": "user",
    "seat": "terra"
  },
  "actor": {
    "permissions": ["seat.transition"]
  },
  "context": {}
}
```

### CAS Version Always Zero

**Known Issue:**
Current implementation doesn't extract CAS versions from `TriadRepository.UpdateSeatStateCAS()`.

**TODO:**
- Modify `UpdateSeatStateCAS` to return `(newState string, prevVersion int64, newVersion int64, error)`
- Update `mutation_engine.go` to pass versions to `AuditAttrs`

**Impact:**
Audit trail records CAS decisions but can't reconstruct exact version sequences (non-blocking).

## Testing

### Smoke Test Script

```bash
export IDENTITY_ID="test-identity-456"
export ACTOR_ID="$IDENTITY_ID"
export DOMAIN_ID="terra"

# Run with policy+audit enabled
DIS_POLICY_ENABLED=true DIS_AUDIT_ENABLED=true \
  ./scripts/dev/authority_mutation_smoke.sh
```

### Query Audit Trail

```sql
-- Recent decisions
SELECT id, domain_id, seat, from_state, to_state, allow, reason, created_at
FROM authority_decisions
ORDER BY created_at DESC
LIMIT 10;

-- Decisions by actor
SELECT seat, from_state, to_state, allow, COUNT(*) as total
FROM authority_decisions
WHERE actor_id = 'identity-123'
GROUP BY seat, from_state, to_state, allow
ORDER BY total DESC;

-- CAS conflicts
SELECT * FROM authority_decisions_cas_conflicts;
```

### SSE Event Stream

```bash
# Subscribe to live seat changes
curl -N http://localhost:8080/api/authority/events

# Expected events:
event: seat.change
data: {"identity_id":"...","seat":"terra","from":"guest","to":"user","decision_id":"abc-123","receipt_id":"xyz-456"}
```

## Configuration Reference

| Environment Variable         | Default Value           | Description                                    |
|------------------------------|-------------------------|------------------------------------------------|
| `DIS_POLICY_ENABLED`         | `true`                  | Enable OPA policy evaluation                   |
| `DIS_POLICY_LOAD`            | `fs`                    | Policy load mode (only `fs` supported)         |
| `DIS_POLICY_BUNDLE_DIR`      | `./policies`            | Directory containing REGO files                |
| `DIS_AUDIT_ENABLED`          | `true`                  | Enable database audit logging                  |
| `DIS_AUDIT_RECEIPT_KIND`     | `ci.call.v1`            | Receipt type for provenance                    |
| `DIS_DECISIONS_TABLE`        | `authority_decisions`   | Audit table name                               |

## Next Steps (GOV-5 Preview)

- **Parent Override Policy**: Upward/downward flow authorization
- **Bounded Cascade**: Automatic seat transitions with maxDepth/TTL
- **Simulation Mode**: Dry-run transitions without CAS mutation
- **Policy Composition**: Multi-package REGO evaluation
- **CAS Version Extraction**: Full version tracking in audit trail

## References

- **GOV-3**: Seat mutation write APIs (foundation)
- **GOV-2**: Read-only authority console APIs
- **Phase 9C**: Receipt verification & provenance continuity
- **REGO Policy Docs**: https://www.openpolicyagent.org/docs/latest/policy-language/

## Files Created

```
internal/authority/mutation/adapters.go       # Adapter interfaces
internal/authority/mutation/opa_policy.go     # OPA evaluator
internal/authority/mutation/audit_logger.go   # DB audit logger
db/migrations/20251113_gov4_audit_indexes.sql # Schema migration
internal/config/policy_audit.go               # Config parsing
internal/boot/wiring_gov4.go                  # Dependency injection
internal/authority/mutation_engine.go         # Updated to use adapters
```

**Total Lines Added**: ~800 lines
**Compilation Status**: ✅ All files compile without errors
**Migration Status**: Ready for deployment
**API Compatibility**: Backward-compatible with GOV-3
