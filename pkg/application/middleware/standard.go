package middleware

import (
	"compress/gzip"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
)

// StandardSetup returns a chain of commonly used middleware
func StandardSetup() application.Middleware {
	return NewChain(
		NewRecovery(true),
		NewLogger(),
		NewSecurityHeaders([]string{}),
		NewCompression(gzip.DefaultCompression),
		NewTimeout(30*time.Second),
	)
}
