# Phase 10G: Federation Verification Policy
package federation

default allow = false

# Allow proof synchronization between trusted domains
allow {
    input.proof.source_domain in data.trusted_domains
    input.proof.target_domain in data.trusted_domains
    input.proof.federation_hash == input.expected_hash
    input.trust_level != "none"
}

# Allow high trust level domains with less strict verification
allow {
    input.trust_level == "high"
    input.proof.source_domain in data.trusted_domains
    input.proof.target_domain in data.trusted_domains
}

# Reject if proof hash doesn't match expected
deny {
    input.proof.federation_hash != input.expected_hash
    input.trust_level != "high"
}

# Reject if domains are not in trusted list
deny {
    not input.proof.source_domain in data.trusted_domains
}

deny {
    not input.proof.target_domain in data.trusted_domains
}

# Define trust levels
trust_levels := ["none", "low", "medium", "high"]

# Helper to validate trust level
valid_trust_level(level) {
    level in trust_levels
}

# Calculate trust score based on verification history
trust_score(domain) := score {
    verified_count := count([p | p := data.verification_history[domain][_]; p.status == "verified"])
    total_count := count(data.verification_history[domain])
    score := (verified_count / total_count) * 100
}

# Recommend trust level based on verification history
recommended_trust_level(domain) := level {
    score := trust_score(domain)
    score >= 95
    level := "high"
} else := level {
    score := trust_score(domain)
    score >= 80
    level := "medium"
} else := level {
    score := trust_score(domain)
    score >= 60
    level := "low"
} else := "none" {
    true
}
