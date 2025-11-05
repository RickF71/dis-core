package canon

// DomainSeed represents a built-in domain and its CSS theme.
type DomainSeed struct {
	ID     string
	Parent string
	CSS    string
}

// SeedAll gathers all canonical domain themes.
func SeedAll() []DomainSeed {
	var seeds []DomainSeed
	seedTerraTheme(&seeds)
	seedUSADomainTheme(&seeds)
	return seeds
}
