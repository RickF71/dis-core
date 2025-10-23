package gates

default allow = {"allow": true, "reasons": []}

allow = {"allow": false, "reasons": ["deny:gates:untrusted_actor"]} {
  not startswith(input.event.actor, "by:domain.")
}
