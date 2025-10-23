package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) registerImportListRoute() {
	s.mux.HandleFunc("/api/import/list", func(w http.ResponseWriter, r *http.Request) {
		recs, err := s.Ledger.ListImports(10)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(recs)
	})
}
