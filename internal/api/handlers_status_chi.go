package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"dis-core/internal/api/middleware"
	"dis-core/internal/receipts"
)

// handlePingChi handles GET /api/ping with format support
func (s *Server) handlePingChi(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"status":  "ok",
		"message": "DIS-Core alive",
		"time":    time.Now().Format(time.RFC3339),
	}

	Respond(w, r, resp)
}

// handleHealthChi handles GET /api/health with format support
func (s *Server) handleHealthChi(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	Respond(w, r, resp)
}

// handleStatusChi handles GET /api/status with format support
func (s *Server) handleStatusChi(w http.ResponseWriter, r *http.Request) {
	// Phase 6: Use middleware context helpers
	db := middleware.FromDB(r.Context())
	prov, _ := middleware.FromProvenance(r.Context())
	engine := middleware.FromPolicyEngine(r.Context())

	// Get status data using existing logic
	statusData := s.getSystemStatus()

	// Add middleware-provided data to status
	statusData["request_id"] = prov.RequestID
	statusData["timestamp"] = prov.Timestamp
	statusData["has_database"] = db != nil
	statusData["has_policy_engine"] = engine != nil

	Respond(w, r, statusData)
}

// handleAuthorityStatusChi handles GET /api/authority/status with format support
func (s *Server) handleAuthorityStatusChi(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// Get authority status using existing logic
	authStatus := s.getAuthorityStatus()

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, authStatus)
	case FormatText:
		// Convert authority status to text format
		textStatus := s.formatAuthorityStatusAsText(authStatus)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(textStatus))
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// Helper methods for status data

// getSystemStatus retrieves system status information
func (s *Server) getSystemStatus() map[string]any {
	return map[string]any{
		"core":        "DIS-Core",
		"status":      "ok",
		"version":     "v0.9.7",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"environment": "development",
		"domains":     0,
		"receipts":    0,
		"schemas":     0,
		"db_latency":  -1,
	}
}

// getAuthorityStatus retrieves authority status information
func (s *Server) getAuthorityStatus() map[string]any {
	status := map[string]any{
		"authority": "active",
		"policies":  []string{"gates.rego", "risk.rego"},
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Phase 10E: Add policy continuity status and risk assessment
	if s.db != nil {
		ctx := context.Background()
		if globalStats, err := s.getGlobalPolicyContinuityStats(ctx); err == nil {
			thresholds := receipts.DefaultContinuityThresholds()
			riskLevel := receipts.GetContinuityRiskLevel(globalStats.ContinuityRate)

			status["continuity"] = map[string]interface{}{
				"rate":              globalStats.ContinuityRate,
				"risk_level":        riskLevel,
				"total_receipts":    globalStats.TotalReceipts,
				"orphan_receipts":   globalStats.OrphanRefs,
				"remediation_ready": globalStats.OrphanRefs > 0,
				"thresholds":        thresholds,
			}

			// Adjust overall status based on continuity risk
			if riskLevel == "critical" {
				status["status"] = "degraded"
				status["alert"] = "Critical policy continuity issues detected"
			} else if riskLevel == "warning" {
				status["alert"] = "Policy continuity warnings detected"
			}
		}
	}

	return status
}

// formatStatusAsText converts status data to human-readable text
func (s *Server) formatStatusAsText(status map[string]any) string {
	return fmt.Sprintf(`DIS-Core Status Report
======================
Core: %v
Status: %v
Version: %v
Environment: %v
Timestamp: %v

Database:
- Domains: %v
- Receipts: %v
- Schemas: %v
- Latency: %v ms
`,
		status["core"],
		status["status"],
		status["version"],
		status["environment"],
		status["timestamp"],
		status["domains"],
		status["receipts"],
		status["schemas"],
		status["db_latency"])
}

// formatAuthorityStatusAsText converts authority status to text
func (s *Server) formatAuthorityStatusAsText(authStatus map[string]any) string {
	policies := authStatus["policies"].([]string)
	policyList := ""
	for _, policy := range policies {
		policyList += fmt.Sprintf("- %s\n", policy)
	}

	result := fmt.Sprintf(`Authority Status Report
=======================
Authority: %v
Status: %v
Timestamp: %v

Active Policies:
%s`,
		authStatus["authority"],
		authStatus["status"],
		authStatus["timestamp"],
		policyList)

	// Phase 10E: Add continuity information to text output
	if continuity, ok := authStatus["continuity"].(map[string]interface{}); ok {
		result += "\nPolicy Continuity:\n"
		result += fmt.Sprintf("- Rate: %.1f%%\n", continuity["rate"])
		result += fmt.Sprintf("- Risk Level: %v\n", continuity["risk_level"])
		result += fmt.Sprintf("- Total Receipts: %v\n", continuity["total_receipts"])
		result += fmt.Sprintf("- Orphan Receipts: %v\n", continuity["orphan_receipts"])
		result += fmt.Sprintf("- Remediation Ready: %v\n", continuity["remediation_ready"])
	}

	// Add any alerts
	if alert, ok := authStatus["alert"].(string); ok {
		result += fmt.Sprintf("\nALERT: %s\n", alert)
	}

	return result
}
