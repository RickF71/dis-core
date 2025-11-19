# GOV-5: CAS Version Tracking + Finagler SSE Integration

## Overview

GOV-5 extends the Authority Console with **Compare-And-Swap (CAS) version tracking** for all seat mutations and propagates decision/receipt IDs and CAS results to Finagler via **Server-Sent Events (SSE)**.

**Key Features:**
- ✅ CAS version tracking in database (identity_seats.version column)
- ✅ CAS success/failure indicators (authority_decisions.cas_ok)
- ✅ Enhanced SSE events with CAS metadata (cas_prev, cas_new, cas_ok)
- ✅ Real-time Finagler UI updates showing version transitions
- ✅ Visual CAS conflict indicators and alerts
- ✅ Decision and receipt ID display in UI

## Architecture

```
┌─────────────────────────────────────────┐
│  Identity Seats Table                   │
│  ┌───────────────────────────────────┐  │
│  │ id, identity_id, seat_type        │  │
│  │ state, version ◄─────────────────┼──┼── GOV-5: Optimistic lock
│  │ assigned_at, occupied_at          │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  UpdateSeatStateCAS (GOV-5 Enhanced)    │
│  ┌───────────────────────────────────┐  │
│  │ 1. SELECT version FOR UPDATE      │  │
│  │ 2. Compare state (from == current)│  │
│  │ 3. UPDATE version = version + 1   │  │
│  │ 4. Return CASUpdateResult         │  │
│  │    - NewState                     │  │
│  │    - PreviousVersion              │  │
│  │    - NewVersion                   │  │
│  │    - Success (bool)               │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Authority Decisions Audit Trail        │
│  ┌───────────────────────────────────┐  │
│  │ cas_version_prev, cas_version_new │  │
│  │ cas_ok ◄──────────────────────────┼──┼── GOV-5: Conflict tracking
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  SSE: /api/authority/events             │
│  ┌───────────────────────────────────┐  │
│  │ event: seat.change                │  │
│  │ data: {                           │  │
│  │   cas_prev: 17,                   │  │
│  │   cas_new: 18,                    │  │
│  │   cas_ok: true,                   │  │
│  │   decision_id: "...",             │  │
│  │   receipt_id: "..."               │  │
│  │ }                                 │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│  Finagler AuthorityPanel (UI)           │
│  ┌───────────────────────────────────┐  │
│  │ • CAS version display per seat    │  │
│  │ • Conflict warning alert (⚠️)     │  │
│  │ • Decision/Receipt ID badges      │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

## Quick Start

### 1. Database Migration

Apply the GOV-5 schema:

```bash
psql "$DIS_DB_DSN" < db/migrations/20251114_gov5_cas_tracking.sql
```

**What it does:**
- Adds `version BIGINT DEFAULT 1` to `identity_seats` table
- Adds `cas_ok BOOLEAN DEFAULT TRUE` to `authority_decisions` table
- Creates indexes for efficient CAS conflict queries
- Creates `authority_decisions_cas_failures` view for monitoring

### 2. Restart Backend

No configuration changes needed — GOV-5 is backward-compatible:

```bash
cd dis-core
go build -o dis-core cmd/dis-core/main.go
./dis-core --schemas=schemas --domains=domains
```

Check logs for CAS tracking confirmation:

```
✅ GOV-5: CAS version tracking enabled
   - identity_seats.version column active
   - authority_decisions.cas_ok tracking enabled
```

### 3. Restart Finagler

```bash
cd finagler
npm install  # If dependencies changed
npm run dev
```

Open http://localhost:5173 and navigate to Authority Console.

### 4. Test CAS Tracking

**Trigger a seat transition:**

```bash
curl -X POST http://localhost:8080/api/authority/seat/transition \
  -H 'Content-Type: application/json' \
  -d '{
    "domain_id": "terra",
    "identity_id": "test-identity-123",
    "layer": "terra",
    "from": "EMPTY",
    "to": "ASSIGNED",
    "actor_id": "admin-456",
    "reason": "GOV-5 test",
    "context": {}
  }'
```

**Expected Response (GOV-5 Enhanced):**

```json
{
  "ok": true,
  "decision_id": "abc-123-def-456",
  "receipt_id": "xyz-789-uvw-012",
  "before_state": "EMPTY",
  "after_state": "ASSIGNED"
}
```

**Observe in Finagler UI:**
1. Authority Panel shows: `CAS: v1 → v2` under the terra seat
2. Decision ID and Receipt ID visible in status footer
3. No conflict indicator (✅ successful transition)

### 5. Simulate CAS Conflict

**Concurrent requests (same identity):**

Terminal 1:
```bash
curl -X POST http://localhost:8080/api/authority/seat/transition \
  -d '{"domain_id":"terra","identity_id":"test-123","layer":"terra","from":"ASSIGNED","to":"OCCUPIED","actor_id":"user1","reason":"concurrent 1"}'
```

Terminal 2 (immediately after):
```bash
curl -X POST http://localhost:8080/api/authority/seat/transition \
  -d '{"domain_id":"terra","identity_id":"test-123","layer":"terra","from":"ASSIGNED","to":"FROZEN","actor_id":"user2","reason":"concurrent 2"}'
```

**Expected Behavior:**
- One request succeeds with `cas_ok: true`
- Other request fails with CAS mismatch error
- Audit log records both attempts with version numbers
- Finagler shows ⚠️ conflict indicator

## Database Schema Changes

### identity_seats Table (GOV-5 Addition)

| Column  | Type   | Description                                    |
|---------|--------|------------------------------------------------|
| version | BIGINT | Optimistic lock counter (incremented on update)|

**Index:**
- `idx_identity_seats_version ON identity_seats(version)`

### authority_decisions Table (GOV-5 Addition)

| Column  | Type    | Description                              |
|---------|---------|------------------------------------------|
| cas_ok  | BOOLEAN | Whether CAS update succeeded (GOV-5)     |

**Indexes:**
- `idx_authority_cas_ok ON authority_decisions(cas_ok) WHERE cas_ok = FALSE` (conflicts only)
- `idx_authority_cas_versions ON authority_decisions(cas_version_prev, cas_version_new)` (range queries)

### New Monitoring View

```sql
SELECT * FROM authority_decisions_cas_failures
ORDER BY created_at DESC
LIMIT 20;
```

Shows all CAS conflicts with full context for debugging race conditions.

## SSE Event Schema (GOV-5)

```typescript
interface SeatChangeEvent {
  type: 'seat.change';
  identity_id: string;
  layer: string;
  from: string;
  to: string;
  decision_id?: string;
  receipt_id?: string;
  cas_prev?: number;     // ← GOV-5: Previous version
  cas_new?: number;      // ← GOV-5: New version
  cas_ok: boolean;       // ← GOV-5: Success indicator
  timestamp: string;
}
```

**Example Event:**

```json
{
  "type": "seat.change",
  "identity_id": "identity-123",
  "layer": "terra",
  "from": "ASSIGNED",
  "to": "OCCUPIED",
  "decision_id": "dec-abc-123",
  "receipt_id": "rec-xyz-789",
  "cas_prev": 17,
  "cas_new": 18,
  "cas_ok": true,
  "timestamp": "2025-11-14T10:30:45Z"
}
```

## Finagler UI Enhancements

### Authority Panel Updates

1. **CAS Version Display** (per seat card):
   ```
   CAS: v17 → v18
   ```

2. **Conflict Indicator** (when cas_ok = false):
   ```
   CAS: v17 → v17 ⚠️
   ```

3. **Alert Banner** (on conflict detection):
   ```
   ⚠️ CAS Conflict Detected
   A concurrent modification was detected. The operation may need to be retried.
   ```

4. **Decision/Receipt IDs** (in status footer):
   ```
   ⚡ terra → OCCUPIED at 10:30:45
   Decision: dec-abc-1... Receipt: rec-xyz-7...
   ```

### Visual Feedback

- **Green border**: Successful CAS transition
- **Red ⚠️ icon**: CAS conflict detected
- **Yellow alert box**: Temporary notification (5 seconds)
- **Console logs**: Full event details for debugging

## Monitoring & Debugging

### Query CAS Conflicts

```sql
-- Recent CAS failures
SELECT
  domain_id,
  seat,
  from_state,
  to_state,
  cas_version_prev,
  cas_version_new,
  reason,
  created_at
FROM authority_decisions_cas_failures
ORDER BY created_at DESC
LIMIT 10;
```

### Query Version Progression

```sql
-- Version history for a specific identity seat
SELECT
  seat,
  from_state,
  to_state,
  cas_version_prev,
  cas_version_new,
  cas_ok,
  created_at
FROM authority_decisions
WHERE domain_id = 'terra'
  AND actor_id = 'identity-123'
ORDER BY created_at ASC;
```

### Check Current Versions

```sql
-- Current version numbers for all seats
SELECT
  identity_id,
  seat_type,
  state,
  version
FROM identity_seats
WHERE identity_id = 'test-identity-123';
```

## Troubleshooting

### Issue: Version column missing

**Error:**
```
column "version" does not exist
```

**Solution:**
Run the GOV-5 migration:
```bash
psql "$DIS_DB_DSN" < db/migrations/20251114_gov5_cas_tracking.sql
```

### Issue: CAS always succeeds (no conflicts detected)

**Cause:** Requests are not concurrent enough.

**Solution:** Use parallel curl with `&`:
```bash
curl ... & curl ... & wait
```

Or use load testing tools like `ab` or `hey`.

### Issue: Frontend doesn't show CAS info

**Check:**
1. Finagler version up-to-date (sse.ts includes cas_prev/cas_new fields)
2. Backend SSE events include CAS data (check browser DevTools Network tab)
3. AuthorityPanel component imported correctly

**Debug:**
```javascript
// In browser console:
eventSource = new EventSource('http://localhost:8080/api/authority/events');
eventSource.addEventListener('seat', (e) => console.log(JSON.parse(e.data)));
```

### Issue: cas_ok always true even on conflicts

**Cause:** Mutation engine not catching CAS mismatch errors.

**Check:**
- `UpdateSeatStateCAS` returns `CASUpdateResult` with `Success: false` on mismatch
- Mutation engine logs failed CAS attempts with `CASApplied: false`

## Code Changes Summary

### Backend (dis-core)

| File                                         | Changes                                      |
|----------------------------------------------|----------------------------------------------|
| `db/migrations/20251114_gov5_cas_tracking.sql` | Added version, cas_ok columns + indexes     |
| `internal/identity/triad_repo.go`            | CASUpdateResult struct, version tracking    |
| `internal/authority/mutation_engine.go`      | Extract CAS versions, pass to audit logger  |
| `internal/authority/mutation/audit_logger.go`| Insert cas_ok into authority_decisions      |
| `internal/authority/events.go`               | Add cas_prev, cas_new, cas_ok to SSE events |

### Frontend (finagler)

| File                                       | Changes                                  |
|--------------------------------------------|------------------------------------------|
| `src/lib/api/sse.ts`                       | Add cas_prev, cas_new, cas_ok to interface|
| `src/components/AuthorityPanel.jsx`        | Display CAS versions, conflict alerts    |

**Total Lines Changed:** ~200 lines (backend + frontend)

## Next Phase Recommendations

### GOV-6: Policy Reload & Bundle Lifecycle Control

- **Hot reload REGO policies** without restarting server
- **Policy versioning** with rollback capability
- **Bundle validation** before activation
- **Audit trail** for policy changes

### GOV-7: Bounded Cascade & Simulation Mode

- **Automatic seat transitions** based on parent seat changes
- **Max depth** and **TTL limits** for cascades
- **Dry-run mode** to preview effects without CAS mutation
- **Cascade decision tree** visualization in Finagler

## References

- **GOV-4**: Policy evaluation & audit logging foundation
- **GOV-3**: Write APIs & SSE events
- **GOV-2**: Read-only authority console
- **PostgreSQL Optimistic Locking**: https://www.postgresql.org/docs/current/mvcc-intro.html

## Verification Checklist

- [x] Migration applied successfully
- [x] Backend compiles without errors
- [x] Frontend compiles and runs
- [x] SSE events include CAS fields
- [x] Authority Panel displays CAS versions
- [x] CAS conflicts trigger visual alerts
- [x] Decision/Receipt IDs visible in UI
- [x] authority_decisions_cas_failures view populated on conflicts
- [x] No breaking changes to existing GOV-3 functionality

## Files Modified

### Backend (dis-core)
```
db/migrations/20251114_gov5_cas_tracking.sql              [NEW]
internal/identity/triad_repo.go                           [MODIFIED]
internal/authority/mutation_engine.go                     [MODIFIED]
internal/authority/mutation/audit_logger.go               [MODIFIED]
internal/authority/events.go                              [MODIFIED]
```

### Frontend (finagler)
```
src/lib/api/sse.ts                                        [MODIFIED]
src/components/AuthorityPanel.jsx                         [MODIFIED]
```

**Total Files**: 7 modified, 1 new migration
**Compilation Status**: ✅ All files compile without errors
**Migration Status**: Ready for deployment
**API Compatibility**: Backward-compatible with GOV-3/GOV-4
**UI Compatibility**: Enhanced, no breaking changes
