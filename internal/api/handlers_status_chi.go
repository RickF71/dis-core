package api

import (
	"fmt"
	"net/http"
	"time"
)

// handlePingFormatAware handles GET /api/ping with format support
func (s *Server) handlePingFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	resp := map[string]any{
		"status":  "ok",
		"message": "DIS-Core alive",
		"time":    time.Now().Format(time.RFC3339),
	}

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, resp)
	case FormatText:
		ServeAsText(w, resp)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleHealthFormatAware handles GET /api/health with format support
func (s *Server) handleHealthFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	resp := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, resp)
	case FormatText:
		ServeAsText(w, resp)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleStatusFormatAware handles GET /api/status with format support
func (s *Server) handleStatusFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// Get status data using existing logic
	statusData := s.getSystemStatus()

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, statusData)
	case FormatText:
		// Convert status to human-readable text format
		textStatus := s.formatStatusAsText(statusData)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(textStatus))
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleAuthorityStatusFormatAware handles GET /api/authority/status with format support
func (s *Server) handleAuthorityStatusFormatAware(w http.ResponseWriter, r *http.Request) {
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
	return map[string]any{
		"authority": "active",
		"policies":  []string{"gates.rego", "risk.rego"},
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
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

	return fmt.Sprintf(`Authority Status Report
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
}
