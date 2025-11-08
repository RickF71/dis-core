package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

// ConnectPostgres opens a PostgreSQL connection using pgx.
func ConnectPostgres(dsn string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return conn, nil
}

// ConnectFromEnv connects using DATABASE_URL environment variable
func ConnectFromEnv() (*pgx.Conn, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DIS_DB_DSN")
	}
	if dsn == "" {
		dsn = "postgres://dis_user:card567@localhost:5432/dis_core?sslmode=disable"
	}
	return ConnectPostgres(dsn)
}
