package api

import (
	"dis-core/internal/db"
	"net/http"
)

// Route describes a single HTTP endpoint.
type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Description string
}

// RegisterAllRoutes binds every API endpoint to the server’s mux.
func (s *Server) RegisterAllRoutes() {
	routes := []Route{
		// 🟢 Core information
		{"GET", "/ping", s.handlePing, "Simple health check"},
		{"GET", "/info", s.handleInfo, "DIS node configuration summary"},
		{"GET", "/verify", HandleExternalVerify, "Verify external receipt"},
		{"GET", "/receipts", s.handleReceipts, "List stored receipts"},

		// 🔐 Authentication and identity
		{"POST", "/api/auth/revoke", s.HandleAuthRevoke, "Revoke handshake / session"},
		{"POST", "/api/auth/handshake", s.HandleDISAuthHandshake, "Create handshake token"},
		{"POST", "/api/auth/virtual_usgov", HandleVirtualUSGovCredential, "Simulated USGOV credential issuer"},
		{"POST", "/api/auth/console", HandleConsoleAuth, "Console login using handshake token"},
		{"POST", "/api/identity/register", HandleIdentityRegister(db.DefaultDB), "Register a new identity"},
		{"GET", "/api/identity/list", HandleIdentityList(db.DefaultDB), "List identities"},

		// 🧩 Self-maintenance
		{"GET", "/api/status", HandleStatus, "Diagnostic counts: receipts, handshakes, revocations, identities"},
	}

	for _, r := range routes {
		s.mux.HandleFunc(r.Path, r.Handler)
	}
}
