package db

import (
	"database/sql"
	"dis-core/internal/config"
	"fmt"
	"os"

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

func Connect(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.DatabaseDSN
	if dsn == "" {
		dsn = os.Getenv("DIS_DB_DSN")
		if dsn == "" {
			dsn = "postgres://dis_user:card567@localhost:5432/dis_core?sslmode=disable"
		}
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

func EnsureTables(db *sql.DB, cfg *config.Config) error {
	// TODO: Bootstrap/migrate all tables, call schema bootstrapper
	return nil
}
