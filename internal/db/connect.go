package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// ConnectPostgres opens a PostgreSQL connection pool.
func ConnectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return db, nil
}
