package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectPostgres opens a PostgreSQL connection pool using pgxpool.
func ConnectPostgres(dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

// ConnectFromEnv connects using DATABASE_URL environment variable
func ConnectFromEnv() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DISCORE_DSN")
	}
	if dsn == "" {
		dsn = os.Getenv("DIS_DB_DSN")
	}
	if dsn == "" {
		dsn = "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"
	}
	return ConnectPostgres(dsn)
}
