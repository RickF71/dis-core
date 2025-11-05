package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// Config interface for database configuration
type ConfigProvider interface {
	DatabaseURL() string
}

var DefaultDB *sql.DB

// Connect establishes a database connection using config-based DSN
func Connect(cfg ConfigProvider) (*sql.DB, error) {
	dsn := cfg.DatabaseURL()
	return ConnectDSN(dsn)
}

func SetupDatabase() (*sql.DB, error) {
	dsn := os.Getenv("DIS_DB_DSN")
	if dsn == "" {
		dsn = "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"
		fmt.Println("⚠️ Using default Postgres DSN:", dsn)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	DefaultDB = db
	fmt.Println("✅ Connected to PostgreSQL:", dsn)
	return db, nil
}

func CloseDatabase() {
	if DefaultDB != nil {
		_ = DefaultDB.Close()
		DefaultDB = nil
	}
}

func ConnectDSN(dsn string) (*sql.DB, error) {
	if dsn == "" {
		// fallback to env or default
		dsn = os.Getenv("DIS_DB_DSN")
		if dsn == "" {
			dsn = "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"
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
