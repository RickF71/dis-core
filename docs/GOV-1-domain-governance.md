# GOV-1: Domain Governance Foundation

**Version:** 1.0
**Date:** November 12, 2025
**Status:** Foundational Specification

This document defines the mathematical and structural foundation of DIS domain governance, establishing the identity triad, seat hierarchy, domain creation lifecycle, authority flow model, and REGO policy binding.

---

## 1. Identity Triad (Universal Assignment)

Every person in the DIS system is automatically assigned three foundational seats that represent their existence, meaning, and consent. These seats bind to the person's identity and follow them across all domains.

### 1.1 Terra.Member (Existence Seat)

**Purpose:** Establishes the physical/existential binding of a person to the DIS system.

**Authority:** None (existence only)

**Lifecycle:**
- `EMPTY` → No person bound
- `ASSIGNED` → Person bound, identity verified
- `OCCUPIED` → Person actively exists in the system

**REGO Outline (terra.policy):**
```rego
package dis.terra

# Terra policy: Existence rules
default allow = false

# Person exists if they have a valid terra.member binding
person_exists {
    input.identity_id != null
    input.terra_seat_id != null
    data.identities[input.identity_id].terra_bound == true
}

# Deny if terra seat is not ASSIGNED or OCCUPIED
deny {
    input.action == "domain.join"
    not person_exists
}
```

### 1.2 Numen.Member (Meaning Seat)

**Purpose:** Establishes the semantic/meaning layer for a person's actions and data.

**Authority:** None (meaning interpretation only)

**Lifecycle:**
- `EMPTY` → No meaning context assigned
- `ASSIGNED` → Meaning context bound
- `OCCUPIED` → Person actively using meaning layer

**REGO Outline (numen.policy):**
```rego
package dis.numen

# Numen policy: Meaning constraints
default allow = false

# Meaning must be consistent with schema
meaning_valid {
    input.schema_ref != null
    input.payload_type != null
    data.schemas[input.schema_ref].valid == true
}

# Deny if meaning context is inconsistent
deny {
    not meaning_valid
}

# Deny if canonical slots are missing
deny {
    input.payload
    not input.payload.canonical_greedy
}
```

### 1.3 Lima.Member (Consent Seat)

**Purpose:** Establishes consent and authority request capability.

**Authority:** Can request authority actions (upward reporting, lateral consent)

**Lifecycle:**
- `EMPTY` → No consent binding
- `ASSIGNED` → Consent capability bound
- `OCCUPIED` → Person actively exercising consent

**REGO Outline (lima.policy):**
```rego
package dis.lima

# Lima policy: Consent and authority requests
default allow = false

# Allow authority requests if lima seat is OCCUPIED
allow {
    input.action == "authority.request"
    input.lima_seat_state == "OCCUPIED"
    input.requester_id == input.identity_id
}

# Allow upward reporting (no approval needed)
allow {
    input.direction == "upward"
    input.lima_seat_state != "EMPTY"
}

# Deny downward authority unless approved
deny {
    input.direction == "downward"
    not input.parent_approved
}
```

---

## 2. Required Seats for Every Domain

Every domain must have two core seats that govern membership and authority.

### 2.1 Domain.Member (Membership Seat)

**Purpose:** Binds a person's identity to a specific domain.

**Authority:** None (membership only)

**Lifecycle:**
- `EMPTY` → No member bound
- `ASSIGNED` → Member invited/pending
- `OCCUPIED` → Member actively participating

**Behaviors:**
- Member can read domain schema
- Member can report upward to root
- Member can request actions (requires root approval)

### 2.2 Domain.Root (Governance Seat)

**Purpose:** Domain-level governance and authority.

**Authority:** Full domain governance (approve actions, manage members, create child domains)

**Lifecycle:**
- `EMPTY` → No root assigned (domain inactive)
- `ASSIGNED` → Root appointed/pending
- `OCCUPIED` → Root actively governing

**Behaviors:**
- Root can approve member requests
- Root can freeze/unfreeze seats
- Root can appoint additional members
- Root can create child domains (with parent approval)
- Root can tighten domain policies

---

## 3. Seat State Machine

All seats follow a universal state machine:

```
     ┌─────────┐
     │  EMPTY  │  (No identity bound, no authority)
     └────┬────┘
          │ assign(identity_id, seat_type)
          ▼
   ┌──────────┐
   │ ASSIGNED │  (Identity bound, upward authority only)
   └─────┬────┘
         │ occupy()
         ▼
   ┌──────────┐
   │ OCCUPIED │  (Full authority: up + down)
   └─────┬────┘
         │ freeze() / detach()
         ▼
   ┌──────────┐
   │ FROZEN/  │  (Authority suspended)
   │ DETACHED │
   └──────────┘
```

**Authority by State:**

| State | Upward Authority | Downward Authority | Governance |
|-------|------------------|-------------------|------------|
| `EMPTY` | ❌ None | ❌ None | ❌ None |
| `ASSIGNED` | ✅ Reporting only | ❌ None | ❌ None |
| `OCCUPIED` | ✅ Full | ✅ Full (with approval) | ✅ Full |
| `FROZEN` | ❌ None | ❌ None | ❌ None |
| `DETACHED` | ❌ None | ❌ None | ❌ None |

---

## 4. Domain Creation Rules

Domain creation follows a **biological cell-division model** where new domains are born from parent domains and inherit their foundational structure.

### 4.1 Prerequisites (Requester Must Have)

1. **Terra.Member seat** → ASSIGNED or OCCUPIED (existence verified)
2. **Lima.Member seat** → ASSIGNED or OCCUPIED (consent capability)
3. **Membership in parent domain** → Must be domain.member in the parent
4. **Parent root approval** → Parent's domain.root seat must approve the creation

### 4.2 Domain Creation Process

```
1. Requester submits domain creation request
   ├─ Validates terra.member exists
   ├─ Validates lima.member exists
   └─ Validates parent.member exists

2. Parent root evaluates request
   ├─ Checks requester authorization
   ├─ Evaluates risk and policy constraints
   └─ Issues approval or denial

3. Domain instantiation (if approved)
   ├─ Create domain record (UUID, name, parent_id)
   ├─ Inherit schema references from parent
   ├─ Inherit base policies from parent
   ├─ Instantiate domain.member seat (EMPTY)
   ├─ Instantiate domain.root seat (EMPTY)
   ├─ Assign requester to domain.member → OCCUPIED
   ├─ Assign requester to domain.root → OCCUPIED
   └─ Issue creation receipt

4. Receipt chain established
   └─ Links: requester → parent.root → new_domain
```

### 4.3 Inheritance Model

**Schema Inheritance:**
- Child domain **references** parent schema (no duplication)
- Child domain may **extend** schema with additional fields
- Child domain **cannot remove** parent schema fields

**Policy Inheritance:**
- Child domain inherits parent's base policies
- Child domain may **tighten** (add denials)
- Child domain **cannot loosen** (remove denials)

**Authority Inheritance:**
- Child domain.root has full authority within the child
- Child domain.root reports upward to parent.root
- Parent.root retains oversight authority

### 4.4 Cell-Division Principle

> Domain creation is analogous to biological cell division: the child domain inherits the DNA (schema + policy) of the parent, starts with a single root cell (requester), and cannot exist without the parent's blessing. **No domain can bypass parent approval.**

---

## 5. Authority Flow Model

Authority in DIS flows through seats based on their state and the direction of the action.

### 5.1 Authority Directions

**Upward Authority (Reporting):**
- Always allowed for ASSIGNED and OCCUPIED seats
- Used for: status reports, requests, audit trails
- No approval required

**Downward Authority (Governance):**
- Only allowed for OCCUPIED seats
- Requires: seat not frozen, parent approval (for cross-domain actions)
- Used for: approvals, seat assignments, policy enforcement

**Lateral Authority (Peer Actions):**
- Only allowed for OCCUPIED seats within the same domain
- Used for: member collaboration, data sharing
- May require root approval depending on action risk

### 5.2 Authority Flow Rules

```
┌──────────────────────────────────────────────────────┐
│ Authority Flow Decision Tree                         │
└──────────────────────────────────────────────────────┘

Is seat OCCUPIED?
├─ NO → Deny (except upward reporting if ASSIGNED)
└─ YES
    │
    Is seat FROZEN?
    ├─ YES → Deny
    └─ NO
        │
        Is action direction UPWARD?
        ├─ YES → Allow (reporting always allowed)
        └─ NO (DOWNWARD or LATERAL)
            │
            Does action require parent approval?
            ├─ YES
            │   └─ Is parent approval granted?
            │       ├─ YES → Allow
            │       └─ NO → Deny
            └─ NO → Allow (within-domain governance)
```

### 5.3 Authority Flow REGO

```rego
package dis.authority

# Authority flow evaluation
default allow = false

# Upward reporting always allowed if seat is not EMPTY
allow {
    input.direction == "upward"
    input.seat_state != "EMPTY"
}

# Downward authority requires OCCUPIED + not frozen
allow {
    input.direction == "downward"
    input.seat_state == "OCCUPIED"
    input.seat_status != "frozen"
    downward_approved
}

# Downward approval logic
downward_approved {
    # Within-domain actions allowed
    input.action_domain == input.seat_domain
}

downward_approved {
    # Cross-domain actions require parent approval
    input.action_domain != input.seat_domain
    input.parent_approved == true
}

# Deny if seat is frozen
deny {
    input.seat_status == "frozen"
}
```

---

## 6. REGO Structure (Three-Layer Stack)

DIS policies are evaluated in three stacked layers, each tightening the constraints of the layer below.

### 6.1 Layer 1: Terra.Policy (Existence)

**Purpose:** Ensures existential validity (identity exists, seats bound).

**Location:** `policy/terra.rego`

**Scope:** Universal (all actions)

**Example:**
```rego
package dis.terra

default allow = false

# Person must have terra.member seat
allow {
    input.identity_id
    data.identities[input.identity_id].terra_bound == true
}

deny {
    not input.identity_id
}
```

### 6.2 Layer 2: Numen.Policy (Meaning)

**Purpose:** Ensures semantic validity (schema compliance, canonical slots).

**Location:** `policy/numen.rego`

**Scope:** All data actions

**Example:**
```rego
package dis.numen

import future.keywords.if

default allow = false

# Schema must be valid
allow if {
    input.schema_ref
    data.schemas[input.schema_ref].valid == true
}

# Canonical greedy slots must exist
deny if {
    input.payload
    not input.payload.canonical_greedy
}

deny if {
    not input.schema_ref
}
```

### 6.3 Layer 3: Lima.Policy (Consent/Authority)

**Purpose:** Ensures consent and authority flow compliance.

**Location:** `policy/lima.rego`

**Scope:** All authority actions

**Example:**
```rego
package dis.lima

default allow = false

# Lima seat must be OCCUPIED for authority actions
allow {
    input.action_type == "authority"
    input.lima_seat_state == "OCCUPIED"
    input.direction == "upward"
}

allow {
    input.action_type == "authority"
    input.lima_seat_state == "OCCUPIED"
    input.direction == "downward"
    input.parent_approved == true
}

deny {
    input.lima_seat_state == "EMPTY"
}
```

### 6.4 Domain.Policy (Tightening)

**Purpose:** Domain-specific rules that **tighten** parent policies.

**Location:** `domains/{domain_id}/policy.rego`

**Tightening Rules:**
- Can add `deny` statements
- **Cannot** remove parent `deny` statements
- Can add conditional `allow` statements (narrower scope)
- **Cannot** add unconditional `allow` that bypasses parent denials

**Example (Tightening Parent Policy):**
```rego
package dis.domain.example

import data.dis.lima as parent

# Inherit parent policy (implicitly evaluated first)

# Tighten: Deny high-risk actions even if parent allows
deny {
    input.risk > 30
}

# Tighten: Require additional approval for member appointments
deny {
    input.action == "seat.appoint"
    input.seat_type == "member"
    not input.root_approved
}

# Allow can narrow scope but not bypass parent denials
allow {
    parent.allow  # Parent must allow first
    input.action == "data.read"
    input.risk < 10
}
```

### 6.5 Policy Evaluation Order

```
1. Terra.Policy (existence check)
   └─ DENY → Stop, return DENY
   └─ ALLOW → Continue

2. Numen.Policy (meaning check)
   └─ DENY → Stop, return DENY
   └─ ALLOW → Continue

3. Lima.Policy (authority check)
   └─ DENY → Stop, return DENY
   └─ ALLOW → Continue

4. Domain.Policy (domain-specific tightening)
   └─ DENY → Stop, return DENY
   └─ ALLOW → Continue

5. Per-Seat.Policy (seat-specific tightening)
   └─ DENY → Stop, return DENY
   └─ ALLOW → Return ALLOW

RESULT: Action is ALLOWED only if all layers allow
```

---

## 7. Schema Stability Principle

**Schema** and **Policy** are separate concerns in DIS and must remain decoupled.

### 7.1 Schema = Structure

- Defines **what fields exist** and their **types**
- Evaluated **locally** (no parent lookup required)
- Inherited by reference (not duplicated)
- Stored in: `schemas/{schema_id}.yaml`

**Example:**
```yaml
schema_id: "ci.call.v1"
version: "1.0"
fields:
  - name: "identity_id"
    type: "string"
    required: true
  - name: "timestamp"
    type: "timestamp"
    required: true
  - name: "canonical_greedy"
    type: "object"
    required: true  # Always required
```

### 7.2 Policy = Behavior

- Defines **what actions are allowed** and **under what conditions**
- May require **parent lookup** (authority flow, inheritance checks)
- Inherited with tightening (child adds denials)
- Stored in: `policy/*.rego` and `domains/{domain_id}/policy.rego`

**Example:**
```rego
package dis.policy

# Policy may reference parent domain
allow {
    input.domain_id
    parent_domain = data.domains[input.domain_id].parent_id
    data.domains[parent_domain].policy_allows == true
}
```

### 7.3 No Schema Duplication

Child domains **do not copy** parent schemas. Instead, they **reference** them.

**Parent schema:**
```yaml
# schemas/base.v1.yaml
schema_id: "base.v1"
fields:
  - name: "identity_id"
    type: "string"
```

**Child schema (extension):**
```yaml
# domains/child-domain/schema.yaml
schema_id: "child.extended.v1"
extends: "base.v1"  # Reference, not copy
fields:
  - name: "custom_field"
    type: "string"
```

### 7.4 Canonical Greedy Slots (Mandatory)

All data payloads **must** include canonical greedy slots for future extensibility.

**Required fields:**
- `canonical_greedy`: Object containing all canonical data
- Must be present in all schemas
- Cannot be removed by child domains

**Example payload:**
```json
{
  "identity_id": "alice@example.com",
  "timestamp": "2025-11-12T10:00:00Z",
  "canonical_greedy": {
    "action": "ci.call.v1",
    "risk": 25,
    "domain_id": "uuid-here"
  }
}
```

---

## 8. Final Summary

**GOV-1 establishes the foundational governance model for DIS, including:**

### ✅ Identity Triad Seats
- **Terra.Member** → Existence binding (no authority)
- **Numen.Member** → Meaning binding (no authority)
- **Lima.Member** → Consent binding (can request authority)

### ✅ Domain Member/Root Seats
- **Domain.Member** → Membership binding (no authority)
- **Domain.Root** → Governance seat (full domain authority)

### ✅ Domain Creation Lifecycle
- Requester must have terra + lima + parent membership
- Parent root must approve
- Child inherits schema (by reference) and policy (with tightening)
- Follows biological cell-division model

### ✅ Full Authority Flow Math
- **EMPTY** → No authority
- **ASSIGNED** → Upward reporting only
- **OCCUPIED** → Full authority (up + down with approval)
- **FROZEN** → Authority suspended

### ✅ Core REGO Binding Model
- Three-layer policy stack: Terra → Numen → Lima
- Domain policies tighten parent policies
- Per-seat policies tighten domain policies
- Evaluation order ensures cascading constraints

### ✅ Schema-Policy Separation
- Schema = structure (local evaluation)
- Policy = behavior (may require parent lookup)
- No schema duplication (inheritance by reference)
- Canonical greedy slots mandatory

---

**This document forms the mathematical and structural foundation of DIS domain governance going forward.**

**Status:** ✅ FOUNDATIONAL SPECIFICATION
**Next Steps:**
1. Implement GOV-1 rules in bootstrap process
2. Add GOV-1 validation to domain creation API
3. Update REGO policies to enforce GOV-1 constraints
4. Create test suite for GOV-1 compliance

**Version History:**
- v1.0 (2025-11-12): Initial specification

---

**Related Documents:**
- `PHASE_S_COMPLETE.md` - Seat management implementation
- `phase_s_summary.md` - Seat system technical details
- `policy/*.rego` - REGO policy implementations
- `schemas/*.yaml` - Schema definitions

**Governance Authority:**
This specification is binding for all DIS implementations and cannot be modified without multi-domain consensus.

