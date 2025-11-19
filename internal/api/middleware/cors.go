package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func BuildAllowedOrigins() []string {
	env := os.Getenv("DIS_ALLOWED_ORIGINS")
	if env != "" {
		parts := strings.Split(env, ",")
		out := []string{}
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
		if len(out) > 0 {
			log.Printf("[CORS] Using DIS_ALLOWED_ORIGINS=%v\n", out)
			return out
		}
	}

	defaults := []string{
		"http://localhost:5173", // Finagler
		"http://localhost:5174", // FinaglerNone
	}
	log.Printf("[CORS] No DIS_ALLOWED_ORIGINS set, falling back to defaults: %v\n", defaults)
	return defaults
}

func CORS(next http.Handler) http.Handler {
	allowed := BuildAllowedOrigins()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowOrigin := ""
		for _, o := range allowed {
			if o == origin {
				allowOrigin = origin
				break
			}
		}

		// Always set CORS headers for all responses
		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
