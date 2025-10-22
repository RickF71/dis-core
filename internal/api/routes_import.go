package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func (s *Server) registerImportRoutes() {
	s.mux.HandleFunc("/api/import", s.handleImport)
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// handle both JSON and multipart
	var filename, category, content string

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		file, hdr, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		buf, _ := io.ReadAll(file)
		content = string(buf)
		filename = hdr.Filename
		category = r.FormValue("category")
	} else {
		var payload struct {
			Filename string `json:"filename"`
			Category string `json:"category"`
			Content  string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		filename, category, content = payload.Filename, payload.Category, payload.Content
	}

	// fallback: infer category from filename
	if category == "" && strings.Contains(filename, ".") {
		switch {
		case strings.Contains(filename, "domain."):
			category = "domain"
		case strings.Contains(filename, "schema."):
			category = "schema"
		case strings.Contains(filename, "overlay."):
			category = "overlay"
		case strings.Contains(filename, "policy."):
			category = "policy"
		case strings.Contains(filename, "receipt."):
			category = "receipt"
		default:
			category = "unknown"
		}
	}

	node := make(map[string]any)
	if err := yaml.Unmarshal([]byte(content), &node); err != nil {
		http.Error(w, fmt.Sprintf("yaml parse error: %v", err), http.StatusBadRequest)
		return
	}

	var err error
	switch category {
	case "domain":
		err = s.DomainManager.ImportFromYAML(node)
	case "schema":
		err = s.SchemaManager.ImportFromYAML(node)
	case "overlay":
		err = s.OverlayManager.ImportFromYAML(node)
	case "policy":
		err = s.PolicyManager.ImportFromYAML(node)
	case "receipt":
		// err = s.Ledger.ImportReceiptFromYAML(node) // No-op for now
	default:
		http.Error(w, fmt.Sprintf("unknown category: %s", category), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("import failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a placeholder receipt ID for now
	receiptID := fmt.Sprintf("rcpt-%d", time.Now().Unix())
	json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"category": category,
		"imported": filename,
		"receipt":  receiptID,
	})
}
