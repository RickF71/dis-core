package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"dis-core/internal/receipts"
)

// handleProofSync handles cross-domain proof synchronization requests
func (s *Server) handleProofSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var syncRequest receipts.ProofSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&syncRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate sync request
	if syncRequest.SourceDomain == "" || syncRequest.TargetDomain == "" {
		http.Error(w, "Source and target domains are required", http.StatusBadRequest)
		return
	}

	// Ensure DB present
	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Initialize proof synchronizer
	synchronizer := receipts.NewProofSynchronizer(db, "local.domain") // TODO: Get from config

	// Perform synchronization
	response, err := synchronizer.SyncProofs(ctx, syncRequest.SourceDomain, syncRequest.TargetDomain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Sync failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Determine response format
	if strings.Contains(r.URL.Path, ".txt") || r.Header.Get("Accept") == "text/plain" {
		s.writeProofSyncText(w, response)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleVerifyProof handles verification of cross-domain proofs
func (s *Server) handleVerifyProof(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	proofID := chi.URLParam(r, "id")

	if proofID == "" {
		http.Error(w, "Proof ID is required", http.StatusBadRequest)
		return
	}

	// Ensure DB present
	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Initialize proof synchronizer
	synchronizer := receipts.NewProofSynchronizer(db, "local.domain") // TODO: Get from config

	// Verify the proof
	verified, err := synchronizer.VerifyForeignProof(ctx, proofID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Verification failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Create verification response
	verificationResponse := map[string]interface{}{
		"proof_id":  proofID,
		"verified":  verified,
		"timestamp": "now", // TODO: Use actual timestamp
	}

	if !verified {
		verificationResponse["reason"] = "Hash verification failed"
	}

	// Determine response format
	if strings.Contains(r.URL.Path, ".txt") || r.Header.Get("Accept") == "text/plain" {
		s.writeVerificationText(w, verificationResponse)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(verificationResponse)
	}
}

// handleFederationSummary provides dashboard metrics for cross-domain verification
func (s *Server) handleFederationSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Ensure DB present
	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Initialize proof synchronizer
	synchronizer := receipts.NewProofSynchronizer(db, "local.domain") // TODO: Get from config

	// Get federation summary
	summary, err := synchronizer.GetFederationSummary(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get federation summary: %v", err), http.StatusInternalServerError)
		return
	}

	// Determine response format
	if strings.Contains(r.URL.Path, ".txt") || r.Header.Get("Accept") == "text/plain" {
		s.writeFederationSummaryText(w, summary)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}

// handleCreateFederationTrust creates or updates trust relationships between domains
func (s *Server) handleCreateFederationTrust(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var trustRequest struct {
		DomainA    string `json:"domain_a"`
		DomainB    string `json:"domain_b"`
		TrustLevel string `json:"trust_level"`
		ExpiresAt  string `json:"expires_at,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&trustRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate trust request
	if trustRequest.DomainA == "" || trustRequest.DomainB == "" {
		http.Error(w, "Both domains are required", http.StatusBadRequest)
		return
	}

	// Validate trust level
	validTrustLevels := []string{receipts.TrustLevelHigh, receipts.TrustLevelMedium, receipts.TrustLevelLow, receipts.TrustLevelNone}
	isValidTrust := false
	for _, level := range validTrustLevels {
		if trustRequest.TrustLevel == level {
			isValidTrust = true
			break
		}
	}

	if !isValidTrust {
		http.Error(w, "Invalid trust level", http.StatusBadRequest)
		return
	}

	// Insert or update trust relationship
	query := `
		INSERT INTO federation_trust (id, domain_a, domain_b, trust_level)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (domain_a, domain_b) DO UPDATE SET
			trust_level = $4,
			established_at = now()
	`

	trustID := fmt.Sprintf("trust-%s-%s", trustRequest.DomainA, trustRequest.DomainB)
	db := s.requireDB(w)
	if db == nil {
		return
	}

	_, err := db.Exec(ctx, query, trustID, trustRequest.DomainA, trustRequest.DomainB, trustRequest.TrustLevel)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create trust relationship: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"status":      "success",
		"trust_id":    trustID,
		"domain_a":    trustRequest.DomainA,
		"domain_b":    trustRequest.DomainB,
		"trust_level": trustRequest.TrustLevel,
		"timestamp":   "now", // TODO: Use actual timestamp
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Text format response writers

func (s *Server) writeProofSyncText(w http.ResponseWriter, response *receipts.ProofSyncResponse) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "=== Proof Synchronization Results ===\n")
	fmt.Fprintf(w, "Status: %s\n", response.Status)
	fmt.Fprintf(w, "Synced Proofs: %d\n", response.SyncedProofs)
	fmt.Fprintf(w, "Failed Proofs: %d\n", response.FailedProofs)
	fmt.Fprintf(w, "Timestamp: %s\n", response.Timestamp.Format("2006-01-02 15:04:05 UTC"))

	if len(response.Discrepancies) > 0 {
		fmt.Fprintf(w, "\nDiscrepancies:\n")
		for _, discrepancy := range response.Discrepancies {
			fmt.Fprintf(w, "  - %s\n", discrepancy)
		}
	}

	if len(response.Details) > 0 {
		fmt.Fprintf(w, "\nDetails:\n")
		for key, value := range response.Details {
			fmt.Fprintf(w, "  %s: %s\n", key, value)
		}
	}
}

func (s *Server) writeVerificationText(w http.ResponseWriter, response map[string]interface{}) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "=== Proof Verification Results ===\n")
	fmt.Fprintf(w, "Proof ID: %s\n", response["proof_id"])
	fmt.Fprintf(w, "Verified: %v\n", response["verified"])
	fmt.Fprintf(w, "Timestamp: %s\n", response["timestamp"])

	if reason, exists := response["reason"]; exists {
		fmt.Fprintf(w, "Reason: %s\n", reason)
	}
}

func (s *Server) writeFederationSummaryText(w http.ResponseWriter, summary *receipts.FederationSummary) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "=== Federation Summary ===\n")
	fmt.Fprintf(w, "Total Federation Proofs: %d\n", summary.TotalFederationProofs)
	fmt.Fprintf(w, "Verified Proofs: %d\n", summary.VerifiedProofs)
	fmt.Fprintf(w, "Pending Proofs: %d\n", summary.PendingProofs)
	fmt.Fprintf(w, "Discrepancies: %d\n", summary.DiscrepancyCount)
	fmt.Fprintf(w, "Verification Rate: %.1f%%\n", summary.VerificationRate)

	if len(summary.TrustedDomains) > 0 {
		fmt.Fprintf(w, "\nTrusted Domains:\n")
		for _, domain := range summary.TrustedDomains {
			fmt.Fprintf(w, "  - %s\n", domain)
		}
	}

	if summary.LastSyncTimestamp != "" {
		fmt.Fprintf(w, "\nLast Sync: %s\n", summary.LastSyncTimestamp)
	}
}
