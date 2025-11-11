package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const provenanceContextKey contextKey = "provenance"

// Provenance contains request tracking information
type Provenance struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

// WithProvenance middleware generates UUID and timestamp for each request
func WithProvenance() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate provenance data
			provenance := Provenance{
				RequestID: uuid.New().String(),
				Timestamp: time.Now().UTC(),
			}

			// Inject provenance into the request context
			ctx := context.WithValue(r.Context(), provenanceContextKey, provenance)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromProvenance retrieves provenance information from the request context
func FromProvenance(ctx context.Context) (Provenance, bool) {
	if prov, ok := ctx.Value(provenanceContextKey).(Provenance); ok {
		return prov, true
	}
	return Provenance{}, false
}
