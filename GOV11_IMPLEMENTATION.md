# GOV-11: Domain-Scoped Identity Projection & Corporeal Authentication

**Implementation Date**: November 14, 2024 (Initial Implementation - Phase A-D)
**Status**: 🔄 Partial Implementation (Core Infrastructure Complete)
**Phase**: GOV-11A, GOV-11B, GOV-11C, GOV-11D (Completed)

## Overview

GOV-11 extends GOV-9 (Authority Continuity) and GOV-10 (Identity Provenance) by introducing domain-scoped identity projections, foreign identity acceptance, and corporeal authentication logging. This phase unifies identity representation, projection, and authentication into a coherent, provenance-backed, sovereignty-preserving system.

## Core Concepts

### 1. dis.id (Sovereign Identity)
- Global, DIS-native identity for a person
- Never directly exposed to arbitrary domains
- Used internally to bind identity_receipts, membership seats, and authority decisions
- Lives in the person's corporeal domain and DIS-Core identity tables

### 2. domain.id (Domain-Scoped Identity Projection)
- A domain's local identifier for an actor
- Created when a person joins or interacts with that domain
- Stored in the domain's membership seat for that actor
- May take any domain-defined form (opaque token, handle, reference)
- Always bound to dis.id via receipts

### 3. Foreign Domain IDs
- A domain may authenticate a person using a domain.id from another domain
- Identity trust relationship, not authority relationship
- Each acceptance must be receipt-backed
- Examples: "Use Google", "Use GitHub", "Use BankID"

### 4. Corporeal Domain as Identity Provider
- Person's corporeal domain (e.g., domain.user.rick) as sovereign identity hub
- Can authenticate locally (password, passkey, biometric)
- Issue identity proofs/claims to other domains (IRL or online)
- Record private logs of authentications
- Manage domain.id mappings and foreign identities

### 5. DIS-Net Boundary
- All identity operations valid only within DIS-Net
- Recorded via dis-core with identity_receipts
- External systems may request/consume proofs
- Canonical source of truth is always DIS-Net

## Implementation Status

### ✅ Completed Components

#### GOV-11A: Data Model Extensions
1. **identity_receipts Schema Extensions**
   - Added columns: `target_domain_id`, `source_domain_id`, `external_subject`, `channel`, `method`, `scope`
   - Updated `identity_lineage_view` with new fields
   - Foreign key constraints for referential integrity
   - Indexes for efficient querying
   - Migration: `20251114_gov11a_identity_projections.sql`

2. **Go Struct Updates**
   - Extended `IdentityReceipt` struct with GOV-11 fields
   - Updated `LineageEntry` for projection data
   - Modified `RecordIdentityReceipt()` to handle 16 fields (was 10)
   - Updated `GetLastReceipt()` and `GetIdentityLineage()` for new columns

#### GOV-11B: New Identity Actions
1. **New IdentityAction Constants**
   ```go
   IdentityDomainIDCreateV1 = "identity.domainid.create.v1"  // Domain creates local identity
   IdentityDomainIDUpdateV1 = "identity.domainid.update.v1"  // Domain updates/rotates identity
   IdentityAcceptV1         = "identity.accept.v1"           // Domain accepts foreign identity
   IdentityAcceptRevokeV1   = "identity.accept.revoke.v1"    // Domain revokes foreign identity
   IdentityIRLAuthV1        = "identity.irlauth.v1"          // Corporeal IRL authentication
   ```

2. **Helper Functions**
   - `RecordDomainIDCreation()` - Create domain.id projection receipt
   - `RecordForeignIdentityAcceptance()` - Accept external identity receipt
   - `RecordIRLAuthentication()` - Log IRL auth event with corporeal domain
   - `RecordForeignIdentityRevocation()` - Revoke foreign identity acceptance
   - All functions maintain hash chain integrity with prev_id linking

#### GOV-11C: Identity Projections API
1. **GET /api/identity/projections/{actor_id}**
   - Returns comprehensive summary of all domain projections
   - Aggregates local domain.ids and foreign acceptances per domain
   - Includes integrity status and receipt counts
   - Response structure:
     ```json
     {
       "actor_id": "uuid",
       "projections": [
         {
           "domain_id": "uuid",
           "domain_name": "string",
           "local_identity": "domain.id value",
           "accepted_identities": [
             {
               "source_domain_id": "uuid",
               "external_subject": "string",
               "scope": "auth-only",
               "receipt_id": "uuid",
               "accepted_at": "timestamp",
               "active": true
             }
           ],
           "receipt_count": 5,
           "integrity_valid": true,
           "last_activity": "timestamp"
         }
       ],
       "total_domains": 3,
       "total_receipts": 15,
       "integrity_status": "all-valid"
     }
     ```

2. **GET /api/domain/{domain_id}/member/{actor_id}/identity**
   - Domain-specific identity view for a member
   - Filters projections summary to single domain
   - Shows local domain.id and accepted foreign identities
   - Returns 404 if no projection found

#### GOV-11D: Corporeal Identity Log
1. **corporeal_identity_log Table**
   - Stores private IRL authentication events
   - Links to identity_receipts via receipt_id
   - Fields: actor_id, corporeal_domain_id, target_domain_id, method, channel, metadata
   - Privacy-aware metadata (JSONB)
   - Access restricted to actor and their corporeal domain

2. **corporeal_identity_log_view**
   - Enriched view with domain names and receipt hash
   - Joins domains and identity_receipts tables
   - Enables efficient filtering by actor, corporeal domain, or target domain

3. **Migration: 20251114_gov11d_corporeal_identity_log.sql**
   - CREATE TABLE with indexes
   - CREATE VIEW with joins
   - Comments documenting access restrictions

### 🔄 Pending Components

#### GOV-11A: Membership Seat Extensions
- **Status**: Not Started
- **Requirements**:
  - Add `domain_identity` field to membership seat representation
  - Add `accepted_identities` array to membership seat
  - Structure for foreign identity acceptance tracking
  - Database schema changes (seats table or JSONB fields)

#### GOV-11B: Mutation Engine Integration
- **Status**: Not Started
- **Requirements**:
  - Emit identity.domainid.create.v1 when membership seats created
  - Emit identity.binding.update.v1 on domain changes
  - Integrate with existing seat transition logic
  - Log to corporeal_identity_log for IRL events

#### GOV-11E: Finagler UI
- **Status**: Not Started
- **Requirements**:
  - Identity Projections view component
  - Membership seat identity panels
  - Corporeal Identity Vault/Log viewer
  - IntegrityBadge for local vs foreign projections
  - Visual distinction for revoked vs active acceptances
  - Export functionality for audit reports

#### GOV-11F: REGO Policy Integration
- **Status**: Not Started
- **Requirements**:
  - Policy rule: no domain action without valid identity projection
  - DIS-Net boundary enforcement in REGO
  - Authentication ≠ Authorization separation
  - Foreign acceptance doesn't grant governance seats
  - Identity mutation checks (authorized routes only)

#### GOV-11: Tests & Documentation
- **Status**: Not Started
- **Requirements**:
  - Unit tests for new identity action helpers
  - Integration tests for API endpoints
  - Policy tests for REGO rules
  - End-to-end scenario tests:
    * Actor joins domain → domain.id creation
    * Domain accepts foreign ID → acceptance receipt
    * Corporeal authenticates IRL → log entry
    * Policy denies action without projection
  - Comprehensive documentation of all components

## Architecture Decisions

### Hash Chain Continuity
- All new identity actions extend the existing hash chain from GOV-10
- `prev_id` links maintained across identity.root.v1 → identity.domainid.create.v1 → identity.accept.v1
- Single actor-centric chain across all identity operations

### Foreign Key Constraints
- `target_domain_id` and `source_domain_id` reference `domains(id)`
- `receipt_id` in corporeal_identity_log references `identity_receipts(id)`
- Ensures referential integrity at database level

### Privacy Considerations
- corporeal_identity_log is **private** (not exposed to general APIs)
- Metadata field is JSONB for flexible privacy-aware context
- Access control required for IRL log endpoints (to be implemented)
- External subject values stored but not exposed without authorization

### Separation of Concerns
- **Identity Projection**: domain.id creation and management
- **Foreign Acceptance**: Trust relationships between domains
- **IRL Authentication**: Corporeal domain as identity provider
- **Policy Enforcement**: Separate REGO layer for presence checks

## Database Schema

### Extended identity_receipts Table
```sql
ALTER TABLE identity_receipts
  ADD COLUMN target_domain_id UUID REFERENCES domains(id),
  ADD COLUMN source_domain_id UUID REFERENCES domains(id),
  ADD COLUMN external_subject TEXT,
  ADD COLUMN channel TEXT,
  ADD COLUMN method TEXT,
  ADD COLUMN scope TEXT;
```

### New corporeal_identity_log Table
```sql
CREATE TABLE corporeal_identity_log (
  id UUID PRIMARY KEY,
  actor_id UUID NOT NULL,
  corporeal_domain_id UUID REFERENCES domains(id),
  target_domain_id UUID REFERENCES domains(id),
  receipt_id UUID REFERENCES identity_receipts(id),
  method TEXT NOT NULL,
  channel TEXT NOT NULL,
  metadata JSONB DEFAULT '{}'::jsonb,
  logged_at TIMESTAMPTZ DEFAULT now()
);
```

## API Examples

### Get Identity Projections
```bash
curl -X GET http://localhost:8080/api/identity/projections/{actor_id}
```

Response shows all domain projections, local identities, and foreign acceptances with integrity status.

### Get Domain Member Identity
```bash
curl -X GET http://localhost:8080/api/domain/{domain_id}/member/{actor_id}/identity
```

Response shows single domain's view of actor's identity (local domain.id and accepted foreign IDs).

### Create Domain ID (Programmatic)
```go
store := identity.NewIdentityReceiptStore(db)
receipt, err := store.RecordDomainIDCreation(
    ctx,
    actorID,
    domainID,
    "user_handle_123",  // domain.id value
    map[string]interface{}{"created_by": "admin"},
)
```

### Accept Foreign Identity (Programmatic)
```go
scope := "auth-only"
receipt, err := store.RecordForeignIdentityAcceptance(
    ctx,
    actorID,
    acceptingDomainID,
    sourceDomainID,  // e.g., Google's domain ID
    "external_token_abc123",
    &scope,
    map[string]interface{}{"trust_level": "verified"},
)
```

### Log IRL Authentication (Programmatic)
```go
receipt, err := store.RecordIRLAuthentication(
    ctx,
    actorID,
    corporealDomainID,
    targetDomainID,  // e.g., bank's domain ID
    "biometric",     // method
    "mobile_app",    // channel
    map[string]interface{}{"session_id": "xyz", "location": "redacted"},
)
```

## Files Created

1. `db/migrations/20251114_gov11a_identity_projections.sql` (97 lines)
2. `db/migrations/20251114_gov11d_corporeal_identity_log.sql` (88 lines)
3. `internal/api/identity_projections.go` (214 lines)
4. Extended `internal/identity/identity_receipts.go` (+193 lines)
5. Extended `internal/identity/identity_lineage.go` (+14 lines)

## Files Modified

1. `internal/identity/identity_receipts.go` - Added GOV-11 actions and helper functions
2. `internal/identity/identity_lineage.go` - Extended LineageEntry with projection fields
3. `internal/api/routes.go` - Registered new projection endpoints

## Migrations Applied

```bash
# GOV-11A: Schema extensions
psql ... -f db/migrations/20251114_gov11a_identity_projections.sql
# Output: ✅ 6 new columns, updated view, 10 existing receipts preserved

# GOV-11D: Corporeal log
psql ... -f db/migrations/20251114_gov11d_corporeal_identity_log.sql
# Output: ✅ Table and view created, 0 initial log entries
```

## Build Verification

```bash
go build -o dis-core-gov11 cmd/dis-core/main.go
```

**Status**: ✅ Build successful

## Integration Points

### With GOV-9 (Authority Continuity)
- Authority receipts track **domain-level** governance decisions
- Identity receipts track **actor-level** identity operations
- Both use same hash chain pattern with prev_id linking
- Complementary audit trails

### With GOV-10 (Identity Provenance)
- Extends existing identity_receipts table and hash chains
- Adds 5 new action types to existing 5 (total: 10 identity actions)
- Reuses GetIdentityLineage() for projection reconstruction
- Maintains backward compatibility with GOV-10 receipts

### Future: With GOV-12 (Cross-Ledger Cohesion)
- Identity projections will enable cross-domain dispute resolution
- Foreign acceptance receipts support identity verification across domains
- Corporeal logs provide IRL authentication proof for legal scenarios

## Security & Privacy

### Immutability
- All identity actions are INSERT-only receipts
- Hash chains prevent tampering
- Foreign key constraints enforce referential integrity

### Privacy Protection
- corporeal_identity_log is private (access control required)
- Metadata stored as JSONB for flexible privacy controls
- External subject values not exposed without authorization
- DIS-Net boundary prevents external mutation

### Consent Tracking
- Every identity action has `consent_by` field (domain ID)
- Foreign acceptance explicitly recorded with scope
- Revocation receipts document withdrawal of consent

## Known Limitations & TODOs

1. **Membership Seat Integration**: Domain identity fields not yet added to seat structures
2. **Mutation Engine**: Not yet emitting domain.id creation receipts automatically
3. **UI Components**: Finagler views for projections and corporeal logs not implemented
4. **REGO Policies**: Identity presence checks not enforced in policy evaluation
5. **Access Control**: IRL log API endpoints need authentication/authorization
6. **Tests**: Comprehensive test suite for GOV-11 features not yet written
7. **Documentation**: API docs and usage guides need expansion

## Next Steps

### Immediate (Complete GOV-11)
1. Extend membership seat model with domain_identity and accepted_identities
2. Integrate identity.domainid.create.v1 emission in mutation engine
3. Create Finagler Identity Projections view component
4. Add REGO policy rules for identity presence enforcement
5. Write comprehensive test suite
6. Create full API documentation

### Future Enhancements
1. Domain.id rotation/update workflows
2. Bulk foreign acceptance management
3. IRL authentication QR code generation
4. Cross-domain identity verification API
5. Corporeal identity vault export (PDF/CSV)
6. Identity projection analytics dashboard

## Related Governance Phases

- **GOV-9**: Authority Continuity & Receipt Lineage (domain-level)
- **GOV-10**: Identity Provenance & Alias Receipt Integration (actor-level)
- **GOV-11**: Domain-Scoped Identity Projection & Corporeal Authentication (this phase)
- **GOV-12** (Planned): Cross-Ledger Cohesion & Global Identity Dispute Resolution

## Key Phrase for Future Reference

**"Invoke GOV-11 Identity Projection Membrane"**

Use this phrase to resume GOV-11 implementation work, particularly for:
- Membership seat model extensions
- Mutation engine integration
- Finagler UI components
- REGO policy enforcement
- Comprehensive testing

## Notes

- GOV-11 introduces a three-layer identity model: dis.id (sovereign), domain.id (projection), foreign acceptance (trust)
- Corporeal domain concept establishes persons as sovereign identity providers
- DIS-Net boundary preserves identity sovereignty (no external unilateral mutation)
- Foreign acceptance enables interoperability without sacrificing sovereignty
- IRL authentication logs bridge physical and digital identity
