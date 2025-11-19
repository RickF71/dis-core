# GOV-6: Auto-Corporeal Domain — Deterministic Consent Simulator

**Phase:** GOV-6
**Status:** ✅ Complete
**Dependencies:** GOV-4 (Policy Evaluation), GOV-5 (CAS Tracking)
**Integration:** Backend (Bootstrap, REGO), Frontend (Finagler UI)

---

## Overview

GOV-6 introduces **`domain.auto.corporeal`**, a self-declaring synthetic domain that provides **deterministic consent decisions** for automated testing and development workflows. This domain:

- **Automatically affirms** when request context contains `[always-affirm]` token
- **Automatically denies** when request context contains `[always-deny]` token
- **Clearly labeled** as synthetic in logs, events, and UI
- **Sandbox-restricted** with all capabilities disabled (no delegation, freeze override, break-glass)

This enables predictable testing without requiring real corporeal consent, while maintaining complete transparency about the synthetic nature of decisions.

---

## Quick Start

### 1. Bootstrap Auto-Corporeal Domain

The domain is automatically registered on server startup:

```bash
# Build and start dis-core
go build -o dis-core cmd/dis-core/main.go
./dis-core --schemas=schemas --domains=domains

# Look for bootstrap log:
# ✅ Created domain: auto.corporeal (synthetic test domain)
```

### 2. Verify Domain Registration

```bash
# Check domain exists
curl http://localhost:8080/api/domain/list | jq '.[] | select(.name=="auto.corporeal")'

# Expected output:
{
  "name": "auto.corporeal",
  "type": "corporeal",
  "state": "active",
  "parent_id": "simula",
  "metadata": {
    "synthetic": true,
    "corporeal_warning": "This domain provides deterministic decisions for testing...",
    "deterministic_test": true
  },
  "capabilities": {
    "delegate": false,
    "freeze_override": false,
    "break_glass": false,
    "authorize": false
  }
}
```

### 3. Test Deterministic Policies

#### Always Affirm:
```bash
curl -X POST http://localhost:8080/api/authority/transition \
  -H "Content-Type: application/json" \
  -d '{
    "identity_id": "test-identity",
    "domain_id": "auto.corporeal",
    "transition": {
      "seat": "member",
      "from": "holder-not-applied",
      "to": "holder-applied"
    },
    "context": {
      "note": "Testing with [always-affirm] token"
    }
  }'

# Expected response:
{
  "decision_id": "...",
  "result": "approve",
  "allow": true,
  "reason": "auto-affirm",
  "cas_ok": true
}
```

#### Always Deny:
```bash
curl -X POST http://localhost:8080/api/authority/transition \
  -H "Content-Type: application/json" \
  -d '{
    "identity_id": "test-identity",
    "domain_id": "auto.corporeal",
    "transition": {
      "seat": "member",
      "from": "holder-not-applied",
      "to": "holder-applied"
    },
    "context": {
      "note": "Testing with [always-deny] token"
    }
  }'

# Expected response:
{
  "decision_id": "...",
  "result": "deny",
  "allow": false,
  "reason": "auto-deny",
  "cas_ok": true
}
```

#### Default Deny (No Token):
```bash
curl -X POST http://localhost:8080/api/authority/transition \
  -H "Content-Type: application/json" \
  -d '{
    "identity_id": "test-identity",
    "domain_id": "auto.corporeal",
    "transition": {
      "seat": "member",
      "from": "holder-not-applied",
      "to": "holder-applied"
    },
    "context": {
      "note": "No token provided"
    }
  }'

# Expected response:
{
  "decision_id": "...",
  "result": "deny",
  "allow": false,
  "reason": "no auto-token provided",
  "cas_ok": true
}
```

---

## Architecture

### Bootstrap Registration

**File:** `cmd/dis-core/bootstrap/domains.go`

```go
func RegisterSyntheticDomains(pool *pgxpool.Pool) error {
    // Check if domains table exists
    var exists bool
    err := pool.QueryRow(context.Background(),
        "SELECT EXISTS(SELECT 1 FROM pg_tables WHERE tablename='domains')").Scan(&exists)

    if !exists {
        return nil // Skip if table not ready
    }

    // Insert auto.corporeal domain if not exists
    _, err = pool.Exec(context.Background(), `
        INSERT INTO domains (name, type, state, parent_id, metadata, capabilities)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (name) DO NOTHING
    `,
        "auto.corporeal",
        "corporeal",
        "active",
        parentID, // Links to 'simula' if exists
        metadataJSON,
        capabilitiesJSON,
    )

    return err
}
```

**Metadata Structure:**
```json
{
  "synthetic": true,
  "corporeal_warning": "This domain provides deterministic decisions for testing. Not a real corporeal authority.",
  "deterministic_test": true
}
```

**Capabilities (All Disabled):**
```json
{
  "delegate": false,
  "freeze_override": false,
  "break_glass": false,
  "authorize": false
}
```

---

### REGO Policy

**File:** `policies/auto.corporeal/seat.rego`

```rego
package dis.seat

# Default deny
default allow := false
default reason := "no auto-token provided"

# Allow rule: [always-affirm] token
allow if {
    contains(input.context.note, "[always-affirm]")
}

reason := "auto-affirm" if {
    contains(input.context.note, "[always-affirm]")
}

# Deny rule: [always-deny] token
allow := false if {
    contains(input.context.note, "[always-deny]")
}

reason := "auto-deny" if {
    contains(input.context.note, "[always-deny]")
}

# Metadata
metadata := {
    "synthetic": true,
    "corporeal_warning": "This domain provides deterministic decisions for testing...",
    "capabilities": {
        "delegate": false,
        "freeze_override": false,
        "break_glass": false,
        "authorize": false
    }
}
```

**Policy Logic:**
1. **Default:** Deny with reason "no auto-token provided"
2. **[always-affirm]:** Allow with reason "auto-affirm"
3. **[always-deny]:** Deny with reason "auto-deny"
4. **Metadata:** Includes synthetic flag and sandbox restrictions

---

### Backend Event Pipeline

#### Mutation Engine Detection

**File:** `internal/authority/mutation_engine.go`

```go
// GOV-6: Detect synthetic domain
if req.DomainID == "auto.corporeal" {
    log.Printf("🔬 [GOV-6] Synthetic domain executing deterministic policy")
    log.Printf("   Context: %+v", req.Context)
}

// Pass domainID to event bus for synthetic flag propagation
engine.events.PublishSeatChange(req.DomainID, decision.ID, receiptID, ...)
```

#### SSE Event Structure

**File:** `internal/authority/events.go`

```go
type SeatChangeEvent struct {
    // ... existing fields ...

    // GOV-6: Synthetic domain markers
    Synthetic bool   `json:"synthetic,omitempty"`
    Message   string `json:"message,omitempty"`
}

func (bus *InMemoryEventBus) PublishSeatChange(domainID string, ...) {
    synthetic := domainID == "auto.corporeal"

    event := SeatChangeEvent{
        // ... existing fields ...
        Synthetic: synthetic,
        Message: "auto.corporeal — synthetic domain executing deterministic decision",
    }

    bus.broadcast(event)
}
```

---

### Frontend Integration

#### TypeScript Interface

**File:** `finagler/src/lib/api/sse.ts`

```typescript
export interface SeatChangeEvent {
  // ... existing fields ...

  // GOV-6: Synthetic domain markers
  synthetic?: boolean;
  message?: string;
}
```

#### Authority Panel UI

**File:** `finagler/src/components/AuthorityPanel.jsx`

```jsx
const [syntheticWarning, setSyntheticWarning] = useState('');

const handleSeatChange = useCallback((event) => {
  if (event.identity_id === identityId) {
    // GOV-6: Synthetic domain handling
    if (event.synthetic) {
      setSyntheticWarning(event.message || 'Decision from synthetic domain');
      setTimeout(() => setSyntheticWarning(''), 8000);
    }
    // ... rest of handler
  }
}, [identityId, refetch]);

// Render yellow warning banner
{syntheticWarning && (
  <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
    <div className="flex items-center gap-2 text-yellow-800">
      <span className="text-xl">⚠️</span>
      <div>
        <div className="font-semibold">Synthetic Domain</div>
        <div className="text-sm text-yellow-700">
          {syntheticWarning}
        </div>
      </div>
    </div>
  </div>
)}
```

**UI Behavior:**
- Yellow warning banner appears when synthetic domain decision is received
- Displays custom message from event or default warning
- Auto-dismisses after 8 seconds
- Styled distinctly from CAS conflict alert (red)

---

## Testing & Verification

### 1. Domain Registration Check

```sql
-- Verify domain exists
SELECT
    name,
    type,
    state,
    parent_id,
    metadata->'synthetic' as synthetic_flag,
    capabilities
FROM domains
WHERE name = 'auto.corporeal';

-- Expected result:
-- name: auto.corporeal
-- type: corporeal
-- state: active
-- synthetic_flag: true
-- capabilities: all false
```

### 2. Policy Execution Test

```bash
# Test all three decision paths
./scripts/test_auto_corporeal.sh

# Script contents:
#!/bin/bash
echo "Testing [always-affirm]..."
curl -X POST .../transition -d '{"context":{"note":"[always-affirm]"}}'

echo "Testing [always-deny]..."
curl -X POST .../transition -d '{"context":{"note":"[always-deny]"}}'

echo "Testing default deny..."
curl -X POST .../transition -d '{"context":{"note":"no token"}}'
```

### 3. Audit Trail Verification

```sql
-- Check synthetic flag in audit log
SELECT
    decision_id,
    identity_id,
    domain_id,
    result,
    reason,
    metadata->'synthetic' as synthetic_flag,
    created_at
FROM authority_decisions
WHERE domain_id = 'auto.corporeal'
ORDER BY created_at DESC
LIMIT 10;
```

### 4. Finagler UI Test

1. Open Finagler: `http://localhost:5173`
2. Navigate to Authority Console
3. Select identity with auto.corporeal seat
4. Trigger transition with `[always-affirm]` in context
5. **Expected:** Yellow warning banner appears with message
6. **Expected:** Banner auto-dismisses after 8 seconds

---

## Security Constraints

### Sandbox Restrictions

All capabilities **disabled** for auto.corporeal domain:

| Capability | Status | Reason |
|------------|--------|--------|
| `delegate` | ❌ Disabled | No authority transfer allowed |
| `freeze_override` | ❌ Disabled | Cannot override state freezes |
| `break_glass` | ❌ Disabled | No emergency access bypass |
| `authorize` | ❌ Disabled | Cannot grant authorizations |

### Transparency Requirements

Every synthetic decision must:
1. ✅ Set `synthetic: true` in domain metadata
2. ✅ Include `corporeal_warning` in metadata
3. ✅ Set `Synthetic: true` in SSE events
4. ✅ Log deterministic execution context
5. ✅ Display yellow warning banner in UI
6. ✅ Store synthetic flag in audit trail

### Audit Compliance

All synthetic decisions are:
- Logged to `authority_decisions` table with domain_id
- Marked with synthetic metadata
- Traceable via decision_id and receipt_id
- Visible in continuity chain
- Distinguishable from real corporeal decisions

---

## Use Cases

### Automated Testing

```javascript
// Integration test example
describe('Authority Transitions', () => {
  it('should approve with [always-affirm] token', async () => {
    const response = await api.transition({
      identity_id: 'test-id',
      domain_id: 'auto.corporeal',
      context: { note: '[always-affirm]' }
    });

    expect(response.result).toBe('approve');
    expect(response.allow).toBe(true);
    expect(response.reason).toBe('auto-affirm');
  });
});
```

### Development Workflows

```bash
# Seed test identities with predictable state
for i in {1..10}; do
  curl -X POST .../transition \
    -d "{\"identity_id\":\"test-$i\",\"domain_id\":\"auto.corporeal\",\"context\":{\"note\":\"[always-affirm]\"}}"
done
```

### CI/CD Pipelines

```yaml
# GitHub Actions example
- name: Test Authority Console
  run: |
    ./dis-core --test-mode &
    sleep 5
    ./scripts/smoke_test_auto_corporeal.sh
    if [ $? -ne 0 ]; then exit 1; fi
```

---

## Troubleshooting

### Domain Not Found

**Symptom:** API returns "domain auto.corporeal not found"

**Solution:**
```bash
# Check bootstrap logs
grep "auto.corporeal" phases/bootstrap.log

# Manually register if missing
psql $DIS_DB_DSN -c "INSERT INTO domains (name, type, state, metadata) VALUES ('auto.corporeal', 'corporeal', 'active', '{\"synthetic\": true}'::jsonb) ON CONFLICT DO NOTHING;"
```

### Policy Not Loading

**Symptom:** Transitions fail with "policy evaluation error"

**Solution:**
```bash
# Verify REGO file exists
ls -la policies/auto.corporeal/seat.rego

# Test policy syntax
opa test policies/auto.corporeal/

# Check OPA bundle loading in logs
grep "auto.corporeal" phases/policy_init.log
```

### Yellow Banner Not Appearing

**Symptom:** UI doesn't show synthetic warning

**Solution:**
```javascript
// Check browser console for SSE events
// Look for: "GOV-5/GOV-6 Event: {..., synthetic: true}"

// Verify event handler is registered
// Check React DevTools for syntheticWarning state

// Test manually:
const testEvent = {
  identity_id: 'test-id',
  synthetic: true,
  message: 'Test synthetic warning'
};
handleSeatChange(testEvent);
```

### Audit Trail Missing Synthetic Flag

**Symptom:** `authority_decisions` records don't have synthetic metadata

**Solution:**
```sql
-- Check if metadata is stored correctly
SELECT decision_id, metadata FROM authority_decisions WHERE domain_id='auto.corporeal' LIMIT 1;

-- If missing, update mutation engine to pass metadata
-- See: internal/authority/mutation_engine.go - buildDecision()
```

---

## Monitoring Queries

### Synthetic Decision Volume

```sql
-- Count synthetic decisions by day
SELECT
    DATE(created_at) as decision_date,
    COUNT(*) as synthetic_decisions,
    SUM(CASE WHEN result='approve' THEN 1 ELSE 0 END) as approvals,
    SUM(CASE WHEN result='deny' THEN 1 ELSE 0 END) as denials
FROM authority_decisions
WHERE domain_id = 'auto.corporeal'
GROUP BY DATE(created_at)
ORDER BY decision_date DESC;
```

### Token Distribution

```sql
-- Analyze which tokens are used most
SELECT
    CASE
        WHEN context->>'note' LIKE '%[always-affirm]%' THEN 'always-affirm'
        WHEN context->>'note' LIKE '%[always-deny]%' THEN 'always-deny'
        ELSE 'no-token'
    END as token_type,
    COUNT(*) as usage_count,
    COUNT(DISTINCT identity_id) as unique_identities
FROM authority_decisions
WHERE domain_id = 'auto.corporeal'
GROUP BY token_type
ORDER BY usage_count DESC;
```

### Provenance Continuity

```sql
-- Verify receipts issued for synthetic decisions
SELECT
    d.decision_id,
    d.receipt_id,
    r.verified,
    r.policy_ref,
    r.redaction_ref
FROM authority_decisions d
LEFT JOIN receipts r ON r.event_id = d.decision_id
WHERE d.domain_id = 'auto.corporeal'
  AND r.receipt_type = 'ci.call.v1'
ORDER BY d.created_at DESC
LIMIT 20;
```

---

## Next Phase: GOV-7

**Policy Reload & Bundle Lifecycle Control**

Building on GOV-6's synthetic domain testing infrastructure:

1. **Hot Reload:** Update REGO policies without server restart
2. **Policy Versioning:** Track policy bundle versions in database
3. **Rollback Capability:** Revert to previous policy versions
4. **Bundle Validation:** Test new policies before activation
5. **Audit Trail:** Log policy changes in `policy_versions` table
6. **Admin API:** Endpoints for policy management (/api/policy/reload, /api/policy/rollback)

---

## Summary

GOV-6 delivers a **deterministic consent simulator** that enables:

- ✅ **Predictable Testing:** [always-affirm] and [always-deny] tokens
- ✅ **Full Transparency:** Synthetic flag in logs, events, and UI
- ✅ **Sandbox Safety:** All capabilities disabled
- ✅ **Audit Compliance:** Complete provenance continuity
- ✅ **Developer Experience:** Rapid iteration without real corporeal consent

The auto.corporeal domain provides a reliable foundation for automated testing, CI/CD pipelines, and development workflows while maintaining clear boundaries between synthetic and real corporeal authority.

---

**Implementation Status:** ✅ Complete
**Phase Files:** `phase_9c.log`, `bootstrap.log`, `policy_init.log`
**Database Objects:** `domains.auto.corporeal`, `policies/auto.corporeal/seat.rego`
**API Endpoints:** `/api/domain/list`, `/api/authority/transition`
**UI Components:** Finagler Authority Panel (synthetic warning banner)
