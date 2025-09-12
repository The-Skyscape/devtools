package middleware

import (
	"net/http"
	"time"
)

// Timeout provides request timeout middleware
type Timeout struct {
	duration time.Duration
	message  string
}

// NewTimeout creates a new timeout middleware
func NewTimeout(duration time.Duration) *Timeout {
	return &Timeout{
		duration: duration,
		message:  "Request timeout",
	}
}

// Handle returns the HTTP handler that enforces request timeouts
func (t *Timeout) Handle(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, t.duration, t.message)
}

// WithMessage sets a custom timeout message
func (t *Timeout) WithMessage(message string) *Timeout {
	t.message = message
	return t
}