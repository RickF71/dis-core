package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"dis-core/internal/corporeal"
)

// CreateCorporealBootstrap handles POST /api/corporeal/bootstrap
func CreateCorporealBootstrap(b *corporeal.Bootstrapper) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ExternalUID string `json:"external_uid"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		if req.ExternalUID == "" {
			http.Error(w, "external_uid is required", http.StatusBadRequest)
			return
		}

		corpDom, actorDom, err := b.EnsureCorporealBootstrap(r.Context(), req.ExternalUID)
		if err != nil {
			log.Printf("❌ Corporeal bootstrap failed for %s: %v", req.ExternalUID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("✅ Phase 0-R.5 — Corporeal bootstrap complete for %s (corporeal=%s, actor=%s)",
			req.ExternalUID, corpDom, actorDom)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"corporeal_domain":     corpDom,
			"default_actor_domain": actorDom,
			"status":               "corporeal_bootstrap_complete",
		})
	})
}
