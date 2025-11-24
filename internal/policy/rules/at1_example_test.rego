package dis.policy.at1_example

import data.dis.policy.at1_example.allow

test_allow_when_not_corrupt {
    allow with input as {"context": {"attrs": {"corrupt": false}}}
}

test_deny_when_corrupt {
    not allow with input as {"context": {"attrs": {"corrupt": true}}}
}
