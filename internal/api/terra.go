package api

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const baseTerraPath = "data/terra/earth"

// RegisterTerraRoutes attaches Terra sync endpoints to the mux.
func RegisterTerraRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/terra/map", handleTerraMap)
	mux.HandleFunc("/api/terra/version", handleTerraVersion)
	mux.HandleFunc("/api/terra/head", handleTerraHead)
}

// --- Handlers ---

func handleTerraMap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	// Optional ?region= parameter, defaults to "world"
	region := r.URL.Query().Get("region")
	if region == "" {
		region = "world"
	}

	filename := map[string]string{
		"world":      "terra_world_clean.geojson",
		"usa":        "usa_clean.geojson",
		"usa_states": "usa_states_clean.geojson",
	}[region]

	if filename == "" {
		http.Error(w, "invalid region", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(baseTerraPath, filename)

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("open %s: %v", filename, err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if fi, err := file.Stat(); err == nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))
	}

	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, fmt.Sprintf("File stream error: %v", err), http.StatusInternalServerError)
	}
}

func handleTerraVersion(w http.ResponseWriter, r *http.Request) {
	hash, mod, err := terraMeta(filepath.Join(baseTerraPath, "terra_world_clean.geojson"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	resp := map[string]interface{}{
		"hash":      hash,
		"modified":  mod.UTC(),
		"sizeBytes": fileSize(filepath.Join(baseTerraPath, "terra_world_clean.geojson")),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleTerraHead(w http.ResponseWriter, r *http.Request) {
	hash, mod, err := terraMeta(filepath.Join(baseTerraPath, "terra_world_clean.geojson"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, hash))
	w.Header().Set("Last-Modified", mod.UTC().Format(http.TimeFormat))
	w.WriteHeader(http.StatusOK)
}

// --- Helpers ---

func terraMeta(path string) (hash string, modTime time.Time, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", time.Time{}, err
	}
	sum := sha1.Sum(data)
	info, _ := os.Stat(path)
	return hex.EncodeToString(sum[:]), info.ModTime(), nil
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
