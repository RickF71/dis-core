package reconcile

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

type PendingFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func HandleListPending(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("_bootstrap")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var pending []PendingFile
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if filepath.Ext(f.Name()) == ".yaml" {
			pending = append(pending, PendingFile{
				Name: f.Name(),
				Path: filepath.Join("_bootstrap", f.Name()),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pending)
}
