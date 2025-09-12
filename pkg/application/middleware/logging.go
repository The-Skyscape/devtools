package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger provides request logging middleware
type Logger struct{}

// NewLogger creates a new logging middleware
func NewLogger() *Logger {
	return &Logger{}
}

// Handle returns the HTTP handler that logs requests
func (l *Logger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status
		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         200, // Default to 200 if not explicitly set
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log the request
		duration := time.Since(start)
		log.Printf("%s %s %d %v %d bytes",
			r.Method,
			r.URL.Path,
			wrapped.status,
			duration,
			wrapped.size,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Flush implements http.Flusher
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
