# GOV-7: Corporeal Authority Policy
# Defines authorization rules for embodied persons and synthetic entities

package dis.seat

# Default deny all actions
default allow := false
default reason := "no explicit authorization"

# GOV-7: Corporeal domains can authorize actions
# These are domains directly under terra.numen.lima.corporeal
allow if {
    input.domain == "terra.numen.lima.corporeal"
    input.action == "authorize"
}

reason := "corporeal sovereignty" if {
    input.domain == "terra.numen.lima.corporeal"
    input.action == "authorize"
}

# GOV-7: Child corporeal domains (personal domains) inherit authorization
allow if {
    startswith(input.domain, "terra.numen.lima.corporeal.")
    not startswith(input.domain, "terra.numen.lima.corporeal.auto")
    input.action == "authorize"
}

reason := "inherited corporeal authority" if {
    startswith(input.domain, "terra.numen.lima.corporeal.")
    not startswith(input.domain, "terra.numen.lima.corporeal.auto")
    input.action == "authorize"
}

# GOV-7: Synthetic domains (.auto) CANNOT self-authorize
# They require explicit proxy consent from a corporeal domain
deny if {
    input.domain == "terra.numen.lima.corporeal.auto"
    input.action == "authorize"
}

deny if {
    startswith(input.domain, "terra.numen.lima.corporeal.auto.")
    input.action == "authorize"
}

reason := "synthetic domain requires corporeal proxy consent" if {
    input.domain == "terra.numen.lima.corporeal.auto"
    input.action == "authorize"
}

reason := "synthetic domain requires corporeal proxy consent" if {
    startswith(input.domain, "terra.numen.lima.corporeal.auto.")
    input.action == "authorize"
}

# GOV-7: Allow synthetic actions with explicit proxy
allow if {
    startswith(input.domain, "terra.numen.lima.corporeal.auto.")
    input.action == "execute"
    input.proxy_for
    startswith(input.proxy_for, "terra.numen.lima.corporeal.")
    not startswith(input.proxy_for, "terra.numen.lima.corporeal.auto")
}

reason := "synthetic action with corporeal proxy" if {
    startswith(input.domain, "terra.numen.lima.corporeal.auto.")
    input.action == "execute"
    input.proxy_for
}

# GOV-7: Prime Seat operations
allow if {
    input.action == "create_prime_seat"
    input.seat_type == "prime"
    input.single_occupancy == true
}

reason := "Prime Seat creation with single occupancy" if {
    input.action == "create_prime_seat"
    input.seat_type == "prime"
}

# GOV-7: Member seat operations require Prime Seat authorization
allow if {
    input.action == "appoint_member"
    input.appointed_by_prime == true
    input.has_prime_seat == true
}

reason := "member appointment by Prime Seat" if {
    input.action == "appoint_member"
    input.appointed_by_prime == true
}

# Metadata for policy introspection
metadata := {
    "policy_id": "gov7.corporeal.seat",
    "version": "1.0.0",
    "description": "GOV-7: Corporeal Authority Policy for embodied persons and synthetic entities",
    "governance": {
        "corporeal": {
            "can_authorize": true,
            "self_governing": true,
            "delegation": "explicit_only"
        },
        "synthetic": {
            "can_authorize": false,
            "requires_proxy": true,
            "inherits_from": "corporeal"
        }
    },
    "lineage": {
        "root": "terra.numen.lima.corporeal",
        "synthetic_branch": "terra.numen.lima.corporeal.auto",
        "human_sovereignty": true
    }
}
