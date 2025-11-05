package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"dis-core/internal/bootstrap" // Import your ImportAll

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	// ⚙️ Configure your Postgres connection here
	dsn := "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	fmt.Println("🧱 Ensuring bootstrap_files table exists...")
	createTable := `
	CREATE TABLE IF NOT EXISTS bootstrap_files (
		id SERIAL PRIMARY KEY,
		rel_path TEXT NOT NULL,
		filename TEXT NOT NULL,
		content BYTEA NOT NULL,
		imported_at TIMESTAMPTZ DEFAULT NOW(),
		exported_to TEXT DEFAULT NULL,
		exported_at TIMESTAMPTZ DEFAULT NULL
	);`
	if _, err := db.Exec(createTable); err != nil {
		log.Fatalf("Table creation failed: %v", err)
	}

	fmt.Println("🚀 Starting import from disyaml/")
	root := filepath.Join(".", "disyaml")

	if err := bootstrap.ImportAll(db, root); err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	fmt.Println("✅ Import complete.")
}
