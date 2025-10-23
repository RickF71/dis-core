package api

import (
	"dis-core/internal/policy"
	"dis-core/internal/receipts"
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) RegisterEvalRoute(engine policy.PolicyEngine) {
	s.mux.HandleFunc("/api/eval", func(w http.ResponseWriter, r *http.Request) {
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
		decision, err := engine.EvaluateAction(input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		receipt := receipts.NewReceipt(
			input["by"].(string),
			input["action"].(string),
			"",          // TODO: frozenCoreHash
			"console-1", // TODO: consoleID
			"seat-1",    // TODO: issuerSeat
		)
		if err := receipt.Save("./receipts"); err != nil {
			log.Printf("receipt save error: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"decision": decision,
			"receipt":  receipt,
		})
	})
}
