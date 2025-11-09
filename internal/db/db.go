package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config interface for database configuration
type ConfigProvider interface {
	DatabaseURL() string
}

var DefaultConn *pgxpool.Pool

// Connect establishes a database connection using config-based DSN
func Connect(cfg ConfigProvider) (*pgxpool.Pool, error) {
	dsn := cfg.DatabaseURL()
	return ConnectDSN(dsn)
}

func SetupDatabase() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DIS_DB_DSN")
	if dsn == "" {
		dsn = "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"
		fmt.Println("⚠️ Using default Postgres DSN:", dsn)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	DefaultConn = pool
	fmt.Println("✅ Connected to PostgreSQL:", dsn)
	return pool, nil
}

func CloseDatabase() {
	if DefaultConn != nil {
		DefaultConn.Close()
		DefaultConn = nil
	}
}

func ConnectDSN(dsn string) (*pgxpool.Pool, error) {
	if dsn == "" {
		// fallback to env or default
		dsn = os.Getenv("DIS_DB_DSN")
		if dsn == "" {
			dsn = "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"
		}
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
