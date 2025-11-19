# GOV-1 Implementation File Tree

All files created or modified for GOV-1 Domain Governance Foundation.

## Documentation (3 files)

```
docs/
└── GOV-1-domain-governance.md          # 671 lines - Complete specification

phases/
├── phase_gov1_instructions.md          # 800+ lines - Implementation guide
├── GOV1_STATUS.md                      # Status report with metrics
└── GOV1_FILES.md                       # This file
```

## Database (1 file)

```
db/migrations/
└── 20251112_add_identity_seats_table.sql   # Identity triad schema
    ├── CREATE TABLE identity_seats
    ├── CREATE INDEX (3 indexes)
    ├── CREATE VIEW identity_triad_status
    └── COMMENT ON TABLE/COLUMN
```

## Go Code (4 files)

```
internal/
├── identity/                           # New package
│   ├── triad_model.go                  # IdentitySeat, IdentityTriad structs
│   │   ├── type IdentitySeat
│   │   ├── type IdentityTriad
│   │   ├── const SeatState* (4 states)
│   │   ├── const SeatType* (3 types)
│   │   └── Methods: HasAuthority, IsOccupied, IsFrozen, IsComplete
│   └── triad_repo.go                   # Repository (200+ lines, 8 methods)
│       ├── type TriadRepository
│       ├── func NewTriadRepository
│       ├── func CreateIdentitySeat
│       ├── func InitializeTriad
│       ├── func GetIdentitySeat
│       ├── func GetIdentityTriad
│       ├── func UpdateSeatState
│       ├── func GetAllIdentities
│       └── func GetMissingTriads
│
└── authority/                          # New package
    └── flow_engine.go                  # Authority flow evaluation
        ├── type AuthorityDirection (upward/downward/lateral)
        ├── type EvaluationResult
        ├── func EvaluateAuthority (GOV-1 rules)
        └── func ValidateSeatForAction

cmd/dis-core/bootstrap/
└── identities.go                       # Bootstrap logic
    └── func BootstrapIdentityTriads    # Auto-create triads
```

## REGO Policies (3 files)

```
policy/
├── terra.rego                          # Layer 1: Existence
│   ├── package dis.terra
│   ├── person_exists
│   ├── allow (ASSIGNED/OCCUPIED)
│   ├── deny (EMPTY/FROZEN/missing)
│   └── export_allow, export_deny, export_reason
│
├── numen.rego                          # Layer 2: Meaning
│   ├── package dis.numen
│   ├── import data.dis.terra
│   ├── schema_valid
│   ├── canonical_valid (canonical_greedy check)
│   ├── numen_seat_valid
│   ├── allow (if terra allows + numen checks pass)
│   ├── deny (numen violations)
│   └── export_allow, export_deny, export_reason
│
└── lima.rego                           # Layer 3: Authority
    ├── package dis.lima
    ├── import data.dis.terra
    ├── import data.dis.numen
    ├── lima_seat_valid
    ├── upward_authority (ASSIGNED+)
    ├── downward_authority (OCCUPIED + approval)
    ├── downward_approved (within-domain or parent-approved)
    ├── lateral_authority (same domain)
    ├── allow (if terra+numen allow + lima authority)
    ├── deny (authority violations)
    └── export_allow, export_deny, export_reason
```

## Scripts (1 file)

```
scripts/
└── verify_gov1.sh                      # 12 automated tests
    ├── Test 1: Database schema
    ├── Test 2: REGO files exist
    ├── Test 3: Go compilation
    ├── Test 4: Package structure
    ├── Test 5: Bootstrap code
    ├── Test 6: Migration file
    ├── Test 7: REGO content
    ├── Test 8: REGO dependencies
    ├── Test 9: Authority engine
    ├── Test 10: Triad repository
    ├── Test 11: Seat state constants
    └── Test 12: Seat type constants
```

## Summary

**Total Files:** 11 (3 docs + 1 SQL + 4 Go + 3 REGO + 1 script)

**Lines of Code:**
- Documentation: ~1,500 lines
- Go: ~600 lines
- REGO: ~300 lines
- SQL: ~50 lines
- Shell: ~180 lines
- **Total: ~2,630 lines**

**New Packages:**
- `internal/identity` (triad management)
- `internal/authority` (flow engine)

**Key Concepts:**
- Identity Triad: terra (existence) + numen (meaning) + lima (consent)
- Seat States: EMPTY → ASSIGNED → OCCUPIED → FROZEN
- Authority Flow: upward / downward / lateral
- Policy Stack: terra → numen → lima → domain → seat

All files follow GOV-1 specification defined in `docs/GOV-1-domain-governance.md`.

