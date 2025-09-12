package middleware

import (
	"net/http"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// Chain combines multiple middleware into a single middleware
// The middleware are applied in the order they are provided
type Chain struct {
	middlewares []application.Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...application.Middleware) *Chain {
	return &Chain{middlewares: middlewares}
}

// Handle returns the HTTP handler that applies all middleware in the chain
func (c *Chain) Handle(final http.Handler) http.Handler {
	// Apply middleware in reverse order so the first middleware
	// in the chain is the outermost handler
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		final = c.middlewares[i].Handle(final)
	}
	return final
}
