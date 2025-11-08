package db

import (
	"database/sql"
	"log"
)

// Peer represents one network peer (trusted node, council seat, etc.)
type Peer struct {
	ID         string
	Name       string
	URL        string
	TrustLevel string
	LastSeen   string
	Status     string
}

// LoadNetworkPeers reads peers directly from the Postgres table.
func LoadNetworkPeers(conn *sql.DB) ([]Peer, error) {
	rows, err := conn.Query(`
		SELECT id, COALESCE(name, id) AS name, url, trust_level,
		       COALESCE(to_char(last_seen, 'YYYY-MM-DD"T"HH24:MI:SSOF'), '') AS last_seen,
		       COALESCE(status, 'unknown')
		FROM peers
		WHERE trust_level = 'trusted'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []Peer
	for rows.Next() {
		var p Peer
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.TrustLevel, &p.LastSeen, &p.Status); err != nil {
			log.Printf("⚠️ failed to scan peer row: %v", err)
			continue
		}
		peers = append(peers, p)
	}
	return peers, nil
}
