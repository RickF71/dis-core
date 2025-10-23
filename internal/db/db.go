package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DefaultDB *sql.DB

func SetupDatabase() (*sql.DB, error) {
	dsn := os.Getenv("DIS_DB_DSN")
	if dsn == "" {
		dsn = "postgres://dis_user:card567@localhost:5432/dis_core?sslmode=disable"
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
