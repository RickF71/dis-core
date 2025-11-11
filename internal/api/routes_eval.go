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
			if saveErr := receipt.Save(ctx, s.DB()); saveErr != nil {
				log.Printf("receipt save error: %v", saveErr)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"decision": decision,
			"receipt":  receipt,
		})
	})
}
