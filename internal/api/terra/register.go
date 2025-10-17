package terra

import (
	"database/sql"
	"net/http"
)

// Register wires Terra sync endpoints into the mux.
func Register(mux *http.ServeMux, _ *sql.DB) {
	mux.HandleFunc("/api/terra/map", handleTerraMap)
	mux.HandleFunc("/api/terra/version", handleTerraVersion)
	mux.HandleFunc("/api/terra/head", handleTerraHead)
}
