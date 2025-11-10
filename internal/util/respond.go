// internal/util/respond.go
package util

import (
	"encoding/json"
	"net/http"
	"strings"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func JSONError(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]any{"error": message})
}

func DetectFormat(r *http.Request) string {
	url := r.URL.Path
	switch {
	case strings.HasSuffix(url, ".json"):
		return "json"
	case strings.HasSuffix(url, ".file"):
		return "file"
	case strings.HasSuffix(url, ".text"):
		return "text"
	default:
		return "json"
	}
}
