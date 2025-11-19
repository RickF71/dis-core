# GOV-12: Alias Canon & DSCI Integration

**Phase:** Canonical Alias Taxonomy (AUTO / RELATIONSHIP / MASK)
**Date:** November 13, 2025
**Status:** ✅ Backend Complete | 📝 Documentation & Testing In Progress

---

## Overview

GOV-12 establishes the **canonical alias taxonomy** for DIS-Core, defining three distinct alias types with clear semantics:

1. **AUTO** - Structural, non-volitional graph bindings
2. **RELATIONSHIP** - Volitional, domain-scoped identities
3. **MASK** - Ephemeral browsing personas with minimal authority

This phase prepares the foundation for full DSCI (Domain-Signed Contract Inheritance) integration in future phases.

---

## Canonical Alias Taxonomy

### AUTO Aliases
**Purpose:** Structural identity bindings created automatically by system flows

**Characteristics:**
- Non-volitional (created by system, not user choice)
- Represents structural membership/relationships
- Cannot be used as user-selectable personas
- One special case: `is_corporeal_auto = true` for corporeal root seat

**Creation Triggers:**
- Corporeal domain initialization
- Seat instantiation (structural membership)
- Domain branching (inherited structural relationships)

**Semantics:**
- `owner_domain_id` = person's corporeal/identity domain
- `target_domain_id` = domain granting structural membership
- For corporeal root: both IDs = corporeal domain (self-referential)

### RELATIONSHIP Aliases
**Purpose:** Volitional identity chosen by person within a specific domain

**Characteristics:**
- User explicitly chooses alias name
- Scoped to specific target domain
- Backed by consent (DSCI contract when available)
- Primary mechanism for domain participation

**Creation Triggers:**
- User joins domain and chooses handle
- Domain issues identifier with user consent

**Semantics:**
- `owner_domain_id` = person's identity domain
- `target_domain_id` = domain being joined
- `alias_name` = user-chosen or domain-proposed identifier
- `dsci_contract_id` = links to consent receipt (future)

### MASK Aliases
**Purpose:** Incognito-but-authentic browsing with minimal authority

**Characteristics:**
- Ephemeral, short-lived
- Default-deny for privileged operations
- DIS_UID binding preserved (authentic but anonymous)
- Can have TTL metadata

**Creation Triggers:**
- User requests exploration/browsing mode
- Temporary participation without full identity revelation

**Semantics:**
- `owner_domain_id` = person's identity domain
- `target_domain_id` = same as owner (self-targeting)
- `alias_name` = auto-generated "mask-XXXX" or user-provided
- `metadata.ttl_days` = expiration hint

---

## Database Schema

### `aliases` Table

```sql
CREATE TABLE aliases (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_domain_id     UUID NOT NULL REFERENCES domains(id),
    target_domain_id    UUID NOT NULL REFERENCES domains(id),
    alias_name          TEXT NOT NULL,
    alias_type          TEXT NOT NULL CHECK (alias_type IN ('AUTO', 'RELATIONSHIP', 'MASK')),
    is_corporeal_auto   BOOLEAN NOT NULL DEFAULT FALSE,
    dsci_contract_id    UUID NULL,  -- Future: links to contracts table
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    retired_at          TIMESTAMPTZ NULL,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb
);
```

**Constraints:**
- `check_corporeal_auto_must_be_auto`: is_corporeal_auto can only be true for AUTO aliases
- `unique_active_alias`: One active alias per (owner, target, name) triple

**Indexes:**
- `idx_aliases_owner_domain(owner_domain_id, alias_type)`
- `idx_aliases_target_domain(target_domain_id, alias_type)`
- `idx_aliases_dsci_contract(dsci_contract_id) WHERE NOT NULL`
- `idx_aliases_active(owner_domain_id, alias_type) WHERE retired_at IS NULL`

### Views

**`active_aliases_by_type`** - Summary statistics by alias type
**`alias_lineage`** - Aliases with resolved domain names

---

## Go Implementation

### Domain Model (`internal/domain/alias.go`)

```go
type AliasType string
const (
    AliasTypeAuto         AliasType = "AUTO"
    AliasTypeRelationship AliasType = "RELATIONSHIP"
    AliasTypeMask         AliasType = "MASK"
)

type Alias struct {
    ID              uuid.UUID
    OwnerDomainID   uuid.UUID
    TargetDomainID  uuid.UUID
    AliasName       string
    AliasType       AliasType
    IsCorporealAuto bool
    DSCIContractID  *uuid.UUID
    CreatedAt       time.Time
    RetiredAt       *time.Time
    Metadata        AliasMetadata
}
```

**Helper Methods:**
- `Status() AliasStatus` - Returns ACTIVE or RETIRED
- `IsActive() bool`
- `IsMask() bool`, `IsAuto() bool`, `IsRelationship() bool`
- `GetTTLDays() int` - For MASK aliases
- `GetDisplayName() string` - From metadata or alias_name

### Repository (`internal/repo/alias_repo.go`)

```go
type AliasRepository struct {
    pool *pgxpool.Pool
}
```

**Methods:**
- `CreateAlias(ctx, *Alias) error`
- `RetireAlias(ctx, id, reason) error`
- `GetAliasesForOwner(ctx, ownerID, includeRetired) ([]Alias, error)`
- `GetAliasesForTarget(ctx, targetID, includeRetired) ([]Alias, error)`
- `GetAliasesByType(ctx, ownerID, type, includeRetired) ([]Alias, error)`
- `GetActiveMaskAliasesForOwner(ctx, ownerID) ([]Alias, error)`
- `GetCorporealAutoAlias(ctx, ownerID) (*Alias, error)`
- `FindActiveAliasByName(ctx, owner, target, name) (*Alias, error)`

### Service Layer (`internal/services/alias_service.go`)

```go
type AliasService struct {
    repo *AliasRepository
}
```

**AUTO Alias Creation:**
- `EnsureCorporealRootAlias(ctx, corporealDomainID) (*Alias, error)`
  - Creates/returns corporeal AUTO alias
  - Sets `is_corporeal_auto = true`
  - Idempotent (returns existing if present)

- `EnsureStructuralAlias(ctx, ownerID, targetID) (*Alias, error)`
  - Creates AUTO alias for structural membership
  - Used during seat instantiation
  - Idempotent

**RELATIONSHIP Alias Creation:**
- `CreateRelationshipAlias(ctx, ownerID, targetID, name, contractID, metadata) (*Alias, error)`
  - User-volitional identity creation
  - Validates name uniqueness
  - Links to DSCI contract (when available)

**MASK Alias Creation:**
- `CreateMaskAlias(ctx, ownerID, requestedName, ttlDays, metadata) (*Alias, error)`
  - Generates random name if not provided ("mask-XXXX")
  - Sets TTL metadata
  - Self-targeting (owner = target)

---

## API Endpoints (Planned)

### GET /api/domain/:id/aliases
Retrieves all aliases for a domain (as owner or target)

**Query Parameters:**
- `includeRetired=true|false` - Include retired aliases
- `type=AUTO|RELATIONSHIP|MASK` - Filter by type

**Response:**
```json
{
  "aliases": [...],
  "count": 5
}
```

### POST /api/domain/:id/alias/relationship
Creates a RELATIONSHIP alias

**Request:**
```json
{
  "target_domain_id": "uuid",
  "alias_name": "myhandle",
  "metadata": {}
}
```

**Response:**
```json
{
  "id": "uuid",
  "alias_type": "RELATIONSHIP",
  "alias_name": "myhandle",
  ...
}
```

### POST /api/domain/:id/alias/mask
Creates a MASK alias

**Request:**
```json
{
  "requested_name": "optional",
  "ttl_days": 30,
  "metadata": {}
}
```

**Response:**
```json
{
  "id": "uuid",
  "alias_type": "MASK",
  "alias_name": "mask-abc123",
  ...
}
```

---

## Receipt Integration

### Existing Receipt Types (GOV-10)
- `identity.alias.add.v1` - Alias addition
- `identity.alias.remove.v1` - Alias removal

### GOV-12 Enhancement
These receipts now link to canonical alias model:

**Payload for alias.add.v1:**
```json
{
  "alias_id": "uuid",
  "owner_domain_id": "uuid",
  "target_domain_id": "uuid",
  "alias_type": "RELATIONSHIP",
  "alias_name": "myhandle",
  "dsci_contract_id": "uuid",
  "created_at": "ISO8601",
  "by": "identity/seat"
}
```

**Payload for alias.retire.v1:**
```json
{
  "alias_id": "uuid",
  "retired_at": "ISO8601",
  "reason": "user request",
  "by": "identity/seat"
}
```

---

## REGO Policy Hooks (Future)

### AUTO Aliases
**Policy Semantics:**
- Treat as structural edges in identity graph
- Cannot be used for user-selectable persona actions
- Authority flows through seat, not alias directly
- Used for lineage/provenance queries

### RELATIONSHIP Aliases
**Policy Semantics:**
- Grant domain-specific capabilities based on alias scope
- Must reference DSCI contract for authority
- Can be revoked if contract is terminated
- Primary persona for domain interactions

### MASK Aliases
**Policy Semantics:**
- **Default: DENY ALL privileged operations**
- Allow: read-only discovery, public content access
- Allow: explicitly whitelisted low-risk actions
- Require: explicit DSCI authorization for any privileged action
- No persistent state changes unless explicitly authorized

**Example REGO Rule (Conceptual):**
```rego
allow {
    input.alias_type == "MASK"
    input.action == "read"
    input.resource.public == true
}

deny {
    input.alias_type == "MASK"
    input.action in ["write", "admin", "grant"]
}
```

---

## Migration Path

### Phase 1: Foundation (GOV-12 - Current)
✅ Database schema with canonical taxonomy
✅ Go domain model and repository
✅ Service layer with semantic enforcement
📝 API endpoints (planned)
📝 Receipt integration (pending)

### Phase 2: DSCI Integration (GOV-13)
- Create `contracts` table
- Add `contracts.branch_inheritance` field
- Wire `dsci_contract_id` foreign key
- Implement contract-based seat instantiation
- Link RELATIONSHIP aliases to contracts

### Phase 3: REGO Enforcement (GOV-14)
- Inject alias context into policy evaluation
- Implement default-deny for MASK aliases
- Enforce RELATIONSHIP alias scoping
- Add lineage-based permission inheritance

### Phase 4: UI & Visualization (GOV-15)
- Finagler alias management interface
- Alias type indicators (badges/icons)
- Relationship alias creation flow
- Mask alias ephemeral session UI

---

## Testing Strategy

### Unit Tests
- Alias model validation
- Repository CRUD operations
- Service semantic enforcement
  - AUTO: idempotency, corporeal vs structural
  - RELATIONSHIP: name conflict detection
  - MASK: name generation, TTL handling

### Integration Tests
- Migration round-trip
- Constraint enforcement
- View queries

### API Tests (When Implemented)
- Alias listing with filters
- RELATIONSHIP alias creation
- MASK alias creation
- Retirement flow

---

## Security Considerations

### Invariants
1. **AUTO aliases are system-created only** - No public API endpoint
2. **Corporeal AUTO is unique** - One per corporeal domain, enforced by constraint
3. **Active aliases are unique per scope** - EXCLUDE constraint prevents duplicates
4. **Retired aliases are immutable** - `retired_at` is append-only

### DSCI Linkage
- `dsci_contract_id` field prepared for future contract table
- RELATIONSHIP aliases must eventually require DSCI contract
- Current implementation allows NULL (grace period for migration)

### MASK Alias Risks
- Default-deny for privileged operations (REGO enforcement required)
- TTL metadata is advisory, not enforced yet
- No automatic cleanup mechanism (future: cron job)

---

## Next Steps

### Immediate (GOV-12 Completion)
1. ✅ Database migration
2. ✅ Domain model
3. ✅ Repository layer
4. ✅ Service layer
5. ⏳ API endpoints (stub implementations)
6. ⏳ Receipt integration
7. ⏳ Unit tests
8. ⏳ Documentation complete

### Follow-up Phases
- **GOV-13:** Contracts table + DSCI wiring
- **GOV-14:** REGO policy enforcement
- **GOV-15:** Finagler UI integration

---

## Files Created

### Database
- `db/migrations/20251113_gov12_alias_canon.sql`

### Backend
- `internal/domain/alias.go` (domain model, ~230 lines)
- `internal/repo/alias_repo.go` (repository, ~290 lines)
- `internal/services/alias_service.go` (service layer, ~210 lines)

### Documentation
- `docs/GOV12_ALIAS_CANON.md` (this file)

---

## Commit Message

```
GOV-12: Alias Canon & DSCI-ready alias model

- Created canonical aliases table with AUTO/RELATIONSHIP/MASK taxonomy
- Added alias_type constraint and corporeal_auto enforcement
- Implemented Alias domain model with helper methods
- Created AliasRepository with full CRUD operations
- Implemented AliasService with semantic enforcement:
  - EnsureCorporealRootAlias for corporeal domains
  - EnsureStructuralAlias for seat instantiation
  - CreateRelationshipAlias for volitional identity
  - CreateMaskAlias for ephemeral browsing
- Added database views for alias analytics
- Prepared DSCI contract linkage (pending contracts table)
- Documented REGO policy hooks for future enforcement

All backend foundations complete ✅
```

---

## Phase Handoff Summary

**Phase ID:** GOV-12
**Phase Name:** Alias Canon & DSCI Integration (Core Alias Model)
**Repo:** dis-core
**Status:** Backend Complete, API/Tests Pending

**Backend Complete:**
- ✅ Database schema with constraints
- ✅ Domain model with helpers
- ✅ Repository with CRUD
- ✅ Service with semantic flows

**Pending:**
- API endpoints
- Receipt wiring
- Unit tests
- Integration with existing identity flows

**Next Phase:** GOV-13 will create contracts table and wire DSCI contract IDs to RELATIONSHIP aliases, enabling full consent-based seat instantiation.
