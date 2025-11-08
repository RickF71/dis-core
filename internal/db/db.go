package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

// Config interface for database configuration
type ConfigProvider interface {
	DatabaseURL() string
}

var DefaultConn *pgx.Conn

// Connect establishes a database connection using config-based DSN
func Connect(cfg ConfigProvider) (*pgx.Conn, error) {
	dsn := cfg.DatabaseURL()
	return ConnectDSN(dsn)
}

func SetupDatabase() (*pgx.Conn, error) {
	dsn := os.Getenv("DIS_DB_DSN")
	if dsn == "" {
		dsn = "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"
		fmt.Println("⚠️ Using default Postgres DSN:", dsn)
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		conn.Close(context.Background())
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	DefaultConn = conn
	fmt.Println("✅ Connected to PostgreSQL:", dsn)
	return conn, nil
}

func CloseDatabase() {
	if DefaultConn != nil {
		_ = DefaultConn.Close(context.Background())
		DefaultConn = nil
	}
}

func ConnectDSN(dsn string) (*pgx.Conn, error) {
	if dsn == "" {
		// fallback to env or default
		dsn = os.Getenv("DIS_DB_DSN")
		if dsn == "" {
			dsn = "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"
		}
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		conn.Close(context.Background())
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return conn, nil
}
