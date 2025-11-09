package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// copilot: Implement Authority Console endpoint.
// Create GET /api/authority/status returning AuthorityStatus struct with version, domains, policies, freeze_map, and identities.
// Use db queries to fill summaries, marshal to JSON, and write to ResponseWriter.

// AuthorityStatus represents the complete authority console status
type AuthorityStatus struct {
	Version     string                  `json:"version"`
	Domains     []AuthorityDomainInfo   `json:"domains"`
	Policies    []AuthorityPolicyInfo   `json:"policies"`
	FreezeMap   map[string]interface{}  `json:"freeze_map"`
	Identities  []AuthorityIdentityInfo `json:"identities"`
	LastUpdated time.Time               `json:"last_updated"`
	SystemInfo  AuthoritySystemInfo     `json:"system_info"`
}

// AuthorityDomainInfo provides summary information about a domain
type AuthorityDomainInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ParentID    *string   `json:"parent_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Active      bool      `json:"active"`
	RecordCount int       `json:"record_count"`
}

// AuthorityPolicyInfo provides summary information about policies
type AuthorityPolicyInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	IsActive  bool      `json:"is_active"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthorityIdentityInfo provides summary information about identities
type AuthorityIdentityInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	IsActive    bool      `json:"is_active"`
	LastSeen    time.Time `json:"last_seen"`
	DomainCount int       `json:"domain_count"`
}

// AuthoritySystemInfo provides system-level information
type AuthoritySystemInfo struct {
	DatabaseConnections int       `json:"database_connections"`
	UptimeSeconds       float64   `json:"uptime_seconds"`
	MemoryUsageMB       float64   `json:"memory_usage_mb"`
	SchemaVersion       string    `json:"schema_version"`
	StartTime           time.Time `json:"start_time"`
}

// HandleAuthorityStatus implements GET /api/authority/status
func HandleAuthorityStatus(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	status, err := buildAuthorityStatus(ctx, db)
	if err != nil {
		http.Error(w, "failed to build authority status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// buildAuthorityStatus collects all data for the authority status response
func buildAuthorityStatus(ctx context.Context, db *pgxpool.Pool) (*AuthorityStatus, error) {
	status := &AuthorityStatus{
		Version:     "0.9.7", // Current DIS-Core version
		LastUpdated: time.Now(),
	}

	// Load domain summaries
	domains, err := loadAuthorityDomainSummaries(ctx, db)
	if err != nil {
		return nil, err
	}
	status.Domains = domains

	// Load policy summaries
	policies, err := loadAuthorityPolicySummaries(ctx, db)
	if err != nil {
		return nil, err
	}
	status.Policies = policies

	// Load freeze map (placeholder for now)
	status.FreezeMap = make(map[string]interface{})

	// Load identity summaries
	identities, err := loadAuthorityIdentitySummaries(ctx, db)
	if err != nil {
		return nil, err
	}
	status.Identities = identities

	// Load system info
	sysInfo, err := loadAuthoritySystemInfo(ctx, db)
	if err != nil {
		return nil, err
	}
	status.SystemInfo = *sysInfo

	return status, nil
}

// loadAuthorityDomainSummaries queries domain information from the database
func loadAuthorityDomainSummaries(ctx context.Context, db *pgxpool.Pool) ([]AuthorityDomainInfo, error) {
	query := `
		SELECT id, name, parent_id, created_at, updated_at
		FROM domains
		ORDER BY created_at DESC`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []AuthorityDomainInfo
	for rows.Next() {
		var d AuthorityDomainInfo
		var parentID *string

		err := rows.Scan(&d.ID, &d.Name, &parentID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}

		d.ParentID = parentID
		d.Active = true // Default to active

		// Count records for this domain (if canon table exists)
		recordQuery := `SELECT COUNT(*) FROM canon WHERE content->>'domain_id' = $1`
		err = db.QueryRow(ctx, recordQuery, d.ID).Scan(&d.RecordCount)
		if err != nil {
			d.RecordCount = 0 // Default if canon table doesn't exist or query fails
		}

		domains = append(domains, d)
	}

	return domains, nil
}

// loadAuthorityPolicySummaries queries policy information from the database
func loadAuthorityPolicySummaries(ctx context.Context, db *pgxpool.Pool) ([]AuthorityPolicyInfo, error) {
	query := `
		SELECT id, name, version, is_active, updated_at
		FROM policies
		ORDER BY updated_at DESC`

	rows, err := db.Query(ctx, query)
	if err != nil {
		// If policies table doesn't exist, return empty slice
		return []AuthorityPolicyInfo{}, nil
	}
	defer rows.Close()

	var policies []AuthorityPolicyInfo
	for rows.Next() {
		var p AuthorityPolicyInfo
		err := rows.Scan(&p.ID, &p.Name, &p.Version, &p.IsActive, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// loadAuthorityIdentitySummaries queries identity information from the database
func loadAuthorityIdentitySummaries(ctx context.Context, db *pgxpool.Pool) ([]AuthorityIdentityInfo, error) {
	query := `
		SELECT id, name, type, is_active, last_seen
		FROM identities
		ORDER BY last_seen DESC`

	rows, err := db.Query(ctx, query)
	if err != nil {
		// If identities table doesn't exist, return empty slice
		return []AuthorityIdentityInfo{}, nil
	}
	defer rows.Close()

	var identities []AuthorityIdentityInfo
	for rows.Next() {
		var i AuthorityIdentityInfo
		err := rows.Scan(&i.ID, &i.Name, &i.Type, &i.IsActive, &i.LastSeen)
		if err != nil {
			return nil, err
		}

		// Count domains associated with this identity
		domainQuery := `SELECT COUNT(*) FROM domains WHERE content->>'owner_id' = $1`
		err = db.QueryRow(ctx, domainQuery, i.ID).Scan(&i.DomainCount)
		if err != nil {
			i.DomainCount = 0 // Default if query fails
		}

		identities = append(identities, i)
	}

	return identities, nil
}

// loadAuthoritySystemInfo collects system-level metrics and information
func loadAuthoritySystemInfo(ctx context.Context, db *pgxpool.Pool) (*AuthoritySystemInfo, error) {
	info := &AuthoritySystemInfo{
		StartTime:           time.Now().Add(-time.Hour), // Placeholder
		UptimeSeconds:       3600,                       // Placeholder
		MemoryUsageMB:       128.5,                      // Placeholder
		DatabaseConnections: int(db.Stat().TotalConns()),
		SchemaVersion:       "0.9.7",
	}

	return info, nil
}
