package db

import (
	"database/sql"
	"dis-core/internal/config"
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

func EnsureTables(db *sql.DB, cfg *config.Config) error {
	// TODO: Bootstrap/migrate all tables, call schema bootstrapper
	return nil
}
