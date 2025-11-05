package canon

import "log"

// BootstrapThemes ensures Terra + USA domain themes exist in the database.
func BootstrapThemes(reg *Registry) {
	if reg == nil || reg.DB == nil {
		log.Println("⚠️ canon.BootstrapThemes: registry or DB is nil, skipping")
		return
	}

	seeds := SeedAll()

	for _, s := range seeds {
		err := reg.InsertDomainIfMissing(s.ID, s.Parent, s.CSS)
		if err != nil {
			log.Printf("❌ failed to insert %s: %v", s.ID, err)
		} else {
			log.Printf("🌱 ensured %s (parent=%s)", s.ID, s.Parent)
		}
	}
}
