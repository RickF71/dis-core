package cors

import (
	"net/http"
	"os"
	"strings"
)

var allowedDevOrigins = []string{
	"http://localhost:5173",
	"http://127.0.0.1:5173",
	"http://[::1]:5173",
}

func loadEnvOrigins() []string {
	raw := os.Getenv("DIS_ALLOWED_ORIGINS")
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		cleaned := strings.TrimSpace(p)
		if cleaned != "" {
			out = append(out, cleaned)
		}
	}
	return out
}

func isAllowedOrigin(origin string) bool {
	for _, o := range allowedDevOrigins {
		if origin == o {
			return true
		}
	}
	for _, o := range loadEnvOrigins() {
		if origin == o {
			return true
		}
	}
	return false
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers",
				"Content-Type, Authorization, X-External-User, X-Acting-As, X-Requested-With, Accept")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
