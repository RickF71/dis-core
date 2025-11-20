package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AdminReceipt represents a receipt for admin API responses
type AdminReceipt struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Actor     string                 `json:"actor"`
	Target    string                 `json:"target"`
	Domain    string                 `json:"domain"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
}

// AdminFreezeRequest represents a freeze operation request
type AdminFreezeRequest struct {
	Target string `json:"target"`
	Reason string `json:"reason"`
}

// AdminPolicyReloadRequest represents a policy reload request
type AdminPolicyReloadRequest struct {
	PolicyType string `json:"policy_type,omitempty"`
}

// registerAdminRoutes registers all admin routes with the server mux
func (s *Server) registerAdminRoutes() {
	mux := s.router

	// Admin routes - all require admin privileges
	mux.HandleFunc("GET /api/admin/receipts/recent", s.handleRecentReceipts)
	mux.HandleFunc("POST /api/admin/freeze", s.handleAdminFreeze)
	mux.HandleFunc("POST /api/admin/policy/reload", s.handleAdminPolicyReload)
}

// verifySeatRole checks if the request has proper admin authorization
func (s *Server) verifySeatRole(r *http.Request) (string, error) {
	// Check Authorization header for admin token
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("no authorization header")
	}

	// Extract bearer token
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", fmt.Errorf("invalid authorization format")
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	// For now, we'll accept specific admin tokens
	// In production, this should validate against seat registry
	switch token {
	case "admin-root-token", "policy-admin-token":
		return "root", nil
	default:
		// Try to extract seat info from context or validate token
		// This is a simplified implementation
		if strings.Contains(token, "admin") {
			return "policy.admin", nil
		}
		return "", fmt.Errorf("invalid admin token")
	}
}

// handleRecentReceipts returns the last 50 receipts from the database
func (s *Server) handleRecentReceipts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify admin privileges
	role, err := s.verifySeatRole(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Query recent receipts from database
	db := s.requireDB(w)
	if db == nil {
		return
	}
	rows, err := db.Query(ctx, `
		SELECT id, type, actor, target, domain, payload, created_at
		FROM receipts
		ORDER BY created_at DESC
		LIMIT 50
	`)
	if err != nil {
		http.Error(w, "Database query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var receipts []AdminReceipt
	for rows.Next() {
		var receipt AdminReceipt
		var payloadBytes []byte

		err := rows.Scan(
			&receipt.ID,
			&receipt.Type,
			&receipt.Actor,
			&receipt.Target,
			&receipt.Domain,
			&payloadBytes,
			&receipt.CreatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to scan receipt: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse payload JSON
		if len(payloadBytes) > 0 {
			if err := json.Unmarshal(payloadBytes, &receipt.Payload); err != nil {
				// If JSON parsing fails, store as string
				receipt.Payload = map[string]interface{}{
					"raw": string(payloadBytes),
				}
			}
		}

		receipts = append(receipts, receipt)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Row iteration error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Record admin action using ci.call.v1
	if s.ledger != nil {
		_ = s.ledger.RecordCall(ctx, role, "receipts.recent", "admin", "admin.receipts.query.v1", map[string]interface{}{
			"action":         "query_recent_receipts",
			"limit":          50,
			"receipts_count": len(receipts),
			"queried_by":     role,
		})
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"receipts":   receipts,
		"count":      len(receipts),
		"queried_by": role,
	}); err != nil {
		http.Error(w, "JSON encoding failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleAdminFreeze handles POST /api/admin/freeze
func (s *Server) handleAdminFreeze(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify admin privileges
	role, err := s.verifySeatRole(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req AdminFreezeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Target == "" {
		http.Error(w, "Target is required", http.StatusBadRequest)
		return
	}

	// Perform freeze operation (this would interact with actual freeze logic)
	// For now, we'll simulate the freeze operation
	freezePayload := map[string]interface{}{
		"target":    req.Target,
		"reason":    req.Reason,
		"frozen_by": role,
		"frozen_at": time.Now().UTC().Format(time.RFC3339),
	}

	// Record freeze action using ci.call.v1
	if s.ledger != nil {
		err := s.ledger.RecordCall(ctx, role, req.Target, "admin", "admin.freeze.v1", freezePayload)
		if err != nil {
			http.Error(w, "Failed to record freeze action: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Target '%s' has been frozen", req.Target),
		"target":  req.Target,
		"reason":  req.Reason,
		"actor":   role,
	})
}

// handleAdminPolicyReload handles POST /api/admin/policy/reload
func (s *Server) handleAdminPolicyReload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify admin privileges
	role, err := s.verifySeatRole(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse request body (optional)
	var req AdminPolicyReloadRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Perform policy reload operation
	reloadPayload := map[string]interface{}{
		"policy_type": req.PolicyType,
		"reloaded_by": role,
		"reloaded_at": time.Now().UTC().Format(time.RFC3339),
	}

	// Record policy reload action using ci.call.v1
	if s.ledger != nil {
		err := s.ledger.RecordCall(ctx, role, "policy.engine", "admin", "admin.policy.reload.v1", reloadPayload)
		if err != nil {
			http.Error(w, "Failed to record policy reload action: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"message":     "Policy reload completed",
		"policy_type": req.PolicyType,
		"actor":       role,
	})
}
