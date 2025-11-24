package dis.root_pseat

default allow = false

allow {
    input.action == "root_pseat_claim"
    input.domain_id == "null"
    input.actor_id == input.bootstrap.actor_id
}
