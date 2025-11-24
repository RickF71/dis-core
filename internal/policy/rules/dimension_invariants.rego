package dis.policy.dimension_invariants

default valid = true

# Prevent domains from overriding 0D (execution core)
valid {
    not input.modifies_0d
}

deny[msg] {
    input.modifies_0d
    msg = "0D invariants cannot be modified by any domain"
}
