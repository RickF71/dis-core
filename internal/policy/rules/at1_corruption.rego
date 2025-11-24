package dis.policy.at1_corruption

default allow = true

# If input.context.attrs.corrupt is explicitly true, deny.
allow {
    not input.context.attrs.corrupt
}

# Optional: defensive default if attrs missing (treat as not corrupt)
allow {
    not input.context.attrs
}
