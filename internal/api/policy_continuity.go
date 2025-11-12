package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"

	"dis-core/internal/receipts"
)

// handlePolicyContinuity handles GET /api/policy/continuity/{domain} for Phase 10D policy continuity validation
func (s *Server) handlePolicyContinuity(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "domain")

	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "Domain ID required")
		case FormatText:
			http.Error(w, "error: Domain ID required", http.StatusBadRequest)
		default:
			http.Error(w, "Domain ID required", http.StatusBadRequest)
		}
		return
	}

	ctx := r.Context()
	result, err := receipts.VerifyPolicyContinuity(ctx, s.db, domainID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Policy continuity validation failed: %s", err.Error()))
		case FormatText:
			http.Error(w, fmt.Sprintf("error: Policy continuity validation failed: %s", err.Error()), http.StatusInternalServerError)
		default:
			http.Error(w, fmt.Sprintf("Policy continuity validation failed: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	// Return response based on format
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		writePolicyContinuityText(w, result)
	default:
		// Default to JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// writePolicyContinuityText writes the policy continuity result in human-readable text format
func writePolicyContinuityText(w http.ResponseWriter, result receipts.PolicyContinuityResult) {
	fmt.Fprintf(w, "Policy Continuity Report for Domain: %s\n", result.DomainRef)
	fmt.Fprintf(w, "Generated: %s\n\n", result.Timestamp)

	fmt.Fprintf(w, "Overall Status: ")
	if result.ContinuityOK {
		fmt.Fprintf(w, "✅ PASS - Policy continuity maintained\n")
	} else {
		fmt.Fprintf(w, "❌ FAIL - Policy continuity issues detected\n")
	}

	fmt.Fprintf(w, "\nReceipt Statistics:\n")
	fmt.Fprintf(w, "  Total Receipts: %d\n", result.TotalReceipts)
	fmt.Fprintf(w, "  Valid Policy Refs: %d\n", result.ValidRefs)
	fmt.Fprintf(w, "  Orphan Refs: %d\n", result.OrphanRefs)
	fmt.Fprintf(w, "  Invalid Refs: %d\n", result.InvalidRefs)

	if len(result.PolicyMappings) > 0 {
		fmt.Fprintf(w, "\nPolicy Reference Mappings:\n")
		for _, mapping := range result.PolicyMappings {
			status := "❌"
			if mapping.PolicyExists {
				status = "✅"
			}
			fmt.Fprintf(w, "  %s %s (%d receipts, last seen: %s)\n",
				status, mapping.PolicyRef, mapping.ReceiptCount, mapping.LastSeen)
		}
	}

	if len(result.OrphanReceipts) > 0 {
		fmt.Fprintf(w, "\nOrphan/Invalid Receipts:\n")
		for _, orphan := range result.OrphanReceipts {
			fmt.Fprintf(w, "  Receipt: %s | Event: %s | Issue: %s\n",
				orphan.ReceiptID, orphan.EventID, orphan.IssueReason)
		}
	}

	fmt.Fprintf(w, "\n")
}

// handlePolicyContinuityGlobal handles GET /api/policy/continuity for global policy continuity overview
func (s *Server) handlePolicyContinuityGlobal(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	ctx := r.Context()

	// Get global policy continuity statistics
	globalStats, err := s.getGlobalPolicyContinuityStats(ctx)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get global stats: %s", err.Error()))
		case FormatText:
			http.Error(w, fmt.Sprintf("error: Failed to get global stats: %s", err.Error()), http.StatusInternalServerError)
		default:
			http.Error(w, fmt.Sprintf("Failed to get global stats: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	// Return response based on format
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(globalStats)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		yamlData, err := yaml.Marshal(globalStats)
		if err != nil {
			http.Error(w, "error: Failed to marshal YAML", http.StatusInternalServerError)
			return
		}
		w.Write(yamlData)
	default:
		// Default to JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(globalStats)
	}
}

// GlobalPolicyContinuityStats represents global policy continuity statistics
type GlobalPolicyContinuityStats struct {
	Timestamp      string                        `json:"timestamp"`
	TotalReceipts  int                           `json:"total_receipts"`
	TotalDomains   int                           `json:"total_domains"`
	ValidRefs      int                           `json:"valid_refs"`
	OrphanRefs     int                           `json:"orphan_refs"`
	InvalidRefs    int                           `json:"invalid_refs"`
	ContinuityRate float64                       `json:"continuity_rate"`
	DomainStats    []receipts.DomainReceiptStats `json:"domain_stats"`
}

// getGlobalPolicyContinuityStats computes global policy continuity statistics
func (s *Server) getGlobalPolicyContinuityStats(ctx context.Context) (GlobalPolicyContinuityStats, error) {
	stats := GlobalPolicyContinuityStats{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		DomainStats: []receipts.DomainReceiptStats{},
	}

	// Get total receipts
	if s.db != nil {
		err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM receipts").Scan(&stats.TotalReceipts)
		if err != nil {
			return stats, fmt.Errorf("failed to count total receipts: %w", err)
		}

		// Count receipts with valid policy refs
		err = s.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM receipts
			WHERE policy_ref IS NOT NULL AND policy_ref != ''
		`).Scan(&stats.ValidRefs)
		if err != nil {
			return stats, fmt.Errorf("failed to count valid refs: %w", err)
		}

		// Count orphan receipts
		stats.OrphanRefs = stats.TotalReceipts - stats.ValidRefs

		// Calculate continuity rate
		if stats.TotalReceipts > 0 {
			stats.ContinuityRate = float64(stats.ValidRefs) / float64(stats.TotalReceipts)
		}

		// Get unique domains from receipts (simplified - extract from event_id patterns)
		rows, err := s.db.Query(ctx, `
			SELECT DISTINCT
				CASE
					WHEN event_id ILIKE '%domain.%' THEN
						substring(event_id FROM 'domain\.([^\.]+)')
					WHEN policy_ref ILIKE '%domain.%' THEN
						substring(policy_ref FROM 'domain\.([^\.]+)')
					ELSE 'unknown'
				END as domain_name,
				COUNT(*) as receipt_count
			FROM receipts
			WHERE event_id ILIKE '%domain.%' OR policy_ref ILIKE '%domain.%'
			GROUP BY 1
			HAVING COUNT(*) > 0
			ORDER BY 2 DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var domainName string
				var receiptCount int
				if err := rows.Scan(&domainName, &receiptCount); err != nil {
					continue
				}
				if domainName != "" && domainName != "unknown" {
					stats.TotalDomains++
					// Get detailed stats for this domain
					if domainStats, err := receipts.GetDomainReceiptStats(ctx, s.db, domainName); err == nil {
						stats.DomainStats = append(stats.DomainStats, domainStats)
					}
				}
			}
		}
	}

	return stats, nil
}

// handleRemediateOrphans handles POST /api/policy/continuity/remediate for Phase 10E orphan remediation
func (s *Server) handleRemediateOrphans(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := r.URL.Query().Get("domain")
	dryRunParam := r.URL.Query().Get("dry_run")

	dryRun := dryRunParam == "true" || dryRunParam == "1"

	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "Domain parameter required")
		case FormatText:
			http.Error(w, "error: Domain parameter required", http.StatusBadRequest)
		default:
			http.Error(w, "Domain parameter required", http.StatusBadRequest)
		}
		return
	}

	ctx := r.Context()
	result, err := receipts.RemediateOrphans(ctx, s.db, domainID, dryRun)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Remediation failed: %s", err.Error()))
		case FormatText:
			http.Error(w, fmt.Sprintf("error: Remediation failed: %s", err.Error()), http.StatusInternalServerError)
		default:
			http.Error(w, fmt.Sprintf("Remediation failed: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	// Return response based on format
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		writeRemediationText(w, result, dryRun)
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// handleContinuityDashboard handles GET /api/dashboard/continuity for Phase 10E dashboard visualization
func (s *Server) handleContinuityDashboard(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	ctx := r.Context()

	// Get global continuity overview
	globalStats, err := s.getGlobalPolicyContinuityStats(ctx)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get dashboard data: %s", err.Error()))
		case FormatText:
			http.Error(w, fmt.Sprintf("error: Failed to get dashboard data: %s", err.Error()), http.StatusInternalServerError)
		default:
			http.Error(w, fmt.Sprintf("Failed to get dashboard data: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	// Add threshold information and risk assessments
	thresholds := receipts.DefaultContinuityThresholds()

	// Phase 10F: Get lineage proof data
	lineageSummary, err := s.getLineageSummary(ctx)
	if err != nil {
		// Log error but continue - lineage data is optional
		lineageSummary = receipts.LineageSummary{}
	}

	// Phase 10G: Get federation summary data
	federationSummary, err := s.getFederationSummary(ctx)
	if err != nil {
		// Log error but continue - federation data is optional
		federationSummary = receipts.FederationSummary{}
	}

	dashboard := map[string]interface{}{
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"thresholds": thresholds,
		"overview":   globalStats,
		"alerts":     generateContinuityAlerts(globalStats, thresholds),
		"lineage":    lineageSummary,    // Phase 10F: Add lineage proof data
		"federation": federationSummary, // Phase 10G: Add federation data
		"metrics": map[string]interface{}{
			"total_receipts":               globalStats.TotalReceipts,
			"total_domains":                globalStats.TotalDomains,
			"orphan_rate":                  calculateOrphanRate(globalStats),
			"remediation_potential":        calculateRemediationPotential(globalStats),
			"total_fix_receipts":           lineageSummary.TotalFixReceipts,       // Phase 10F
			"pending_verifications":        lineageSummary.PendingVerifications,   // Phase 10F
			"verified_chain_rate":          lineageSummary.VerifiedChainRate,      // Phase 10F
			"federation_verification_rate": federationSummary.VerificationRate,    // Phase 10G
			"cross_domain_discrepancies":   federationSummary.DiscrepancyCount,    // Phase 10G
			"trusted_domain_count":         len(federationSummary.TrustedDomains), // Phase 10G
		},
	}

	// Return response based on format
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dashboard)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		writeDashboardText(w, dashboard)
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dashboard)
	}
}

// writeRemediationText formats remediation results as plain text
func writeRemediationText(w http.ResponseWriter, result receipts.RemediationResult, dryRun bool) {
	mode := "LIVE"
	if dryRun {
		mode = "DRY RUN"
	}

	fmt.Fprintf(w, "=== Policy Continuity Remediation (%s) ===\n", mode)
	fmt.Fprintf(w, "Timestamp: %s\n", result.Timestamp)
	fmt.Fprintf(w, "Total Orphans: %d\n", result.TotalOrphans)
	fmt.Fprintf(w, "Remediated: %d\n", result.Remediated)
	fmt.Fprintf(w, "Failed: %d\n", result.Failed)
	fmt.Fprintf(w, "Continuity: %.1f%% → %.1f%%\n", result.ContinuityBefore, result.ContinuityAfter)

	if len(result.Strategies) > 0 {
		fmt.Fprintf(w, "\nStrategies Used:\n")
		for _, strategy := range result.Strategies {
			fmt.Fprintf(w, "  • %s\n", strategy)
		}
	}

	if len(result.Details) > 0 {
		fmt.Fprintf(w, "\nRemediation Details:\n")
		for i, detail := range result.Details {
			if i >= 10 { // Limit text output
				fmt.Fprintf(w, "  ... and %d more\n", len(result.Details)-i)
				break
			}
			status := "✓"
			if !detail.Success {
				status = "✗"
			}
			fmt.Fprintf(w, "  %s %s: %s → %s (%s)\n",
				status, detail.ReceiptID[:8], detail.OldPolicyRef, detail.NewPolicyRef, detail.Strategy)
		}
	}
}

// writeDashboardText formats dashboard data as plain text
func writeDashboardText(w http.ResponseWriter, dashboard map[string]interface{}) {
	fmt.Fprintf(w, "=== Policy Continuity Dashboard ===\n")

	if overview, ok := dashboard["overview"].(GlobalPolicyContinuityStats); ok {
		fmt.Fprintf(w, "Total Receipts: %d\n", overview.TotalReceipts)
		fmt.Fprintf(w, "Total Domains: %d\n", overview.TotalDomains)

		if metrics, ok := dashboard["metrics"].(map[string]interface{}); ok {
			if rate, ok := metrics["orphan_rate"].(float64); ok {
				fmt.Fprintf(w, "Orphan Rate: %.1f%%\n", rate)
			}
		}
	}

	// Phase 10F: Add lineage proof information
	if lineage, ok := dashboard["lineage"].(receipts.LineageSummary); ok {
		fmt.Fprintf(w, "\nLineage Proofs:\n")
		fmt.Fprintf(w, "  Fix Receipts: %d\n", lineage.TotalFixReceipts)
		fmt.Fprintf(w, "  Pending Verifications: %d\n", lineage.PendingVerifications)
		fmt.Fprintf(w, "  Verified Chain Rate: %.1f%%\n", lineage.VerifiedChainRate)
		if lineage.LastFixTimestamp != "" {
			fmt.Fprintf(w, "  Last Fix: %s\n", lineage.LastFixTimestamp)
		}
	}

	// Phase 10G: Add federation information
	if federation, ok := dashboard["federation"].(receipts.FederationSummary); ok {
		fmt.Fprintf(w, "\nCross-Domain Federation:\n")
		fmt.Fprintf(w, "  Federation Proofs: %d\n", federation.TotalFederationProofs)
		fmt.Fprintf(w, "  Verified Proofs: %d\n", federation.VerifiedProofs)
		fmt.Fprintf(w, "  Pending Proofs: %d\n", federation.PendingProofs)
		fmt.Fprintf(w, "  Discrepancies: %d\n", federation.DiscrepancyCount)
		fmt.Fprintf(w, "  Verification Rate: %.1f%%\n", federation.VerificationRate)
		if len(federation.TrustedDomains) > 0 {
			fmt.Fprintf(w, "  Trusted Domains: %s\n", strings.Join(federation.TrustedDomains, ", "))
		}
		if federation.LastSyncTimestamp != "" {
			fmt.Fprintf(w, "  Last Sync: %s\n", federation.LastSyncTimestamp)
		}
	}

	if alerts, ok := dashboard["alerts"].([]ContinuityAlert); ok && len(alerts) > 0 {
		fmt.Fprintf(w, "\nAlerts:\n")
		for _, alert := range alerts {
			fmt.Fprintf(w, "  %s: %s\n", strings.ToUpper(alert.Level), alert.Message)
		}
	}

	if thresholds, ok := dashboard["thresholds"].(receipts.ContinuityThresholds); ok {
		fmt.Fprintf(w, "\nThresholds:\n")
		fmt.Fprintf(w, "  Critical: <%.1f%%\n", thresholds.Critical)
		fmt.Fprintf(w, "  Warning: <%.1f%%\n", thresholds.Warning)
		fmt.Fprintf(w, "  Healthy: >%.1f%%\n", thresholds.Healthy)
	}
}

// ContinuityAlert represents a continuity risk alert
type ContinuityAlert struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Domain  string `json:"domain,omitempty"`
}

// generateContinuityAlerts creates alerts based on continuity thresholds
func generateContinuityAlerts(overview GlobalPolicyContinuityStats, thresholds receipts.ContinuityThresholds) []ContinuityAlert {
	alerts := []ContinuityAlert{}

	// Check each domain for threshold violations
	for _, domain := range overview.DomainStats {
		rate := float64(domain.WithPolicyRef) / float64(domain.Total) * 100.0

		if rate < thresholds.Critical {
			alerts = append(alerts, ContinuityAlert{
				Level:   "critical",
				Message: fmt.Sprintf("Domain %s continuity at %.1f%% (critical threshold)", domain.DomainRef, rate),
				Domain:  domain.DomainRef,
			})
		} else if rate < thresholds.Warning {
			alerts = append(alerts, ContinuityAlert{
				Level:   "warning",
				Message: fmt.Sprintf("Domain %s continuity at %.1f%% (warning threshold)", domain.DomainRef, rate),
				Domain:  domain.DomainRef,
			})
		}
	}

	// Add global alert if needed
	if overview.TotalReceipts > 0 {
		globalRate := float64(overview.TotalReceipts-overview.OrphanRefs) / float64(overview.TotalReceipts) * 100.0
		if globalRate < thresholds.Critical {
			alerts = append(alerts, ContinuityAlert{
				Level:   "critical",
				Message: fmt.Sprintf("Global continuity at %.1f%% requires immediate attention", globalRate),
			})
		}
	}

	return alerts
}

// calculateOrphanRate calculates the global orphan rate
func calculateOrphanRate(overview GlobalPolicyContinuityStats) float64 {
	if overview.TotalReceipts == 0 {
		return 0.0
	}
	return float64(overview.OrphanRefs) / float64(overview.TotalReceipts) * 100.0
}

// calculateRemediationPotential estimates how many orphans could be remediated
func calculateRemediationPotential(overview GlobalPolicyContinuityStats) float64 {
	// Simplified estimation: assume 80% of orphans can be remediated with patterns
	return float64(overview.OrphanRefs) * 0.8
}

// getFederationSummary retrieves cross-domain federation metrics for Phase 10G
func (s *Server) getFederationSummary(ctx context.Context) (receipts.FederationSummary, error) {
	synchronizer := receipts.NewProofSynchronizer(s.db, "local.domain") // TODO: Get from config
	summary, err := synchronizer.GetFederationSummary(ctx)
	if err != nil {
		return receipts.FederationSummary{}, err
	}
	return *summary, nil
}
