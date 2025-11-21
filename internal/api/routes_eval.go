package api

import (
	"dis-core/internal/ledger"
	"dis-core/internal/policy"
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) RegisterEvalRoute(engine policy.PolicyEngine) {
	s.router.HandleFunc("/api/eval", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		// TODO: Check domain freeze and BreakGlassToken here
		// If a DB is available and the input includes actor_id + domain_id in
		// its context, attempt to build a richer, transactional policy context
		// so evaluations observe consistent role/seat information.
		if db := s.DB(); db != nil {
			if c, ok := input["context"].(map[string]interface{}); ok {
				if actorID, ok2 := c["actor_id"].(string); ok2 && actorID != "" {
					if domainIDStr, ok3 := c["domain_id"].(string); ok3 && domainIDStr != "" {
						tx, txErr := db.Begin(ctx)
						if txErr == nil {
							if ctxMap, berr := policy.BuildContextTx(ctx, tx, actorID, domainIDStr); berr == nil {
								input["context"] = ctxMap
								if roles, ok := ctxMap["roles"]; ok {
									input["roles"] = roles
								}
								if domainID, ok := ctxMap["domain_id"]; ok {
									input["domain_id"] = domainID
								}
							}
							_ = tx.Rollback(ctx)
						}
					}
				}
			}
		}
		decision, err := engine.EvaluateAction(ctx, input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		receipt, err := ledger.NewReceipt(
			input["by"].(string),
			input["action"].(string),
			"",          // TODO: frozenCoreHash
			"console-1", // TODO: consoleID
			"seat-1",    // TODO: issuerSeat
		)
		if err != nil {
			log.Printf("receipt creation error: %v", err)
		} else {
			// Save receipt only if a DB pool is available. Saving is best-effort
			// and should not make the eval endpoint fail when running in test
			// mode without a DB configured.
			if s.DB() != nil {
				if saveErr := receipt.Save(ctx, s.DB()); saveErr != nil {
					log.Printf("receipt save error: %v", saveErr)
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"decision": decision,
			"receipt":  receipt,
		})
	})
}
