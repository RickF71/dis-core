package server

import "net/http"

// WithCORS adds full CORS support for all DIS endpoints.
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow any origin during local dev
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Explicitly list common HTTP verbs
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		// Permit standard headers used by Finagler/React fetch calls
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization, X-Requested-With")
		// Expose headers you might inspect in JS
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		// Allow credentials if ever needed (e.g., cookies)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests directly
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
