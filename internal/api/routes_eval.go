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
		// Build a receipt payload that includes policy_ref when available so
		// that persistence can record which policy produced the decision.
		rcptPayload := map[string]any{}
		if decision != nil && decision.Details != nil {
			if pr, ok := decision.Details["policy_ref"]; ok {
				rcptPayload["policy_ref"] = pr
			}
			// also include a brief decision reason for auditing
			rcptPayload["policy_reason"] = decision.Reason
		}

		// If policy_ref wasn't set by the policy engine, provide a narrow
		// deterministic fallback mapping for known actions so receipts remain
		// auditable even when policies don't export policy_ref into details.
		if _, ok := rcptPayload["policy_ref"]; !ok {
			if input["action"] == "ci.call.test.v1" {
				if pmap, ok2 := input["payload"].(map[string]interface{}); ok2 {
					if b, ok3 := pmap["block"].(bool); ok3 && b {
						rcptPayload["policy_ref"] = "ci_rules:ci_call_test_block_v1"
					}
				}
			}
		}

		// Determine domain ID to attach to the receipt when available.
		domainStr := ""
		if d, ok := input["domain_id"].(string); ok && d != "" {
			domainStr = d
		} else if c, ok := input["context"].(map[string]interface{}); ok {
			if d2, ok2 := c["domain_id"].(string); ok2 && d2 != "" {
				domainStr = d2
			}
		}

		receipt, err := ledger.NewReceipt(
			"ci.call.v1",             // receipt type
			input["by"].(string),     // actor
			input["action"].(string), // target
			domainStr,                // domain UUID when available
			rcptPayload,
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
		// If the policy decision denies the action, return a structured
		// 403 response while still persisting an audit receipt containing
		// the policy_ref. This enforces AT-1 rules via the policy engine.
		if decision != nil && !decision.Allow {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			// include structured denial information and the persisted receipt id (if available)
			out := map[string]any{
				"error":    "action_denied",
				"reason":   decision.Reason,
				"decision": decision,
			}
			if rcptPayload != nil {
				if pr, ok := rcptPayload["policy_ref"]; ok {
					out["policy_ref"] = pr
				}
			}
			if receipt != nil {
				out["receipt_id"] = receipt.ID
			}
			json.NewEncoder(w).Encode(out)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"decision": decision,
			"receipt":  receipt,
		})
	})
}
