package authorityapi

import (
	"net/http"
	"strings"
)

// AuthConfig holds configuration for placeholder authentication
type AuthConfig struct {
	SharedSecret string
}

// SimpleAuthMiddleware provides basic shared secret authentication
// This is a placeholder implementation - Phase 9B will implement proper RBAC
func SimpleAuthMiddleware(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract authorization header
			authHeader := r.Header.Get("Authorization")

			// For Phase 9A, we'll allow requests without auth for testing
			// In Phase 9B, this will be properly implemented
			if authHeader == "" {
				// Allow unauthenticated access for mock endpoints in Phase 9A
				next.ServeHTTP(w, r)
				return
			}

			// Simple bearer token check
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")

				// Check against shared secret from config
				if token == config.SharedSecret {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Invalid or missing authentication
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

// GetAuthConfigFromYAML reads shared secret from config.yaml
// This is a placeholder - will read actual config in Phase 9B
func GetAuthConfigFromYAML() AuthConfig {
	// Placeholder implementation - in Phase 9B this will read from actual config.yaml
	return AuthConfig{
		SharedSecret: "dev-secret-phase9a",
	}
}
