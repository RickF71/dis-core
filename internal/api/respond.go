package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSON writes a structured response with a status code.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		log.Printf("❌ JSON encode error: %v", err)
	}
}

// Error writes a standardized error message.
func Error(w http.ResponseWriter, status int, err error) {
	JSON(w, status, map[string]any{
		"error":  err.Error(),
		"status": status,
	})
}
