package domain

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "dis-core/internal/schema"
)

// fileExists reports whether path exists and is a regular file.
func fileExists(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return !info.IsDir()
}

// LookupByCode searches for a domain file by ISO code.
// Search order:
//  1) /domains/countries/{iso_lower}/domain.{iso_lower}.yaml
//  2) /domains/terra/{iso_lower}.yaml
// Returns a parsed DomainDoc or a clear error with 400/404 prefixes.
func LookupByCode(code string, repoRoot string, reg *schema.Registry) (*DomainDoc, error) {
    if code == "" {
        return nil, fmt.Errorf("400 Bad Request: missing domain code")
    }

    iso := strings.ToLower(strings.TrimSpace(code))
    if len(iso) < 2 || len(iso) > 5 {
        return nil, fmt.Errorf("400 Bad Request: invalid domain code: %q", code)
    }

    // 1) countries layout
    p1 := filepath.Join(repoRoot, "domains", "countries", iso, fmt.Sprintf("domain.%s.yaml", iso))
    if fileExists(p1) {
        d, err := LoadAndValidate(p1, reg)
        if err != nil {
            return nil, fmt.Errorf("400 Bad Request: failed to load domain %s: %v", code, err)
        }
        return d, nil
    }

    // 2) terra fallback
    p2 := filepath.Join(repoRoot, "domains", "terra", fmt.Sprintf("%s.yaml", iso))
    if fileExists(p2) {
        d, err := LoadAndValidate(p2, reg)
        if err != nil {
            return nil, fmt.Errorf("400 Bad Request: failed to load domain %s: %v", code, err)
        }
        return d, nil
    }

    // not found — return clear 404
    return nil, fmt.Errorf("404 Not Found: domain for code %q not found", code)
}
