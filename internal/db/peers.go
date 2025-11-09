package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Peer struct {
	ID         string
	Name       string
	URL        string
	TrustLevel string
	LastSeen   string
	Status     string
}

func LoadNetworkPeers(conn *pgxpool.Pool) ([]Peer, error) {
	ctx := context.Background()
	rows, err := conn.Query(ctx, `
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
			log.Printf("failed to scan peer row: %v", err)
			continue
		}
		peers = append(peers, p)
	}
	return peers, nil
}
