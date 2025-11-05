package canon

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// Registry holds a database handle for canonical operations.
type Registry struct {
	DB *sql.DB
}

// InsertDomainIfMissing ensures a domain row exists in the database,
// inserting it with name/resolved_name/css if not found.
func (r *Registry) InsertDomainIfMissing(id, parent, css string) error {
	if r.DB == nil {
		return fmt.Errorf("no database connection in canon registry")
	}

	// 1. Check if the record already exists
	var exists bool
	err := r.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM domains WHERE id=$1)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check domain existence: %w", err)
	}
	if exists {
		return nil // nothing to do
	}

	// 2. Compute derived names
	short := shortName(id)
	resolved := resolvedName(id)

	// 3. Build INSERT statement with fallback:
	//    if columns name/resolved_name exist → use them,
	//    otherwise only insert id, parent_id, css.
	insertStmt := `
		INSERT INTO domains (id, parent_id, name, resolved_name, css)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = r.DB.Exec(insertStmt, id, parent, short, resolved, css)
	if err != nil {
		// Try again with a minimal insert (for older tables)
		if strings.Contains(err.Error(), "column") {
			log.Printf("⚠️ Falling back to minimal insert for %s: %v", id, err)
			_, err = r.DB.Exec(
				`INSERT INTO domains (id, parent_id, css) VALUES ($1, $2, $3)`,
				id, parent, css,
			)
		}
		if err != nil {
			return fmt.Errorf("insert domain %s: %w", id, err)
		}
	}

	log.Printf("🌱 inserted domain %s (parent=%s)", id, parent)
	return nil
}

// shortName gives a canonical short name for a domain id.
func shortName(id string) string {
	switch id {
	case "domain.null":
		return "Null"
	case "domain.terra":
		return "Terra"
	case "domain.usa":
		return "USA"
	default:
		// fallback: take the last token
		parts := strings.Split(id, ".")
		return strings.Title(parts[len(parts)-1])
	}
}

// resolvedName gives a human-readable resolved name for a domain id.
func resolvedName(id string) string {
	switch id {
	case "domain.null":
		return "Null Domain (Ground State)"
	case "domain.terra":
		return "Terra Domain"
	case "domain.usa":
		return "United States Domain"
	default:
		return fmt.Sprintf("%s Domain", shortName(id))
	}
}
