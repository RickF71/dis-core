package iso

import (
	"strings"
)

// iso2to3 is a static map of ISO2 to ISO3 codes (partial, add more as needed)
var iso2to3 = map[string]string{
	"FR": "FRA",
	"US": "USA",
	"GB": "GBR",
	"DE": "DEU",
	"CN": "CHN",
	// ... (add all needed codes)
}

// NormalizeISO3 returns a canonical ISO3 code for a given input code and admin name.
func NormalizeISO3(code string, admin string) string {
	c := strings.ToUpper(strings.TrimSpace(code))
	if c == "-99" {
		if strings.EqualFold(admin, "France") {
			return "FRA"
		}
		return "UNK"
	}
	if len(c) == 2 {
		if iso, ok := iso2to3[c]; ok {
			return iso
		}
		return "UNK"
	}
	if len(c) == 3 {
		return c
	}
	return "UNK"
}
