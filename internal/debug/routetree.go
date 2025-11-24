package debug

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	"github.com/go-chi/chi/v5"
)

// DebugRoutesHandler returns an http.HandlerFunc which walks the provided
// chi.Router and writes method, pattern and handler type for each route.
func DebugRoutesHandler(r chi.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// Walk the router and print each route
		_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			// Try to resolve a human-friendly handler name when possible.
			name := fmt.Sprintf("%T", handler)
			// If the handler is an http.HandlerFunc, get the underlying function name.
			if hf, ok := handler.(http.HandlerFunc); ok {
				ptr := reflect.ValueOf(hf).Pointer()
				if f := runtime.FuncForPC(ptr); f != nil {
					name = f.Name()
				}
			}
			fmt.Fprintf(w, "%s %s %s\n", method, route, name)
			return nil
		})
	}
}
