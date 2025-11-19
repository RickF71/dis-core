# GOV-11H: Domain Branching Policy Considerations

## Overview
GOV-11H canonizes domain branching as the ONLY mechanism for domain realignment.
Parent relationships (`parent_id`) are now IMMUTABLE after domain creation.

## Future REGO Policy Requirements

### 1. Branch Lineage Context
REGO policies will need access to branch lineage information:

```rego
# Example future policy structure
package domain.authorization

# Branch lineage context will be injected
import data.domain.branch_lineage

allow_action[msg] {
    # Policy can evaluate based on:
    # - input.domain.branch_of
    # - input.domain.branch_depth
    # - branch_lineage[domain_id]  # Full ancestry chain

    # Example: Restrict certain actions in deeply branched domains
    input.domain.branch_depth < 3
    msg := "Action allowed: branch depth within limits"
}
```

### 2. DSCI (Domain-Signed Contract Inheritance)
Policies need to evaluate contract inheritance across branches:

```rego
# Example: Seat instantiation policy
can_instantiate_seat {
    # User has contract with original domain
    contract := data.contracts[input.user_id][input.original_domain_id]
    contract.active == true

    # Contract allows branch inheritance
    contract.branch_inheritance != false

    # Target domain is in branch lineage
    is_branch_of(input.target_domain_id, input.original_domain_id)
}

is_branch_of(branch, original) {
    branch_lineage[branch][_] == original
}
```

### 3. Branch Permission Inheritance
Policies should handle permission inheritance from original domain:

```rego
# Example: Permission check with branch inheritance
has_permission[perm] {
    # Direct permission in branch domain
    domain_permissions[input.domain_id][perm]
}

has_permission[perm] {
    # Inherited permission from original domain (via DSCI)
    domain := data.domains[input.domain_id]
    domain.branch_of != null

    original := data.domains[domain.branch_of]
    original_permissions[domain.branch_of][perm]

    # Check if permission is inheritable
    perm.inheritable == true
}
```

### 4. Branch Creation Authorization
Policies to control who can create branches:

```rego
# Example: Branch authorization
can_create_branch {
    # User must have admin seat in original domain
    has_seat(input.user_id, input.original_domain_id, "admin")

    # Original domain policy allows branching
    domain_policy := data.domains[input.original_domain_id].policy
    domain_policy.allow_branching != false

    # Branch depth limit check
    original_depth := data.domains[input.original_domain_id].branch_depth
    original_depth < max_branch_depth
}
```

### 5. Cross-Branch Data Access
Policies for accessing data across branch boundaries:

```rego
# Example: Data visibility policy
can_read_data {
    # Data in same domain
    data_domain_id == input.domain_id
}

can_read_data {
    # Data in original domain (branch can read original)
    branch := data.domains[input.domain_id]
    branch.branch_of == data_domain_id

    # Original domain allows branch read access
    original_policy[data_domain_id].branch_read_access == true
}
```

## Policy Data Requirements

### Domain Context Extension
```json
{
  "domain": {
    "id": "uuid",
    "name": "string",
    "parent_id": "uuid",
    "branch_of": "uuid | null",
    "branch_depth": "integer",
    "created_at": "timestamp"
  },
  "branch_lineage": [
    "uuid1",  // Current domain
    "uuid2",  // Immediate branch_of
    "uuid3"   // Original domain
  ]
}
```

### Contract Context Extension
```json
{
  "contract": {
    "id": "uuid",
    "user_id": "uuid",
    "domain_id": "uuid",
    "branch_inheritance": true,
    "active": true,
    "created_at": "timestamp"
  }
}
```

## Implementation Notes

1. **Policy Injection Points**
   - Branch creation requests
   - Seat instantiation requests
   - Cross-domain data access
   - Permission evaluation

2. **Data Loading**
   - Branch lineage must be precomputed/cached
   - Contract lookups should be indexed by user_id + domain_id
   - Branch depth checks should be efficient (indexed column)

3. **Performance Considerations**
   - Recursive lineage queries limited to depth 10
   - Branch info cached in domain metadata
   - Contract inheritance flags indexed

4. **Security Invariants**
   - parent_id MUST be immutable after creation
   - Branch relationships enforced by FK constraints
   - Receipt generated for every branch operation
   - DSCI eligibility checked before seat creation

## Migration Path

1. **Phase 1 (GOV-11H - Current)**
   - Database schema with branch fields
   - Basic branch creation API
   - DSCI eligibility helper (stub)

2. **Phase 2 (Future)**
   - REGO policy integration
   - Contract table with branch_inheritance
   - Full DSCI seat instantiation

3. **Phase 3 (Future)**
   - Branch lineage caching
   - Advanced permission inheritance
   - UI for branch management

## Testing Scenarios

1. Create branch with valid lineage
2. Attempt to change parent_id (should fail)
3. Check DSCI eligibility across branches
4. Verify receipt generation
5. Test branch depth limits
6. Cross-branch permission evaluation
