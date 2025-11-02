package schema

import "fmt"

// RegisterAllStructs attaches known Go exemplar structs to an existing Registry.
// It’s purely optional — useful only if you want runtime decoding of Go types.
func (r *Registry) RegisterAllStructs() {
	if r == nil {
		return
	}

	fmt.Println("🧩 Registering Go exemplar structs into schema.Registry...")

	// These entries won’t collide with your YAML-based loader, since they
	// are added with synthetic paths and hashes for in-memory usage.
	r.byKey[r.key("authoritant.definition.v1", "v0.9.5")] = Entry{
		ID:      "authoritant.definition.v1",
		Version: "v0.9.5",
		Path:    "(compiled: AuthoritantDefinition)",
	}
	r.byKey[r.key("authoritant.single.v1", "v0.9.5")] = Entry{
		ID:      "authoritant.single.v1",
		Version: "v0.9.5",
		Path:    "(compiled: AuthoritantSingle)",
	}
	r.byKey[r.key("authoritant.collective.v1", "v0.9.5")] = Entry{
		ID:      "authoritant.collective.v1",
		Version: "v0.9.5",
		Path:    "(compiled: AuthoritantCollective)",
	}
	r.byKey[r.key("authoritant.delegated.v1", "v0.9.5")] = Entry{
		ID:      "authoritant.delegated.v1",
		Version: "v0.9.5",
		Path:    "(compiled: AuthoritantDelegated)",
	}
	r.byKey[r.key("jikka.shared-core.triad.v1", "v0.9.5")] = Entry{
		ID:      "jikka.shared-core.triad.v1",
		Version: "v0.9.5",
		Path:    "(compiled: SharedCoreTriad)",
	}
	r.byKey[r.key("authoritant.link.v1", "v0.9.5")] = Entry{
		ID:      "authoritant.link.v1",
		Version: "v0.9.5",
		Path:    "(compiled: AuthoritantLink)",
	}
}
