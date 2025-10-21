package terra

import (
	"dis-core/internal/util/iso"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// ...existing code...

func handleTerraMap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

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

	path := filepath.Join(baseTerraPath, filename)
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("open %s: %v", filename, err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var geo map[string]interface{}
	dec := json.NewDecoder(file)
	if err := dec.Decode(&geo); err != nil {
		http.Error(w, fmt.Sprintf("decode geojson: %v", err), http.StatusInternalServerError)
		return
	}

	features, ok := geo["features"].([]interface{})
	if ok {
		unkCount := 0
		for _, f := range features {
			feat, ok := f.(map[string]interface{})
			if !ok {
				continue
			}
			props, ok := feat["properties"].(map[string]interface{})
			if !ok {
				continue
			}
			iso2 := ""
			admin := ""
			if v, ok := props["ISO_A2"].(string); ok {
				iso2 = v
			}
			if v, ok := props["ADMIN"].(string); ok {
				admin = v
			}
			iso3 := iso.NormalizeISO3(iso2, admin)
			props["ISO3"] = iso3
			if iso3 == "UNK" {
				unkCount++
			}
		}
		if unkCount > 0 {
			fmt.Printf("[terra overlay] %d features with UNK ISO3 in %s\n", unkCount, filename)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(geo)
}
