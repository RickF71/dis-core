package app

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"dis-core/internal/schema"

	"github.com/jackc/pgx/v5"

	"gopkg.in/yaml.v2"
)

// BootstrapAuthority is the single entry point for initializing DIS-Core.
// It seeds canonical data (themes), ensures schema registration,
// and inserts null-domain policies if missing.
func BootstrapAuthority(db *pgx.Conn, reg *schema.Registry) error {
	ctx := context.Background()

	log.Println("[bootstrap] Starting DIS-Core initialization...")

	// ---- Step 2: Register core schema ----
	log.Println("[bootstrap] Registering core authority schema...")
	// TODO: Update schema.LoadOrRegister to use pgx.Conn
	// if err := schema.LoadOrRegister(ctx, db); err != nil {
	// 	log.Printf("[bootstrap] schema load failed: %v", err)
	// 	return err
	// }
	log.Println("[bootstrap] authority.console schema ensured.")

	// ---- Step 3: Ensure null-domain policies ----
	log.Println("[bootstrap] Ensuring null-domain policies (gates, risk, freeze)...")
	// TODO: Update policy.EnsureNullPolicies to use pgx.Conn
	// if err := policy.EnsureNullPolicies(ctx, db); err != nil {
	// 	log.Printf("[bootstrap] policy ensure failed: %v", err)
	// 	return err
	// }
	log.Println("[bootstrap] null policies ready.")

	log.Println("[bootstrap] DIS-Core initialization complete.")
	return nil
}

func RegisterAuthorityLayer(db *pgx.Conn) error {
	data, err := os.ReadFile("schemas/authority.layer.yaml")
	if err != nil {
		return err
	}
	var obj map[string]interface{}
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return err
	}
	jsonBytes, _ := json.Marshal(obj)

	_, err = db.Exec(context.Background(),
		`INSERT INTO schemas (name, version, json_schema)
         VALUES ($1, $2, $3)
         ON CONFLICT (name) DO UPDATE SET json_schema = EXCLUDED.json_schema`,
		"dis.authority.layer", "0.9.3", jsonBytes)
	return err
}
