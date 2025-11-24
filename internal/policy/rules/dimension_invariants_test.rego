package dis.policy.dimension_invariants

import data.dis.policy.dimension_invariants.deny

test_deny_modification {
    deny[msg] with input as {"modifies_0d": true}
}

test_allow_no_modification {
    not deny[msg] with input as {"modifies_0d": false}
}
