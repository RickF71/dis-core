package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/receipts"
)

// CreateFixReceiptHandler handles POST /api/receipts/fix for Phase 10F lineage proof creation
func (s *Server) CreateFixReceiptHandler(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	var fix receipts.FixReceipt
	if err := json.NewDecoder(r.Body).Decode(&fix); err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %s", err.Error()))
		case FormatText:
			http.Error(w, fmt.Sprintf("error: Invalid JSON: %s", err.Error()), http.StatusBadRequest)
		default:
			http.Error(w, fmt.Sprintf("Invalid JSON: %s", err.Error()), http.StatusBadRequest)
		}
		return
	}

	// Generate fix receipt ID with rcpt-fix- prefix
	fix.ID = "rcpt-fix-" + strings.ReplaceAll(uuid.New().String(), "-", "")
	fix.Timestamp = time.Now()

	// Validate required fields
	if fix.AuthorizedBy == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing authorized_by field")
		case FormatText:
			http.Error(w, "error: missing authorized_by field", http.StatusBadRequest)
		default:
			http.Error(w, "missing authorized_by field", http.StatusBadRequest)
		}
		return
	}

	if fix.OriginalReceipt == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing original_receipt field")
		case FormatText:
			http.Error(w, "error: missing original_receipt field", http.StatusBadRequest)
		default:
			http.Error(w, "missing original_receipt field", http.StatusBadRequest)
		}
		return
	}

	// Set default verification status
	if fix.Verification == "" {
		fix.Verification = receipts.VerificationPending
	}

	// Insert into database
	ctx := r.Context()
	db := s.requireDB(w)
	if db == nil {
		return
	}

	_, err := db.Exec(ctx,
		`INSERT INTO fix_receipts (id, original_receipt, domain_ref, action_ref, policy_ref, fix_method, authorized_by, timestamp, verification)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		fix.ID, fix.OriginalReceipt, fix.DomainRef, fix.ActionRef, fix.PolicyRef, fix.FixMethod, fix.AuthorizedBy, fix.Timestamp, fix.Verification)

	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %s", err.Error()))
		case FormatText:
			http.Error(w, fmt.Sprintf("error: Database error: %s", err.Error()), http.StatusInternalServerError)
		default:
			http.Error(w, fmt.Sprintf("Database error: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	// Return response based on format
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fix)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "✅ Fix Receipt Created: %s\n", fix.ID)
		fmt.Fprintf(w, "Original Receipt: %s\n", fix.OriginalReceipt)
		fmt.Fprintf(w, "Domain: %s\n", fix.DomainRef)
		fmt.Fprintf(w, "Fix Method: %s\n", fix.FixMethod)
		fmt.Fprintf(w, "Authorized By: %s\n", fix.AuthorizedBy)
		fmt.Fprintf(w, "Verification: %s\n", fix.Verification)
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fix)
	}
}

// ListFixReceiptsHandler handles GET /api/receipts/fix for Phase 10F fix receipt listing
func (s *Server) ListFixReceiptsHandler(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	ctx := r.Context()

	// Query with optional filtering
	filterType := r.URL.Query().Get("type")
	domain := r.URL.Query().Get("domain")
	verification := r.URL.Query().Get("verification")

	query := `SELECT id, original_receipt, domain_ref, action_ref, policy_ref, fix_method, authorized_by, timestamp, verification
			  FROM fix_receipts`
	args := []interface{}{}
	conditions := []string{}

	if filterType == "fix" {
		// Already filtering for fix receipts by table
	}

	if domain != "" {
		conditions = append(conditions, fmt.Sprintf("domain_ref = $%d", len(args)+1))
		args = append(args, domain)
	}

	if verification != "" {
		conditions = append(conditions, fmt.Sprintf("verification = $%d", len(args)+1))
		args = append(args, verification)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY timestamp DESC"

	db := s.requireDB(w)
	if db == nil {
		return
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %s", err.Error()))
		case FormatText:
			http.Error(w, fmt.Sprintf("error: Database error: %s", err.Error()), http.StatusInternalServerError)
		default:
			http.Error(w, fmt.Sprintf("Database error: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}
	defer rows.Close()

	var list []receipts.FixReceipt
	for rows.Next() {
		var f receipts.FixReceipt
		if err := rows.Scan(&f.ID, &f.OriginalReceipt, &f.DomainRef, &f.ActionRef, &f.PolicyRef, &f.FixMethod, &f.AuthorizedBy, &f.Timestamp, &f.Verification); err != nil {
			switch format {
			case FormatJSON:
				JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Scan error: %s", err.Error()))
			case FormatText:
				http.Error(w, fmt.Sprintf("error: Scan error: %s", err.Error()), http.StatusInternalServerError)
			default:
				http.Error(w, fmt.Sprintf("Scan error: %s", err.Error()), http.StatusInternalServerError)
			}
			return
		}
		list = append(list, f)
	}

	// Return response based on format
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"fix_receipts": list,
			"count":        len(list),
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "=== Fix Receipts (%d found) ===\n\n", len(list))

		if len(list) == 0 {
			fmt.Fprintf(w, "No fix receipts found.\n")
		} else {
			for i, fix := range list {
				fmt.Fprintf(w, "%d. %s\n", i+1, fix.ID)
				fmt.Fprintf(w, "   Original: %s\n", fix.OriginalReceipt)
				fmt.Fprintf(w, "   Domain: %s\n", fix.DomainRef)
				fmt.Fprintf(w, "   Method: %s\n", fix.FixMethod)
				fmt.Fprintf(w, "   Status: %s\n", fix.Verification)
				fmt.Fprintf(w, "   Authorized: %s\n", fix.AuthorizedBy)
				fmt.Fprintf(w, "   Created: %s\n", fix.Timestamp.Format(time.RFC3339))
				fmt.Fprintf(w, "\n")
			}
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"fix_receipts": list,
			"count":        len(list),
		})
	}
}

// GetLineageProofHandler handles GET /api/receipts/lineage/{id} for Phase 10F lineage proof retrieval
func (s *Server) GetLineageProofHandler(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	originalReceiptID := strings.TrimPrefix(r.URL.Path, "/api/receipts/lineage/")

	if originalReceiptID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "Receipt ID required")
		case FormatText:
			http.Error(w, "error: Receipt ID required", http.StatusBadRequest)
		default:
			http.Error(w, "Receipt ID required", http.StatusBadRequest)
		}
		return
	}

	ctx := r.Context()
	db := s.requireDB(w)
	if db == nil {
		return
	}

	proof, err := s.getLineageProof(ctx, db, originalReceiptID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, fmt.Sprintf("Lineage proof not found: %s", err.Error()))
		case FormatText:
			http.Error(w, fmt.Sprintf("error: Lineage proof not found: %s", err.Error()), http.StatusNotFound)
		default:
			http.Error(w, fmt.Sprintf("Lineage proof not found: %s", err.Error()), http.StatusNotFound)
		}
		return
	}

	// Return response based on format
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(proof)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "=== Lineage Proof ===\n")
		fmt.Fprintf(w, "Original Receipt: %s\n", proof.OriginalReceiptID)
		fmt.Fprintf(w, "Fix Receipt: %s\n", proof.FixReceiptID)
		fmt.Fprintf(w, "Verified: %t\n", proof.Verified)
		fmt.Fprintf(w, "Created: %s\n", proof.CreatedAt.Format(time.RFC3339))
		if proof.VerifiedAt != nil {
			fmt.Fprintf(w, "Verified At: %s\n", proof.VerifiedAt.Format(time.RFC3339))
		}
		if len(proof.ProofChain) > 0 {
			fmt.Fprintf(w, "\nProof Chain:\n")
			for i, step := range proof.ProofChain {
				fmt.Fprintf(w, "  %d. %s\n", i+1, step)
			}
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(proof)
	}
}

// getLineageProof retrieves lineage proof for a receipt
func (s *Server) getLineageProof(ctx context.Context, db *pgxpool.Pool, originalReceiptID string) (receipts.LineageProof, error) {
	var proof receipts.LineageProof

	query := `
		SELECT
			fix_receipt_id,
			original_receipt_id,
			verified,
			fix_timestamp as created_at,
			CASE WHEN verified THEN fix_timestamp ELSE NULL END as verified_at
		FROM lineage_proofs
		WHERE original_receipt_id = $1
		ORDER BY fix_timestamp DESC
		LIMIT 1
	`

	var verifiedAt *time.Time
	err := db.QueryRow(ctx, query, originalReceiptID).Scan(
		&proof.FixReceiptID,
		&proof.OriginalReceiptID,
		&proof.Verified,
		&proof.CreatedAt,
		&verifiedAt,
	)

	if err != nil {
		return proof, fmt.Errorf("no lineage proof found for receipt %s", originalReceiptID)
	}

	proof.VerifiedAt = verifiedAt

	// Build proof chain
	proof.ProofChain = []string{
		fmt.Sprintf("Original receipt: %s", proof.OriginalReceiptID),
		fmt.Sprintf("Fix receipt: %s", proof.FixReceiptID),
	}

	if proof.Verified {
		proof.ProofChain = append(proof.ProofChain, "Verification: Complete")
	} else {
		proof.ProofChain = append(proof.ProofChain, "Verification: Pending")
	}

	return proof, nil
}

// getLineageSummary provides dashboard metrics for lineage proof system
func (s *Server) getLineageSummary(ctx context.Context, db *pgxpool.Pool) (receipts.LineageSummary, error) {
	summary := receipts.LineageSummary{
		RecentFixes: []receipts.FixReceipt{},
	}

	// Count total fix receipts
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM fix_receipts").Scan(&summary.TotalFixReceipts)
	if err != nil {
		return summary, err
	}

	// Count pending verifications
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM fix_receipts WHERE verification = 'pending'").Scan(&summary.PendingVerifications)
	if err != nil {
		return summary, err
	}

	// Count verified chains
	summary.VerifiedChains = summary.TotalFixReceipts - summary.PendingVerifications

	// Calculate verified chain rate
	if summary.TotalFixReceipts > 0 {
		summary.VerifiedChainRate = float64(summary.VerifiedChains) / float64(summary.TotalFixReceipts) * 100.0
	}

	// Get last fix timestamp
	var lastTimestamp *time.Time
	err = db.QueryRow(ctx, "SELECT timestamp FROM fix_receipts ORDER BY timestamp DESC LIMIT 1").Scan(&lastTimestamp)
	if err == nil && lastTimestamp != nil {
		summary.LastFixTimestamp = lastTimestamp.Format(time.RFC3339)
	}

	// Get recent fixes (last 5)
	rows, err := db.Query(ctx, `
		SELECT id, original_receipt, domain_ref, action_ref, policy_ref, fix_method, authorized_by, timestamp, verification
		FROM fix_receipts
		ORDER BY timestamp DESC
		LIMIT 5
	`)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	for rows.Next() {
		var f receipts.FixReceipt
		if err := rows.Scan(&f.ID, &f.OriginalReceipt, &f.DomainRef, &f.ActionRef, &f.PolicyRef, &f.FixMethod, &f.AuthorizedBy, &f.Timestamp, &f.Verification); err == nil {
			summary.RecentFixes = append(summary.RecentFixes, f)
		}
	}

	return summary, nil
}
