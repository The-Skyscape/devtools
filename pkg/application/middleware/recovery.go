package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recovery provides panic recovery middleware
type Recovery struct {
	stackTrace bool
}

// NewRecovery creates a new recovery middleware
func NewRecovery(stackTrace bool) *Recovery {
	return &Recovery{stackTrace}
}

// Handle returns the HTTP handler that recovers from panics
func (r *Recovery) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				if r.stackTrace {
					log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				} else {
					log.Printf("Panic recovered: %v", err)
				}

				// Call error function
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, req)
	})
}
