# Phase GOV-2: Authority Console Integration - COMPLETE

**Status**: ✅ Implemented and Integrated
**Date**: 2024
**Scope**: Read-only API endpoints + minimal Finagler UI for identity triad visualization

---

## Overview

Phase GOV-2 adds Authority Console integration on top of the GOV-1 foundation, providing read-only API endpoints for querying identity triads and evaluating authority flows. This phase includes both backend (dis-core) and frontend (Finagler) components.

**Key Principle**: Read-only operations only. Write APIs (seat state transitions) are reserved for GOV-3.

---

## Backend Implementation (dis-core)

### 1. API Types (`internal/api/authority/types.go`)

**Package**: `authorityapi` (matches existing authority API package)

**DTOs Created**:
- `IdentitySeatDTO` - Represents a single seat (terra/numen/lima) with ID, layer, state, timestamps
- `IdentityTriadDTO` - Complete triad with identity_id, completeness flag, and 3 seats
- `FlowPreview` - 6 sample authority scenarios with expected results
- `FlowEvalInput` - Authority evaluation request (identity_id, direction, target)
- `FlowEvalResult` - Evaluation response with allow/deny, reason, seat statuses
- `TriadSeatStatus` - Seat status helper for evaluation responses

**Helper Functions**:
- `GetTimestamp()` - Returns current RFC3339 timestamp

### 2. API Handler (`internal/api/authority/handler.go`)

**Package**: `authorityapi`

**TriadHandler Struct**:
```go
type TriadHandler struct {
    triadRepo *identity.TriadRepository
}
```

**Constructor**:
```go
func NewTriadHandler(triadRepo *identity.TriadRepository) *TriadHandler
```

**Endpoints Implemented**:

#### a. `GET /api/authority/triad/{identityId}`
- Fetches complete identity triad (terra/numen/lima seats)
- Extracts identityId from Chi URL params
- Returns `IdentityTriadDTO` with all 3 seats
- 404 if identity not found or triad incomplete

#### b. `GET /api/authority/flow/preview`
- Returns 6 sample authority flow scenarios
- Each scenario includes: scenario name, identity_id, direction, target_domain, expected result
- Scenarios cover:
  - Upward authority (ASSIGNED allowed)
  - Upward authority (OCCUPIED allowed)
  - Downward authority (ASSIGNED blocked)
  - Downward authority (OCCUPIED allowed)
  - Frozen seat (blocked)
  - Cross-domain (requires parent approval)

#### c. `POST /api/authority/flow/eval`
- Evaluates authority flow for a specific identity
- Request body: `FlowEvalInput` (identity_id, direction, target_domain, parent_approved)
- Fetches triad, builds seat statuses
- Calls `authority.EvaluateAuthority()` with direction and triad
- Returns: `FlowEvalResult` with allow/deny, reason, seat states, timestamp

**Dependencies**:
- Uses `internal/identity.TriadRepository` for data access
- Uses `internal/authority.EvaluateAuthority()` for flow evaluation
- Chi router for URL parameter extraction

### 3. Route Registration (`internal/api/routes.go`)

**Integration Point**: `RegisterAllRoutes()` method

**Routes Added**:
```go
if s.triadRepo != nil {
    triadHandler := authorityapi.NewTriadHandler(s.triadRepo)
    r.Get("/api/authority/triad/{identityId}", triadHandler.GetTriadByIdentity)
    r.Get("/api/authority/flow/preview", triadHandler.PreviewFlow)
    r.Post("/api/authority/flow/eval", triadHandler.EvaluateFlow)
}
```

**Safety**: Nil-check ensures routes only registered if TriadRepository is initialized

### 4. Server Initialization (`internal/api/server.go`)

**Changes**:

**Import Added**:
```go
"dis-core/internal/identity"
```

**Field Added to Server struct**:
```go
// GOV-1: Identity triad management
triadRepo *identity.TriadRepository
```

**Initialization in `NewWithPolicy()`**:
```go
// GOV-1: Initialize identity triad repository
s.triadRepo = identity.NewTriadRepository(db)
```

### 5. Bootstrap Integration (`cmd/dis-core/main.go`)

**Added Call**:
```go
// GOV-1: Bootstrap identity triads (terra/numen/lima)
if err := bootstrap.BootstrapIdentityTriads(ctx, dbComponents.Database); err != nil {
    log.Printf("Warning: GOV-1 triad bootstrap failed: %v", err)
}
```

**Location**: Between Phase S0 (root seat setup) and Authority Console initialization

**Effect**: Auto-creates terra/numen/lima seats for all existing identities on server startup

### 6. Testing Script (`scripts/dev/authority_smoke.sh`)

**Purpose**: Comprehensive API smoke testing for GOV-2 endpoints

**Tests**:
1. **Health Check** - Verifies `/api/health` returns 200
2. **Triad Fetch** - Tests `/api/authority/triad/{id}` with sample identity
3. **Flow Preview** - Validates `/api/authority/flow/preview` returns 6 scenarios
4. **Upward Authority Eval** - Tests `POST /api/authority/flow/eval` with upward direction
5. **Downward Authority Eval** - Tests downward authority with lima seat check

**Features**:
- Color-coded output (GREEN/RED/YELLOW/BLUE)
- jq-based JSON parsing and validation
- Checks for expected fields and values
- Validates authority engine logic (lima seat state affects downward authority)
- Detailed error reporting

**Usage**:
```bash
chmod +x scripts/dev/authority_smoke.sh
./scripts/dev/authority_smoke.sh
```

---

## Frontend Implementation (Finagler)

### 1. TypeScript API Client (`src/lib/api/authorityClient.ts`)

**Interfaces**:
```typescript
interface IdentitySeat {
  seat_id: string;
  layer: "terra" | "numen" | "lima";
  state: "EMPTY" | "ASSIGNED" | "OCCUPIED" | "FROZEN";
  updated_at?: string;
}

interface IdentityTriad {
  identity_id: string;
  complete: boolean;
  seats: IdentitySeat[];
  terra?: IdentitySeat;
  numen?: IdentitySeat;
  lima?: IdentitySeat;
  timestamp: string;
}

interface FlowPreview {
  scenarios: Array<{
    scenario: string;
    identity_id: string;
    direction: string;
    target_domain: string;
    expected: string;
  }>;
}

interface FlowEvalInput {
  identity_id: string;
  direction: "upward" | "downward" | "lateral";
  target_domain?: string;
  parent_approved?: boolean;
}

interface FlowEvalResult {
  identity_id: string;
  direction: string;
  target_domain?: string;
  allow: boolean;
  reason: string;
  terra: { state: string; has_authority: boolean };
  numen: { state: string; has_authority: boolean };
  lima: { state: string; has_authority: boolean };
  timestamp: string;
}
```

**Functions**:
```typescript
async function fetchTriad(baseUrl: string, identityId: string): Promise<IdentityTriad>
async function previewFlow(baseUrl: string): Promise<FlowPreview>
async function evalFlow(baseUrl: string, input: FlowEvalInput): Promise<FlowEvalResult>
```

**Constants**:
- `DEFAULT_BASE_URL = 'http://localhost:8080'`

**Error Handling**: All functions throw descriptive errors with HTTP status codes

### 2. React Hook (`src/hooks/useTriad.ts`)

**Signature**:
```typescript
function useTriad(
  baseUrl: string,
  identityId?: string
): UseTriadResult

interface UseTriadResult {
  data: IdentityTriad | null;
  loading: boolean;
  error: Error | null;
  refetch: () => void;
}
```

**Features**:
- Automatic fetch on mount when identityId provided
- Loading state management
- Error handling
- Cancellation on unmount (prevents memory leaks)
- Manual refetch capability via `refetch()`
- TypeScript type safety throughout

**Behavior**:
- Returns `null` data when no identityId provided
- Clears data/error when identityId changes
- Uses internal trigger counter for refetch mechanism

### 3. UI Component (`src/components/AuthorityPanel.jsx`)

**Component**: `AuthorityPanel`

**Props**:
```javascript
{
  baseUrl?: string,           // Default: 'http://localhost:8080'
  identityId?: string         // Required for data fetch
}
```

**Features**:

**Visual Design**:
- 3-column grid layout (responsive: stacks on mobile)
- Color-coded seat states:
  - EMPTY: Gray (#6b7280)
  - ASSIGNED: Blue (#3b82f6)
  - OCCUPIED: Green (#10b981)
  - FROZEN: Red (#ef4444)
- Emoji labels: 🌍 Terra, ✨ Numen, 🤝 Lima

**Data Display**:
- Identity ID header with refresh button
- Per-seat cards showing:
  - Seat type and emoji
  - Current state (color-coded badge)
  - Seat ID (first 8 chars)
  - Assignment timestamp
- Last updated timestamp (bottom right)

**State Handling**:
- No identity selected → Gray placeholder
- Loading → Blue loading indicator
- Error → Red error card with retry button
- No data → Yellow warning card
- Success → Full triad display

**Interactions**:
- Refresh button (top right) triggers manual refetch
- Retry button on errors

**Dependencies**:
- Uses `useTriad` hook for data fetching
- Tailwind CSS for styling
- React for component lifecycle

---

## Integration Flow

### Startup Sequence

1. **Database Connection** - `InitializeDatabase()`
2. **Table Bootstrap** - `BootstrapTables()` (creates identity_seats table if not exists)
3. **Phase S0** - Root seat setup
4. **GOV-1 Bootstrap** - `BootstrapIdentityTriads()` (creates terra/numen/lima for all identities)
5. **Authority Console Init** - Console components
6. **API Server Init** - TriadRepository created, routes registered
7. **HTTP Server Start** - Ready to serve requests

### Request Flow

**Triad Fetch**:
```
Browser → AuthorityPanel → useTriad → authorityClient.fetchTriad()
  → GET /api/authority/triad/{id}
  → TriadHandler.GetTriadByIdentity()
  → TriadRepository.GetIdentityTriad()
  → PostgreSQL query
  → Response (IdentityTriadDTO)
```

**Authority Evaluation**:
```
Browser → authorityClient.evalFlow()
  → POST /api/authority/flow/eval (JSON body)
  → TriadHandler.EvaluateFlow()
  → TriadRepository.GetIdentityTriad()
  → authority.EvaluateAuthority(direction, triad)
  → REGO policy evaluation (terra → numen → lima layers)
  → Response (FlowEvalResult)
```

---

## File Summary

**Backend (dis-core)**:
- `internal/api/authority/types.go` - DTOs and interfaces (70 lines)
- `internal/api/authority/handler.go` - API endpoints (230 lines)
- `internal/api/routes.go` - Route registration (modified)
- `internal/api/server.go` - Server initialization (modified)
- `cmd/dis-core/main.go` - Bootstrap integration (modified)
- `scripts/dev/authority_smoke.sh` - Smoke test script (180 lines)

**Frontend (Finagler)**:
- `src/lib/api/authorityClient.ts` - TypeScript API client (120 lines)
- `src/hooks/useTriad.ts` - React data-fetching hook (65 lines)
- `src/components/AuthorityPanel.jsx` - UI component (120 lines)

**Total**: ~800 new lines of code, 6 new files, 4 modified files

---

## Testing

### Smoke Test Results

**Command**: `./scripts/dev/authority_smoke.sh`

**Expected Output**:
```
=== Authority Console API Smoke Tests ===
[1/5] Health check... ✅ PASSED
[2/5] Triad fetch (identity: test-id-001)... ✅ PASSED
  Complete: true
  Seats: terra (OCCUPIED), numen (OCCUPIED), lima (OCCUPIED)
[3/5] Flow preview... ✅ PASSED
  Scenarios: 6
[4/5] Upward authority evaluation... ✅ PASSED
  Allow: true, Reason: "upward authority granted for ASSIGNED seat"
[5/5] Downward authority evaluation... ✅ PASSED
  Allow: true (if OCCUPIED) / false (if not OCCUPIED)
  Lima state: OCCUPIED

All tests passed! ✅
```

### Manual Testing

**Using curl**:
```bash
# Fetch triad
curl http://localhost:8080/api/authority/triad/test-id-001

# Preview flows
curl http://localhost:8080/api/authority/flow/preview

# Evaluate authority
curl -X POST http://localhost:8080/api/authority/flow/eval \
  -H "Content-Type: application/json" \
  -d '{
    "identity_id": "test-id-001",
    "direction": "upward",
    "target_domain": "parent-domain"
  }'
```

**Using Browser/Finagler**:
1. Import AuthorityPanel in any view
2. Pass identityId prop: `<AuthorityPanel identityId="test-id-001" />`
3. Observe 3-seat display with color-coded states
4. Click refresh to re-fetch data

---

## API Reference

### GET /api/authority/triad/{identityId}

**Description**: Fetch complete identity triad (terra/numen/lima)

**Response**:
```json
{
  "identity_id": "test-id-001",
  "complete": true,
  "seats": [
    {
      "seat_id": "abc123...",
      "layer": "terra",
      "state": "OCCUPIED",
      "updated_at": "2024-01-01T12:00:00Z"
    },
    {
      "seat_id": "def456...",
      "layer": "numen",
      "state": "OCCUPIED",
      "updated_at": "2024-01-01T12:00:00Z"
    },
    {
      "seat_id": "ghi789...",
      "layer": "lima",
      "state": "ASSIGNED",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Status Codes**:
- 200: Success
- 400: Missing identityId
- 404: Identity not found or incomplete triad

---

### GET /api/authority/flow/preview

**Description**: Get sample authority flow scenarios

**Response**:
```json
{
  "scenarios": [
    {
      "scenario": "Upward authority with ASSIGNED seat",
      "identity_id": "test-id-assigned",
      "direction": "upward",
      "target_domain": "parent-domain",
      "expected": "allow"
    },
    ...
  ]
}
```

**Status Codes**:
- 200: Success

---

### POST /api/authority/flow/eval

**Description**: Evaluate authority flow for identity

**Request**:
```json
{
  "identity_id": "test-id-001",
  "direction": "upward",
  "target_domain": "parent-domain",
  "parent_approved": false
}
```

**Response**:
```json
{
  "identity_id": "test-id-001",
  "direction": "upward",
  "target_domain": "parent-domain",
  "allow": true,
  "reason": "upward authority granted for ASSIGNED seat",
  "terra": {
    "state": "OCCUPIED",
    "has_authority": true
  },
  "numen": {
    "state": "OCCUPIED",
    "has_authority": true
  },
  "lima": {
    "state": "ASSIGNED",
    "has_authority": true
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Status Codes**:
- 200: Success
- 400: Invalid request body
- 404: Identity not found

---

## Dependencies

**Backend**:
- GOV-1 foundation (identity_seats table, TriadRepository, authority.EvaluateAuthority)
- Chi router v5 (URL parameter extraction)
- pgxpool (PostgreSQL connection)
- encoding/json (response serialization)

**Frontend**:
- React (hooks, component lifecycle)
- TypeScript (type safety)
- Tailwind CSS (styling)
- fetch API (HTTP requests)

---

## Next Steps (GOV-3)

**Write APIs** (Seat State Transitions):
- POST `/api/authority/triad/{id}/occupy` - Transition to OCCUPIED
- POST `/api/authority/triad/{id}/freeze` - Transition to FROZEN
- POST `/api/authority/triad/{id}/unfreeze` - Transition back to previous state
- POST `/api/authority/triad/{id}/reset` - Return to ASSIGNED

**Authorization Layer**:
- Who can modify triads? (governance rules)
- Audit logging for all seat changes
- Rate limiting and security controls

**Extended Visualization**:
- Historical state transitions
- Authority flow diagrams
- Bulk operations UI

**Testing**:
- Go unit tests for handlers
- TypeScript tests for client
- Integration tests for full stack
- Load testing for authority evaluation

---

## Verification Checklist

✅ DTO types created (types.go)
✅ API handlers implemented (handler.go)
✅ Routes registered in server (routes.go)
✅ TriadRepository initialized (server.go)
✅ Bootstrap integration (main.go)
✅ Smoke test script created (authority_smoke.sh)
✅ TypeScript client implemented (authorityClient.ts)
✅ React hook created (useTriad.ts)
✅ UI component implemented (AuthorityPanel.jsx)
✅ All compile errors resolved
✅ Documentation complete

**Status**: Phase GOV-2 implementation complete and ready for testing.

---

## Notes

**Design Decisions**:

1. **Read-Only First**: GOV-2 intentionally excludes write operations to ensure proper authority model validation before allowing state mutations

2. **Chi Router Usage**: Backend uses Chi for URL parameters (matches existing codebase patterns)

3. **Nil-Safe Registration**: Routes only registered if TriadRepository is non-nil (fail-safe initialization)

4. **Color-Coded UI**: Seat states use consistent color scheme across all views for quick visual recognition

5. **TypeScript Safety**: Full type definitions prevent common runtime errors in frontend

6. **Bootstrap Integration**: Auto-creates triads for all identities on startup (idempotent operation)

**Known Limitations**:

- No real-time updates (requires manual refresh)
- No historical state tracking in UI
- No batch operations support
- No WebSocket integration for live updates
- No authority flow visualization diagrams

**Future Enhancements**:

- Real-time triad updates via WebSocket
- Historical timeline view
- Authority flow graph visualization
- Batch triad operations
- Export/import functionality
- Advanced filtering and search
