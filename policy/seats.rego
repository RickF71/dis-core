package dis.seats

# Phase S2: Core Seats Policy
# Defines baseline allow rules for Prime Seat and member seats

default allow := false

# Allow if active Prime Seat exists and not frozen/detached
allow if {
    input.seat_context.active_prime
    not input.seat_context.frozen
    not input.seat_context.detached
}

# Allow if active member seat exists and not frozen/detached
allow if {
    input.seat_context.active_member
    not input.seat_context.frozen
    not input.seat_context.detached
}

# Deny frozen seats explicitly
deny["seat_frozen"] if {
    input.seat_context.frozen
}

# Deny detached seats explicitly
deny["seat_detached"] if {
    input.seat_context.detached
}
