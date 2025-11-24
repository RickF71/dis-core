package dis.policy.at1_corruption

import data.dis.policy.at1_corruption.allow

test_allow_when_not_corrupt {
    allow with input as {"context": {"attrs": {"corrupt": false}}}
}

test_deny_when_corrupt {
    not allow with input as {"context": {"attrs": {"corrupt": true}}}
}

test_allow_when_missing_attrs {
    allow with input as {"context": {}}  # no attrs → allowed
}
